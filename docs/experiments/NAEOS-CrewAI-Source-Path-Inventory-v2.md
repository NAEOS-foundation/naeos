# NAEOS CrewAI Source Execution-Path Inventory v2

**Version:** 0.1  
**Status:** REACT AND NATIVE TOOL-CALLING PASS; FLOW TOOLACTION AND OPENAI AGENTS ADAPTER FAIL; LANGGRAPH ADAPTER UNKNOWN  
**Branch:** `feat/crewai-v2-harness-foundation`  
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

**Observation: hook-mediated in both inspected versions; public runtime DENY enforcement confirmed in Section 10.**

- In 1.15.16, `CrewAgentExecutor` in `lib/crewai/src/crewai/agents/crew_agent_executor.py` delegates parsed `AgentAction` execution to `execute_tool_and_check_finality`.
- In 1.15.16 and 1.15.44, `lib/crewai/src/crewai/utilities/tool_utils.py` constructs `ToolCallHookContext`, calls `run_before_tool_call_hooks`, and only proceeds to `ToolUsage.use/ause` if the hook did not block. It then invokes after-tool hooks.
- In 1.15.44, `AgentExecutor.execute_tool_action` delegates to the same utility function.

Source:
- [1.15.16 crew_agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/crew_agent_executor.py)
- [1.15.16 tool_utils.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/utilities/tool_utils.py)
- [1.15.44 agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_executor.py)
- [1.15.44 tool_utils.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/utilities/tool_utils.py)

**Interpretation:** the public `Crew.kickoff()` ReAct path was exercised with a deterministic LLM. The denying hook was observed and the independent ledger remained empty in both pinned versions; see Section 10.

### 2.2 Native tool-calling path

**Observation: a separate hook seam exists in both inspected versions.**

In `lib/crewai/src/crewai/agents/agent_executor.py`, native tool-call handling:
1. parses tool arguments;
2. resolves the original/structured tool;
3. creates `ToolCallHookContext`;
4. invokes `run_before_tool_call_hooks`;
5. invokes the resolved native function only when not blocked;
6. invokes after-tool hooks and emits lifecycle events.

The inspected method invokes the resolved function with the parsed `args_dict` after the before-hook decision.

Sources:
- [1.15.16 crew_agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/crew_agent_executor.py)
- [1.15.44 agent_executor.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_executor.py)

**Interpretation:** native tool calling was exercised separately from the ReAct utility path through public `Crew.kickoff()`. The denying hook was observed and the independent ledger remained empty in both pinned versions; see Section 11.

### 2.3 Flow `ToolAction`

**Observation: direct tool invocation; no CrewAI tool-hook dispatcher call is present in the inspected `ToolAction.run` implementation.**

- 1.15.16: `ToolAction.run` renders `with_` arguments and calls `self.tool.run(...)`.
- 1.15.44: `ToolAction.run` renders arguments and calls `self.tool.run(**rendered)`; additional failure-policy handling follows.

Sources:
- [1.15.16 flow/runtime/_actions.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/flow/runtime/_actions.py)
- [1.15.44 flow/runtime/_actions.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/flow/runtime/_actions.py)

**Interpretation:** runtime confirmation is recorded in Section 9. Both pinned versions execute the Flow ToolAction body and produce an independent side effect while the global before-tool hook is not dispatched.

### 2.4 OpenAI Agents tool adapter

**Observation: direct underlying tool invocation; no CrewAI tool-hook dispatcher call is present in the inspected wrapper.**

The adapter wrapper normalizes the arguments and calls `tool._run(**args_dict)`, awaiting the result if needed.

Sources:
- [1.15.16 openai_agent_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/agent_adapters/openai_agents/openai_agent_tool_adapter.py)
- [1.15.44 openai_agent_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_adapters/openai_agents/openai_agent_tool_adapter.py)

**Interpretation:** the converted SDK FunctionTool.on_invoke_tool callback was invoked in both pinned versions. It called BaseTool._run directly, did not dispatch the CrewAI before-tool hook, and produced one independent side effect; see Section 12.

### 2.5 LangGraph tool adapter

**Observation: direct BaseTool invocation; no CrewAI tool-hook dispatcher call is present in the inspected wrapper.**

The adapter wrapper dispatches to `tool.run(args[0])`, `tool.run(kwargs["input"])`, or `tool.run(**kwargs)` based on argument shape, then awaits if necessary.

