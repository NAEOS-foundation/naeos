"use client";

import { useMemo, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

interface Template {
  name: string;
  version: string;
  description: string;
  author: string;
  tags: string[];
  repo_url: string;
  download_url: string;
  languages: string[];
  frameworks?: string[];
  downloads: number;
  updated_at: string;
}

function formatMonthYear(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return `${d.toLocaleString("en-US", { month: "short" })} ${d.getFullYear()}`;
}

export default function TemplatesGrid({
  templates,
  lang,
}: {
  templates: Template[];
  lang: Lang;
}) {
  const { t } = useTranslation(lang);
  const [query, setQuery] = useState("");
  const [tag, setTag] = useState<"all" | "official" | "community">("all");
  const [copied, setCopied] = useState("");

  const filtered = useMemo(
    () =>
      templates.filter((p) => {
        const q = query.toLowerCase();
        const matchQ =
          !q ||
          p.name.toLowerCase().includes(q) ||
          p.tags.join(",").toLowerCase().includes(q);
        const tagsStr = p.tags.join(",");
        const matchTag =
          tag === "all" ||
          (tag === "community" ? !tagsStr.includes("official") : tagsStr.includes(tag));
        return matchQ && matchTag;
      }),
    [templates, query, tag],
  );

  const copyCmd = (name: string) => {
    void navigator.clipboard.writeText(`naeos template init ${name}`).then(() => {
      setCopied(name);
      window.setTimeout(() => setCopied(""), 2000);
    });
  };

  return (
    <>
      <div className="plugins-toolbar">
        <input
          type="text"
          className="input"
          placeholder={lang === "id" ? "Cari template..." : "Search templates..."}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          aria-label="Search templates"
        />
        <div className="plugins-tags">
          {(["all", "official", "community"] as const).map((tg) => (
            <button
              key={tg}
              className={`tag-btn${tag === tg ? " active" : ""}`}
              onClick={() => setTag(tg)}
            >
              {tg.charAt(0).toUpperCase() + tg.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div className="plugins-grid">
        {filtered.map((p) => {
          const official = p.tags.includes("official");
          return (
            <div key={p.name} className="plugin-card">
              <div className="plugin-card-header">
                <div className="plugin-icon">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                    <rect x="3" y="3" width="7" height="7" /><rect x="14" y="3" width="7" height="7" />
                    <rect x="14" y="14" width="7" height="7" /><rect x="3" y="14" width="7" height="7" />
                  </svg>
                </div>
                <div>
                  <h3 className="plugin-name">{p.name}</h3>
                  <span className="plugin-version">v{p.version}</span>{" "}
                  <span className={`badge ${official ? "badge-green" : "badge-blue"}`}>
                    {official ? "Official" : "Community"}
                  </span>
                </div>
              </div>
              <p className="plugin-desc">{p.description}</p>
              <div className="plugin-meta">
                {p.languages.length > 0 && (
                  <span className="plugin-meta-item">{p.languages.join(", ")}</span>
                )}
                <span className="plugin-meta-item">{p.downloads}</span>
                <span className="plugin-meta-item">{formatMonthYear(p.updated_at)}</span>
              </div>
              <div className="plugin-tags-row">
                {[...(p.frameworks ?? []), ...p.tags].map((tagItem) => (
                  <span key={tagItem} className="plugin-tag">{tagItem}</span>
                ))}
              </div>
              <div className="plugin-actions">
                <a href={p.repo_url} className="btn btn-secondary btn-sm" target="_blank" rel="noopener">GitHub</a>
                <button className="btn btn-primary btn-sm" onClick={() => copyCmd(p.name)}>
                  {copied === p.name ? "Copied!" : t("templates_use")}
                </button>
              </div>
            </div>
          );
        })}
      </div>
      {filtered.length === 0 && (
        <p style={{ color: "var(--color-text-muted)", textAlign: "center", padding: "2rem 0" }}>
          {lang === "id" ? "Tidak ada template ditemukan." : "No templates found."}
        </p>
      )}
    </>
  );
}
