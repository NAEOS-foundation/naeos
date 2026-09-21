// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/governance/review"
)

const okHeader = "// SPDX-License-Identifier: Apache-2.0\npackage main\n"

// reviewerScenarios attacks the heuristics in internal/governance/review/reviewer.go.
func reviewerScenarios() []Result {
	return []Result{
		scnToDoObfuscation(),
		scnPlaceholderObfuscation(),
		scnLicenseHeaderKeywordSpoof(),
	}
}

func scnToDoObfuscation() Result {
	rv := review.NewReviewer()
	// Baseline: the canonical TODO marker is caught by the no-todo rule.
	control, _ := rv.ReviewArtifact("main.go", okHeader+"// TODO: fix this\n", []string{"no-todo", "has-license-header"})
	var buf strings.Builder
	for _, variant := range []string{"to-do", "T0D0", "to do"} {
		r, err := rv.ReviewArtifact("main.go", okHeader+"// "+variant+": fix this\n", []string{"no-todo", "has-license-header"})
		if err != nil {
			fmt.Fprintf(&buf, "%s:error ", variant)
			continue
		}
		fmt.Fprintf(&buf, "%s=%s ", variant, r.Status)
	}
	evaded := control.Status != review.StatusApproved &&
		!strings.Contains(buf.String(), "changes_requested")
	return Result{
		Layer:    LayerReviewer,
		Scenario: "TODO obfuscation evades no-todo",
		Attack:   "no-todo uses Contains(lower(content), \"todo\"); 'to-do', 'T0D0', 'to do' retain a real TODO meaning yet pass approval",
		Bypassed: evaded, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("control=%s; evasions: %s", control.Status, strings.TrimSpace(buf.String())),
		Risk:     Medium,
	}
}

func scnPlaceholderObfuscation() Result {
	rv := review.NewReviewer()
	control, _ := rv.ReviewArtifact("main.go", okHeader+"// REPLACE_ME with value\n", []string{"no-placeholder", "has-license-header"})
	var buf strings.Builder
	for _, variant := range []string{"replace-me", "change.me", "CHANGE ME", "REPLACE-ME"} {
		r, err := rv.ReviewArtifact("main.go", okHeader+"// "+variant+" with value\n", []string{"no-placeholder", "has-license-header"})
		if err != nil {
			fmt.Fprintf(&buf, "%s:error ", variant)
			continue
		}
		fmt.Fprintf(&buf, "%s=%s ", variant, r.Status)
	}
	evaded := control.Status != review.StatusApproved &&
		!strings.Contains(strings.ToLower(buf.String()), "changes_requested")
	return Result{
		Layer:    LayerReviewer,
		Scenario: "placeholder obfuscation evades no-placeholder",
		Attack:   "no-placeholder only matches the literal lowercased set {placeholder, changeme, replace_me}; 'replace-me', 'change.me', 'CHANGE ME' are undetected",
		Bypassed: evaded, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("control=%s; evasions: %s", control.Status, strings.TrimSpace(buf.String())),
		Risk:     Medium,
	}
}

func scnLicenseHeaderKeywordSpoof() Result {
	rv := review.NewReviewer()
	r, err := rv.ReviewArtifact("main.go", "// Licensed under a fake MIT note\npackage main\nfunc main(){}\n", []string{"has-license-header", "no-todo", "no-placeholder"})
	if err != nil {
		return Result{Layer: LayerReviewer, Scenario: "license header keyword spoof", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: Medium}
	}
	return Result{
		Layer:    LayerReviewer,
		Scenario: "license header keyword spoof",
		Attack:   "has-license-header is satisfied by any of license/apache/mit/copyright within the first 20 lines; a fabricated marker passes",
		Bypassed: r.Status == review.StatusApproved, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("content without real header -> status=%s", r.Status),
		Risk:     Low,
	}
}
