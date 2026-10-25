// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"context"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/pkg/sdk"
)

const testSpecification = "project: sdk-check\nmodules:\n  - name: core\n    path: ./core\n"

func TestRunReturnsPublicResult(t *testing.T) {
	client, err := sdk.New(sdk.Config{Name: "sdk-check", Languages: []string{"go"}, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Run(context.Background(), testSpecification)
	if err != nil {
		t.Fatal(err)
	}
	if result.NEIRHash == "" || len(result.Artifacts) == 0 || result.NEIR == nil || len(result.NEIR.Modules) == 0 {
		t.Fatalf("expected public hash and artifacts, got hash %q and %d artifacts", result.NEIRHash, len(result.Artifacts))
	}
}

func TestValidateReturnsHashes(t *testing.T) {
	client, err := sdk.New(sdk.Config{Name: "sdk-check", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Validate(context.Background(), testSpecification)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpecificationHash == "" || result.NEIRHash == "" || result.NEIR == nil || len(result.NEIR.Modules) == 0 {
		t.Fatalf("expected validation hashes, got specification=%q neir=%q", result.SpecificationHash, result.NEIRHash)
	}
}

func TestGovernedModeFailsClosedWithoutEnabledPolicy(t *testing.T) {
	client, err := sdk.New(sdk.Config{Mode: "GOVERNED", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Run(context.Background(), testSpecification); err == nil || !strings.Contains(err.Error(), "no effective policies") {
		t.Fatalf("expected fail-closed policy error, got %v", err)
	}
}

func TestGovernedModeEvaluatesPublicPolicy(t *testing.T) {
	client, err := sdk.New(sdk.Config{
		Mode:     "governed",
		DryRun:   true,
		Policies: []sdk.PolicyRule{{ID: "project-exists", Condition: "exists:project", Enabled: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Run(context.Background(), testSpecification)
	if err != nil {
		t.Fatal(err)
	}
	if result.Governance.Status != "evaluated" || len(result.PolicyResults) != 1 || !result.PolicyResults[0].Passed {
		t.Fatalf("expected passing policy evaluation, got governance %#v, results %#v", result.Governance, result.PolicyResults)
	}
}

func TestRunReturnsCanceledContext(t *testing.T) {
	client, err := sdk.New(sdk.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Run(ctx, testSpecification); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
