# NAEOS CrewAI Flow ToolAction Runtime Evidence v2

**Schema:** naeos.crewai-compatibility-v2.flow-tool-action/0.2  
**Run:** [GitHub Actions 44591572137](https://github.com/NAEOS-foundation/naeos/actions/runs/44591572137)  
**PR head used by run:** a20767368f372e3307e813eb940c66c384225556  
**GitHub Actions merge SHA:** c8978d764da525672dc6765c90e1f5a144d21c4c  
**Python:** 3.12.16  
**Execution path:** public Flow.kickoff -> @start method -> build_action -> ToolAction.run -> BaseTool.run

> This is a negative compatibility result for CrewAI's global before_tool_call hook on the Flow ToolAction path. It does not assert that all NAEOS policy integrations are bypassed; it establishes that this particular CrewAI hook is not an enforcement boundary for this path.

## Results

| CrewAI version | Hook dispatched | Tool-body entries | Independent side effects | Result |
|---|---:|---:|---:|---|
| 1.15.16 | No | 1 | 1 | **FAIL** |
| 1.15.44 | No | 1 | 1 | **FAIL** |

Both matrix jobs successfully installed the pinned CrewAI version, entered the public Flow kickoff path, completed the Flow, and observed the test tool body execute. The registered DENY hook was not called. The independent ledger recorded exactly one side effect for each version.

## Evidence integrity

| Version | Side-effect ledger checksum | Artifact |
|---|---|---|
| 1.15.16 | sha256:f05e1bfa5cdce83f301c8619973c4d530369a51c612b7345679b72c7d4c5f5d1 | [Artifact 14563688021](https://github.com/NAEOS-foundation/naeos/actions/runs/44591572137/artifacts/14563688021) |
| 1.15.44 | sha256:8d73939d620779616e03f9cdfead1ac550dbdc3bd140a45ba6dbb29b58f5dc1f | [Artifact 14564410230](https://github.com/NAEOS-foundation/naeos/actions/runs/44591572137/artifacts/14564410230) |

The run logs report:
- decision = DENY
- hook_dispatch_observed = false
- tool_body_entries = 1
- side_effect_count = 1
- execution_completed = true
- compatibility_result = FAIL

## Interpretation

The experiment confirms the suspected Flow ToolAction hook bypass in both pinned versions. A green workflow is not expected for this negative test: the matrix job is intentionally red when the DENY contract is violated. The evidence artifact is still uploaded using if: always().

## Required remediation / design consequence

- Do not treat CrewAI's global before_tool_call hook as a universal policy boundary.
- Any NAEOS-controlled Flow action must invoke an explicit NAEOS authorization boundary before delegating to BaseTool.run(), or use a policy-enforcing tool wrapper that cannot be bypassed by direct Flow invocation.
- Preserve the independent side-effect ledger in regression tests.
- Do not suppress the observed FAIL or convert it into a warning-only PASS.
- Keep ReAct, native tool calling, OpenAI Agents, and LangGraph as separate execution-path tests; their outcomes must not be inferred from this Flow result.
