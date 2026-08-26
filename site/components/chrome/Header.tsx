"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

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
  ["templates", null],
  ["schemaregistry", null],
  ["showcase", "nav_showcase"],
  ["cookbook", "nav_cookbook"],
  ["download", "nav_download"],
  ["community", "nav_community"],
] as const;

function label(t: (key: string) => string, seg: string, key: string | null): string {
  if (key) return t(key);
  return seg === "templates" ? "Templates" : "Schema Registry";
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

  const altHref =
    lang === "en"
      ? pathname === "/" || pathname === ""
        ? "/id/"
        : `/id${pathname}`
      : pathname.replace(/^\/id\/?/, "/") || "/";

  return (
    <header className="site-header" role="banner">
      <div className="container header-inner">
        <Link href={base === "" ? "/" : `${base}/`} className="logo" aria-label="NAEOS Home">
          <img className="logo-icon" src="/images/Logo.png" width={32} height={32} loading="lazy" alt="" aria-hidden="true" />
          <span className="logo-text">NAEOS</span>
        </Link>
        <nav className="site-nav" role="navigation" aria-label="Main navigation">
          <Link href={`${base}/features`} className="nav-link" {...(isActive(`${base}/features`) ? { "aria-current": "page" as const } : {})}>
            {t("nav_features")}
          </Link>
          <div className="nav-dropdown">
            <Link href={`${base}/docs`} className="nav-link" {...(isActive(`${base}/docs`) ? { "aria-current": "page" as const } : {})}>
              {t("nav_docs")}
              <svg width="10" height="6" viewBox="0 0 10 6" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ marginLeft: 2, verticalAlign: "middle" }} aria-hidden="true"><path d="M1 1l4 4 4-4" /></svg>
            </Link>
            <div className="nav-dropdown-content">
              {DOCS_DROPDOWN.map(([path, key]) => (
                <Link key={path} href={`${base}/${path}`} className="nav-dropdown-link">{t(key)}</Link>
              ))}
            </div>
          </div>
          {MAIN_LINKS.map(([seg, key]) => (
            <Link key={seg} href={`${base}/${seg}`} className="nav-link" {...(isActive(`${base}/${seg}`) ? { "aria-current": "page" as const } : {})}>
              {label(t, seg, key)}
            </Link>
          ))}
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
            <a href={pathname} className={`lang-link${lang === "en" ? " active" : ""}`} aria-label="English">EN</a>
            <a href={altHref} className={`lang-link${lang === "id" ? " active" : ""}`} aria-label="Bahasa Indonesia">ID</a>
          </div>
          <Link href={`${base}/download`} className="btn btn-primary btn-sm header-cta">{t("cta_get_started")}</Link>
          <button
            className="mobile-menu-btn"
            aria-label={t("toggle_menu")}
            aria-expanded={menuOpen}
            onClick={() => setMenuOpen((v) => !v)}
          >
            <span /><span /><span />
          </button>
        </div>
      </div>
      <div className={`mobile-menu${menuOpen ? " open" : ""}`} role="navigation" aria-label="Mobile navigation">
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
          <a href={pathname} className={`lang-link${lang === "en" ? " active" : ""}`} aria-label="English">EN</a>
          <a href={altHref} className={`lang-link${lang === "id" ? " active" : ""}`} aria-label="Bahasa Indonesia">ID</a>
        </div>
        <div className="mobile-menu-actions">
          <button className="btn btn-secondary btn-sm" onClick={() => toggleTheme()} style={{ flex: 1 }}>{t("toggle_theme")}</button>
          <button className="btn btn-secondary btn-sm" onClick={() => { setMenuOpen(false); openSearch(); }} style={{ flex: 1 }}>{t("nav_search")}</button>
        </div>
      </div>
    </header>
  );
}

const MOBILE_SUB_DOCS = [
  ["docs/getting-started", "footer_getting_started", true],
  ["docs/installation", "footer_installation", true],
  ["docs/architecture", "footer_architecture", true],
  ["docs/cli-reference", "footer_cli_reference", true],
] as const;

const MOBILE_LINKS = (
  [
    ["features", "nav_features"],
    ["docs", "nav_docs"],
    ...MOBILE_SUB_DOCS,
    ...MAIN_LINKS,
  ] as const
).map(([seg, key, indent]) => [seg, key ?? null, Boolean(indent)] as const);

