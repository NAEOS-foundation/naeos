// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import { getCloudflareContext } from "@opennextjs/cloudflare";

const MAX_REQUEST_BYTES = 8 * 1024;

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

async function readBoundedBody(request: Request): Promise<
  | { ok: true; body: string }
  | { ok: false; status: 400 | 413; message: string }
> {
  const declaredLength = request.headers.get("content-length");
  if (declaredLength !== null) {
    const length = Number(declaredLength);
    if (Number.isFinite(length) && length > MAX_REQUEST_BYTES) {
      return { ok: false, status: 413, message: "Request payload is too large." };
    }
  }

  if (!request.body) {
    return { ok: false, status: 400, message: "A valid JSON request body is required." };
  }

  const reader = request.body.getReader();
  const chunks: Uint8Array[] = [];
  let totalBytes = 0;

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      totalBytes += value.byteLength;
      if (totalBytes > MAX_REQUEST_BYTES) {
        await reader.cancel();
        return { ok: false, status: 413, message: "Request payload is too large." };
      }
      chunks.push(value);
    }
  } catch {
    return { ok: false, status: 400, message: "Unable to read request body." };
  } finally {
    reader.releaseLock();
  }

  const bodyBytes = new Uint8Array(totalBytes);
  let offset = 0;
  for (const chunk of chunks) {
    bodyBytes.set(chunk, offset);
    offset += chunk.byteLength;
  }

  return { ok: true, body: new TextDecoder().decode(bodyBytes) };
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

export async function POST(request: Request) {
  const boundedBody = await readBoundedBody(request);
  if (!boundedBody.ok) {
    return errorResponse(boundedBody.message, boundedBody.status);
  }

  let body: unknown;
  try {
    body = JSON.parse(boundedBody.body);
  } catch {
    return errorResponse("A valid JSON request body is required.", 400);
  }

  if (!body || typeof body !== "object" || Array.isArray(body)) {
    return errorResponse("A JSON object is required.", 400);
  }

  let token: string | undefined;
  let configuredUrl: string | undefined;
  try {
    const context = await getCloudflareContext({ async: true });
    const env = context.env as ControlPlaneEnv;
    token = env.NAEOS_CONTROLPLANE_API_TOKEN;
    configuredUrl = env.NAEOS_CONTROLPLANE_API_URL;
  } catch {
    // Missing runtime context must fail closed; never fall back to a hard-coded
    // production upstream or a browser-provided credential.
  }

  if (!token?.trim()) {
    return errorResponse("Control Plane server authentication is not configured.", 503);
  }

  const upstreamUrl = getUpstreamUrl(configuredUrl);
  if (!upstreamUrl) {
    return errorResponse("Control Plane upstream is not configured correctly.", 503);
  }

  try {
    const upstream = await fetch(
      new URL("/api/control-plane/decision", upstreamUrl).toString(),
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token.trim()}`,
        },
        body: boundedBody.body,
        signal: AbortSignal.timeout(10_000),
        cache: "no-store",
      },
    );
    const payload = await upstream.text();
    return new Response(payload, {
      status: upstream.status,
      headers: {
        "Content-Type": upstream.headers.get("content-type") ?? "application/json",
        "Cache-Control": "no-store",
        "X-Content-Type-Options": "nosniff",
      },
    });
  } catch {
    return errorResponse("Control Plane upstream is unavailable.", 502);
  }
}
