// Copyright 2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/dependencyrisk"
)

const trustedReviewRegistryPath = "governance/dependency-review-registry.json"

var trustedGitObjectID = regexp.MustCompile("^(?:[0-9a-fA-F]{40}|[0-9a-fA-F]{64})$")

type trustedReviewRecordEvidence struct {
	RecordKey       string `json:"matched_review_record"`
	Reviewer        string `json:"reviewer"`
	ReviewReference string `json:"review_reference"`
	ReviewedAt      string `json:"reviewed_at"`
	ReviewExpiresAt string `json:"review_expires_at"`
}

type trustedReviewEvidence struct {
	SourceRevision string                        `json:"trusted_registry_revision"`
	Records        []trustedReviewRecordEvidence `json:"matched_reviews"`
}

// loadTrustedReview reads the registry from the protected base revision, never
// from the pull-request worktree. A record is evidence only for an exact
// ecosystem/name/version match; it does not lower the classifier's risk.
func loadTrustedReview(base, policyVersion string, req requestFile, now time.Time) (*trustedReviewEvidence, error) {
	if !trustedGitObjectID.MatchString(strings.TrimSpace(base)) {
		return nil, errors.New("trusted base revision must be a full hexadecimal Git object ID")
	}
	// #nosec G204 -- base is validated as a full Git object ID; arguments are passed without a shell.
	raw, err := exec.CommandContext(context.Background(), "git", "show", base+":"+trustedReviewRegistryPath).Output()
	if err != nil {
		return nil, fmt.Errorf("read trusted registry from base revision %s: %w", base, err)
	}
	author := pullRequestAuthor()
	if strings.EqualFold(os.Getenv("GITHUB_EVENT_NAME"), "pull_request") && author == "" {
		return nil, errors.New("pull request author identity unavailable; trusted review fails closed")
	}
	return validateTrustedReview(raw, base, policyVersion, req, now, author)
}

func validateTrustedReview(raw []byte, base, policyVersion string, req requestFile, now time.Time, author string) (*trustedReviewEvidence, error) {
	var registry dependencyrisk.ReviewRegistry
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return nil, fmt.Errorf("parse trusted registry: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("trusted registry contains trailing or malformed JSON data")
	}
	if err := dependencyrisk.ValidateReviewRegistry(registry, policyVersion, now); err != nil {
		return nil, fmt.Errorf("validate trusted registry: %w", err)
	}
	if len(req.Dependencies) == 0 {
		return nil, errors.New("changed dependency identities are unavailable; trusted review cannot be matched")
	}
	evidence := &trustedReviewEvidence{SourceRevision: base, Records: make([]trustedReviewRecordEvidence, 0, len(req.Dependencies))}
	seen := make(map[string]struct{}, len(req.Dependencies))
	for _, dep := range req.Dependencies {
		if strings.TrimSpace(dep.Version) == "" {
			return nil, fmt.Errorf("dependency %s %s has no exact version; trusted review cannot be matched", dep.Ecosystem, dep.Name)
		}
		key := strings.ToLower(strings.TrimSpace(dep.Ecosystem)) + ":" + strings.TrimSpace(dep.Name) + "@" + strings.TrimSpace(dep.Version)
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		record, ok := registry.FindExactReview(dep.Ecosystem, dep.Name, dep.Version)
		if !ok {
			return nil, fmt.Errorf("no trusted review for exact dependency %s", key)
		}
		if author != "" && strings.EqualFold(strings.TrimSpace(author), strings.TrimSpace(record.Reviewer)) {
			return nil, fmt.Errorf("reviewer %q cannot self-approve pull-request author %q", record.Reviewer, author)
		}
		evidence.Records = append(evidence.Records, trustedReviewRecordEvidence{
			RecordKey:       key,
			Reviewer:        record.Reviewer,
			ReviewReference: record.ReviewReference,
			ReviewedAt:      record.ReviewedAt,
			ReviewExpiresAt: record.ReviewExpiresAt,
		})
	}
	return evidence, nil
}

func pullRequestAuthor() string {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return ""
	}
	// #nosec G703 -- GITHUB_EVENT_PATH is assigned by the GitHub Actions runner, not a pull-request input.
	raw, err := os.ReadFile(eventPath)
	if err != nil {
		return ""
	}
	var event struct {
		PullRequest struct {
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"pull_request"`
	}
	if json.Unmarshal(raw, &event) != nil {
		return ""
	}
	return event.PullRequest.User.Login
}
