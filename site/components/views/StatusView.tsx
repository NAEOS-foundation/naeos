// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

"use client";

import { useEffect, useState } from "react";
import type { Lang } from "@/lib/site";
import statusData from "@/data/status.json";

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
  github: { stars: number; forks: number; openIssues: number; version: string; contributors?: number };
  services: ServiceStatus[];
}

const fallback = statusData as StatusPayload;

const LABELS: Record<Lang, { updated: string; github: string; stars: string; forks: string; issues: string; version: string; operational: string; degraded: string; down: string; latency: string; live: string; static: string; liveNote: string }> = {
  en: {
    updated: "Last checked", github: "GitHub", stars: "Stars", forks: "Forks", issues: "Open issues",
    version: "Latest release", operational: "Operational", degraded: "Degraded", down: "Down", latency: "latency",
    live: "Live", static: "Static",
    liveNote: "Live telemetry is fetched at page load. If the telemetry endpoint is unavailable or invalid, this page uses the versioned static baseline.",
  },
  id: {
    updated: "Terakhir diperiksa", github: "GitHub", stars: "Bintang", forks: "Fork", issues: "Issue terbuka",
    version: "Rilis terbaru", operational: "Beroperasi", degraded: "Menurun", down: "Gangguan", latency: "latensi",
    live: "Live", static: "Statis",
    liveNote: "Telemetri live diambil saat halaman dimuat. Jika endpoint telemetri tidak tersedia atau tidak valid, halaman menggunakan baseline statis berversi.",
  },
};

export default function StatusView({ lang }: { lang: Lang }) {
  const t = LABELS[lang];
  const [data, setData] = useState<StatusPayload>(fallback);

  useEffect(() => {
    let active = true;
    fetch("/api/status", { cache: "no-store" })
      .then((response) => {
        if (!response.ok) throw new Error("status endpoint unavailable");
        return response.json() as Promise<StatusPayload>;
      })
      .then((payload) => {
        if (
          active &&
          payload?.source === "live" &&
          typeof payload.updatedAt === "string" &&
          Array.isArray(payload.services) &&
          payload.github
        ) {
          setData(payload);
        }
      })
      .catch(() => {
        // Keep the deterministic versioned fallback.
      });
    return () => {
      active = false;
    };
  }, []);

  const updated = data.updatedAt ? new Date(data.updatedAt) : null;
  const isLive = data.source === "live";

  return (
    <section className="content-section">
      <div className="container container-narrow">
        <div className="single-content">
          <div className="status-summary">
            <span className={`badge ${isLive ? "badge-success" : "badge-orange"}`}>
              {isLive ? `● ${t.live}` : `● ${t.static}`}
            </span>
            <span className="status-summary-updated">
              {t.updated}:{" "}
              {updated ? updated.toLocaleString(lang === "id" ? "id-ID" : "en-US", { dateStyle: "medium", timeStyle: "short" }) : "—"}
            </span>
          </div>

          <p className="status-note">{t.liveNote}</p>

          <h2>{t.github}</h2>
          <div className="status-stats-grid">
            <div className="status-stat"><span className="status-stat-value">{data.github.stars?.toLocaleString() ?? "—"}</span><span className="status-stat-label">{t.stars}</span></div>
            <div className="status-stat"><span className="status-stat-value">{data.github.forks?.toLocaleString() ?? "—"}</span><span className="status-stat-label">{t.forks}</span></div>
            <div className="status-stat"><span className="status-stat-value">{data.github.openIssues?.toLocaleString() ?? "—"}</span><span className="status-stat-label">{t.issues}</span></div>
            <div className="status-stat"><span className="status-stat-value">v{data.github.version ?? "—"}</span><span className="status-stat-label">{t.version}</span></div>
          </div>

          <h2>{lang === "id" ? "Status Layanan" : "Service Status"}</h2>
          <div className="status-grid">
            {data.services.map((s) => {
              const label = s.status === "down" ? t.down : s.status === "degraded" ? t.degraded : t.operational;
              const badgeClass = s.status === "down" ? "badge-red" : s.status === "degraded" ? "badge-orange" : "badge-success";
              return (
                <div key={s.id} className="status-card">
                  <span className={`status-dot ${s.status}`} />
                  <div>
                    <h4>{s.name}</h4>
                    <p>{s.kind === "ws" ? "WebSocket" : "HTTP"}{s.latencyMs > 0 ? ` · ${s.latencyMs}ms ${t.latency}` : ""}</p>
                  </div>
                  <span className={`badge ${badgeClass}`}>{label}</span>
                </div>
              );
            })}
          </div>

          <h2>{lang === "id" ? "Riwayat Insiden" : "Incident History"}</h2>
          <p>{lang === "id" ? "Belum ada insiden besar yang dilaporkan. Halaman ini akan diperbarui jika terjadi gangguan layanan." : "No major incidents have been reported. This page will be updated if any service disruptions occur."}</p>

          <h2>{lang === "id" ? "Jaminan Uptime" : "Uptime Guarantee"}</h2>
          <p>{lang === "id" ? "NAEOS adalah tool CLI open-source yang berjalan di lokal. Website dan infrastruktur dikelola dengan upaya terbaik." : "NAEOS is an open-source CLI tool that runs locally. The website and infrastructure are maintained on a best-effort basis."}</p>

          <hr />

          <p>
            {lang === "id" ? (
              <><strong>Mengalami masalah?</strong>{" "}<a href="https://github.com/NAEOS-foundation/naeos/issues/new">Buka issue GitHub</a>{" "}atau hubungi kami di <code>support@naeos.dev</code>.</>
            ) : (
              <><strong>Experiencing an issue?</strong>{" "}<a href="https://github.com/NAEOS-foundation/naeos/issues/new">Open a GitHub issue</a>{" "}or contact us at <code>support@naeos.dev</code>.</>
            )}
          </p>
        </div>
      </div>
    </section>
  );
}
