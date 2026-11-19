# NAEOS CrewAI Source Execution-Path Inventory v2

**Version:** 0.1  
**Status:** SOURCE INSPECTION COMPLETE — RUNTIME VALIDATION NOT EXECUTED  
**Branch:** `feat/crewai-compatibility-validation-v2`  
**Related plan:** [NAEOS CrewAI Compatibility Validation v2](./NAEOS-CrewAI-Compatibility-Validation-v2.md)

> This inventory records static source observations for two pinned upstream tags. It is not a runtime conformance result, security certification, or proof that all downstream integrations use these paths exactly as inspected.

## 1. Source revisions inspected

| CrewAI tag | Tag object | Commit |
|---|---|---|
| `1.15.16` | `78d47e4232020ce5b66303ccdf53c4fda7334803` | `808ecc95f51d4923a3635f6209909f4f77fef9ed` |
| `1.15.44` | `6f3a1e66052dbbe0d6335ccdb4a43c15a9b2a949` | `374c05392021271cabb809868282d949e0efbbe2` |

Tags are named `1.15.16` and `1.15.44` (without a `v` prefix). The commit SHAs above are the immutable source revisions used for this inspection.

## 2. Execution-path findings

### 2.1 Standard ReAct / text tool-call path

**Observation: hook-mediated in both inspected versions.**

- In 1.15.16, `CrewAgentExecutor` in `lib/crewai/src/crewai/agents/crew_agent_executor.py` delegates parsed `AgentAction` execution to `execute_tool_and_check_finality`.
- In 1.15.16 and 1.15.44, `lib/crewai/src/crewai/utilities/tool_utils.py` constructs `ToolCallHookContext`, calls `run_before_tool_call_hooks`, and only proceeds to `ToolUsage.use/ause` if the hook did not block. It then invokes after-tool hooks.
- In 1.15.44, `AgentExecutor.execute_tool_action` delegates to the same utility function.

Source:
- [1.15.16 crew_agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/crew_agent_executor.py)
- [1.15.16 tool_utils.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/utilities/tool_utils.py)
- [1.15.44 agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_executor.py)
- [1.15.44 tool_utils.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/utilities/tool_utils.py)

**Interpretation:** this is a candidate path for an actual-runtime hook test. A direct call to the dispatcher alone is insufficient; the harness must enter through the public agent/crew execution path and independently count tool-body side effects.

### 2.2 Native tool-calling path

**Observation: a separate hook seam exists in both inspected versions.**

In `lib/crewai/src/crewai/agents/agent_executor.py`, native tool-call handling:
1. parses tool arguments;
2. resolves the original/structured tool;
3. creates `ToolCallHookContext`;
4. invokes `run_before_tool_call_hooks`;
5. invokes the resolved native function only when not blocked;
6. invokes after-tool hooks and emits lifecycle events.

The inspected method contains a direct call through `self._available_functions[func_name](**args_dict)` after the before-hook decision.

Sources:
- [1.15.16 crew_agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/crew_agent_executor.py)
- [1.15.44 agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_executor.py)

**Interpretation:** test native tool calling separately from the ReAct utility path in each version. Static source inspection indicates both versions have a before-hook seam; runtime enforcement remains unverified.

### 2.3 Flow `ToolAction`

**Observation: direct tool invocation; no CrewAI tool-hook dispatcher call is present in the inspected `ToolAction.run` implementation.**

- 1.15.16: `ToolAction.run` renders `with_` arguments and calls `self.tool.run(...)`.
- 1.15.44: `ToolAction.run` renders arguments and calls `self.tool.run(**rendered)`; additional failure-policy handling follows.

Sources:
- [1.15.16 flow/runtime/_actions.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/flow/runtime/_actions.py)
- [1.15.44 flow/runtime/_actions.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/flow/runtime/_actions.py)

**Interpretation:** classify as **HOOK BYPASS RISK / NOT COVERED BY THE OBSERVED TOOL-HOOK SEAM**. Confirm with a runtime test that registers a denying before-tool hook, executes a Flow `ToolAction`, and checks an independent side-effect counter. Do not claim a bypass test has passed until executed.

### 2.4 OpenAI Agents tool adapter

**Observation: direct underlying tool invocation; no CrewAI tool-hook dispatcher call is present in the inspected wrapper.**

