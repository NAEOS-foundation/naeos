#!/bin/sh
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -eu
: "${NAEOS_API_JWT_SECRET:?NAEOS_API_JWT_SECRET is required}"
if [ -n "${NAEOS_SANDBASE_AUTHZ_CONFIG_JSON:-}" ]; then
  : "${NAEOS_SANDBASE_AUTHZ_TOKEN:?NAEOS_SANDBASE_AUTHZ_TOKEN is required when SandBase config is set}"
  umask 077
  printf '%s' "$NAEOS_SANDBASE_AUTHZ_CONFIG_JSON" > /tmp/sandbase-authz.json
  export NAEOS_SANDBASE_AUTHZ_CONFIG=/tmp/sandbase-authz.json
elif [ -n "${NAEOS_SANDBASE_AUTHZ_TOKEN:-}" ]; then
  echo "NAEOS_SANDBASE_AUTHZ_CONFIG_JSON is required when NAEOS_SANDBASE_AUTHZ_TOKEN is set" >&2
  exit 1
fi
exec /usr/local/bin/naeos serve --port 8080 --auth --jwt-secret "$NAEOS_API_JWT_SECRET"
