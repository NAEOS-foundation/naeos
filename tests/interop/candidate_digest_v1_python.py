#!/usr/bin/env python3
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
"""Independent Python RFC 8785/SHA-256 conformance check for NAEOS digest v1.

Uses the third-party rfc8785 package rather than the Go implementation under test.
"""
from __future__ import annotations

import hashlib
import json
import math
import pathlib
import sys

import rfc8785

PROFILE = "naeos-candidate-jcs-sha256-v1"
PREFIX = b"NAEOS-CANDIDATE-DIGEST-V1\n"
VECTORS = pathlib.Path(__file__).with_name("handoff-candidate-digest-v1-vectors.json")


def reject_duplicate_keys(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate JSON object key: {key!r}")
        result[key] = value
    return result


def reject_non_json_constant(value):
    raise ValueError(f"non-JSON numeric constant: {value}")


def reject_non_finite(value):
    if isinstance(value, float) and not math.isfinite(value):
        raise ValueError("non-finite number is not permitted")
    if isinstance(value, dict):
        for item in value.values():
            reject_non_finite(item)
    elif isinstance(value, list):
        for item in value:
            reject_non_finite(item)


def digest_candidate(raw: str) -> tuple[bytes, str]:
    candidate = json.loads(
        raw,
        object_pairs_hook=reject_duplicate_keys,
        parse_constant=reject_non_json_constant,
    )
    if not isinstance(candidate, dict) or set(candidate) != {"tool", "input"}:
        raise ValueError("candidate must contain exactly tool and input")
    if not isinstance(candidate["tool"], str) or not candidate["tool"]:
        raise ValueError("tool must be a non-empty string")
    reject_non_finite(candidate["input"])
    canonical = rfc8785.dumps(candidate)
    digest = hashlib.sha256(PREFIX + canonical).hexdigest()
    return canonical, digest


def main() -> int:
    fixture = json.loads(VECTORS.read_text(encoding="utf-8"))
    if fixture.get("profile") != PROFILE or fixture.get("domain_prefix_utf8") != PREFIX.decode():
        raise SystemExit("profile or domain prefix mismatch")
    for vector in fixture["vectors"]:
        raw = json.dumps(vector["candidate"], ensure_ascii=False, separators=(",", ":"))
        canonical, digest = digest_candidate(raw)
        expected_bytes = vector["canonical_json_utf8"].encode("utf-8")
        if canonical != expected_bytes:
            raise AssertionError(
                f"{vector['id']}: canonical mismatch: {canonical!r} != {expected_bytes!r}"
            )
        if digest != vector["expected_digest_hex"]:
            raise AssertionError(
                f"{vector['id']}: digest mismatch: {digest} != {vector['expected_digest_hex']}"
            )
        print(f"PASS {vector['id']}")
    print(f"PASS all {len(fixture['vectors'])} golden vectors via independent Python RFC 8785 implementation")
    return 0


if __name__ == "__main__":
    sys.exit(main())
