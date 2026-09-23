"use client";
// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0


import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";
import { SITE } from "@/lib/site";

export function openSearch() {
  window.dispatchEvent(new CustomEvent("open-search"));
}

export function toggleTheme() {
  const el = document.documentElement;
  const next = el.getAttribute("data-theme") === "light" ? "dark" : "light";
  document.body.classList.add("theme-transitioning");
  el.setAttribute("data-theme", next);
  try {
    localStorage.setItem("theme", next);
  } catch {
    /* ignore */
  }
  window.setTimeout(() => document.body.classList.remove("theme-transitioning"), 400);
}

const DOCS_DROPDOWN = [
  ["docs/getting-started", "footer_getting_started"],
  ["docs/installation", "footer_installation"],
  ["docs/architecture", "footer_architecture"],
  ["docs/cli-reference", "footer_cli_reference"],
  ["docs/spec-language", "nav_spec_language"],
  ["docs/ai-compiler", "nav_ai_compiler"],
] as const;

const MAIN_LINKS = [
  ["blog", "nav_blog"],
  ["plugins", "nav_plugins"],
  ["investor-deck", "nav_investor_deck"],
  ["launch-announcement", "nav_launch_announcement"],
  ["community", "nav_community"],
] as const;

function label(t: (key: string) => string, seg: string, key: string | null): string {
  if (key) return t(key);
  const labels: Record<string, string> = {
    features: "Product",
    "docs/architecture": "Architecture",
    "docs/getting-started": "Developers",
    plugins: "Ecosystem",
    blog: "Resources",
    templates: "Templates",
  };
  return labels[seg] ?? "Schema Registry";
}

interface Props {
  lang: Lang;
}

