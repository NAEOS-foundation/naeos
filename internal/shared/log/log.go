package log

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	mu            sync.RWMutex
)

func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func SetLevel(level slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

func SetJSON() {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func SetLogger(l *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = l
}

func Logger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLogger
}

func With(args ...any) *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLogger.With(args...)
}

func Info(msg string, args ...any) {
	mu.RLock()
	l := defaultLogger
	mu.RUnlock()
	l.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	mu.RLock()
	l := defaultLogger
	mu.RUnlock()
	l.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	mu.RLock()
	l := defaultLogger
	mu.RUnlock()
	l.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	mu.RLock()
	l := defaultLogger
	mu.RUnlock()
	l.Debug(msg, args...)
}
