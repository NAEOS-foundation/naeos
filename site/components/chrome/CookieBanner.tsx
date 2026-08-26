"use client";

import { useEffect, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

export default function CookieBanner({ lang }: { lang: Lang }) {
  const { t } = useTranslation(lang);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    try {
      if (!localStorage.getItem("cookies-accepted") && !localStorage.getItem("cookies-declined")) {
        setVisible(true);
      }
    } catch {
      /* ignore */
    }
  }, []);

  function choose(key: "cookies-accepted" | "cookies-declined") {
    try {
      localStorage.setItem(key, "1");
    } catch {
      /* ignore */
    }
    setVisible(false);
  }

  if (!visible) return null;

  return (
    <div className="cookie-banner show" role="region" aria-label="Cookie consent">
      <div className="container">
        <p>{t("cookie_consent")}</p>
        <div style={{ display: "flex", gap: "0.75rem", flexShrink: 0 }}>
          <button className="btn btn-secondary btn-sm" onClick={() => choose("cookies-declined")}>
            {t("cookie_decline")}
          </button>
          <button className="btn btn-primary btn-sm" onClick={() => choose("cookies-accepted")}>
            {t("cookie_accept")}
          </button>
        </div>
      </div>
    </div>
  );
}
