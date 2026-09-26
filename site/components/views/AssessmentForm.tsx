"use client";

import { FormEvent, useState } from "react";

const API_BASE = (process.env.NEXT_PUBLIC_CRM_API_URL ?? "").replace(/\/$/, "");

export default function AssessmentForm({ lang }: { lang: "en" | "id" }) {
  const id = lang === "id";
  const [status, setStatus] = useState<"idle" | "sending" | "success" | "error">("idle");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("sending");
    const form = new FormData(event.currentTarget);
    const payload = Object.fromEntries(form.entries());

    try {
      const response = await fetch(`${API_BASE}/api/v1/public/assessment-intake`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error("Submission failed");
      event.currentTarget.reset();
      setStatus("success");
    } catch {
      setStatus("error");
    }
  }

  const field = (name: string, label: string, placeholder: string, required = false) => (
    <label style={{ display: "grid", gap: "0.4rem" }}>
      <span>{label}{required ? " *" : ""}</span>
      <input
        name={name}
        required={required}
        placeholder={placeholder}
        maxLength={2000}
        style={{ width: "100%", padding: "0.75rem", borderRadius: "0.5rem", border: "1px solid var(--border-color, #ccc)", background: "var(--surface, transparent)", color: "inherit" }}
      />
    </label>
  );

  return (
    <form onSubmit={submit} className="community-card" style={{ display: "grid", gap: "1rem", marginTop: "2rem" }}>
      {field("name", id ? "Nama" : "Name", id ? "Nama lengkap" : "Your name", true)}
      {field("email", "Email", "you@company.com", true)}
      {field("company", id ? "Perusahaan" : "Company", id ? "Nama perusahaan" : "Company name", true)}
      {field("agents", id ? "AI coding agents" : "AI coding agents", id ? "Contoh: Codex, Claude Code, Copilot" : "e.g. Codex, Claude Code, Copilot")}
      {field("workflows", id ? "Workflow utama" : "Main workflows", id ? "Contoh: PR, CI/CD, deployment" : "e.g. PR, CI/CD, deployment")}
      {field("controls", id ? "Control yang sudah ada" : "Existing controls", id ? "RBAC, approval, policy, audit..." : "RBAC, approvals, policy, audit...")}
      {field("risk", id ? "Aksi agent paling berisiko" : "Highest-risk agent actions", id ? "Apa yang paling ingin dikontrol?" : "What do you most want to control?")}
      {field("goals", id ? "Target 30 hari" : "30-day goals", id ? "Apa yang ingin divalidasi?" : "What do you want to validate?")}
      <input name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" style={{ position: "absolute", left: "-10000px", opacity: 0 }} />
      <button type="submit" className="btn btn-primary" disabled={status === "sending"}>
        {status === "sending" ? (id ? "Mengirim..." : "Sending...") : (id ? "Kirim Assessment Request" : "Request Assessment")}
      </button>
      {status === "success" && <p role="status">{id ? "Terima kasih. Request Anda sudah masuk ke CRM NAEOS." : "Thank you. Your request is now in the NAEOS CRM."}</p>}
      {status === "error" && <p role="alert">{id ? "Request gagal dikirim. Silakan coba lagi atau gunakan email." : "Submission failed. Please try again or use email."}</p>}
    </form>
  );
}
