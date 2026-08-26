"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslation } from "@/lib/useTranslation";
import { DEFAULT_LANG, type Lang } from "@/lib/site";

function CopyIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="1" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

export function CopyCodeButton({ getText, lang = DEFAULT_LANG }: { getText: () => string; lang?: Lang }) {
  const [copied, setCopied] = useState(false);
  const { t } = useTranslation(lang);
  return (
    <button
      className={`copy-btn${copied ? " copied" : ""}`}
      aria-label={t("copy_code")}
      onClick={() => {
        navigator.clipboard.writeText(getText()).then(() => {
          setCopied(true);
          window.setTimeout(() => setCopied(false), 2000);
        });
      }}
    >
      <CopyIcon />
      {copied ? t("copy_copied") : t("copy")}
    </button>
  );
}

/** Adds header (language label + copy) to a highlighted code block. */
export function CodeBlock({
  children,
  lang,
}: {
  children: React.ReactNode;
  lang: string;
}) {
  const preRef = useRef<HTMLPreElement>(null);
  return (
    <div className="code-block highlight">
      <div className="code-block-header">
        <span>{lang || "text"}</span>
        <CopyCodeButton getText={() => preRef.current?.textContent ?? ""} />
      </div>
      <pre ref={preRef} data-lang={lang}>
        {children}
      </pre>
    </div>
  );
}

export function MermaidDiagram({ chart }: { chart: string }) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    let cancelled = false;
    import("mermaid")
      .then(({ default: mermaid }) => {
        if (cancelled || !ref.current) return;
        mermaid.initialize({
          startOnLoad: false,
          theme: "dark",
          themeVariables: {
            primaryColor: "#111122",
            lineColor: "#08d6ff",
          },
          securityLevel: "strict",
        });
        void mermaid
          .render(`mmd-${Math.random().toString(36).slice(2)}`, chart)
          .then(({ svg }) => {
            if (!cancelled && ref.current) ref.current.innerHTML = svg;
          });
      })
      .catch(() => {
        if (ref.current) ref.current.textContent = chart;
      });
    return () => {
      cancelled = true;
    };
  }, [chart]);
  return (
    <div className="mermaid-wrapper">
      <div ref={ref}>{chart}</div>
    </div>
  );
}

/**
 * Attaches progressive enhancements to server-rendered markdown content:
 * FAQ accordions, image lightbox, and copy-on-hover for pre blocks.
 */
export default function ContentEffects() {
  useEffect(() => {
    const root = document.getElementById("main-content");
    if (!root) return;

    const isId = window.location.pathname.startsWith("/id");
    const copyLabel = isId ? "Salin" : "Copy";
    const copiedLabel = isId ? "Tersalin!" : "Copied!";

    const faqHandlers: [HTMLElement, () => void][] = [];
    root.querySelectorAll<HTMLElement>(".faq-item").forEach((item) => {
      const btn = item.querySelector<HTMLElement>(".faq-question");
      const answer = item.querySelector<HTMLElement>(".faq-answer");
      if (!btn || !answer) return;
      const handler = () => {
        const open = item.classList.toggle("open");
        answer.style.maxHeight = open ? `${answer.scrollHeight}px` : "";
      };
      btn.addEventListener("click", handler);
      faqHandlers.push([btn, handler]);
    });

    let overlay: HTMLElement | null = null;
    function closeOverlay() {
      overlay?.remove();
      overlay = null;
    }
    const imgHandler = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName !== "IMG" || !target.closest(".single-content")) return;
      closeOverlay();
      overlay = document.createElement("div");
      overlay.className = "lightbox-overlay";
      const img = document.createElement("img");
      img.src = (target as HTMLImageElement).src;
      overlay.appendChild(img);
      overlay.addEventListener("click", closeOverlay);
      document.body.appendChild(overlay);
    };
    root.addEventListener("click", imgHandler);

    const hoverCleanups: (() => void)[] = [];
    root.querySelectorAll(".content-section pre").forEach((pre) => {
      if (pre.closest(".code-block")) return;
      const btn = document.createElement("button");
      btn.className = "copy-hover-btn";
      btn.textContent = copyLabel;
      btn.addEventListener("click", () => {
        navigator.clipboard.writeText(pre.textContent ?? "").then(() => {
          btn.textContent = copiedLabel;
          window.setTimeout(() => {
            btn.textContent = copyLabel;
          }, 1500);
        });
      });
      (pre as HTMLElement).style.position = "relative";
      pre.appendChild(btn);
      hoverCleanups.push(() => btn.remove());
    });

    return () => {
      for (const [btn, handler] of faqHandlers) btn.removeEventListener("click", handler);
      root.removeEventListener("click", imgHandler);
      for (const fn of hoverCleanups) fn();
      closeOverlay();
    };
  }, []);

  return null;
}
