// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type ErrorCode string

const (
	ErrParse      ErrorCode = "PARSE_ERROR"
	ErrValidation ErrorCode = "VALIDATION_ERROR"
	ErrCloud      ErrorCode = "CLOUD_ERROR"
	ErrPlugin     ErrorCode = "PLUGIN_ERROR"
	ErrAuth       ErrorCode = "AUTH_ERROR"
	ErrPipeline   ErrorCode = "PIPELINE_ERROR"
	ErrConfig     ErrorCode = "CONFIG_ERROR"
	ErrDatabase   ErrorCode = "DATABASE_ERROR"
	ErrNetwork    ErrorCode = "NETWORK_ERROR"
	ErrInternal   ErrorCode = "INTERNAL_ERROR"
	ErrNotFound   ErrorCode = "NOT_FOUND"
	ErrConflict   ErrorCode = "CONFLICT"
	ErrTimeout    ErrorCode = "TIMEOUT"
	ErrRateLimit  ErrorCode = "RATE_LIMIT"
	ErrPermDenied ErrorCode = "PERMISSION_DENIED"
	ErrCanceled   ErrorCode = "CANCELED"
)

type StackFrame struct {
	File string
	Line int
	Func string
}

type NaeosError struct {
	Code      ErrorCode
	Message   string
	Inner     error
	stack     []StackFrame
	context   map[string]any
	retryable bool
}

func (e *NaeosError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Inner)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *NaeosError) Unwrap() error {
	return e.Inner
}

func (e *NaeosError) Is(target error) bool {
	if t, ok := target.(*NaeosError); ok {
		return e.Code == t.Code
	}
	return false
}

func (e *NaeosError) Stack() []StackFrame {
	return e.stack
}

func (e *NaeosError) Context() map[string]any {
	return e.context
}

func (e *NaeosError) IsRetryable() bool {
	return e.retryable
}

func (e *NaeosError) WithContext(key string, value any) *NaeosError {
	if e.context == nil {
		e.context = make(map[string]any)
	}
	e.context[key] = value
	return e
}

func (e *NaeosError) WithRetry() *NaeosError {
	e.retryable = true
	return e
}

func captureStack(skip int) []StackFrame {
	var frames []StackFrame
	pcs := make([]uintptr, 10)
	n := runtime.Callers(skip, pcs)
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pcs[i])
		frames = append(frames, StackFrame{
			File: file,
			Line: line,
			Func: fn.Name(),
		})
	}
	return frames
}

func New(code ErrorCode, msg string) *NaeosError {
	return &NaeosError{
		Code:    code,
		Message: msg,
		stack:   captureStack(3),
	}
}

func Wrap(code ErrorCode, msg string, err error) *NaeosError {
	return &NaeosError{
		Code:    code,
		Message: msg,
		Inner:   err,
		stack:   captureStack(3),
	}
}

func Wrapf(err error, code ErrorCode, format string, args ...any) *NaeosError {
	return &NaeosError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Inner:   err,
		stack:   captureStack(3),
	}
}

func Is(err error, code ErrorCode) bool {
	for err != nil {
		var ne *NaeosError
		if As(err, &ne) && ne.Code == code {
			return true
		}
		uw, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = uw.Unwrap()
	}
	return false
}

func IsAny(err error, codes ...ErrorCode) bool {
	for _, code := range codes {
		if Is(err, code) {
			return true
		}
	}
	return false
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func IsRetryable(err error) bool {
	var ne *NaeosError
	if As(err, &ne) && ne.retryable {
		return true
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return IsRetryable(u.Unwrap())
	}
	return false
}

func CodeOf(err error) ErrorCode {
	var ne *NaeosError
	if As(err, &ne) {
		return ne.Code
	}
	return ""
}

func ContextOf(err error) map[string]any {
	var ne *NaeosError
	if As(err, &ne) {
		return ne.context
	}
	return nil
}

type ErrorGroup struct {
	errors []error
}

func Group(errs ...error) *ErrorGroup {
	var valid []error
	for _, e := range errs {
		if e != nil {
			valid = append(valid, e)
		}
	}
	return &ErrorGroup{errors: valid}
}

func (g *ErrorGroup) Error() string {
	if len(g.errors) == 0 {
		return ""
	}
	if len(g.errors) == 1 {
		return g.errors[0].Error()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d errors:", len(g.errors))
	for _, e := range g.errors {
		fmt.Fprintf(&b, "\n  - %v", e)
	}
	return b.String()
}

func (g *ErrorGroup) Errors() []error {
	return g.errors
}

func (g *ErrorGroup) Unwrap() []error {
	return g.errors
}

func (g *ErrorGroup) Is(target error) bool {
	for _, e := range g.errors {
		if errors.Is(e, target) {
			return true
		}
		if sub, ok := e.(*ErrorGroup); ok && sub.Is(target) {
			return true
		}
	}
	return false
}

func (g *ErrorGroup) HasCode(code ErrorCode) bool {
	for _, e := range g.errors {
		if Is(e, code) {
			return true
		}
	}
	return false
}

func (g *ErrorGroup) Codes() []ErrorCode {
	seen := make(map[ErrorCode]bool)
	var codes []ErrorCode
	for _, e := range g.errors {
		c := CodeOf(e)
		if c != "" && !seen[c] {
			seen[c] = true
			codes = append(codes, c)
		}
	}
	return codes
}

func (g *ErrorGroup) Len() int {
	count := 0
	for _, e := range g.errors {
		var sub *ErrorGroup
		if errors.As(e, &sub) {
			count += sub.Len()
		} else {
			count++
		}
	}
	return count
}

func (g *ErrorGroup) Empty() bool {
	return len(g.errors) == 0
}

var (
	ErrNotConnected    = New(ErrNetwork, "not connected")
	ErrInvalidSpec     = New(ErrValidation, "invalid spec")
	ErrPluginNotFound  = New(ErrPlugin, "plugin not found")
	ErrDeployFailed    = New(ErrCloud, "deploy failed")
	ErrTimedOut        = New(ErrTimeout, "operation timed out")
	ErrRateLimited     = New(ErrRateLimit, "rate limited")
	ErrUnauthorized    = New(ErrAuth, "unauthorized")
	ErrForbidden       = New(ErrPermDenied, "forbidden")
	ErrAlreadyExists   = New(ErrConflict, "resource already exists")
	ErrInvalidConfig   = New(ErrConfig, "invalid configuration")
	ErrInternalFailed  = New(ErrInternal, "internal error")
	ErrNotImplemented  = New(ErrInternal, "not implemented")
	ErrDependencyCycle = New(ErrPipeline, "dependency cycle detected")
)
