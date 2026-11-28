// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0
"use client";

import { useState, type FormEvent } from "react";

const roles = [
  "Technical Founding Partner / CTO",
  "AI Systems & Engineering Partner",
  "Product & Engineering Strategy Partner",
  "Commercial & GTM Partner",
  "Governance & Security Specialist",
];
const areas = ["AI agents & orchestration", "Backend / distributed systems", "Frontend & UX", "DevOps / platform", "Security & governance", "Product strategy", "Sales / GTM", "Partnerships", "Research / evaluation"];

type Props = { lang: "en" | "id" };
const copy = {
  en: {
    title: "Founding Partner Application", intro: "Help build the core of NAEOS. Share your experience, the role you are interested in, and how you would contribute.",
    profile: "Profile", name: "Full name", email: "Email", location: "Location", linkedin: "LinkedIn URL", portfolio: "GitHub / portfolio URL", professional: "Current professional role",
    role: "Primary role", alternate: "Alternative role (optional)", expertise: "Areas of expertise", experience: "Relevant experience (min. 20 characters)", portfolioSummary: "Portfolio / projects summary", contribution: "What would you contribute to NAEOS? (min. 20 characters)", motivation: "Why do you want to build NAEOS with us? (min. 20 characters)",
    availability: "Availability", hours: "Hours per week", model: "Collaboration model", part: "Part-time", flexible: "Flexible", other: "Other / discuss", date: "Earliest availability", trial: "I can participate in a scoped trial collaboration up to 30 days, with written scope and agreements before work begins.",
    expectations: "Expectations", partnership: "Partnership expectations", equity: "Equity expectations (optional)", consent: "I consent to NAEOS using my application data to evaluate my application and contact me about it. I understand the data will be used for recruitment purposes.",
    submit: "Submit application", sending: "Submitting…", success: "Thank you. Your application has been received.", failure: "Submission failed. Please try again shortly.", required: "Required", noticeTitle: "Transparent early-stage conditions", notice: "NAEOS is pre-incorporation and pre-funding. This is part-time collaboration; no salary is currently available. Equity may be discussed but is not guaranteed and is subject to future legal structuring and a written agreement. No shares have been issued or guaranteed. Any trial (up to 30 days) requires written scope, deliverables, IP, confidentiality, and boundaries agreed in advance.", privacy: "We use your information only to evaluate this application and contact you about recruitment. Do not include confidential information in your submission.",
    honeypot: "Leave this blank",
  },
  id: {
    title: "Formulir Founding Partner", intro: "Bantu membangun core NAEOS. Ceritakan pengalaman, posisi yang diminati, dan kontribusi yang ingin Anda bawa.",
    profile: "Profil", name: "Nama lengkap", email: "Email", location: "Lokasi", linkedin: "URL LinkedIn", portfolio: "URL GitHub / portofolio", professional: "Peran profesional saat ini",
    role: "Posisi utama", alternate: "Posisi alternatif (opsional)", expertise: "Bidang keahlian", experience: "Pengalaman relevan (min. 20 karakter)", portfolioSummary: "Ringkasan portofolio / proyek", contribution: "Kontribusi apa yang akan Anda bawa untuk NAEOS? (min. 20 karakter)", motivation: "Mengapa ingin membangun NAEOS bersama kami? (min. 20 karakter)",
    availability: "Ketersediaan", hours: "Jam per minggu", model: "Model kolaborasi", part: "Paruh waktu", flexible: "Fleksibel", other: "Lainnya / diskusikan", date: "Ketersediaan mulai", trial: "Saya dapat mengikuti trial kolaborasi terstruktur hingga 30 hari, dengan scope dan kesepakatan tertulis sebelum pekerjaan dimulai.",
    expectations: "Ekspektasi", partnership: "Ekspektasi kemitraan", equity: "Ekspektasi equity (opsional)", consent: "Saya menyetujui NAEOS menggunakan data aplikasi untuk mengevaluasi aplikasi dan menghubungi saya. Saya memahami data digunakan untuk proses rekrutmen.",
    submit: "Kirim aplikasi", sending: "Mengirim…", success: "Terima kasih. Aplikasi Anda telah diterima.", failure: "Pengiriman gagal. Silakan coba kembali.", required: "Wajib", noticeTitle: "Transparansi tahap awal", notice: "NAEOS masih pre-incorporation dan pre-funding. Kolaborasi paruh waktu; saat ini belum ada gaji. Equity dapat dibahas tetapi tidak dijamin dan bergantung pada struktur legal serta perjanjian tertulis di masa depan. Belum ada saham yang diterbitkan atau dijamin. Trial hingga 30 hari memerlukan scope, deliverables, IP, confidentiality, dan batas kerja yang disepakati tertulis sebelumnya.", privacy: "Informasi digunakan hanya untuk menilai aplikasi dan menghubungi Anda terkait rekrutmen. Jangan sertakan informasi rahasia.",
    honeypot: "Kosongkan field ini",
  },
};
const inputStyle = { width: "100%", padding: "0.8rem", borderRadius: "8px", border: "1px solid var(--color-border-light)", background: "var(--color-bg-surface)", color: "var(--color-text)", font: "inherit" };
const labelStyle = { display: "block", fontWeight: 600, marginBottom: "0.45rem" };
const fieldStyle = { marginBottom: "1rem" };
const panelStyle = { padding: "clamp(1rem, 3vw, 2rem)", border: "1px solid var(--color-border)", borderRadius: "14px", background: "var(--color-bg-surface-alt)" };

