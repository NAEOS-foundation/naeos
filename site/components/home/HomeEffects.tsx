"use client";

import { useEffect, useRef, useState } from "react";

function animateCounter(el: HTMLElement, target: number) {
  const duration = 1500;
  let startTime: number | null = null;
  function step(timestamp: number) {
    if (!startTime) startTime = timestamp;
    const progress = Math.min((timestamp - startTime) / duration, 1);
    const eased = 1 - Math.pow(1 - progress, 3);
    el.textContent = String(Math.floor(eased * target));
    if (progress < 1) requestAnimationFrame(step);
    else el.textContent = String(target);
  }
  requestAnimationFrame(step);
}

export function CountUpNumber({ target }: { target: number }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            animateCounter(el, target);
            observer.unobserve(el);
          }
        }
      },
      { threshold: 0.5 },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [target]);

  return (
    <>
      <div className="stat-number" ref={ref} aria-hidden="true">
        0
      </div>
      <span
        style={{
          position: "absolute",
          width: "1px",
          height: "1px",
          padding: "0",
          margin: "-1px",
          overflow: "hidden",
          clip: "rect(0 0 0 0)",
          whiteSpace: "nowrap",
          border: "0",
        }}
      >
        {target}
      </span>
    </>
  );
}

interface GhStats {
  stars?: number;
  forks?: number;
  issues?: number;
  contributors?: number;
}

const CACHE_KEY = "gh-stats-cache";
const CACHE_TTL = 60 * 60 * 1000;

function getCachedStats(): GhStats | null {
  try {
    const raw = localStorage.getItem(CACHE_KEY);
    if (!raw) return null;
    const { data, ts } = JSON.parse(raw) as { data: GhStats; ts: number };
    if (Date.now() - ts > CACHE_TTL) return null;
    return data;
  } catch {
    return null;
  }
}

function setCachedStats(data: GhStats) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({ data, ts: Date.now() }));
  } catch {
    /* ignore */
  }
}

export function GithubStats({
  labels,
  initial,
}: {
  labels: string[];
  initial?: GhStats;
}) {
  const refs = useRef<(HTMLDivElement | null)[]>([]);
  const [stats, setStats] = useState<GhStats>(() => initial ?? getCachedStats() ?? {});

  useEffect(() => {
    if (initial) return;
    if (getCachedStats()) return;
    let cancelled = false;

    Promise.all([
      fetch("https://api.github.com/repos/NAEOS-foundation/naeos")
        .then((r) => r.json())
        .then((data: Record<string, number>) => ({
          stars: data.stargazers_count,
          forks: data.forks_count,
          issues: data.open_issues_count,
        })),
      fetch("https://api.github.com/repos/NAEOS-foundation/naeos/contributors?per_page=1&anon=true")
        .then((r) => {
          const link = r.headers.get("Link");
          if (link) {
            const m = link.match(/page=(\d+)>; rel="last"/);
            if (m) return Number(m[1]);
          }
          return undefined;
        }),
    ])
      .then(([repoData, contributors]) => {
        if (cancelled) return;
        const merged: GhStats = { ...repoData, contributors };
        setStats(merged);
        setCachedStats(merged);
      })
      .catch(() => {});

    return () => {
      cancelled = true;
    };
  }, [initial]);

  useEffect(() => {
    const values = [stats.stars, stats.forks, stats.issues, stats.contributors];
    values.forEach((v, i) => {
      const el = refs.current[i];
      if (el && v !== undefined) animateCounter(el, v);
    });
  }, [stats]);

  const valueKey = ["stars", "forks", "issues", "contributors"] as const;

  return (
    <>
      {labels.map((label, i) => (
        <div key={label} className="github-stat">
          <div
            className="github-stat-number"
            ref={(el) => {
              refs.current[i] = el;
            }}
            aria-hidden="true"
          >
            {stats[valueKey[i]] ?? 0}
          </div>
          <div className="github-stat-label">{label}</div>
        </div>
      ))}
    </>
  );
}
