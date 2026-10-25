// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package sdk provides the initial additive public Go integration API for
// NAEOS. Its exported types are independent of packages under internal/.
package sdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/internal/neir/publicadapter"
	neirv1 "github.com/NAEOS-foundation/naeos/pkg/neir/v1"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

// Config contains the supported, implementation-independent pipeline options.
// Mode "governed" enables the fail-closed governance requirement.
type Config struct {
	Name              string
	Mode              string
	Verbose           bool
	DryRun            bool
	OutputDir         string
	Languages         []string
	Parallel          *bool
	RequireGovernance bool
	Policies          []PolicyRule
}

// PolicyRule defines a governance rule accepted by the public SDK.
type PolicyRule struct {
	ID        string
	Condition string
	Priority  int
	Action    string
	Scope     string
	Enabled   bool
}

// Client runs and validates NAEOS specifications using the supported SDK API.
type Client struct {
	pipeline *pipeline.Pipeline
}

// New creates a client from public configuration values.
func New(cfg Config) (*Client, error) {
	rules := make([]policy.Rule, len(cfg.Policies))
	for i, rule := range cfg.Policies {
		rules[i] = policy.Rule{
			RuleID: rule.ID, Condition: rule.Condition, Priority: rule.Priority,
			Action: rule.Action, Scope: rule.Scope, Enabled: rule.Enabled,
		}
	}

	impl, err := pipeline.New(pipeline.Config{
		Name: cfg.Name, Mode: cfg.Mode, Verbose: cfg.Verbose,
		DryRun: cfg.DryRun, OutputDir: cfg.OutputDir,
		Languages: append([]string(nil), cfg.Languages...), Parallel: cfg.Parallel,
		RequireGovernance: cfg.RequireGovernance || strings.EqualFold(cfg.Mode, "governed"),
		Policies:          rules,
	})
	if err != nil {
		return nil, fmt.Errorf("create NAEOS SDK client: %w", err)
	}
	return &Client{pipeline: impl}, nil
}

// Run executes the configured pipeline for specification input. It checks for
// cancellation before processing and between pipeline stages. A currently
// executing stage is not interrupted and must finish before cancellation is
// observed.
func (c *Client) Run(ctx context.Context, input string) (*Result, error) {
	if err := c.validate(ctx); err != nil {
		return nil, err
	}
	result, err := c.pipeline.RunContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return publicResult(result)
}

// Validate validates input without generating output artifacts. It checks for
// cancellation before and after validation; validation itself is not
// interrupted once started.
func (c *Client) Validate(ctx context.Context, input string) (*Result, error) {
	if err := c.validate(ctx); err != nil {
		return nil, err
	}
	result, err := c.pipeline.ValidateContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return publicResult(result)
}

func (c *Client) validate(ctx context.Context) error {
	if c == nil || c.pipeline == nil {
		return errors.New("NAEOS SDK client is nil")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	return ctx.Err()
}

// Result is the public, implementation-independent summary of one operation.
// Run and Validate populate the specification and NEIR hashes and return the
// versioned public NEIR document without exposing internal runtime types.
type Result struct {
	RunID             string
	SpecificationHash string
	NEIRHash          string
	NEIR              *neirv1.Document
	Artifacts         []Artifact
	Tasks             []TaskSummary
	Reviews           []ReviewSummary
	PolicyResults     []PolicyEvaluation
	Governance        GovernanceSummary
	Graph             GraphSummary
}

// Artifact contains a generated file. Content is copied from the pipeline result.
type Artifact struct {
	Path    string
	Content []byte
}

// TaskSummary describes one planned pipeline task.
type TaskSummary struct {
	ID           string
	Name         string
	Dependencies []string
	Priority     int
}

// ReviewSummary describes one artifact review and its findings.
type ReviewSummary struct {
	Status  string
	RuleID  string
	Message string
}

// PolicyEvaluation is the public outcome of evaluating one policy rule.
type PolicyEvaluation struct {
	Passed   bool
	RuleID   string
	Message  string
	Action   string
	Priority int
}

// GovernanceSummary records governance state and policy-context provenance.
type GovernanceSummary struct {
	Mode                 string
	Status               string
	EffectivePolicyCount int
	DisabledPolicyRules  []string
	PolicyContextVersion string
	PolicyContextDigest  string
}

// GraphSummary exposes graph size without exposing the internal graph model.
type GraphSummary struct {
	Nodes int
	Edges int
}

func publicResult(source *pipeline.Result) (*Result, error) {
	if source == nil {
		return nil, nil
	}
	publicNEIR, err := publicadapter.ToPublic(source.NEIR)
	if err != nil {
		return nil, fmt.Errorf("convert pipeline NEIR result: %w", err)
	}
	result := &Result{
		RunID:             source.RunID,
		SpecificationHash: source.SpecificationHash, NEIRHash: source.NEIRHash,
		NEIR: publicNEIR,
		Governance: GovernanceSummary{
			Mode: source.GovernanceMode, Status: source.GovernanceStatus,
			EffectivePolicyCount: source.EffectivePolicyCount,
			DisabledPolicyRules:  append([]string(nil), source.DisabledPolicyRules...),
			PolicyContextVersion: source.PolicyContextVersion,
			PolicyContextDigest:  source.PolicyContextDigest,
		},
	}
	if source.Graph != nil {
		result.Graph = GraphSummary{Nodes: source.Graph.NodeCount(), Edges: source.Graph.EdgeCount()}
	}
	for _, artifact := range source.Artifacts {
		result.Artifacts = append(result.Artifacts, Artifact{
			Path: artifact.Path, Content: append([]byte(nil), artifact.Content...),
		})
	}
	for _, task := range source.Tasks {
		result.Tasks = append(result.Tasks, TaskSummary{
			ID: task.ID, Name: task.Name,
			Dependencies: append([]string(nil), task.Dependencies...), Priority: task.Priority,
		})
	}
	for _, review := range source.Reviews {
		if review == nil {
			continue
		}
		if len(review.Comments) == 0 {
			result.Reviews = append(result.Reviews, ReviewSummary{
				Status: string(review.Status), Message: review.Summary,
			})
			continue
		}
		for _, comment := range review.Comments {
			result.Reviews = append(result.Reviews, ReviewSummary{
				Status: string(review.Status), RuleID: comment.RuleID, Message: comment.Message,
			})
		}
	}
	for _, evaluation := range source.PolicyResults {
		result.PolicyResults = append(result.PolicyResults, PolicyEvaluation{
			Passed: evaluation.Passed, RuleID: evaluation.RuleID,
			Message: evaluation.Message, Action: evaluation.Action, Priority: evaluation.Priority,
		})
	}
	return result, nil
}
