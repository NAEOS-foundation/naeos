#!/usr/bin/env python3
"""Real Temporal crash/recovery validation.

A worker process is deliberately terminated after a side effect but before
the Activity result is committed. A replacement worker then receives the
same Activity task. The experiment records the logical invocation ID and
checks whether recovery causes a duplicate side effect.

This is runtime evidence, not a claim that Temporal provides exactly-once
side effects: Temporal documents Activity execution as effectively-once,
with multiple task executions possible across retries.
"""
from __future__ import annotations

import asyncio
import json
import os
import signal
import subprocess
import sys
import tempfile
import time
from pathlib import Path

from temporalio import activity, workflow
from temporalio.client import Client
from temporalio.worker import Worker

TASK_QUEUE = "naeos-crash-recovery-003"
WORKFLOW_ID = "naeos-crash-recovery-003"
ACTIVITY_NAME = "consequential_side_effect"
EVIDENCE = Path("external-validation-evidence/temporal-evidence.json")


@activity.defn(name=ACTIVITY_NAME)
async def consequential_side_effect(invocation_id: str) -> str:
    state = Path(os.environ["NAEOS_STATE_FILE"])
    state.parent.mkdir(parents=True, exist_ok=True)
    with state.open("a", encoding="utf-8") as f:
        f.write(json.dumps({"invocation_id": invocation_id, "pid": os.getpid()}) + "\n")
        f.flush()
        os.fsync(f.fileno())

    # First execution crashes the worker after the side effect and before
    # returning the Activity result. The replacement worker must recover the
    # same logical Activity task.
    marker = state.with_suffix(".crashed")
    if not marker.exists():
        marker.write_text("crashed\n", encoding="utf-8")
        os.kill(os.getpid(), signal.SIGKILL)
    return "completed"


@workflow.defn
class CrashRecoveryWorkflow:
    @workflow.run
    async def run(self, invocation_id: str) -> str:
        from datetime import timedelta
        return await workflow.execute_activity(
            consequential_side_effect,
            invocation_id,
            start_to_close_timeout=timedelta(seconds=10),
        )


async def run_worker(client: Client, state: Path, crash: bool) -> int:
    os.environ["NAEOS_STATE_FILE"] = str(state)
    async with Worker(client, task_queue=TASK_QUEUE, workflows=[CrashRecoveryWorkflow], activities=[consequential_side_effect]):
        if crash:
            await asyncio.sleep(30)
        else:
            result = await client.execute_workflow(
                CrashRecoveryWorkflow.run,
                "inv-recovery-003",
                id=WORKFLOW_ID,
                task_queue=TASK_QUEUE,
            )
            assert result == "completed"
            return 0
    return 0


async def main() -> None:
    with tempfile.TemporaryDirectory() as td:
        state = Path(td) / "effects.jsonl"
        env = os.environ.copy()
        env["NAEOS_STATE_FILE"] = str(state)

        # Worker 1 owns the first Activity attempt and is expected to die.
        p1 = subprocess.Popen(
            [sys.executable, __file__, "worker1", str(state)],
            env=env,
        )
        await asyncio.sleep(2)

        client = await Client.connect("127.0.0.1:7233")
        try:
            await client.execute_workflow(
                CrashRecoveryWorkflow.run,
                "inv-recovery-003",
                id=WORKFLOW_ID,
                task_queue=TASK_QUEUE,
            )
        except Exception:
            # The client call may observe the first attempt's crash. Recovery
            # is driven by Temporal when worker 2 is available.
            pass

        for _ in range(30):
            if p1.poll() is not None:
                break
            await asyncio.sleep(0.5)

        # Worker 2 resumes the same Activity task.
        p2 = subprocess.Popen(
            [sys.executable, __file__, "worker2", str(state)],
            env=env,
        )
        try:
            for _ in range(60):
                try:
                    result = await client.get_workflow_handle(WORKFLOW_ID).result()
                    if result == "completed":
                        break
                except Exception:
                    pass
                await asyncio.sleep(0.5)
            else:
                raise AssertionError("workflow did not recover")
        finally:
            p2.terminate()
            p2.wait(timeout=5)

        records = [json.loads(line) for line in state.read_text(encoding="utf-8").splitlines()]
        invocation_ids = [r["invocation_id"] for r in records]
        print(json.dumps({"records": records, "execution_count": len(records), "logical_invocation_count": len(set(invocation_ids))}, indent=2))

        # Crucial finding: the same logical invocation may execute more than
        # once after a crash. NAEOS must therefore not equate authorization
        # with exactly-once side effects. Evidence must expose attempt/recovery
        # state and bind the outcome to the logical invocation.
        if len(records) < 2:
            raise AssertionError("expected a post-crash retry attempt")
        EVIDENCE.parent.mkdir(parents=True, exist_ok=True)
        EVIDENCE.write_text(json.dumps({
            "schema": "naeos.external-validation-temporal.v1",
            "status": "PASS",
            "runtime": "Temporal local dev server",
            "logical_invocation_id": "inv-recovery-003",
            "worker_boundary": "worker-1 -> worker-2",
            "side_effect_attempts": len(records),
            "logical_invocations": len(set(invocation_ids)),
            "finding": "crash after side effect before Activity completion caused a retry; evidence must distinguish logical invocation from execution attempt",
            "authorization_and_execution_are_distinct": True,
        }, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] in {"worker1", "worker2"}:
        asyncio.run(run_worker(
            asyncio.run(Client.connect("127.0.0.1:7233")),
            Path(sys.argv[2]),
            sys.argv[1] == "worker1",
        ))
    else:
        asyncio.run(main())
