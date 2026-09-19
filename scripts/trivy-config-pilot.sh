#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET="${TRIVY_TARGET:-${ROOT}/examples/plugins/trivy-config/testdata/Dockerfile}"
OUTPUT="${TRIVY_OUTPUT:-${ROOT}/tmp/trivy-config-report.json}"

if ! command -v trivy >/dev/null 2>&1; then
	printf 'Trivy is required on PATH.\n' >&2
	exit 1
fi

mkdir -p "$(dirname "${OUTPUT}")"
trivy config --format json --quiet --output "${OUTPUT}" "${TARGET}"
printf 'Trivy report written to %s\n' "${OUTPUT}"