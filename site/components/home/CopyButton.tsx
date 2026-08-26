"use client";

import { useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

export default function CopyButton({
  text,
  label,
  lang = "en",
}: {
  text: string;
  label: string;
  lang?: Lang;
}) {
  const { t } = useTranslation(lang);
  const [copied, setCopied] = useState(false);

  function handleCopy() {
    void navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    }).catch(() => {});
  }

  return (
    <button
      className={`copy-btn${copied ? " copied" : ""}`}
      aria-label={label}
      onClick={handleCopy}
    >
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2" /><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" /></svg>
      {copied ? t("copy_copied", "Copied!") : t("copy")}
    </button>
  );
}