export default function FoundingPartnerIntake({ lang }: Props) {
  const t = copy[lang];
  const [state, setState] = useState<"idle" | "sending" | "success" | "error">("idle");
  const [error, setError] = useState("");
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (state === "sending") return;
    setState("sending"); setError("");
    const formElement = event.currentTarget;
    const fd = new FormData(formElement);
    const payload: Record<string, unknown> = Object.fromEntries(fd.entries());
    payload.expertise_areas = fd.getAll("expertise_areas");
    payload.trial_available = fd.get("trial_available") === "on";
    payload.consent = fd.get("consent") === "on";
    try {
      const res = await fetch("https://kqmhpsftijskesjdovjm.supabase.co/functions/v1/founding-partner-intake", {
        method: "POST", headers: { "Content-Type": "application/json", "apikey": "sb_publishable_PkjsKuqVuy5BdjikdbNpNg_tsYL66Pq" }, body: JSON.stringify(payload),
      });
      const result = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(typeof result.error === "string" ? result.error : t.failure);
      setState("success");
      formElement.reset();
    } catch (err) { setState("error"); setError(err instanceof Error ? err.message : t.failure); }
  };
  const field = (name: string, label: string, opts: { required?: boolean; type?: string; placeholder?: string; maxLength?: number } = {}) => (
    <div style={fieldStyle}><label htmlFor={name} style={labelStyle}>{label}{opts.required ? " *" : ""}</label><input id={name} name={name} type={opts.type ?? "text"} required={opts.required} maxLength={opts.maxLength} placeholder={opts.placeholder} style={inputStyle}/></div>
  );
  const textarea = (name: string, label: string, required = false, rows = 4) => (
    <div style={fieldStyle}><label htmlFor={name} style={labelStyle}>{label}{required ? " *" : ""}</label><textarea id={name} name={name} required={required} minLength={required ? 20 : undefined} rows={rows} style={{ ...inputStyle, resize: "vertical" }}/></div>
  );
  return (
    <section className="container" style={{ maxWidth: "860px", padding: "2rem 1.25rem 5rem" }}>
      <div className="badge badge-green">NAEOS · THE INTELLIGENT CORE</div>
      <h1 style={{ fontSize: "clamp(2rem, 5vw, 3.5rem)", lineHeight: 1.1, marginTop: "1rem" }}>{t.title}</h1>
      <p style={{ color: "var(--color-text-muted)", marginBottom: "2rem" }}>{t.intro}</p>
      {state === "success" ? <div role="status" style={{ ...panelStyle, borderColor: "var(--color-success)" }}><h2>{t.success}</h2><p>Submission recorded for review. Submitted applications do not guarantee an interview or partnership.</p></div> : (
        <form onSubmit={submit}>
          <div style={panelStyle}><h2>{t.profile}</h2>
            {field("applicant_name", t.name, { required: true, maxLength: 120 })}
            {field("email", t.email, { required: true, type: "email", maxLength: 254 })}
            {field("location", t.location, { maxLength: 150 })}
            {field("linkedin_url", t.linkedin, { type: "url", placeholder: "https://", maxLength: 500 })}
            {field("portfolio_url", t.portfolio, { type: "url", placeholder: "https://", maxLength: 500 })}
            {field("professional_role", t.professional, { maxLength: 200 })}
            <div style={fieldStyle}><label htmlFor="primary_role" style={labelStyle}>{t.role} *</label><select id="primary_role" name="primary_role" required style={inputStyle}><option value="">—</option>{roles.map((r) => <option key={r} value={r}>{r}</option>)}</select></div>
            <div style={fieldStyle}><label htmlFor="alternate_role" style={labelStyle}>{t.alternate}</label><select id="alternate_role" name="alternate_role" style={inputStyle}><option value="">—</option>{roles.map((r) => <option key={r} value={r}>{r}</option>)}</select></div>
            <div style={fieldStyle}><label style={labelStyle}>{t.expertise}</label><div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit,minmax(220px,1fr))", gap: "0.5rem" }}>{areas.map((a) => <label key={a} style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}><input name="expertise_areas" type="checkbox" value={a}/>{a}</label>)}</div></div>
            {textarea("relevant_experience", t.experience, true)}
            {textarea("portfolio_summary", t.portfolioSummary, false, 3)}
            {textarea("contribution_proposal", t.contribution, true)}
            {textarea("motivation", t.motivation, true)}
          </div>
          <div style={{ ...panelStyle, marginTop: "1rem" }}><h2>{t.availability}</h2>
            <div style={fieldStyle}><label htmlFor="weekly_hours" style={labelStyle}>{t.hours} *</label><select id="weekly_hours" name="weekly_hours" required style={inputStyle}><option value="">—</option>{["2–5","5–10","10–15","15–20","20+"].map((v) => <option key={v}>{v}</option>)}</select></div>
            <div style={fieldStyle}><label htmlFor="collaboration_model" style={labelStyle}>{t.model} *</label><select id="collaboration_model" name="collaboration_model" required style={inputStyle}><option value="">—</option><option value="Part-time">{t.part}</option><option value="Flexible">{t.flexible}</option><option value="Other">{t.other}</option></select></div>
            <div style={fieldStyle}><label htmlFor="earliest_availability" style={labelStyle}>{t.date}</label><input id="earliest_availability" name="earliest_availability" type="date" style={inputStyle}/></div>
            <label style={{ display: "flex", alignItems: "start", gap: "0.6rem" }}><input type="checkbox" name="trial_available" required style={{ marginTop: "0.3rem" }}/><span>{t.trial} *</span></label>
          </div>
          <div style={{ ...panelStyle, marginTop: "1rem" }}><h2>{t.expectations}</h2>
            {textarea("partnership_expectations", t.partnership, true)}
            {textarea("equity_expectations", t.equity, false, 3)}
            <label style={{ display: "flex", alignItems: "start", gap: "0.6rem", marginTop: "1rem" }}><input type="checkbox" name="consent" required style={{ marginTop: "0.3rem" }}/><span>{t.consent} *</span></label>
          </div>
          <div style={{ ...panelStyle, marginTop: "1rem", borderColor: "var(--color-warning)" }}><h3>{t.noticeTitle}</h3><p>{t.notice}</p><small>{t.privacy}</small></div>
          <div aria-hidden="true" style={{ position: "absolute", left: "-10000px", height: "1px", width: "1px", overflow: "hidden" }}><label>{t.honeypot}<input name="website" tabIndex={-1} autoComplete="off"/></label></div>
          {state === "error" && <p role="alert" style={{ color: "var(--color-danger)" }}>{error || t.failure}</p>}
          <button type="submit" disabled={state === "sending"} className="btn btn-primary" style={{ marginTop: "1.25rem", padding: "0.9rem 1.5rem", opacity: state === "sending" ? 0.7 : 1 }}>{state === "sending" ? t.sending : t.submit}</button>
        </form>
      )}
    </section>
  );
}
