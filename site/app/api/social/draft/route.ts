// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { NextResponse } from "next/server";

const API = "https://api.github.com";
const REPO = "NAEOS-foundation/naeos";
const headers = {
  Accept: "application/vnd.github+json",
  "User-Agent": "NAEOS-Social-Signal/1.0",
};

type GitHubIssue = {
  number: number;
  title: string;
  html_url: string;
  labels?: Array<{ name?: string }>;
  pull_request?: { url?: string };
};

type GitHubPull = GitHubIssue & {
  merged_at?: string | null;
  state: string;
};

type GitHubRun = {
  name?: string;
  status?: string;
  conclusion?: string | null;
  html_url?: string;
};

type GitHubCommit = {
  sha: string;
  html_url: string;
  commit?: { message?: string };
};

type GitHubRelease = {
  tag_name?: string;
  html_url?: string;
  name?: string;
};

async function github<T>(path: string): Promise<T> {
  const response = await fetch(`${API}${path}`, {
    headers,
    cache: "no-store",
    signal: AbortSignal.timeout(5000),
  });
  if (!response.ok) throw new Error(`GitHub returned ${response.status} for ${path}`);
  return response.json() as Promise<T>;
}

function isActualIssue(issue: GitHubIssue) {
  return !issue.pull_request;
}

function isBug(issue: GitHubIssue) {
  return issue.labels?.some((label) => label.name?.toLowerCase() === "bug") ?? false;
}

function classify(
  openIssues: GitHubIssue[],
  openPulls: GitHubPull[],
  mergedPulls: GitHubPull[],
  runs: GitHubRun[],
  release?: GitHubRelease,
) {
  const signals: string[] = [];

  if (release?.tag_name) signals.push("release");
  if (mergedPulls.length > 0) signals.push("merged-pr");
  if (openPulls.length > 0 && openPulls.length <= 5) signals.push("active-pr");
  if (openIssues.some(isBug)) signals.push("bug");
  if (runs[0]?.conclusion === "failure") signals.push("ci-failure");

  return signals;
}

function buildPost(
  openIssues: GitHubIssue[],
  openPulls: GitHubPull[],
  mergedPulls: GitHubPull[],
  runs: GitHubRun[],
  release?: GitHubRelease,
) {
  const bullets: string[] = [];

  if (release?.tag_name) bullets.push(`Release ${release.tag_name} is available.`);
  if (mergedPulls[0]) bullets.push(`PR #${mergedPulls[0].number} merged: ${mergedPulls[0].title}.`);
  if (openPulls[0]) bullets.push(`Active engineering work: PR #${openPulls[0].number} — ${openPulls[0].title}.`);
  if (openIssues.some(isBug)) bullets.push(`Bug-tracked work remains visible in the repository (${openIssues.filter(isBug).length} open bug issue(s)).`);
  if (runs[0]?.conclusion === "failure") bullets.push(`Latest CI signal is ${runs[0].conclusion}; investigation remains part of the engineering state.`);

  return [
    "NAEOS Engineering Update",
    "",
    "NAEOS is building its engineering system in public.",
    "",
    ...bullets.slice(0, 4).map((item) => `• ${item}`),
    "",
    "The public status page tracks the repository's engineering reality rather than a manually maintained snapshot.",
    "",
    "#NAEOS #AIEngineering #SoftwareEngineering",
  ].join("\n");
}

export const dynamic = "force-dynamic";
export const revalidate = 0;

export async function GET() {
  try {
    const [rawIssues, pulls, closedPulls, runs, commits, release] = await Promise.all([
      github<GitHubIssue[]>(`/repos/${REPO}/issues?state=open&per_page=20`),
      github<GitHubPull[]>(`/repos/${REPO}/pulls?state=open&sort=updated&direction=desc&per_page=10`),
      github<GitHubPull[]>(`/repos/${REPO}/pulls?state=closed&sort=updated&direction=desc&per_page=10`),
      github<GitHubRun[]>(`/repos/${REPO}/actions/runs?per_page=5`),
      github<GitHubCommit[]>(`/repos/${REPO}/commits?per_page=5`),
      github<GitHubRelease>(`/repos/${REPO}/releases/latest`).catch(() => undefined),
    ]);

    const issues = rawIssues.filter(isActualIssue);
    const mergedPulls = closedPulls.filter((pull) => pull.merged_at).slice(0, 5);
    const signals = classify(issues, pulls, mergedPulls, runs, release);

    // Only emit a social candidate for meaningful engineering signals.
    const publishable = signals.some((signal) =>
      ["release", "merged-pr", "bug", "ci-failure"].includes(signal),
    );

    const body = buildPost(issues, pulls, mergedPulls, runs, release);

    return NextResponse.json(
      {
        source: "github-live",
        generatedAt: new Date().toISOString(),
        publishable,
        signals,
        sourceEvents: {
          latestCommit: commits[0]
            ? { sha: commits[0].sha, url: commits[0].html_url, message: commits[0].commit?.message }
            : undefined,
          latestRelease: release
            ? { tag: release.tag_name, name: release.name, url: release.html_url }
            : undefined,
          latestWorkflow: runs[0],
        },
        counts: {
          openIssues: issues.length,
          openBugs: issues.filter(isBug).length,
          openPullRequests: pulls.length,
          recentlyMergedPullRequests: mergedPulls.length,
        },
        drafts: {
          linkedin: body,
          x: body.length > 280 ? `${body.slice(0, 260).trim()}…\n\n#NAEOS #AIEngineering` : body,
        },
      },
      { headers: { "Cache-Control": "no-store, max-age=0" } },
    );
  } catch {
    return NextResponse.json(
      {
        source: "unavailable",
        generatedAt: new Date().toISOString(),
        publishable: false,
        signals: [],
        drafts: {},
      },
      { status: 503, headers: { "Cache-Control": "no-store, max-age=0" } },
    );
  }
}
