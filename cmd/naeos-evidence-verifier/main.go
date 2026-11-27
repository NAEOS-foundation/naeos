// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0
package main
import (
 "context"
 "errors"
 "log"
 "net/http"
 "os"
 "os/signal"
 "strings"
 "syscall"
 "time"
 "github.com/NAEOS-foundation/naeos/internal/controlplane"
)
type policy struct{version string}
func(p policy)Version()string{return p.version}
func(p policy)EvaluateExecutionEvidence(_ context.Context,o controlplane.AuthoritativeExecutionObservation,e controlplane.ExecutionOutcome)error{if o.Outcome!=e{return controlplane.ErrAuthoritativeEvidenceNotVerified};return nil}
func main(){if err:=run();err!=nil{log.Printf("verifier stopped: %v",err);os.Exit(1)}}
func run()error{
 sourceURL:=os.Getenv("NAEOS_EVIDENCE_SOURCE_URL");sourceToken:=os.Getenv("NAEOS_EVIDENCE_SOURCE_TOKEN");serviceToken:=os.Getenv("NAEOS_EVIDENCE_VERIFIER_TOKEN");databaseURL:=os.Getenv("NAEOS_EVIDENCE_VERIFIER_DATABASE_URL");scope:=strings.TrimSpace(os.Getenv("NAEOS_EVIDENCE_VERIFIER_SCOPE"));version:=strings.TrimSpace(os.Getenv("NAEOS_EVIDENCE_POLICY_VERSION"));addr:=os.Getenv("NAEOS_EVIDENCE_VERIFIER_ADDR");if addr==""{addr="127.0.0.1:8081"}
 if sourceURL==""||sourceToken==""||serviceToken==""||databaseURL==""||scope==""||version==""{return errors.New("required verifier configuration is missing")}
 ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer stop()
 guard,err:=controlplane.NewPostgresEvidenceVerifierReplayGuard(ctx,databaseURL);if err!=nil{return err};defer guard.Close()
 source,err:=controlplane.NewHTTPAuthoritativeEvidenceSource(sourceURL,sourceToken);if err!=nil{return err}
 service,err:=controlplane.NewProviderNeutralEvidenceVerificationService(source,policy{version},guard,scope,serviceToken);if err!=nil{return err}
 mux:=http.NewServeMux();mux.Handle("/v1/verify",service.Handler())
 mux.HandleFunc("/healthz",func(w http.ResponseWriter,_ *http.Request){w.WriteHeader(http.StatusOK);_,_=w.Write([]byte("ok"))})
 mux.HandleFunc("/readyz",func(w http.ResponseWriter,r *http.Request){if _,err:=guard.ClaimReadiness(r.Context());err!=nil{w.WriteHeader(http.StatusServiceUnavailable);_,_=w.Write([]byte("not ready"));return};w.WriteHeader(http.StatusOK);_,_=w.Write([]byte("ready"))})
 server:=&http.Server{Addr:addr,Handler:mux,ReadHeaderTimeout:3*time.Second,ReadTimeout:10*time.Second,WriteTimeout:15*time.Second,IdleTimeout:60*time.Second,MaxHeaderBytes:32*1024}
 errs:=make(chan error,1);go func(){errs<-server.ListenAndServe()}()
 select{case<-ctx.Done():shutdown,cancel:=context.WithTimeout(context.Background(),8*time.Second);defer cancel();return server.Shutdown(shutdown);case e:=<-errs:if errors.Is(e,http.ErrServerClosed){return nil};return e}
}
