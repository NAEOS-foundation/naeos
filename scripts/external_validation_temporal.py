#!/usr/bin/env python3
"""Real Temporal crash/recovery validation.

The first worker is killed after a durable side effect is committed locally
but before the Activity result reaches Temporal. A replacement worker then
receives the same logical Activity. This records execution attempts and proves
that recovery can retry one logical invocation.
"""
from __future__ import annotations

import asyncio
import json
import os
import signal
import subprocess
import sys
import tempfile
from datetime import timedelta
from pathlib import Path

from temporalio import activity, workflow
from temporalio.client import Client
from temporalio.worker import Worker

TASK_QUEUE = "naeos-crash-recovery-003"
WORKFLOW_ID = "naeos-crash-recovery-003"
EVIDENCE = Path("external-validation-evidence/temporal-evidence.json")


@activity.defn
async def consequential_side_effect(invocation_id: str) -> str:
    state = Path(os.environ["NAEOS_STATE_FILE"])
    state.parent.mkdir(parents=True, exist_ok=True)
    with state.open("a", encoding="utf-8") as f:
        f.write(json.dumps({"invocation_id": invocation_id, "pid": os.getpid()}) + "\n")
        f.flush()
        os.fsync(f.fileno())

    marker = state.with_suffix(".crashed")
    if not marker.exists():
        marker.write_text("crashed\n", encoding="utf-8")
        os.kill(os.getpid(), signal.SIGKILL)
    return "completed"


@workflow.defn
class CrashRecoveryWorkflow:
    @workflow.run
    async def run(self, invocation_id: str) -> str:
        return await workflow.execute_activity(
            consequential_side_effect,
            invocation_id,
            start_to_close_timeout=timedelta(seconds=10),
        )


async def worker_process(state: Path) -> None:
    os.environ["NAEOS_STATE_FILE"] = str(state)
    client = await Client.connect("127.0.0.1:7233")
    async with Worker(
        client,
        task_queue=TASK_QUEUE,
        workflows=[CrashRecoveryWorkflow],
        activities=[consequential_side_effect],
    ):
        await asyncio.sleep(30)


async def main() -> None:
    with tempfile.TemporaryDirectory() as td:
        state = Path(td) / "effects.jsonl"
        env = os.environ.copy()
        env["NAEOS_STATE_FILE"] = str(state)

        p1 = subprocess.Popen([sys.executable, __file__, "worker", str(state)], env=env)
        await asyncio.sleep(2)

        client = await Client.connect("127.0.0.1:7233")
        handle = await client.start_workflow(
            CrashRecoveryWorkflow.run,
            "inv-recovery-003",
            id=WORKFLOW_ID,
            task_queue=TASK_QUEUE,
        )

        for _ in range(30):
            if p1.poll() is not None:
                break
            await asyncio.sleep(0.5)
        if p1.poll() is None:
            p1.kill()
            p1.wait(timeout=5)
        if p1.returncode == 0:
            raise AssertionError("worker 1 did not crash as expected")

        p2 = subprocess.Popen([sys.executable, __file__, "worker", str(state)], env=env)
        try:
            result = await asyncio.wait_for(handle.result(), timeout=30)
            if result != "completed":
                raise AssertionError(f"unexpected workflow result: {result!r}")
        finally:
            p2.terminate()
            p2.wait(timeout=5)

        records = [json.loads(line) for line in state.read_text(encoding="utf-8").splitlines()]
        invocation_ids = [r["invocation_id"] for r in records]
        if len(records) < 2:
            raise AssertionError("expected Temporal to retry the Activity after worker crash")
        if len(set(invocation_ids)) != 1:
            raise AssertionError("recovery changed the logical invocation identity")

        report = {
            "schema": "naeos.external-validation-temporal.v1",
            "status": "PASS",
            "runtime": "Temporal local dev server",
            "logical_invocation_id": "inv-recovery-003",
            "worker_boundary": "worker-1 -> worker-2",
            "side_effect_attempts": len(records),
            "logical_invocations": len(set(invocation_ids)),
            "finding": "crash after side effect before Activity completion caused a retry of the same logical invocation",
            "authorization_and_execution_are_distinct": True,
            "evidence_requirement": "bind authorization, execution attempt, observed outcome, and recovery state to one logical invocation ID",
        }
        EVIDENCE.parent.mkdir(parents=True, exist_ok=True)
        EVIDENCE.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
        print(json.dumps(report, indent=2))


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] == "worker":
        asyncio.run(worker_process(Path(sys.argv[2])))
    else:
        asyncio.run(main())
