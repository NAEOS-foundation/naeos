// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "fmt"

// CLI exit codes distinguish policy outcomes from infrastructure failures.
const (
	ExitCodePolicyDeny       = 10
	ExitCodeExecutionFailure = 11
	ExitCodeVerificationFail = 12
)

// ExitCodeError lets command handlers preserve a stable process exit contract.
type ExitCodeError struct {
	Code int
	Err  error
}

func (e *ExitCodeError) Error() string { return e.Err.Error() }
func (e *ExitCodeError) Unwrap() error { return e.Err }

func newExitCodeError(code int, format string, args ...any) error {
	return &ExitCodeError{Code: code, Err: fmt.Errorf(format, args...)}
}