The adapter wrapper normalizes the arguments and calls `tool._run(**args_dict)`, awaiting the result if needed.

Sources:
- [1.15.16 openai_agent_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/agent_adapters/openai_agents/openai_agent_tool_adapter.py)
- [1.15.44 openai_agent_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_adapters/openai_agents/openai_agent_tool_adapter.py)

**Interpretation:** classify as **HOOK BYPASS RISK / NOT COVERED BY THE OBSERVED TOOL-HOOK SEAM**. Validate through the adapter's real conversion and invocation API, with a denying hook and independent tool-body counter. Confirm supported optional dependencies and public construction API before writing the test.

### 2.5 LangGraph tool adapter

**Observation: direct BaseTool invocation; no CrewAI tool-hook dispatcher call is present in the inspected wrapper.**

The adapter wrapper dispatches to `tool.run(args[0])`, `tool.run(kwargs["input"])`, or `tool.run(**kwargs)` based on argument shape, then awaits if necessary.

Sources:
- [1.15.16 langgraph_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/agent_adapters/langgraph/langgraph_tool_adapter.py)
- [1.15.44 langgraph_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_adapters/langgraph/langgraph_tool_adapter.py)

**Interpretation:** classify as **HOOK BYPASS RISK / NOT COVERED BY THE OBSERVED TOOL-HOOK SEAM**. Validate through the actual LangGraph adapter output and invocation path; a unit test of the wrapper alone is not an end-to-end integration result.

## 3. Consolidated static inventory

| Path | 1.15.16 source observation | 1.15.44 source observation | Runtime status |
|---|---|---|---|
| ReAct / text tool call | Before/after hooks in `tool_utils.py` | Before/after hooks in `tool_utils.py` | NOT EXECUTED |
| Native tool calling | Before-hook seam in `crew_agent_executor.py` | Before-hook seam in `agent_executor.py` | NOT EXECUTED |
| Flow `ToolAction` | Direct `tool.run` | Direct `tool.run` | NOT EXECUTED |
| OpenAI Agents adapter | Direct `tool._run` | Direct `tool._run` | NOT EXECUTED |
| LangGraph adapter | Direct `tool.run` | Direct `tool.run` | NOT EXECUTED |

## 4. Consequences for the v2 harness

1. Keep ReAct and native tool calling as separate test entry points.
2. Treat Flow `ToolAction`, OpenAI Agents adapter, and LangGraph adapter as independent suspected bypass paths.
3. For every path, register a denying hook and assert both the hook observation and the independent side-effect ledger.
4. A missing hook callback plus a tool-body counter increment is a failure of the tested enforcement contract.
5. A missing callback with no side effect is not a pass by itself; classify the path as blocked/inconclusive until the harness proves the execution boundary and observation are valid.
6. Do not patch or monkey-patch the framework's internal hook dispatcher in a way that makes a bypass path appear covered.
7. Pin adapter dependencies and record their exact versions in the run manifest. If an optional adapter dependency is unavailable, mark that path `BLOCKED`, not `PASS`.

## 5. Relationship to existing NAEOS v1 harness

The current `scripts/external_validation_crewai_boundaries.py` registers hooks and calls `run_before_tool_call_hooks` directly in its own `invoke` helper, then invokes `tool.run`. It therefore validates the hook dispatcher and tool body in that controlled harness, but it does not enter through CrewAI's public Agent/Crew executor, Flow runtime, OpenAI Agents adapter, or LangGraph adapter. Its v1 report remains valid within its stated scope; it must not be described as full execution-path coverage.

## 6. Scope and limitations

- This is static source inspection of the two commits listed above.
- It does not establish that every Flow action, third-party adapter, user-defined wrapper, or downstream integration follows the inspected path.
- It does not establish whether a future CrewAI release changes these seams.
- It does not prove zero side effects, exactly-once execution, durable review/resume, or policy freshness.
- No runtime tests, integration tests, or side-effect assertions were executed as part of producing this inventory.

## 7. Next implementation step

Implement the smallest pinned harness that first proves the **ReAct path** and **Flow ToolAction path** against both versions. Then add the OpenAI Agents and LangGraph adapter paths using their real public adapter APIs. Native tool calling in both versions remains its own path, separate from ReAct. Execute the v2 scenario matrix only after the path inventory and harness entry points are verified.
