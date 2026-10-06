// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
 "bytes"
 "crypto/sha256"
 "encoding/base64"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "strconv"
 "strings"
 "sync"
 "time"
)

type RFC7807Problem struct {
 Type string `json:"type"`
 Title string `json:"title"`
 Status int `json:"status"`
 Detail string `json:"detail,omitempty"`
 Instance string `json:"instance,omitempty"`
 RequestID string `json:"request_id,omitempty"`
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, detail string) {
 w.Header().Set("Content-Type", "application/problem+json")
 w.Header().Set("API-Version", "2")
 p := RFC7807Problem{
  Type: "https://naeos.dev/problems/" + strings.ToLower(strings.ReplaceAll(http.StatusText(status), " ", "-")),
  Title: http.StatusText(status), Status: status, Detail: detail,
  Instance: r.URL.Path, RequestID: r.Header.Get("X-Request-ID"),
 }
 w.WriteHeader(status)
 _ = json.NewEncoder(w).Encode(p)
}

type v2Cursor struct { Offset int `json:"offset"`; Limit int `json:"limit"` }

func encodeV2Cursor(c v2Cursor) string {
 raw, _ := json.Marshal(c)
 return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeV2Cursor(s string) (v2Cursor, error) {
 raw, err := base64.RawURLEncoding.DecodeString(s)
 if err != nil { return v2Cursor{}, fmt.Errorf("invalid cursor") }
 var c v2Cursor
 if err := json.Unmarshal(raw, &c); err != nil || c.Offset < 0 || c.Limit < 1 || c.Limit > 100 {
  return v2Cursor{}, fmt.Errorf("invalid cursor")
 }
 return c, nil
}

func parseV2Cursor(r *http.Request) (v2Cursor, error) {
 limit := 50
 if raw := r.URL.Query().Get("limit"); raw != "" {
  n, err := strconv.Atoi(raw)
  if err != nil || n < 1 || n > 100 { return v2Cursor{}, fmt.Errorf("limit must be between 1 and 100") }
  limit = n
 }
 if raw := r.URL.Query().Get("cursor"); raw != "" {
  c, err := decodeV2Cursor(raw)
  if err != nil { return v2Cursor{}, err }
  if r.URL.Query().Get("limit") != "" && c.Limit != limit { return v2Cursor{}, fmt.Errorf("cursor limit mismatch") }
  return c, nil
 }
 return v2Cursor{Offset: 0, Limit: limit}, nil
}

func (s *Server) handleV2Version(w http.ResponseWriter, r *http.Request) {
 if r.Method != http.MethodGet { writeProblem(w,r,http.StatusMethodNotAllowed,"method not allowed"); return }
 w.Header().Set("API-Version","2")
 s.writeJSON(w,http.StatusOK,map[string]any{"version":"2","api":"v2"})
}

func (s *Server) handleV2Pipelines(w http.ResponseWriter, r *http.Request) {
 if r.Method != http.MethodGet { writeProblem(w,r,http.StatusMethodNotAllowed,"method not allowed"); return }
 cursor, err := parseV2Cursor(r)
 if err != nil { writeProblem(w,r,http.StatusBadRequest,err.Error()); return }
 s.pipelinesMu.RLock()
 items := make([]pipelineRun,len(s.pipelines)); copy(items,s.pipelines)
 s.pipelinesMu.RUnlock()
 for i,j:=0,len(items)-1;i<j;i,j=i+1,j-1 { items[i],items[j]=items[j],items[i] }
 start:=cursor.Offset; if start>len(items){start=len(items)}
 end:=start+cursor.Limit; if end>len(items){end=len(items)}
 page:=items[start:end]
 nextCursor:=""
 if end<len(items){nextCursor=encodeV2Cursor(v2Cursor{Offset:end,Limit:cursor.Limit})}
 w.Header().Set("API-Version","2")
 s.writeJSON(w,http.StatusOK,map[string]any{"data":page,"count":len(page),"next_cursor":nextCursor})
}

type idempotencyEntry struct {
 fingerprint string
 status int
 header http.Header
 body []byte
 expiresAt time.Time
}

var v2Idempotency = struct {
 sync.Mutex
 items map[string]idempotencyEntry
}{items:make(map[string]idempotencyEntry)}

type v2ResponseRecorder struct { header http.Header; body bytes.Buffer; status int }

func newV2ResponseRecorder()*v2ResponseRecorder{return &v2ResponseRecorder{header:make(http.Header),status:http.StatusOK}}
func(w *v2ResponseRecorder)Header()http.Header{return w.header}
func(w *v2ResponseRecorder)WriteHeader(status int){if w.status!=http.StatusOK{return};w.status=status}
func(w *v2ResponseRecorder)Write(p []byte)(int,error){return w.body.Write(p)}
func(w *v2ResponseRecorder)commit(dst http.ResponseWriter){
 for k,values:=range w.header{for _,value:=range values{dst.Header().Add(k,value)}}
 dst.WriteHeader(w.status);_,_=dst.Write(w.body.Bytes())
}

func fingerprintRequest(r *http.Request,body []byte)string{
 h:=sha256.New();_,_=io.WriteString(h,r.Method+"\n"+r.URL.RequestURI()+"\n");_,_=h.Write(body)
 return hex.EncodeToString(h.Sum(nil))
}

func(s *Server)v2IdempotencyMiddleware(next http.Handler)http.Handler{
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  key:=strings.TrimSpace(r.Header.Get("Idempotency-Key"))
  mutating:=r.Method==http.MethodPost||r.Method==http.MethodPut||r.Method==http.MethodPatch||r.Method==http.MethodDelete
  if !mutating{next.ServeHTTP(w,r);return}
  if key==""{writeProblem(w,r,http.StatusBadRequest,"Idempotency-Key header is required for mutating API v2 requests");return}
  if len(key)>255{writeProblem(w,r,http.StatusBadRequest,"Idempotency-Key must be 255 characters or fewer");return}
  body,err:=io.ReadAll(r.Body);if err!=nil{writeProblem(w,r,http.StatusBadRequest,"unable to read request body");return}
  r.Body=io.NopCloser(bytes.NewReader(body));fp:=fingerprintRequest(r,body)
  v2Idempotency.Lock()
  for k,entry:=range v2Idempotency.items{if time.Now().After(entry.expiresAt){delete(v2Idempotency.items,k)}}
  entry,exists:=v2Idempotency.items[key];v2Idempotency.Unlock()
  if exists{
   if entry.fingerprint!=fp{writeProblem(w,r,http.StatusConflict,"Idempotency-Key was already used with a different request");return}
   for k,values:=range entry.header{for _,value:=range values{w.Header().Add(k,value)}}
   w.WriteHeader(entry.status);_,_=w.Write(entry.body);return
  }
  rec:=newV2ResponseRecorder();next.ServeHTTP(rec,r);rec.commit(w)
  if rec.status>=200&&rec.status<500{
   h:=make(http.Header);for k,values:=range rec.header{h[k]=append([]string(nil),values...)}
   v2Idempotency.Lock();v2Idempotency.items[key]=idempotencyEntry{fingerprint:fp,status:rec.status,header:h,body:append([]byte(nil),rec.body.Bytes()...),expiresAt:time.Now().Add(24*time.Hour)};v2Idempotency.Unlock()
  }
 })
}
