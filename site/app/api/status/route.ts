// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { NextResponse } from "next/server";
import fallback from "@/data/status.json";

interface ServiceStatus {
  id: string;
  name: string;
  url: string;
  target: string;
  kind: "http" | "ws";
  status: "operational" | "degraded" | "down";
  latencyMs: number;
}

interface StatusPayload {
  updatedAt: string;
  source: "static" | "live";
  github: {
    stars: number;
    forks: number;
    openIssues: number;
    version: string;
    contributors?: number;
  };
  services: ServiceStatus[];
}

const FALLBACK = fallback as StatusPayload;

async function probe(url: string, kind: "http" | "ws"): Promise<ServiceStatus> {
  const started = Date.now();
  try {
    const response = await fetch(url, {
      method: "GET",
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    const latencyMs = Date.now() - started;
    const reachable = response.status < 500;
    return {
      id: kind === "ws" ? "realtime" : url.includes("docs.") ? "docs" : url.includes("registry.") ? "registry" : "web",
      name: kind === "ws" ? "Realtime" : url.includes("docs.") ? "Docs" : url.includes("registry.") ? "Registry" : "Website",
      url,
      target: kind === "ws" ? "wss://ws.naeos.dev/ws" : url,
      kind,
      status: reachable ? "operational" : "down",
      latencyMs,
    };
  } catch {
    return {
      id: kind === "ws" ? "realtime" : url.includes("docs.") ? "docs" : url.includes("registry.") ? "registry" : "web",
      name: kind === "ws" ? "Realtime" : url.includes("docs.") ? "Docs" : url.includes("registry.") ? "Registry" : "Website",
      url,
      target: kind === "ws" ? "wss://ws.naeos.dev/ws" : url,
      kind,
      status: "down",
      latencyMs: Date.now() - started,
    };
  }
}

export const dynamic = "force-dynamic";
export const revalidate = 0;

export async function GET() {
  try {
    const githubResponse = await fetch("https://api.github.com/repos/NAEOS-foundation/naeos", {
      headers: { Accept: "application/vnd.github+json", "User-Agent": "NAEOS-Status/1.0" },
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    if (!githubResponse.ok) throw new Error(`GitHub repository probe returned ${githubResponse.status}`);

    const github = (await githubResponse.json()) as {
      stargazers_count: number;
      forks_count: number;
      open_issues_count: number;
    };

    const releaseResponse = await fetch(
      "https://api.github.com/repos/NAEOS-foundation/naeos/releases/latest",
      {
        headers: { Accept: "application/vnd.github+json", "User-Agent": "NAEOS-Status/1.0" },
        cache: "no-store",
        signal: AbortSignal.timeout(5000),
      },
    );
    const release = releaseResponse.ok
      ? ((await releaseResponse.json()) as { tag_name?: string })
      : undefined;

    const services = await Promise.all([
      probe("https://naeos.dev", "http"),
      probe("https://docs.naeos.dev", "http"),
      probe("https://registry.naeos.dev", "http"),
      probe("https://ws.naeos.dev/ws", "ws"),
    ]);

    const payload: StatusPayload = {
      updatedAt: new Date().toISOString(),
      source: "live",
      github: {
        stars: github.stargazers_count,
        forks: github.forks_count,
        openIssues: github.open_issues_count,
        version: (release?.tag_name ?? FALLBACK.github.version).replace(/^v/, ""),
        contributors: FALLBACK.github.contributors,
      },
      services,
    };

    return NextResponse.json(payload, {
      headers: { "Cache-Control": "no-store, max-age=0" },
    });
  } catch {
    return NextResponse.json(
      { ...FALLBACK, source: "static" },
      { status: 200, headers: { "Cache-Control": "no-store, max-age=0" } },
    );
  }
}
