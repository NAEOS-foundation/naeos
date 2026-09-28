// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestSelectHighestRiskRequestRejectsNewDependency(t *testing.T) {
	req := selectHighestRiskRequest(
		map[string]string{"safe": "v1.2.3"},
		map[string]string{"safe": "v1.2.4", "evil": "v1.0.0"},
	)
	if req.Name != "evil" || req.VersionChange != "unknown" || req.KnownDependency {
		t.Fatalf("new dependency must dominate aggregate: %+v", req)
	}
}

func TestSelectHighestRiskRequestRejectsUnknownAmongExisting(t *testing.T) {
	req := selectHighestRiskRequest(
		map[string]string{"safe": "v1.2.3", "broken": "not-a-version"},
		map[string]string{"safe": "v1.2.4", "broken": "still-invalid"},
	)
	if req.Name != "broken" || req.VersionChange != "unknown" || !req.KnownDependency {
		t.Fatalf("unknown version change must dominate aggregate: %+v", req)
	}
}

func TestSelectHighestRiskRequestSelectsMajorOverPatch(t *testing.T) {
	req := selectHighestRiskRequest(
		map[string]string{"safe": "v1.2.3", "major": "v1.2.3"},
		map[string]string{"safe": "v1.2.4", "major": "v2.0.0"},
	)
	if req.Name != "major" || req.VersionChange != "major" || !req.KnownDependency {
		t.Fatalf("major change must dominate patch change: %+v", req)
	}
}

func TestSelectHighestRiskRequestAllowsAllSafeChangesToRemainPatch(t *testing.T) {
	req := selectHighestRiskRequest(
		map[string]string{"one": "v1.2.3", "two": "v2.4.1"},
		map[string]string{"one": "v1.2.4", "two": "v2.4.2"},
	)
	if req.VersionChange != "patch" || !req.KnownDependency {
		t.Fatalf("all-patch changes should remain patch: %+v", req)
	}
}

func TestSelectHighestRiskRequestSelectsSecurityCriticalityOverGeneralPatch(t *testing.T) {
	req := requestFile{
		Ecosystem:       "go",
		Name:            "general-patch",
		VersionChange:   "patch",
		Evidence:        true,
		KnownDependency: true,
	}
	security := requestFile{
		Ecosystem:       "go",
		Name:            "security-patch",
		VersionChange:   "patch",
		Paths:           []string{"internal/security/token.go"},
		Evidence:        true,
		KnownDependency: true,
	}
	if requestRiskRank(security) <= requestRiskRank(req) {
		t.Fatalf(
			"security-critical patch must outrank general patch: security=%d general=%d",
			requestRiskRank(security),
			requestRiskRank(req),
		)
	}
}