export default function Header({ lang }: Props) {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);
  const { t } = useTranslation(lang);
  const base = lang === "en" ? "" : "/id";

  useEffect(() => setMenuOpen(false), [pathname]);

  useEffect(() => {
    if (!menuOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenuOpen(false);
    };
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", onKey);
    };
  }, [menuOpen]);

  const isActive = (href: string) =>
    href !== "" && href !== "/" && (pathname === href || pathname.startsWith(`${href}/`));

  const langlessPath = pathname.replace(/^\/(?:en|id)(?=\/|$)/, "") || "/";
  const enHref = langlessPath;
  const idHref = `/id${langlessPath}`;

  return (
    <header className="site-header" role="banner">
      <div className="container header-inner">
        <Link href={base === "" ? "/" : `${base}/`} className="logo" aria-label="NAEOS Home">
          <img className="logo-icon" src="/images/Logo.png" width={32} height={32} loading="lazy" alt="" aria-hidden="true" />
          <span className="logo-text">NAEOS</span>
        </Link>
        <nav className="site-nav" role="navigation" aria-label="Main navigation">
          <Link href={`${base}/features`} className="nav-link" {...(isActive(`${base}/features`) ? { "aria-current": "page" as const } : {})}>Product</Link>
          <Link href={`${base}/docs/architecture`} className="nav-link" {...(isActive(`${base}/docs/architecture`) ? { "aria-current": "page" as const } : {})}>Architecture</Link>
          <div className="nav-dropdown">
            <Link href={`${base}/docs/getting-started`} className="nav-link" {...(isActive(`${base}/docs`) ? { "aria-current": "page" as const } : {})}>
              Developers
              <svg width="10" height="6" viewBox="0 0 10 6" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ marginLeft: 2, verticalAlign: "middle" }} aria-hidden="true"><path d="M1 1l4 4 4-4" /></svg>
            </Link>
            <div className="nav-dropdown-content">
              <Link href={`${base}/docs/getting-started`} className="nav-dropdown-link">{t("footer_getting_started")}</Link>
              <Link href={`${base}/docs/installation`} className="nav-dropdown-link">{t("footer_installation")}</Link>
              <Link href={`${base}/docs/architecture`} className="nav-dropdown-link">{t("footer_architecture")}</Link>
              <Link href={`${base}/docs/cli-reference`} className="nav-dropdown-link">{t("footer_cli_reference")}</Link>
              <Link href={`${base}/docs/spec-language`} className="nav-dropdown-link">{t("nav_spec_language")}</Link>
              <Link href={`${base}/docs/ai-compiler`} className="nav-dropdown-link">{t("nav_ai_compiler")}</Link>
            </div>
          </div>
          <Link href={`${base}/plugins`} className="nav-link" {...(isActive(`${base}/plugins`) ? { "aria-current": "page" as const } : {})}>Ecosystem</Link>
          <div className="nav-dropdown">
            <Link href={`${base}/blog`} className="nav-link" {...(isActive(`${base}/blog`) ? { "aria-current": "page" as const } : {})}>
              Resources
              <svg width="10" height="6" viewBox="0 0 10 6" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ marginLeft: 2, verticalAlign: "middle" }} aria-hidden="true"><path d="M1 1l4 4 4-4" /></svg>
            </Link>
            <div className="nav-dropdown-content">
              <Link href={`${base}/blog`} className="nav-dropdown-link">Blog</Link>
              <Link href={`${base}/docs`} className="nav-dropdown-link">{t("nav_docs")}</Link>
              <Link href={`${base}/investor-deck`} className="nav-dropdown-link">{t("nav_investor_deck")}</Link>
              <Link href={`${base}/community`} className="nav-dropdown-link">{t("nav_community")}</Link>
            </div>
          </div>
        </nav>
        <div className="header-actions">
          <button className="search-toggle" onClick={() => openSearch()} aria-label={t("nav_search")}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" /></svg>
          </button>
          <button className="theme-toggle" onClick={() => toggleTheme()} aria-label={t("toggle_theme")}>
            <svg className="sun-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><circle cx="12" cy="12" r="5" /><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" /></svg>
            <svg className="moon-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" /></svg>
          </button>
          <div className="lang-switcher">
            <a href={enHref} className={`lang-link${lang === "en" ? " active" : ""}`} aria-label="English">EN</a>
            <a href={idHref} className={`lang-link${lang === "id" ? " active" : ""}`} aria-label="Bahasa Indonesia">ID</a>
          </div>
          <a href={SITE.repo} className="nav-link header-github" target="_blank" rel="noopener">GitHub</a>
          <Link href={`${base}/download`} className="btn btn-primary btn-sm header-cta" data-umami-event="header-get-started">{t("cta_get_started")}</Link>
          <button
            className="mobile-menu-btn"
            aria-label={t("toggle_menu")}
            aria-expanded={menuOpen}
            aria-haspopup="true"
            aria-controls="mobile-navigation"
            onClick={() => setMenuOpen((v) => !v)}
          >
            <span /><span /><span />
          </button>
        </div>
      </div>
      <div id="mobile-navigation" className={`mobile-menu${menuOpen ? " open" : ""}`} role="navigation" aria-label="Mobile navigation" aria-hidden={!menuOpen}>
        {MOBILE_LINKS.map(([seg, key, indent]) => (
          <Link
            key={seg}
            href={`${base}/${seg}`}
            className="nav-link"
            style={indent ? { paddingLeft: "2rem", fontSize: "var(--font-size-sm)" } : undefined}
            {...(isActive(`${base}/${seg}`) && !indent ? { "aria-current": "page" as const } : {})}
          >
            {label(t, seg, key ?? null)}
          </Link>
        ))}
        <div className="lang-switcher mobile" style={{ padding: "0.75rem 1rem" }}>
          <a href={enHref} className={`lang-link${lang === "en" ? " active" : ""}`} aria-label="English">EN</a>
          <a href={idHref} className={`lang-link${lang === "id" ? " active" : ""}`} aria-label="Bahasa Indonesia">ID</a>
        </div>
        <div className="mobile-menu-actions">
          <button className="btn btn-secondary btn-sm" onClick={() => toggleTheme()} style={{ flex: 1 }}>{t("toggle_theme")}</button>
          <button className="btn btn-secondary btn-sm" onClick={() => { setMenuOpen(false); openSearch(); }} style={{ flex: 1 }}>{t("nav_search")}</button>
        </div>
      </div>
    </header>
  );
}

const MOBILE_LINKS = [
  ["features", null, false],
  ["docs/architecture", null, false],
  ["docs/getting-started", null, false],
  ["plugins", null, false],
  ["blog", null, false],
  ["community", "nav_community", false],
] as const;

