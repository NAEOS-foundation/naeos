	}

	return selectHighestRiskRequest(oldv, newv), nil
}

// deriveEmptyNewGoModule recognizes a newly added nested Go module with no
// declared dependencies. This is a narrowly-scoped, evidence-backed no-dependency
// change; modules with requirements or accompanying go.sum changes remain on the
// normal fail-closed dependency path.
func deriveEmptyNewGoModule(base, manifest string, changed []string) (requestFile, bool, error) {
	// Only accept the empty-module classification when this is the sole
	// dependency manifest change; otherwise a second manifest could be masked.
	if len(changed) != 1 || changed[0] != manifest {
		return requestFile{}, false, nil
	}
	if _, err := gitShowFile(base, manifest); err == nil {
		return requestFile{}, false, nil
	}
	current, err := os.ReadFile(manifest)
	if err != nil {
		return requestFile{}, false, fmt.Errorf("read new Go module %s: %w", manifest, err)
	}
	parsed, err := modfile.Parse(manifest, current, nil)
	if err != nil {
		return requestFile{}, false, fmt.Errorf("parse new Go module %s: %w", manifest, err)
	}
	if parsed.Module == nil || len(parsed.Requires) != 0 {
		return requestFile{}, false, nil
	}
	return requestFile{
		PolicyVersion: "1.0.0", SchemaVersion: "1.0.0",
		Ecosystem: "go-module-empty", Name: "new-empty-module",
		VersionChange: "patch", Paths: []string{manifest},
		Evidence: automaticEvidenceAvailable(), KnownDependency: true,
	}, true, nil
}

func changedDependencyManifests(base string) ([]string, error) {
	raw, err := exec.CommandContext(context.Background(), "git", "diff", "--name-only", base+"...HEAD").Output() //nolint:gosec // BASE_SHA is supplied by trusted CI
	if err != nil {
		return nil, fmt.Errorf("list changed files: %w", err)
	}
	var out []string
	for _, path := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if path == "" {
			continue
		}
		baseName := filepath.Base(path)
		switch baseName {
		case "go.mod", "go.sum", "package.json", "package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "requirements.txt", "pyproject.toml", "Cargo.toml", "Cargo.lock":
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out, nil
}

func gitShowFile(base, path string) ([]byte, error) {
	out, err := exec.CommandContext(context.Background(), "git", "show", base+":"+path).Output() //nolint:gosec // refs are supplied by trusted CI
	if err != nil {
		return nil, err
	}
	return out, nil
}

func deriveNPMRequest(base, path string) (requestFile, bool, error) {
	newBytes, err := os.ReadFile(path)
	if err != nil {
		return requestFile{}, false, fmt.Errorf("read npm manifest %s: %w", path, err)
	}
	oldBytes, oldErr := gitShowFile(base, path)
	if oldErr != nil {
		oldBytes = []byte("{}")
	}
	var oldPkg, newPkg map[string]any
	if err := json.Unmarshal(oldBytes, &oldPkg); err != nil {
		return requestFile{}, false, fmt.Errorf("parse base npm manifest %s: %w", path, err)
	}
	if err := json.Unmarshal(newBytes, &newPkg); err != nil {
		return requestFile{}, false, fmt.Errorf("parse npm manifest %s: %w", path, err)
	}
	lockPath, err := npmLockPathForManifest(base, path)
	if err != nil {
		return requestFile{}, false, fmt.Errorf("find npm lockfile for %s: %w", path, err)
	}
	var oldLockBytes, newLockBytes []byte
	if lockPath != "" {
		lockInfo, statErr := os.Lstat(lockPath) //nolint:gosec // lockPath is constrained to a workspace-relative lockfile below
		if statErr != nil {
			return requestFile{}, false, fmt.Errorf("inspect npm lockfile %s: %w", lockPath, statErr)
		}
		if !lockInfo.Mode().IsRegular() {
			return requestFile{}, false, fmt.Errorf("npm lockfile %s is not a regular file", lockPath)
		}
		newLockBytes, err = os.ReadFile(lockPath) //nolint:gosec // path was constrained to a workspace-relative lockfile above
		if err != nil {
			return requestFile{}, false, fmt.Errorf("read npm lockfile %s: %w", lockPath, err)
		}
		oldLockBytes, _ = gitShowFile(base, lockPath)
	}

	sections := []string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies"}
	directChanges := make(map[string]struct{})
	for _, section := range sections {
		oldDeps, _ := oldPkg[section].(map[string]any)
		newDeps, _ := newPkg[section].(map[string]any)
		names := make(map[string]struct{}, len(oldDeps)+len(newDeps))
		for name := range oldDeps {
			names[name] = struct{}{}
		}
		for name := range newDeps {
			names[name] = struct{}{}
		}
		for name := range names {
			oldSpec, oldOK := oldDeps[name].(string)
			newSpec, newOK := newDeps[name].(string)
			if !oldOK || !newOK || oldSpec != newSpec {
				directChanges[name] = struct{}{}
			}
		}
	}
	directNames := make([]string, 0, len(directChanges))
	for name := range directChanges {
		directNames = append(directNames, name)
	}
	sort.Strings(directNames)

	best := requestFile{
		PolicyVersion: "1.0.0", SchemaVersion: "1.0.0", Ecosystem: "npm", Name: "dependency-change",
		VersionChange: "unknown", Evidence: automaticEvidenceAvailable(), KnownDependency: false,
	}