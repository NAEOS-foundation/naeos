"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import { openSearch } from "@/components/chrome/Header";
import { DEFAULT_LANG, type Lang } from "@/lib/site";

export default function NotFound() {
  const [lang, setLang] = useState<Lang>(DEFAULT_LANG);

  useEffect(() => {
    if (window.location.pathname.startsWith("/id")) setLang("id");
  }, []);

  const { t } = useTranslation(lang);
  const homeHref = lang === "id" ? "/id/" : "/";

  return (
    <section className="section section-first">
      <div className="container">
        <div className="not-found">
          <div className="not-found-art" aria-hidden="true">
            <span className="nf-card nf-card-a">4</span>
            <span className="nf-lens" />
            <span className="nf-card nf-card-b">4</span>
          </div>
          <h1 className="not-found-code">404</h1>
          <p className="not-found-desc">{t("page_not_found_desc")}</p>
          <p className="not-found-hint">{t("page_not_found_hint")}</p>
          <div className="not-found-actions">
            <Link href={homeHref} className="btn btn-primary">{t("back_home")}</Link>
            <button className="btn btn-secondary" onClick={() => openSearch()}>{t("nav_search")}</button>
            <Link href={lang === "id" ? "/id/docs/" : "/docs/"} className="btn btn-secondary">{t("nav_docs")}</Link>
          </div>
        </div>
      </div>
    </section>
  );
}
