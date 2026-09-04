package serve

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

// Server is the production NAEOS daemon. It owns one or more HTTP/HTTPS
// listeners and coordinates graceful shutdown across all of them.
type Server struct {
	cfg      *Config
	api      *api.Server
	servers  []*http.Server
	logLevel slog.Level
	mu       sync.Mutex
	stopped  bool
}

// New builds a daemon from a validated Config. API listeners are backed by a
// single shared api.Server instance. The returned Server is not started until
// Start is called.
func New(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var apiServer *api.Server
	for _, l := range cfg.Listeners {
		if l.API {
			auth := &api.AuthConfig{Enabled: cfg.Auth.Enabled}
			if cfg.Auth.Enabled {
				auth.JWTSecret = cfg.Auth.JWTSecret
			}
			apiServer = api.NewServer(l.Addr, auth)
			break
		}
	}

	s := &Server{
		cfg:      cfg,
		api:      apiServer,
		logLevel: parseLogLevel(cfg.LogLevel),
	}

	if apiServer != nil {
		for key, rps := range cfg.APIKeys {
			apiServer.RegisterAPIKey(key, rps)
		}
	}

	return s, nil
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Listeners returns a copy of the configured listener definitions.
func (s *Server) Listeners() []Listener {
	return append([]Listener(nil), s.cfg.Listeners...)
}

// Start opens every configured listener, configures graceful shutdown, and
// blocks until a shutdown signal is received. It returns nil on a clean
// shutdown, or the first fatal error encountered.
func (s *Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return s.StartWithContext(ctx)
}

// StartWithContext is like Start but uses the provided context for shutdown
// instead of OS signals. Cancelling the context triggers a graceful shutdown.
func (s *Server) StartWithContext(ctx context.Context) error {
	if len(s.cfg.Listeners) == 0 {
		return naeoserr.New(naeoserr.ErrConfig, "server: no listeners configured")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: s.logLevel}))
	slog.SetDefault(logger)

	errCh := make(chan error, len(s.cfg.Listeners))
	var wg sync.WaitGroup

	for _, l := range s.cfg.Listeners {
		var handler http.Handler = http.NotFoundHandler()
		if l.API && s.api != nil {
			handler = s.api.Handler()
		}

		srv := &http.Server{
			Addr:         l.Addr,
			Handler:      handler,
			ReadTimeout:  s.duration(s.cfg.ReadTimeout, 15*time.Second),
			WriteTimeout: s.duration(s.cfg.WriteTimeout, 15*time.Second),
			IdleTimeout:  s.duration(s.cfg.IdleTimeout, 60*time.Second),
		}
		s.mu.Lock()
		s.servers = append(s.servers, srv)
		s.mu.Unlock()

		wg.Add(1)
		go func(srv *http.Server, l Listener) {
			defer wg.Done()
			errCh <- s.serve(srv, l)
		}(srv, l)
	}

	go func() {
		<-ctx.Done()
		slog.Warn("shutting down NAEOS server", "component", "serve")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.duration(s.cfg.ShutdownTimeout, 30*time.Second))
		defer cancel()
		_ = s.Shutdown(shutdownCtx)
	}()

	// A fatal error on any listener ends the daemon, mimicking systemd restarts.
	var firstErr error
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			firstErr = err
		}
	}
	if err := s.Shutdown(context.Background()); err != nil && firstErr == nil {
		firstErr = err
	}
	wg.Wait()
	return firstErr
}

func (s *Server) serve(srv *http.Server, l Listener) error {
	slog.Info("listening", "addr", l.Addr, "tls", l.IsTLS(), "name", orDefault(l.Name, "listener"), "component", "serve")
	if l.IsTLS() {
		cert, err := tls.LoadX509KeyPair(l.TLSCert, l.TLSKey)
		if err != nil {
			return naeoserr.Wrap(naeoserr.ErrConfig, "load tls material", err)
		}
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		err = srv.ListenAndServeTLS("", "")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops all listeners within the given context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return nil
	}
	s.stopped = true
	var errs []error
	for _, srv := range s.servers {
		if srv == nil {
			continue
		}
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			errs = append(errs, err)
		}
	}
	if s.api != nil {
		if err := s.api.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown: %w", errors.Join(errs...))
	}
	return nil
}

func (s *Server) duration(raw string, def time.Duration) time.Duration {
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return def
	}
	return d
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
