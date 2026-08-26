"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

export interface BlogCardData {
  url: string;
  title: string;
  date?: string;
  categories?: string[];
  summary: string;
  readingTime: number;
}

function formatDate(iso: string | undefined, lang: Lang): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(lang === "id" ? "id-ID" : "en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export default function BlogListGrid({
  posts,
  lang,
}: {
  posts: BlogCardData[];
  lang: Lang;
}) {
  const { t } = useTranslation(lang);
  const [category, setCategory] = useState("all");

  const categories = useMemo(() => {
    const set = new Set<string>();
    for (const p of posts) for (const c of p.categories ?? []) set.add(c);
    return [...set];
  }, [posts]);

  const filtered =
    category === "all"
      ? posts
      : posts.filter((p) => (p.categories ?? []).includes(category));

  return (
    <>
      {categories.length > 0 && (
        <div className="blog-filters" style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", marginBottom: "2rem" }}>
          <button className={`category-btn${category === "all" ? " active" : ""}`} onClick={() => setCategory("all")}>
            {t("filter_all")}
          </button>
          {categories.map((c) => (
            <button key={c} className={`category-btn${category === c ? " active" : ""}`} onClick={() => setCategory(c)}>
              {c}
            </button>
          ))}
        </div>
      )}
      <div className="blog-grid stagger-fade">
        {filtered.map((post) => (
          <article key={post.url} className="blog-card">
            <div className="blog-date">{formatDate(post.date, lang)}</div>
            {(post.categories ?? []).length > 0 && (
              <div className="blog-categories">
                {post.categories!.map((c) => (
                  <span key={c} className="category-badge">{c}</span>
                ))}
              </div>
            )}
            <h3>
              <Link href={post.url}>{post.title}</Link>
            </h3>
            <p>{post.summary}</p>
            <span className="reading-time">
              {post.readingTime} {t("blog_min_read")}
            </span>
          </article>
        ))}
      </div>
    </>
  );
}
