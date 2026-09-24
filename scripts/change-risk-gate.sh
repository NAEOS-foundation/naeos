#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_SHA="${BASE_SHA:-${GITHUB_BASE_SHA:-${GITHUB_EVENT_BEFORE:-}}}"
HEAD_SHA="${HEAD_SHA:-HEAD}"
OUTPUT="${NAEOS_CHANGE_RISK_OUTPUT:-${ROOT}/change-risk-evidence.json}"

cd "${ROOT}"

if [[ -z "${BASE_SHA}" ]]; then
  BASE_SHA="$(git rev-parse HEAD^)"
fi

export BASE_SHA
export HEAD_SHA
export NAEOS_CHANGE_RISK_OUTPUT="${OUTPUT}"

go run ./cmd/naeos-change-risk
