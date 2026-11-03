// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const runtimeBridgeProcessHelperEnv = "NAEOS_RUNTIME_BRIDGE_PROCESS_HELPER"
const runtimeBridgeProcessArgsEnv = "NAEOS_RUNTIME_BRIDGE_PROCESS_ARGS"

func TestRuntimeBridgeCLIProcessContractHelper(t *testing.T) {
	if os.Getenv(runtimeBridgeProcessHelperEnv) != "1" {
		return
	}
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv(runtimeBridgeProcessArgsEnv)), &args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	root := NewRootCommand()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exitErr *ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func TestRuntimeBridgeCLIProcessContract(t *testing.T) {
	home := t.TempDir()
	filesystemRoot := filepath.Join(home, "sandbox")
	if err := os.MkdirAll(filesystemRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	request := []byte("{\"request_id\":\"req-cli-contract\",\"invocation_id\":\"inv-cli-contract\",\"actor\":\"test\",\"tool\":\"filesystem\",\"action\":\"write\",\"resource\":\"filesystem\",\"payload\":{\"path\":\"contract-artifact.txt\",\"content\":\"should not be written\"}}\n")

	t.Run("policy denial has distinct exit code and JSON stdout", func(t *testing.T) {
		args := []string{"runtime", "bridge", "--filesystem-root", filesystemRoot}
		stdout, stderr, exitCode := runRuntimeBridgeProcess(t, home, args, request)
		if exitCode != ExitCodePolicyDeny {
			t.Fatalf("exit code = %d, want %d; stderr=%q", exitCode, ExitCodePolicyDeny, stderr)
		}
		var response map[string]any
		if err := json.Unmarshal(bytes.TrimSpace(stdout), &response); err != nil {
			t.Fatalf("stdout must contain a JSON response line: %v; stdout=%q", err, stdout)
		}
		if _, ok := response["result"]; !ok {
			t.Fatalf("JSON response lacks result field: %#v", response)
		}
		if stderr == "" || !bytes.Contains(stderr, []byte("DENY")) {
			t.Fatalf("stderr must identify policy denial, got %q", stderr)
		}
		if _, err := os.Stat(filepath.Join(filesystemRoot, "contract-artifact.txt")); !os.IsNotExist(err) {
			t.Fatalf("denied request caused filesystem side effect; stat err=%v", err)
		}
	})

	t.Run("evidence persistence failure has execution failure exit code", func(t *testing.T) {
		evidencePath := filepath.Join(home, "evidence-target")
		if err := os.Mkdir(evidencePath, 0o700); err != nil {
			t.Fatal(err)
		}
		args := []string{"runtime", "bridge", "--filesystem-root", filesystemRoot, "--evidence-file", evidencePath}
		stdout, stderr, exitCode := runRuntimeBridgeProcess(t, home, args, request)
		if exitCode != ExitCodeExecutionFailure {
			t.Fatalf("exit code = %d, want %d; stderr=%q", exitCode, ExitCodeExecutionFailure, stderr)
		}
		if len(bytes.TrimSpace(stdout)) != 0 {
			t.Fatalf("persistence failure must not emit a success response, stdout=%q", stdout)
		}
		if stderr == "" || !bytes.Contains(stderr, []byte("persist runtime evidence")) {
			t.Fatalf("stderr must identify persistence failure, got %q", stderr)
		}
		if _, err := os.Stat(filepath.Join(filesystemRoot, "contract-artifact.txt")); !os.IsNotExist(err) {
			t.Fatalf("denied request caused filesystem side effect; stat err=%v", err)
		}
	})
}

func runRuntimeBridgeProcess(t *testing.T, home string, args []string, stdin []byte) ([]byte, []byte, int) {
	t.Helper()
	encodedArgs, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestRuntimeBridgeCLIProcessContractHelper$")
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		runtimeBridgeProcessHelperEnv+"=1",
		runtimeBridgeProcessArgsEnv+"="+string(encodedArgs),
	)
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err == nil {
		return stdout.Bytes(), stderr.Bytes(), 0
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("run CLI subprocess: %v", err)
	}
	return stdout.Bytes(), stderr.Bytes(), exitErr.ExitCode()
}
