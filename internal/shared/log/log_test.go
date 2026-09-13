// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestSetOutput(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected log output to contain 'test message', got: %s", buf.String())
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	SetLevel(slog.LevelWarn)
	SetOutput(&buf)

	Warn("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Errorf("expected Warn output, got: %s", buf.String())
	}
}

func TestSetLogger(t *testing.T) {
	var buf bytes.Buffer
	custom := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	SetLogger(custom)

	if Logger() != custom {
		t.Error("expected custom logger to be set")
	}

	Debug("debug msg")
	if !strings.Contains(buf.String(), "debug msg") {
		t.Errorf("expected debug message in output, got: %s", buf.String())
	}
}

func TestLogger(t *testing.T) {
	l := Logger()
	if l == nil {
		t.Error("expected non-nil logger")
	}
}

func TestInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	Info("hello", "key", "value")

	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected 'hello', got: %s", buf.String())
	}
}

func TestErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	Error("fail", "code", 42)

	if !strings.Contains(buf.String(), "fail") {
		t.Errorf("expected 'fail', got: %s", buf.String())
	}
}
