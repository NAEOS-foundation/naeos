"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

interface SearchEntry {
  title: string;
  permalink: string;
  section: string;
  content: string;
  sections: string[];
}

interface MatchInfo {
  key?: string | number;
  value?: string;
  indices: readonly (readonly [number, number])[];
}

interface ResultItem {
  item: SearchEntry;
  matches?: readonly MatchInfo[];
}

function highlight(
  text: string,
  indices: readonly (readonly [number, number])[] | undefined,
): React.ReactNode {
  if (!indices || indices.length === 0) return text;
  const parts: React.ReactNode[] = [];
  let last = 0;
  for (const [start, end] of indices.slice(0, 3)) {
    if (start > last) parts.push(text.slice(last, start));
    parts.push(<mark key={`${start}-${end}`}>{text.slice(start, end + 1)}</mark>);
    last = end + 1;
  }
  if (last < text.length) parts.push(text.slice(last));
  return parts;
}

export default function SearchModal({ lang }: { lang: Lang }) {
  const { t } = useTranslation(lang);
  const pathname = usePathname();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<ResultItem[]>([]);
  const [recent, setRecent] = useState<string[]>([]);
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const fuseRef = useRef<import("fuse.js").default<SearchEntry> | null>(null);
  const indexRef = useRef<SearchEntry[] | null>(null);

  const isId = lang === "id" || pathname.startsWith("/id");
  const indexUrl = isId ? "/id/index.json" : "/index.json";

  useEffect(() => {
    try {
      const saved = localStorage.getItem("recent-searches");
      if (saved) setRecent(JSON.parse(saved) as string[]);
    } catch {
      /* ignore */
    }
  }, []);

  const openModal = useCallback(() => {
    setOpen(true);
    requestAnimationFrame(() => inputRef.current?.focus());
  }, []);

  const closeModal = useCallback(() => {
    setOpen(false);
    setQuery("");
    setResults([]);
    setSelected(0);
  }, []);

  useEffect(() => {
    const onOpen = () => openModal();
    window.addEventListener("open-search", onOpen);
    return () => window.removeEventListener("open-search", onOpen);
  }, [openModal]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        open ? closeModal() : openModal();
      }
      if (e.key === "Escape" && open) closeModal();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, openModal, closeModal]);

  useEffect(() => {
    document.body.style.overflow = open ? "hidden" : "";
    return () => {
      document.body.style.overflow = "";
    };
  }, [open]);

  async function ensureIndex() {
    if (indexRef.current && fuseRef.current) return;
    try {
      const [{ default: Fuse }, res] = await Promise.all([
        import("fuse.js"),
        fetch(indexUrl),
      ]);
      const data = (await res.json()) as SearchEntry[];
      indexRef.current = data;
      fuseRef.current = new Fuse(data, {
        keys: ["title", "sections", "content"],
        threshold: 0.4,
        includeScore: true,
        includeMatches: true,
      });
    } catch {
      indexRef.current = [];
      fuseRef.current = null;
    }
  }

  function runSearch(q: string) {
    setQuery(q);
    setSelected(0);
    if (!q.trim()) {
      setResults([]);
      return;
    }
    void ensureIndex().then(() => {
      const fuse = fuseRef.current;
      if (!fuse) {
        const data = indexRef.current ?? [];
        const lower = q.toLowerCase();
        setResults(
          data
            .filter(
              (item) =>
                item.title.toLowerCase().includes(lower) ||
                item.content.toLowerCase().includes(lower),
            )
            .slice(0, 20)
            .map((item) => ({ item })),
        );
        return;
      }
      setResults(fuse.search(q).slice(0, 20));
    });
  }

  function saveRecent(q: string) {
    const next = [q, ...recent.filter((r) => r !== q)].slice(0, 5);
    setRecent(next);
    try {
      localStorage.setItem("recent-searches", JSON.stringify(next));
    } catch {
      /* ignore */
    }
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelected((s) => Math.min(s + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelected((s) => Math.max(s - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      const hit = results[selected];
      if (hit) {
        saveRecent(query);
        closeModal();
        window.location.href = hit.item.permalink;
      }
    }
  }

  const grouped = new Map<string, ResultItem[]>();
  for (const r of results) {
    const list = grouped.get(r.item.section) ?? [];
    list.push(r);
    grouped.set(r.item.section, list);
  }

  let flatIndex = -1;

  return (
    <>
      {open && <div className="search-overlay open" onClick={closeModal} />}
      <div className={`search-modal${open ? " open" : ""}`} role="dialog" aria-modal="true" aria-label={t("nav_search")}>
        <div className="search-input-wrapper">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" /></svg>
          <input
            ref={inputRef}
            id="search-input"
            type="text"
            placeholder={t("search_placeholder")}
            value={query}
            onChange={(e) => runSearch(e.target.value)}
            onKeyDown={onKeyDown}
            autoComplete="off"
          />
          <button className="search-close" onClick={closeModal} aria-label={t("close")}>✕</button>
        </div>
        <div className="search-results">
          {!query.trim() ? (
            <>
              {t("search_hint")}
              {recent.length > 0 && (
                <div className="search-recent">
                  {recent.map((r) => (
                    <button key={r} className="search-recent-item" onClick={() => runSearch(r)}>
                      {r}
                    </button>
                  ))}
                </div>
              )}
            </>
          ) : results.length === 0 ? (
            <div className="search-no-results">{t("search_no_results")}</div>
          ) : (
            [...grouped.entries()].map(([section, items]) => (
              <div key={section} className="search-group">
                <div className="search-group-title">{section}</div>
                {items.map((r) => {
                  flatIndex += 1;
                  const idx = flatIndex;
                  const match = r.matches?.find((m) => m.key === "title");
                  const contentMatch = r.matches?.find((m) => m.key === "content");
                  const snippetStart = contentMatch ? Math.max(0, (contentMatch.indices[0]?.[0] ?? 0) - 60) : 0;
                  const snippet = contentMatch
                    ? `${snippetStart > 0 ? "…" : ""}${r.item.content.slice(snippetStart, snippetStart + 160)}…`
                    : r.item.content.slice(0, 160);
                  return (
                    <Link
                      key={r.item.permalink}
                      href={r.item.permalink}
                      className={`search-result-item${idx === selected ? " selected" : ""}`}
                      onMouseEnter={() => setSelected(idx)}
                      onClick={() => {
                        saveRecent(query);
                        closeModal();
                      }}
                    >
                      <div className="search-result-title">{highlight(r.item.title, match?.indices)}</div>
                      <div className="search-result-snippet">{contentMatch ? highlight(snippet, undefined) : snippet}</div>
                    </Link>
                  );
                })}
              </div>
            ))
          )}
        </div>
      </div>
    </>
  );
}
