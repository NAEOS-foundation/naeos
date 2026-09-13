// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrValidation, "bad input")
	if err.Code != ErrValidation {
		t.Errorf("expected code %s, got %s", ErrValidation, err.Code)
	}
	if err.Message != "bad input" {
		t.Errorf("expected message 'bad input', got %s", err.Message)
	}
	if len(err.Stack()) == 0 {
		t.Error("expected non-empty stack")
	}
}

func TestWrap(t *testing.T) {
	inner := errors.New("root cause")
	err := Wrap(ErrCloud, "deploy failed", inner)

	if !errors.Is(err, inner) {
		t.Error("expected unwrappable to inner")
	}
	if !Is(err, ErrCloud) {
		t.Error("expected Is to match code")
	}
}

func TestWrapf(t *testing.T) {
	inner := errors.New("disk full")
	err := Wrapf(inner, ErrDatabase, "write to %s failed", "/data")

	if !strings.Contains(err.Message, "/data") {
		t.Errorf("expected formatted message, got %s", err.Message)
	}
}

func TestIsAny(t *testing.T) {
	err := New(ErrNetwork, "timeout")

	if !IsAny(err, ErrNetwork, ErrTimeout) {
		t.Error("expected IsAny to match")
	}
	if IsAny(err, ErrCloud, ErrPlugin) {
		t.Error("expected IsAny not to match")
	}
}

func TestIsRetryable(t *testing.T) {
	err := New(ErrTimeout, "slow").WithRetry()
	if !IsRetryable(err) {
		t.Error("expected retryable")
	}

	err2 := New(ErrInternal, "crash")
	if IsRetryable(err2) {
		t.Error("expected not retryable")
	}

	wrapped := Wrap(ErrNetwork, "fail", err)
	if !IsRetryable(wrapped) {
		t.Error("expected wrapped retryable")
	}
}

func TestCodeOf(t *testing.T) {
	err := New(ErrPlugin, "missing")
	if CodeOf(err) != ErrPlugin {
		t.Errorf("expected %s, got %s", ErrPlugin, CodeOf(err))
	}
	if CodeOf(errors.New("plain")) != "" {
		t.Error("expected empty code for plain error")
	}
}

func TestContextOf(t *testing.T) {
	err := New(ErrValidation, "bad").WithContext("field", "name").WithContext("line", 42)
	ctx := ContextOf(err)

	if ctx["field"] != "name" {
		t.Errorf("expected field=name, got %v", ctx["field"])
	}
	if ctx["line"] != 42 {
		t.Errorf("expected line=42, got %v", ctx["line"])
	}
}

func TestErrorGroup(t *testing.T) {
	e1 := New(ErrValidation, "first")
	e2 := New(ErrCloud, "second")

	group := Group(e1, e2, nil)

	if group.Len() != 2 {
		t.Errorf("expected 2 errors, got %d", group.Len())
	}
	if group.Empty() {
		t.Error("expected non-empty group")
	}
	if !strings.Contains(group.Error(), "2 errors") {
		t.Errorf("expected count in error, got %s", group.Error())
	}
}

func TestErrorGroupHasCode(t *testing.T) {
	group := Group(
		New(ErrNetwork, "timeout"),
		New(ErrCloud, "deploy"),
	)

	if !group.HasCode(ErrNetwork) {
		t.Error("expected HasCode to find ErrNetwork")
	}
	if !group.HasCode(ErrCloud) {
		t.Error("expected HasCode to find ErrCloud")
	}
	if group.HasCode(ErrPlugin) {
		t.Error("expected HasCode not to find ErrPlugin")
	}
}

func TestErrorGroupCodes(t *testing.T) {
	group := Group(
		New(ErrNetwork, "a"),
		New(ErrNetwork, "b"),
		New(ErrCloud, "c"),
	)

	codes := group.Codes()
	if len(codes) != 2 {
		t.Errorf("expected 2 unique codes, got %d", len(codes))
	}
}

func TestErrorGroupUnwrap(t *testing.T) {
	e1 := New(ErrValidation, "one")
	e2 := New(ErrCloud, "two")
	group := Group(e1, e2)

	unwrapped := group.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("expected 2 unwrapped, got %d", len(unwrapped))
	}

	if !group.Is(ErrInvalidSpec) {
		t.Error("expected Is to match sentinel")
	}
}

func TestErrorGroupSingle(t *testing.T) {
	err := New(ErrInternal, "solo")
	group := Group(err)

	if group.Error() != err.Error() {
		t.Errorf("expected single error message, got %s", group.Error())
	}
}

func TestErrorGroupEmpty(t *testing.T) {
	group := Group()
	if !group.Empty() {
		t.Error("expected empty group")
	}
	if group.Error() != "" {
		t.Errorf("expected empty string, got %s", group.Error())
	}
}

func TestErrorIsInterface(t *testing.T) {
	err := New(ErrAuth, "denied")
	target := New(ErrAuth, "other")
	if !errors.Is(err, target) {
		t.Error("expected Is to match by code")
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []struct {
		err  error
		code ErrorCode
	}{
		{ErrNotConnected, ErrNetwork},
		{ErrInvalidSpec, ErrValidation},
		{ErrPluginNotFound, ErrPlugin},
		{ErrDeployFailed, ErrCloud},
		{ErrTimedOut, ErrTimeout},
		{ErrRateLimited, ErrRateLimit},
		{ErrUnauthorized, ErrAuth},
		{ErrForbidden, ErrPermDenied},
		{ErrAlreadyExists, ErrConflict},
		{ErrInvalidConfig, ErrConfig},
		{ErrInternalFailed, ErrInternal},
		{ErrNotImplemented, ErrInternal},
		{ErrDependencyCycle, ErrPipeline},
	}

	for _, s := range sentinels {
		if !Is(s.err, s.code) {
			t.Errorf("sentinel %v: expected Is(%s)", s.err, s.code)
		}
	}
}

func TestStackFrames(t *testing.T) {
	err := New(ErrInternal, "trace")
	frames := err.Stack()
	if len(frames) == 0 {
		t.Fatal("expected non-empty stack")
	}
	if frames[0].File == "" {
		t.Error("expected non-empty file in stack frame")
	}
	if frames[0].Line == 0 {
		t.Error("expected non-zero line in stack frame")
	}
}

func TestNestedGroup(t *testing.T) {
	g1 := Group(New(ErrValidation, "v1"), New(ErrValidation, "v2"))
	g2 := Group(New(ErrCloud, "c1"), g1)

	if g2.Len() != 3 {
		t.Errorf("expected 3 errors (nested), got %d", g2.Len())
	}
}
