"use client";

import { useState, type FormEvent } from "react";
import { useTranslation } from "@/lib/useTranslation";
import type { Lang } from "@/lib/site";

export default function NewsletterForm({ lang }: { lang: Lang }) {
  const { t } = useTranslation(lang);
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = e.currentTarget;
    const emailInput = form.elements.namedItem("email") as HTMLInputElement;
    const email = emailInput.value.trim();
    const website = (form.elements.namedItem("website") as HTMLInputElement).value;

    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setStatus("error");
      setMessage(t("newsletter_invalid_email"));
      return;
    }

    setStatus("loading");
    setMessage(t("newsletter_loading"));
    try {
      const res = await fetch("/api/newsletter", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, locale: lang, website }),
      });
      const data = (await res.json()) as { ok?: boolean };
      if (res.ok && data.ok) {
        setStatus("success");
        setMessage(t("newsletter_success"));
        form.reset();
      } else {
        setStatus("error");
        setMessage(t("newsletter_error"));
      }
    } catch {
      setStatus("error");
      setMessage(t("newsletter_error"));
    }
  }

  return (
    <div className="newsletter-section" style={{ marginTop: "1.25rem" }}>
      <h4 style={{ marginBottom: "0.5rem", color: "var(--color-text)", fontSize: "var(--font-size-sm)" }}>
        {t("footer_newsletter")}
      </h4>
      <form className="newsletter-form" style={{ maxWidth: "100%" }} noValidate onSubmit={onSubmit}>
        <input
          className="newsletter-email"
          type="email"
          name="email"
          placeholder={t("newsletter_email_placeholder")}
          required
          autoComplete="email"
          aria-label={t("footer_newsletter")}
          style={{ padding: "0.5rem 0.75rem", fontSize: "var(--font-size-sm)" }}
        />
        <input className="newsletter-honeypot" type="text" name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" />
        <button type="submit" className="btn btn-primary btn-sm" disabled={status === "loading"}>
          {t("footer_subscribe")}
        </button>
      </form>
      <div className={`newsletter-message${status === "success" ? " is-success" : status === "error" ? " is-error" : ""}`} role="status" aria-live="polite">
        {message}
      </div>
    </div>
  );
}
