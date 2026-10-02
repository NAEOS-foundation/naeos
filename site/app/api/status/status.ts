// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

export interface ServiceStatus {
  id: string;
  name: string;
  url: string;
  target: string;
  kind: "http" | "ws";
  status: "operational" | "degraded" | "down";
  latencyMs: number;
}

export interface StatusPayload {
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

type Fetcher = typeof fetch;

const REQUEST_TIMEOUT_MS = 5000;
const REPO = "NAEOS-foundation/naeos";
const GITHUB_HEADERS = {
  accept: "application/vnd.github+json",
  "user-agent": "NAEOS-status",
};

async function fetchWithTimeout(
  fetcher: Fetcher,
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
  try {
    return await fetcher(input, { ...init, signal: controller.signal });
  } finally {
    clearTimeout(timer);
  }
}

export async function probeHttp(
  url: string,
  fetcher: Fetcher = fetch,
): Promise<{ ok: boolean; latencyMs: number }> {
  const started = Date.now();
  try {
    const response = await fetchWithTimeout(fetcher, url, { method: "GET" });
    return { ok: response.status < 500, latencyMs: Date.now() - started };
  } catch {
    return { ok: false, latencyMs: Date.now() - started };
  }
}

async function fetchJson<T>(
  fetcher: Fetcher,
  url: string,
): Promise<T> {
  const response = await fetchWithTimeout(fetcher, url, {
    headers: GITHUB_HEADERS,
  });
  if (!response.ok) {
    throw new Error(`GitHub request failed: ${response.status}`);
  }
  return response.json() as Promise<T>;
}

function parseContributorCount(link: string | null): number | undefined {
  if (!link) return undefined;
  const match = link.match(/[?&]page=(\d+)>; rel="last"/);
  return match ? Number(match[1]) : undefined;
}

function service(
  id: string,
  name: string,
  url: string,
  kind: "http" | "ws",
  probe: { ok: boolean; latencyMs: number },
): ServiceStatus {
  return {
    id,
    name,
    url,
    target: url,
    kind,
    status: probe.ok ? "operational" : "down",
    latencyMs: probe.latencyMs,
  };
}

export async function getLiveStatus(
  fetcher: Fetcher = fetch,
): Promise<StatusPayload> {
  const [web, docs, registry, realtime, github, release] = await Promise.all([
    probeHttp("https://naeos.dev", fetcher),
    probeHttp("https://docs.naeos.dev", fetcher),
    probeHttp("https://registry.naeos.dev", fetcher),
    probeHttp("https://ws.naeos.dev/ws", fetcher),
    fetchJson<{ stargazers_count: number; forks_count: number; open_issues_count: number }>(
      fetcher,
      `https://api.github.com/repos/${REPO}`,
    ),
    fetchJson<{ tag_name?: string }>(
      fetcher,
      `https://api.github.com/repos/${REPO}/releases/latest`,
    ),
  ]);

  return {
    updatedAt: new Date().toISOString(),
    source: "live",
    github: {
      stars: github.stargazers_count,
      forks: github.forks_count,
      openIssues: github.open_issues_count,
      version: (release.tag_name ?? "unknown").replace(/^v/, ""),
    },
    services: [
      service("web", "Website", "https://naeos.dev", "http", web),
      service("docs", "Docs", "https://docs.naeos.dev", "http", docs),
      service("registry", "Registry", "https://registry.naeos.dev", "http", registry),
      service("realtime", "Realtime", "https://ws.naeos.dev/ws", "ws", realtime),
    ],
  };
}

export { parseContributorCount };