Sources:
- [1.15.16 langgraph_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.16/lib/crewai/src/crewai/agents/agent_adapters/langgraph/langgraph_tool_adapter.py)
- [1.15.44 langgraph_tool_adapter.py](https://github.com/crewAIInc/crewAI/blob/1.15.44/lib/crewai/src/crewai/agents/agent_adapters/langgraph/langgraph_tool_adapter.py)

**Interpretation:** the converted StructuredTool invocation returned an unawaited coroutine in both pinned versions. The CrewAI hook was not dispatched, but the tool body also did not run; classify the policy result as UNKNOWN, not PASS or FAIL. See Section 13. The full LangGraph ToolNode orchestration remains unverified.

## 3. Consolidated static inventory

| Path | 1.15.16 source observation | 1.15.44 source observation | Runtime status |
|---|---|---|---|
| ReAct / text tool call | Before/after hooks in `tool_utils.py` | Before/after hooks in `tool_utils.py` | **PASS — runtime confirmed in both versions** |
| Native tool calling | Before-hook seam in `crew_agent_executor.py` | Before-hook seam in `agent_executor.py` | **PASS — runtime confirmed in both versions** |
| Flow `ToolAction` | Direct `tool.run` | Direct `tool.run` | **FAIL — runtime confirmed in both versions** |
| OpenAI Agents adapter | Direct `tool._run` | Direct `tool._run` | **FAIL — runtime confirmed in both versions** |
| LangGraph adapter | Direct `tool.run` | Direct `tool.run` | **UNKNOWN — unawaited coroutine; tool body not reached** |

## 4. Runtime harness entry points (1.15.44 source)

The static findings above identify concrete entry points for a future integration harness. These are source-derived entry points, not evidence that a runtime test has been run.

### 4.1 Flow ToolAction through the public Flow runtime

- Public runtime class: `crewai.flow.flow.Flow` (re-exporting the runtime implementation).
- Declarative constructor: `Flow.from_declaration(contents=...)`.
- Definition schema: `FlowDefinition` with `FlowMethodDefinition.do` set to `FlowToolActionDefinition`.
- Runtime method binding: `Flow._flow_post_init` -> `_action_bound_methods` -> `build_action`.
- Consequential entry point: `Flow.kickoff()`.
- Tool execution: `ToolAction.run` -> `self.tool.run(**rendered)`.

For the test, create an importable test-only `BaseTool` class with an append-only in-memory side-effect ledger. Build a declarative Flow referencing that class via `module:qualname`, register a denying before-tool hook, and invoke `kickoff()`. Assert whether the hook was called and whether the tool body changed the ledger. Keep the fixture importable by the declarative resolver; avoid a monkey-patched internal method.

### 4.2 OpenAI Agents adapter

- Adapter: `OpenAIAgentToolAdapter`.
- Conversion entry point: `_convert_tools_to_openai_format(tools)`.
- Tool wrapper is supplied to the OpenAI Agents `FunctionTool(on_invoke_tool=...)`.
- The wrapper calls `tool._run(**args_dict)` directly.

Test the converted FunctionTool through its supported invocation interface from the pinned `openai-agents` dependency. Do not merely call a copied wrapper or invoke `tool._run` directly. If the installed package's supported interface cannot be exercised deterministically, mark the path BLOCKED and record the exact dependency version.

### 4.3 LangGraph adapter

- Adapter: `LangGraphToolAdapter`.
- Conversion entry point: `configure_tools(tools)`, followed by `tools()`.
- For CrewAI BaseTool inputs, conversion creates a LangChain `StructuredTool` whose async wrapper calls `tool.run(...)` directly.

Test the returned StructuredTool through its supported async invocation API (for example, `ainvoke`) with the actual pinned `langchain-core` package. Assert the before-hook observation and independent side-effect ledger. If optional dependencies are absent, classify BLOCKED rather than PASS.

### 4.4 ReAct and native tool calling

Use the public Agent/Crew execution route that causes the pinned executor to parse and dispatch a tool call. Do not call `run_before_tool_call_hooks` directly in the v2 test. Use deterministic/mock LLM output if the public runtime supports it, and assert the real tool body counter. Test ReAct and native tool-calling independently for each version.

### 4.5 Version and dependency caveats

The 1.15.44 package metadata inspected declares core dependencies and does not expose dedicated extras named for these OpenAI Agents or LangGraph adapters in the shown optional-dependency list. Adapter dependencies are resolved dynamically through import helpers or inside the adapter. The harness must therefore record installed package versions explicitly and should not assume that installing `crewai` alone makes every adapter test runnable.

## 5. Consequences for the v2 harness

1. Keep ReAct and native tool calling as separate test entry points.
2. Treat Flow `ToolAction`, OpenAI Agents adapter, and LangGraph adapter as independent suspected bypass paths.
3. For every path, register a denying hook and assert both the hook observation and the independent side-effect ledger.
4. A missing hook callback plus a tool-body counter increment is a failure of the tested enforcement contract.
5. A missing callback with no side effect is not a pass by itself; classify the path as blocked/inconclusive until the harness proves the execution boundary and observation are valid.
6. Do not patch or monkey-patch the framework's internal hook dispatcher in a way that makes a bypass path appear covered.
7. Pin adapter dependencies and record their exact versions in the run manifest. If an optional adapter dependency is unavailable, mark that path `BLOCKED`, not `PASS`.

## 6. Relationship to existing NAEOS v1 harness

The current `scripts/external_validation_crewai_boundaries.py` registers hooks and calls `run_before_tool_call_hooks` directly in its own `invoke` helper, then invokes `tool.run`. It therefore validates the hook dispatcher and tool body in that controlled harness, but it does not enter through CrewAI's public Agent/Crew executor, Flow runtime, OpenAI Agents adapter, or LangGraph adapter. Its v1 report remains valid within its stated scope; it must not be described as full execution-path coverage.

## 7. Scope and limitations

- This is static source inspection of the two commits listed above.
- It does not establish that every Flow action, third-party adapter, user-defined wrapper, or downstream integration follows the inspected path.
- It does not establish whether a future CrewAI release changes these seams.
- It does not prove zero side effects, exactly-once execution, durable review/resume, or policy freshness.
- No runtime tests, integration tests, or side-effect assertions were executed as part of producing this inventory.

## 8. Next implementation step

Implement the smallest pinned harness that first proves the **ReAct path** and **Flow ToolAction path** against both versions. Then add the OpenAI Agents and LangGraph adapter paths using their real public adapter APIs. Native tool calling in both versions remains its own path, separate from ReAct. Execute the v2 scenario matrix only after the path inventory and harness entry points are verified.


## 9. Runtime evidence — Flow ToolAction DENY probe

The v2 harness executed the public Flow runtime path on GitHub Actions with Python 3.12.16 and the pinned CrewAI versions below. The test registered a global `before_tool_call` hook that returned DENY, invoked the Flow through `Flow.kickoff()`, and recorded tool-body execution independently in the side-effect ledger.

| CrewAI version | Hook dispatched | Tool-body entries | Independent side effects | Result |
|---|---:|---:|---:|---|
| 1.15.16 | No | 1 | 1 | **FAIL** |
| 1.15.44 | No | 1 | 1 | **FAIL** |

Both probes completed the execution path without a fixture or import error. The tool returned `tool-body-executed`; the ledger recorded exactly one side effect. This is a confirmed enforcement bypass for the tested global tool-hook contract on Flow `ToolAction`, not merely a static-source suspicion.

Evidence and raw JSON artifacts: [GitHub Actions run 44591869537](https://github.com/NAEOS-foundation/naeos/actions/runs/44591869537). The two matrix jobs intentionally exit non-zero on this result. The legacy policy-boundary job in the same workflow passes, but it does not cover the Flow execution path.

**Scope boundary:** this result does not establish behavior for ReAct, native tool calling, OpenAI Agents, or LangGraph. Those paths remain independently testable and must not inherit the Flow result.


## 10. Runtime evidence — ReAct DENY probe

The v2 harness exercised the public `Crew.kickoff()` path with a deterministic `BaseLLM` that emitted one ReAct tool action followed by a final answer. A global `before_tool_call` hook returned DENY, and the test tool body wrote to the independent side-effect ledger only if it actually ran.

| CrewAI version | Hook dispatched | Tool-body entries | Independent side effects | Result |
|---|---:|---:|---:|---|
| 1.15.16 | Yes | 0 | 0 | **PASS** |
| 1.15.44 | Yes | 0 | 0 | **PASS** |

Both versions recorded exactly one DENY hook observation for `naeos_v2_react_probe`, completed the ReAct loop with two LLM calls, and returned `policy test completed`. The tool body was not entered and the independent ledger remained empty.

Evidence and raw JSON artifacts: [GitHub Actions run 44592362102](https://github.com/NAEOS-foundation/naeos/actions/runs/44592362102). Both ReAct matrix jobs completed successfully.

**Scope boundary:** ReAct PASS does not imply Flow ToolAction, native tool calling, OpenAI Agents, or LangGraph are covered. Flow ToolAction is independently confirmed FAIL in Section 9; the remaining paths stay untested until their own runtime evidence exists.


## 11. Runtime evidence — native tool-calling DENY probe

The v2 harness exercised native tool calling independently from the ReAct text path. A deterministic `BaseLLM` explicitly advertised native function-calling support and returned a structured native function-call response through the public `Crew.kickoff()` API. The global before-tool hook returned DENY; the tool body wrote to the independent side-effect ledger only if the resolved function actually ran.

| CrewAI version | Native path selected | Hook dispatched | Tool-body entries | Independent side effects | Result |
|---|---:|---:|---:|---:|---|
| 1.15.16 | Yes | Yes | 0 | 0 | **PASS** |
| 1.15.44 | Yes | Yes | 0 | 0 | **PASS** |

Both versions recorded exactly one DENY observation for `naeos_v2_native_probe`, completed the native tool-call loop with two LLM calls, returned `native tool policy test completed`, and left the independent ledger empty.

Evidence and raw JSON artifacts: [GitHub Actions run 44594967725](https://github.com/NAEOS-foundation/naeos/actions/runs/44594967725). Both native tool-calling matrix jobs completed successfully.

**Scope boundary:** this PASS covers CrewAI's native tool-calling path only. It does not establish compatibility for the OpenAI Agents adapter or LangGraph adapter, which remain separate execution paths and require their own runtime probes.


## 12. Runtime evidence — OpenAI Agents adapter DENY probe

The v2 harness installed the optional OpenAI Agents SDK, converted the CrewAI test tool through `OpenAIAgentToolAdapter.configure_tools()`, and invoked the resulting SDK `FunctionTool.on_invoke_tool` callback with a deterministic JSON argument payload. This tests the actual adapter callback and SDK tool object; it does not claim to exercise the full OpenAI Agents `Runner` model loop.

| CrewAI version | Hook dispatched | Tool-body entries | Independent side effects | Result |
|---|---:|---:|---:|---|
| 1.15.16 | No | 1 | 1 | **FAIL** |
| 1.15.44 | No | 1 | 1 | **FAIL** |

Both adapter callbacks completed and returned `tool-body-executed`. The CrewAI global before-tool hook was not dispatched, and the independent ledger recorded one side effect under DENY.

Evidence and raw JSON artifacts: [GitHub Actions run 44594967725](https://github.com/NAEOS-foundation/naeos/actions/runs/44594967725). Both OpenAI Agents adapter matrix jobs failed intentionally on the observed side effect.

**Scope boundary:** this result applies to the CrewAI OpenAI Agents tool adapter callback. It does not claim a full OpenAI Agents Runner end-to-end conformance test.


## 13. Runtime evidence — LangGraph adapter invocation

The first attempt to exercise the converted tool through `ToolNode.ainvoke()` was BLOCKED before reaching the adapter invocation boundary by a LangGraph configuration error (`Missing required config key 'N/A' for 'tools'`). The probe was then narrowed to the converted LangChain `StructuredTool.ainvoke()` API so the adapter's own asynchronous wrapper could be observed directly.

| CrewAI version | Hook dispatched | Tool-body entries | Independent side effects | Adapter return | Policy result |
|---|---:|---:|---:|---|---|
| 1.15.16 | No | 0 | 0 | Unawaited coroutine object | **UNKNOWN** |
| 1.15.44 | No | 0 | 0 | Unawaited coroutine object | **UNKNOWN** |

Both converted tool invocations returned successfully at the outer `ainvoke()` call, but the result was a coroutine object from the adapter's async wrapper. The wrapper was not awaited, so the underlying `BaseTool.run()` and tool body were never entered. There was no side effect, but there was also no observed DENY enforcement; therefore this is UNKNOWN, not PASS.

The source constructs `StructuredTool(func=tool_wrapper, ...)` while `tool_wrapper` is asynchronous. LangChain's `StructuredTool` separates synchronous `func` from asynchronous `coroutine`; the adapter should be checked for using the appropriate asynchronous field or a synchronous wrapper. This is a likely adapter-construction defect, not proof of policy enforcement.

Evidence and raw JSON artifacts: [GitHub Actions run 44594967725](https://github.com/NAEOS-foundation/naeos/actions/runs/44594967725). This test exercises the converted StructuredTool API, not a full LangGraph graph/ToolNode orchestration.
