			"observer independently checked the filesystem",
			"observer found no side effect",
			"verification preserved the mismatch as a failure",
		},
		EvidenceID: lieRecord.ID,
		Verification: string(lieVerification.Status),
		FailureDetected: lieVerification.Status == verification.StatusFailed,
	})

	// 5. TAMPER: capture an observed artifact, mutate it, then verify again.
	tamperRoot, err := os.MkdirTemp(root, "tamper-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tamperRoot)
	tamperContent := []byte("original\n")
	tamperSandbox := &fileSandbox{root: tamperRoot, content: tamperContent, persist: true}
	tamperGateway, tamperControl := newGateway(policy.DecisionAllow, tamperSandbox)
	tamperResult, err := tamperGateway.Authorize(gateway.ToolRequest{
		Tool: "filesystem", Action: action, Resource: resource,
		Environment: environment, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	tamperObserver := observer{root: tamperRoot}
	tamperObserved, tamperExists, err := tamperObserver.Observe()
	if err != nil || !tamperExists {
		return nil, fmt.Errorf("tamper scenario initial observation failed: %v", err)
	}
	tamperRecord, err := appendEvidence(tamperControl, tamperResult, tamperObserved, tamperObserver)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(tamperRoot, fileName), []byte("mutated\n"), 0o600); err != nil {
		return nil, err
	}
	tamperVerification, err := verify(tamperRecord, tamperObserver)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name: "05-tamper-after-observation",
		Expected: "POST-OBSERVATION MUTATION MUST FAIL VERIFICATION",
		Observed: fmt.Sprintf("verification=%s", tamperVerification.Status),
		Passed: tamperVerification.Status == verification.StatusFailed,
		Checks: []string{
			"artifact was observed and hashed",
			"artifact was mutated after evidence capture",
			"independent verification detected digest mismatch",
		},
		EvidenceID: tamperRecord.ID,
		Verification: string(tamperVerification.Status),
		FailureDetected: tamperVerification.Status == verification.StatusFailed,
	})

	reproducible := reproducibleProbe()
	results = append(results, scenarioResult{
		Name: "06-reproducibility",
		Expected: "REPEATED CONTROLLED SIDE EFFECTS PRODUCE THE SAME OBSERVED DIGEST",
		Observed: fmt.Sprintf("reproducible=%v", reproducible),
		Passed: reproducible,
		Checks: []string{
			"same deterministic input used for repeated local side effects",
			"independent observations produced identical SHA-256 digests",
		},
	})

	return results, nil
}

func reproducibleProbe() bool {
	content := []byte("reproducible side effect\n")
	firstRoot, err := os.MkdirTemp("", "naeos-repro-a-*")
	if err != nil { return false }
	defer os.RemoveAll(firstRoot)
	secondRoot, err := os.MkdirTemp("", "naeos-repro-b-*")
	if err != nil { return false }
	defer os.RemoveAll(secondRoot)

	first := &fileSandbox{root: firstRoot, content: content, persist: true}
	second := &fileSandbox{root: secondRoot, content: content, persist: true}
	if _, err := first.Execute(gateway.ToolRequest{Resource: resource, Action: action}); err != nil { return false }
	if _, err := second.Execute(gateway.ToolRequest{Resource: resource, Action: action}); err != nil { return false }
	firstObserved, firstExists, err := (observer{root: firstRoot}).Observe()
	if err != nil || !firstExists { return false }
	secondObserved, secondExists, err := (observer{root: secondRoot}).Observe()
	if err != nil || !secondExists { return false }
	return evidence.ComputeArtifactHash(firstObserved) == evidence.ComputeArtifactHash(secondObserved)
}

func newGateway(decision policy.Decision, sandbox gateway.Sandbox) (*gateway.ExecutionGateway, *control.ControlPlane) {
	reg := policy.NewRegistry()
	_ = reg.Register(&policy.Policy{
		ID: "level3-" + string(decision), Name: "Level 3 " + string(decision),
		Version: "1.0.0",
		Scope: policy.Scope{Resource: resource, Action: action, Environment: environment},
		Default: decision, Active: true,
	})
	cp := control.New(reg, control.FailClosed(true))
	return gateway.New(cp, sandbox), cp
}

func appendEvidence(cp *control.ControlPlane, result gateway.ExecutionResult, observed []byte, obs observer) (evidence.EvidenceRecord, error) {
	policyVersion := "unknown"
	decisions := cp.ListDecisions()
	if len(decisions) > 0 {
		policyVersion = decisions[len(decisions)-1].PolicyVersion
	}

	rec := evidence.EvidenceRecord{
		Actor: actor, Resource: result.Request.Resource, Action: result.Request.Action,
		Environment: environment, PolicyID: result.PolicyID, PolicyVersion: policyVersion,