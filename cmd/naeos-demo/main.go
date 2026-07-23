//nolint:errcheck
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr      = flag.String("addr", ":9090", "listen address")
	binary    = flag.String("binary", "naeos", "path to naeos binary")
	maxSess   = flag.Int("max-sessions", 10, "max concurrent sessions")
	sessTTL   = flag.Duration("session-ttl", 5*time.Minute, "session idle timeout")
	allowOrig = flag.String("allow-origin", "*", "allowed origin (or * for all)")
	whitelist = flag.String("whitelist", "init,validate,compile,run,version,help,spec", "comma-separated allowed subcommands")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		if *allowOrig == "*" {
			return true
		}
		origin := r.Header.Get("Origin")
		return origin == *allowOrig || strings.HasPrefix(origin, *allowOrig)
	},
}

type session struct {
	id      string
	dir     string
	lastUse time.Time
	cmd     *exec.Cmd
	mu      sync.Mutex
}

type demoServer struct {
	mu       sync.Mutex
	sessions map[string]*session
	naeosBin string
	allowed  map[string]bool
}

func newDemoServer() *demoServer {
	allowed := make(map[string]bool)
	for _, s := range strings.Split(*whitelist, ",") {
		allowed[strings.TrimSpace(s)] = true
	}
	return &demoServer{
		sessions: make(map[string]*session),
		naeosBin: *binary,
		allowed:  allowed,
	}
}

func (s *demoServer) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if time.Since(sess.lastUse) > *sessTTL {
			sess.kill()
			os.RemoveAll(sess.dir)
			delete(s.sessions, id)
		}
	}
}

func (s *demoServer) createSession() (*session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.sessions) >= *maxSess {
		for id, sess := range s.sessions {
			if time.Since(sess.lastUse) > *sessTTL {
				sess.kill()
				os.RemoveAll(sess.dir)
				delete(s.sessions, id)
			}
		}
	}
	if len(s.sessions) >= *maxSess {
		return nil, fmt.Errorf("max sessions (%d) reached, try again later", *maxSess)
	}

	dir, err := os.MkdirTemp("", "naeos-demo-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	sess := &session{
		id:      filepath.Base(dir),
		dir:     dir,
		lastUse: time.Now(),
	}
	s.sessions[sess.id] = sess
	return sess, nil
}

func (sess *session) kill() {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.cmd != nil && sess.cmd.Process != nil {
		sess.cmd.Process.Kill()
	}
}

func (sess *session) touch() {
	sess.mu.Lock()
	sess.lastUse = time.Now()
	sess.mu.Unlock()
}

func (s *demoServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	defer conn.Close()

	sess, err := s.createSession()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("!error "+err.Error()))
		return
	}
	defer func() {
		sess.kill()
		os.RemoveAll(sess.dir)
		s.mu.Lock()
		delete(s.sessions, sess.id)
		s.mu.Unlock()
	}()

	conn.WriteMessage(websocket.TextMessage, []byte("!ready Session created in "+sess.dir))
	conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		line := strings.TrimSpace(string(msg))
		if line == "" {
			continue
		}

		if line == "!reset" {
			sess.kill()
			newDir, err := os.MkdirTemp("", "naeos-demo-*")
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("!error "+err.Error()))
				continue
			}
			os.RemoveAll(sess.dir)
			sess.dir = newDir
			conn.WriteMessage(websocket.TextMessage, []byte("!ready Session reset"))
			conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
			continue
		}

		sess.touch()

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		if len(parts) >= 2 && parts[0] == "naeos" {
			if !s.allowed[parts[1]] {
				conn.WriteMessage(websocket.TextMessage, []byte("!error Command '"+parts[1]+"' not allowed in demo mode"))
				conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
				continue
			}
			s.runCommand(conn, sess, parts[1:])
		} else if parts[0] == "naeos" {
			conn.WriteMessage(websocket.TextMessage, []byte("!error Usage: naeos <command> [args]"))
			conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
		} else {
			conn.WriteMessage(websocket.TextMessage, []byte("!error Unknown command. Try: naeos init, naeos validate, naeos compile"))
			conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
		}
	}
}

func (s *demoServer) runCommand(conn *websocket.Conn, sess *session, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.naeosBin, args...)
	cmd.Dir = sess.dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("!error "+err.Error()))
		conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("!error "+err.Error()))
		conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
		return
	}

	sess.mu.Lock()
	sess.cmd = cmd
	sess.mu.Unlock()

	if err := cmd.Start(); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("!error "+err.Error()))
		conn.WriteMessage(websocket.TextMessage, []byte("!prompt"))
		return
	}

	done := make(chan struct{})
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				conn.WriteMessage(websocket.TextMessage, buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				conn.WriteMessage(websocket.TextMessage, buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		conn.WriteMessage(websocket.TextMessage, []byte("\n!error Command timed out"))
	}

	conn.WriteMessage(websocket.TextMessage, []byte("\n!prompt"))
}

func main() {
	flag.Parse()

	if _, err := exec.LookPath(*binary); err != nil {
		log.Printf("WARNING: naeos binary not found in PATH (%v). Demo will fail on command execution.", err)
	}

	server := newDemoServer()

	go func() {
		for range time.Tick(30 * time.Second) {
			server.cleanup()
		}
	}()

	http.HandleFunc("/ws", server.handleWS)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	fmt.Printf("NAEOS Demo Server listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
