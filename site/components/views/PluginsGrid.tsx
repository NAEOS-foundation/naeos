"use client";

import { useMemo, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

interface Plugin {
  name: string;
  version: string;
  description: string;
  author: string;
  tags: string[];
  type: string;
  repo_url: string;
  license: string;
  downloads: number;
  platform: string;
  download_url: string;
  updated_at: string;
}

const LayerIcon = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
    <path d="M12 2L2 7l10 5 10-5-10-5z" /><path d="M2 17l10 5 10-5" /><path d="M2 12l10 5 10-5" />
  </svg>
);

function formatMonthYear(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return `${d.toLocaleString("en-US", { month: "short" })} ${d.getFullYear()}`;
}

export default function PluginsGrid({
  plugins,
  lang,
}: {
  plugins: Plugin[];
  lang: Lang;
}) {
  const { t } = useTranslation(lang);
  const [query, setQuery] = useState("");
  const [tag, setTag] = useState<"all" | "official" | "community">("all");
  const [copied, setCopied] = useState("");

  const filtered = useMemo(
    () =>
      plugins.filter((p) => {
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
    [plugins, query, tag],
  );

  const copyCmd = (name: string) => {
    void navigator.clipboard
      .writeText(`naeos marketplace plugin install ${name}`)
      .then(() => {
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
          placeholder={lang === "id" ? "Cari plugin..." : "Search plugins..."}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          aria-label="Search plugins"
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
                <div className="plugin-icon"><LayerIcon /></div>
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
                <span className="plugin-meta-item">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M12 2v20M2 12h20" /></svg>
                  {p.type}
                </span>
                <span className="plugin-meta-item">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                  {p.downloads}
                </span>
                <span className="plugin-meta-item">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><circle cx="12" cy="12" r="10" /><path d="M12 6v6l4 2" /></svg>
                  {formatMonthYear(p.updated_at)}
                </span>
              </div>
              <div className="plugin-tags-row">
                {p.tags.map((tagItem) => (
                  <span key={tagItem} className="plugin-tag">{tagItem}</span>
                ))}
              </div>
              <div className="plugin-actions">
                <a href={p.repo_url} className="btn btn-secondary btn-sm" target="_blank" rel="noopener">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" /></svg>
                  GitHub
                </a>
                <button className="btn btn-primary btn-sm" onClick={() => copyCmd(p.name)}>
                  {copied === p.name ? "Copied!" : t("plugins_install")}
                </button>
              </div>
            </div>
          );
        })}
      </div>
      {filtered.length === 0 && (
        <p style={{ color: "var(--color-text-muted)", textAlign: "center", padding: "2rem 0" }}>
          {lang === "id" ? "Tidak ada plugin ditemukan." : "No plugins found."}
        </p>
      )}
    </>
  );
}

export function PluginPublishSteps({ lang }: { lang: Lang }) {
  const { t } = useTranslation(lang);
  return (
    <div className="section" style={{ marginBottom: "2rem" }}>
      <h2 className="section-title">{t("plugins_publish_title")}</h2>
      <p>{t("plugins_publish_intro")}</p>
      <ol>
        <li>{t("plugins_publish_step1")}: <code>naeos plugin init my-plugin --author &quot;You&quot; --desc &quot;My first plugin&quot;</code></li>
        <li>{t("plugins_publish_step2")}: <code>go test -race ./... &amp;&amp; GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .</code></li>
        <li>{t("plugins_publish_step3")}: <code>{t("plugins_publish_cmd")}</code></li>
      </ol>
      <p><em>{t("plugins_publish_official")}</em></p>
    </div>
  );
}
