// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import { getCloudflareContext } from "@opennextjs/cloudflare";

export const dynamic = "force-dynamic";
export const revalidate = 0;

type ControlPlaneEnv = CloudflareEnv & {
  NAEOS_CONTROLPLANE_API_TOKEN?: string;
  NAEOS_CONTROLPLANE_API_URL?: string;
};

function errorResponse(message: string, status: number) {
  return Response.json(
    { message },
    {
      status,
      headers: {
        "Cache-Control": "no-store",
        "X-Content-Type-Options": "nosniff",
      },
    },
  );
}

function getUpstreamUrl(value: string | undefined): URL | undefined {
  if (!value?.trim()) return undefined;

  try {
    const url = new URL(value.trim());
    if (
      url.protocol !== "https:" ||
      !url.hostname.endsWith(".up.railway.app") ||
      url.username ||
      url.password ||
      url.pathname !== "/" ||
      url.search ||
      url.hash
    ) {
      return undefined;
    }
    return url;
  } catch {
    return undefined;
  }
}

export async function GET(request: Request) {
  let token: string | undefined;
  let configuredUrl: string | undefined;
  try {
    const context = await getCloudflareContext({ async: true });
    const env = context.env as ControlPlaneEnv;
    token = env.NAEOS_CONTROLPLANE_API_TOKEN;
    configuredUrl = env.NAEOS_CONTROLPLANE_API_URL;
  } catch {
    // Missing runtime context must fail closed; no hard-coded upstream fallback.
  }

  if (!token?.trim()) {
    return errorResponse("Control Plane server authentication is not configured.", 503);
  }

  const upstreamUrl = getUpstreamUrl(configuredUrl);
  if (!upstreamUrl) {
    return errorResponse("Control Plane upstream is not configured correctly.", 503);
  }

  const target = new URL("/api/control-plane/evidence", upstreamUrl);
  target.search = new URL(request.url).search;

  let upstream: Response;
  try {
    upstream = await fetch(target.toString(), {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token.trim()}`,
      },
      signal: AbortSignal.timeout(10_000),
      cache: "no-store",
    });
  } catch {
    return errorResponse("Control Plane upstream is unavailable.", 502);
  }

  const payload = await upstream.text();
  return new Response(payload, {
    status: upstream.status,
    headers: {
      "Content-Type": upstream.headers.get("content-type") ?? "application/json",
      "Cache-Control": "no-store",
      "X-Content-Type-Options": "nosniff",
    },
  });
}
