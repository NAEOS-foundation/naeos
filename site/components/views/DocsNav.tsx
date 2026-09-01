"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

interface Entry {
  title: string;
  url: string;
}

export function DocsSidebar({
  entries,
  currentUrl,
  title,
  lang = "en",
}: {
  entries: Entry[];
  currentUrl: string;
  title: string;
  lang?: "en" | "id";
}) {
  const [filter, setFilter] = useState("");
  const filtered = useMemo(
    () =>
      entries.filter((e) => e.title.toLowerCase().includes(filter.toLowerCase())),
    [entries, filter],
  );

  return (
    <aside className="doc-sidebar">
      <h4>{title}</h4>
      <input
        type="text"
        className="sidebar-search"
        placeholder="Filter..."
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        aria-label={lang === "id" ? "Filter dokumen" : "Filter docs"}
      />
      <nav id="sidebar-nav">
        {filtered.map((e) => (
          <Link
            key={e.url}
            href={e.url}
            className={e.url === currentUrl ? "active" : ""}
            data-title={e.title}
          >
            {e.title}
          </Link>
        ))}
        {filtered.length === 0 && <span className="sidebar-empty">—</span>}
      </nav>
    </aside>
  );
}

interface TocHeading {
  depth: number;
  text: string;
  id: string;
}

export function Toc({
  headings,
  title,
}: {
  headings: TocHeading[];
  title: string;
}) {
  const [activeId, setActiveId] = useState("");

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) setActiveId(entry.target.id);
        }
      },
      { rootMargin: "-80px 0px -65% 0px" },
    );
    for (const h of headings) {
      const el = document.getElementById(h.id);
      if (el) observer.observe(el);
    }
    return () => observer.disconnect();
  }, [headings]);

  return (
    <div className="toc">
      <div className="toc-title">{title}</div>
      <nav>
        {headings.map((h) => (
          <a
            key={h.id}
            href={`#${h.id}`}
            className={activeId === h.id ? "active" : ""}
            style={{ paddingLeft: `${(h.depth - 2) * 12}px` }}
          >
            {h.text}
          </a>
        ))}
      </nav>
    </div>
  );
}
