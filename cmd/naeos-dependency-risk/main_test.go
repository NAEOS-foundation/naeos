// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (\n\t"os"\n\t"os/exec"\n\t"path/filepath"\n\t"strings"\n\t"testing"\n)

func TestRequireLinePreservesGoVersionPrefix(t *testing.T) {
	match := requireLine.FindStringSubmatch("- github.com/klauspost/compress v1.18.5 // indirect")
	if len(match) != 3 || match[2] != "v1.18.5" {
		t.Fatalf("requireLine match = %#v, want version v1.18.5", match)
	}
}

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
	if req.Name != "broken" || req.VersionChange != "unknown" || req.KnownDependency {
		t.Fatalf("unknown version change must dominate aggregate: %+v", req)
	}
}

func TestSelectHighestRiskRequestSelectsMajorOverPatch(t *testing.T) {
	req := selectHighestRiskRequest(
		map[string]string{"safe": "v1.2.3", "major": "v1.2.3"},
		map[string]string{"safe": "v1.2.4", "major": "v2.0.0"},
	)
	if req.Name != "major" || req.VersionChange != "major" || req.KnownDependency {
		t.Fatalf("major change must dominate patch change: %+v", req)
	}
}

func TestSelectHighestRiskRequestAllowsAllSafeChangesToRemainPatch(t *testing.T) {
	req := selectHighestRiskRequest(
		map[string]string{"one": "v1.2.3", "two": "v2.4.1"},
		map[string]string{"one": "v1.2.4", "two": "v2.4.2"},
	)
	if req.VersionChange != "patch" || req.KnownDependency {
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

func TestNPMVersionChange(t *testing.T) {
	tests := []struct {
		name, oldv, newv, want string
	}{
		{"patch", "0.35.4", "0.35.5", "patch"},
		{"major", "11.17.2", "12.1.0", "major"},
		{"caret-patch", "^1.2.3", "^1.2.4", "patch"},
		{"caret-minor", "^1.2.3", "^1.3.0", "minor"},
		{"unknown-range", "workspace:*", "workspace:*", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := npmVersionChange(tt.oldv, tt.newv); got != tt.want {
				t.Fatalf("npmVersionChange(%q, %q) = %q, want %q", tt.oldv, tt.newv, got, tt.want)
			}
		})
	}
}

func TestDeriveGoToolchainRequestPatchDirective(t *testing.T) {
	request, changed := deriveGoToolchainRequest("module example.com/naeos\n\ngo 1.26.10\n", "module example.com/naeos\n\ngo 1.26.11\n")
	if !changed || request.Ecosystem != "go-toolchain" || request.Name != "go" || request.VersionChange != "patch" || !request.KnownDependency {
		t.Fatalf("expected a known Go toolchain patch classification: changed=%v request=%+v", changed, request)
	}
}

func TestDeriveGoToolchainRequestNoChange(t *testing.T) {
	_, changed := deriveGoToolchainRequest("module example.com/naeos\n\ngo 1.26.11\n", "module example.com/naeos\n\ngo 1.26.11\n")
	if changed {
		t.Fatal("same Go directive should not be classified as a toolchain change")
	}
}
func TestDeriveEmptyNewGoModuleRecognizesDependencyFreeNestedModule(t *testing.T) {
	workspace := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(workspace); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(originalDir) })
	runGitTestCommand(t, workspace, "init")
	runGitTestCommand(t, workspace, "config", "user.email", "test@example.invalid")
	runGitTestCommand(t, workspace, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(workspace, "go.mod"), []byte("module example.com/root\n\ngo 1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitTestCommand(t, workspace, "add", "go.mod")
	runGitTestCommand(t, workspace, "commit", "-m", "base")
	base := strings.TrimSpace(runGitTestCommandOutput(t, workspace, "rev-parse", "HEAD"))
	nested := filepath.Join(workspace, "services", "new-service")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(nested, "go.mod")
	if err := os.WriteFile(manifest, []byte("module example.com/new-service\n\ngo 1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	relative := "services/new-service/go.mod"
	req, ok, err := deriveEmptyNewGoModule(base, relative, []string{relative})
	if err != nil || !ok {
		t.Fatalf("expected dependency-free module classification: ok=%v err=%v", ok, err)
	}
	if req.Ecosystem != "go-module-empty" || req.Name != "new-empty-module" || req.VersionChange != "patch" || !req.KnownDependency {
		t.Fatalf("unexpected classification: %+v", req)
	}
	if _, ok, err := deriveEmptyNewGoModule(base, relative, []string{relative, "package-lock.json"}); err != nil || ok {
		t.Fatalf("empty module must not mask another manifest change: ok=%v err=%v", ok, err)
	}
}

func TestDeriveEmptyNewGoModuleRejectsDependenciesAndGoSum(t *testing.T) {
	workspace := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	if err := os.Chdir(workspace); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(originalDir) })
	runGitTestCommand(t, workspace, "init")
	runGitTestCommand(t, workspace, "config", "user.email", "test@example.invalid")
	runGitTestCommand(t, workspace, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(workspace, "go.mod"), []byte("module example.com/root\n\ngo 1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitTestCommand(t, workspace, "add", "go.mod")
	runGitTestCommand(t, workspace, "commit", "-m", "base")
	base := strings.TrimSpace(runGitTestCommandOutput(t, workspace, "rev-parse", "HEAD"))
	nested := filepath.Join(workspace, "services", "new-service")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(nested, "go.mod")
	if err := os.WriteFile(manifest, []byte("module example.com/new-service\n\ngo 1.26.0\n\nrequire example.com/dependency v1.2.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	relative := "services/new-service/go.mod"
	if _, ok, err := deriveEmptyNewGoModule(base, relative, []string{relative}); err != nil || ok {
		t.Fatalf("module with dependencies must not be classified empty: ok=%v err=%v", ok, err)
	}
	if err := os.WriteFile(filepath.Join(nested, "go.sum"), []byte("evidence checksum\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := deriveEmptyNewGoModule(base, relative, []string{relative, "services/new-service/go.sum"}); err != nil || ok {
		t.Fatalf("module with go.sum change must not be classified empty: ok=%v err=%v", ok, err)
	}
}

func runGitTestCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	if output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func runGitTestCommandOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
	return string(output)
}