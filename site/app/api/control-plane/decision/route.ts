// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import { getCloudflareContext } from "@opennextjs/cloudflare";

const CONTROL_PLANE_API_URL =
  "https://naeos-control-plane-production.up.railway.app";
const MAX_REQUEST_BYTES = 8 * 1024;

export const dynamic = "force-dynamic";
export const revalidate = 0;

type ControlPlaneEnv = CloudflareEnv & {
  NAEOS_CONTROLPLANE_API_TOKEN?: string;
};

export async function POST(request: Request) {
  const contentLength = Number(request.headers.get("content-length") ?? "0");
  if (Number.isFinite(contentLength) && contentLength > MAX_REQUEST_BYTES) {
    return Response.json(
      { message: "Request payload is too large." },
      { status: 413, headers: { "Cache-Control": "no-store" } },
    );
  }

  let body: unknown;
  let rawBody: string;
  try {
    rawBody = await request.text();
    if (new TextEncoder().encode(rawBody).byteLength > MAX_REQUEST_BYTES) {
      return Response.json(
        { message: "Request payload is too large." },
        { status: 413, headers: { "Cache-Control": "no-store" } },
      );
    }
    body = JSON.parse(rawBody);
  } catch {
    return Response.json(
      { message: "A valid JSON request body is required." },
      { status: 400, headers: { "Cache-Control": "no-store" } },
    );
  }

  if (!body || typeof body !== "object" || Array.isArray(body)) {
    return Response.json(
      { message: "A JSON object is required." },
      { status: 400, headers: { "Cache-Control": "no-store" } },
    );
  }

  let token: string | undefined;
  try {
    const context = await getCloudflareContext({ async: true });
    const env = context.env as ControlPlaneEnv;
    token = env.NAEOS_CONTROLPLANE_API_TOKEN;
  } catch {
    token = undefined;
  }

  if (!token?.trim()) {
    return Response.json(
      { message: "Control Plane server authentication is not configured." },
      { status: 503, headers: { "Cache-Control": "no-store" } },
    );
  }

  try {
    const upstream = await fetch(`${CONTROL_PLANE_API_URL}/api/control-plane/decision`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token.trim()}`,
      },
      body: rawBody,
      signal: AbortSignal.timeout(10_000),
      cache: "no-store",
    });
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
    return Response.json(
      { message: "Control Plane upstream is unavailable." },
      { status: 502, headers: { "Cache-Control": "no-store" } },
    );
  }
}
