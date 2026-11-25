// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0
"use client";

import { useState, type FormEvent } from "react";

type Application = {
  id: string;
  applicant_name: string;
  email: string;
  location: string | null;
  linkedin_url: string | null;
  portfolio_url: string | null;
  professional_role: string | null;
  primary_role: string;
  alternate_role: string | null;
  expertise_areas: string[];
  relevant_experience: string;
  portfolio_summary: string | null;
  contribution_proposal: string;
  motivation: string;
  weekly_hours: string;
  collaboration_model: string;
  earliest_availability: string | null;
  partnership_expectations: string;
  equity_expectations: string | null;
  application_status: string;
  created_at: string;
};
type ReviewEvent = {
  id: string;
  reviewer_id: string;
  previous_status: string;
  new_status: string;
  note: string | null;
  created_at: string;
};
type RpcSession = { access_token: string; user: { id: string; email?: string } };
type Props = { lang: "en" | "id" };

const statuses = ["New Application", "Screening", "Shortlisted", "Interview", "Mutual Fit", "Trial", "Partnership Negotiation", "Accepted", "Rejected", "Withdrawn", "On Hold"];
const transitions: Record<string, string[]> = {
  "New Application": ["Screening", "Rejected", "Withdrawn", "On Hold"],
  Screening: ["Shortlisted", "Interview", "Rejected", "Withdrawn", "On Hold"],
  Shortlisted: ["Interview", "Mutual Fit", "Rejected", "Withdrawn", "On Hold"],
  Interview: ["Mutual Fit", "Trial", "Rejected", "Withdrawn", "On Hold"],
  "Mutual Fit": ["Trial", "Partnership Negotiation", "Rejected", "Withdrawn", "On Hold"],
  Trial: ["Partnership Negotiation", "Accepted", "Rejected", "Withdrawn", "On Hold"],
  "Partnership Negotiation": ["Accepted", "Rejected", "Withdrawn", "On Hold"],
  "On Hold": ["Screening", "Shortlisted", "Interview", "Mutual Fit", "Trial", "Partnership Negotiation", "Rejected", "Withdrawn"],
};
const copy = {
  en: {
    title: "Founding Partner Review", subtitle: "Restricted founder workspace. Sign in with an existing reviewer account.",
    email: "Reviewer email", password: "Password", signIn: "Sign in", signingIn: "Verifying…", signOut: "Sign out",
    config: "Dashboard is not configured. Set NEXT_PUBLIC_SUPABASE_URL and NEXT_PUBLIC_SUPABASE_ANON_KEY to the NAEOS staging project.",
    unauthorized: "This account is not enabled as a founding partner reviewer.", loginError: "Sign-in failed. Check your credentials and try again.",
    loading: "Loading applications…", applications: "Applications", all: "All statuses", empty: "No applications found.",
    open: "Open review", applicant: "Applicant", role: "Primary role", status: "Status", submitted: "Submitted",
    details: "Application details", profile: "Profile", experience: "Relevant experience", proposal: "Contribution proposal",
    motivation: "Motivation", expectations: "Partnership expectations", expertise: "Expertise", availability: "Availability",
    hours: "Hours / week", collaboration: "Collaboration", history: "Review history", noHistory: "No review events yet.",
    nextStatus: "Move to status", note: "Review note (optional, max 1,000 characters)", save: "Record transition",
    saving: "Saving…", saved: "Transition recorded and audit event appended.", previous: "Previous", next: "Next",
    page: "Page", of: "of", session: "Session is held in memory only and ends when this page is refreshed.",
    unavailable: "Could not load dashboard data. Your session may have expired.", genericError: "Request failed. No changes were confirmed.",
  },
  id: {
    title: "Review Founding Partner", subtitle: "Workspace founder terbatas. Masuk dengan akun reviewer yang sudah dibuat.",
    email: "Email reviewer", password: "Password", signIn: "Masuk", signingIn: "Memverifikasi…", signOut: "Keluar",
    config: "Dashboard belum dikonfigurasi. Isi NEXT_PUBLIC_SUPABASE_URL dan NEXT_PUBLIC_SUPABASE_ANON_KEY dengan nilai project staging NAEOS.",
    unauthorized: "Akun ini belum diaktifkan sebagai reviewer founding partner.", loginError: "Login gagal. Periksa kredensial lalu coba lagi.",
    loading: "Memuat aplikasi…", applications: "Daftar aplikasi", all: "Semua status", empty: "Tidak ada aplikasi.",
    open: "Buka review", applicant: "Pelamar", role: "Peran utama", status: "Status", submitted: "Dikirim",
    details: "Detail aplikasi", profile: "Profil", experience: "Pengalaman relevan", proposal: "Proposal kontribusi",
    motivation: "Motivasi", expectations: "Ekspektasi kemitraan", expertise: "Keahlian", availability: "Ketersediaan",
    hours: "Jam / minggu", collaboration: "Kolaborasi", history: "Riwayat review", noHistory: "Belum ada event review.",
    nextStatus: "Ubah status ke", note: "Catatan review (opsional, maksimal 1.000 karakter)", save: "Simpan transisi",
    saving: "Menyimpan…", saved: "Transisi tersimpan dan audit event ditambahkan.", previous: "Sebelumnya", next: "Berikutnya",
    page: "Halaman", of: "dari", session: "Sesi hanya disimpan di memori dan berakhir saat halaman dimuat ulang.",
    unavailable: "Data dashboard gagal dimuat. Sesi mungkin telah berakhir.", genericError: "Permintaan gagal. Perubahan belum dikonfirmasi.",
  },
};

export default function FoundingPartnerReviewDashboard({ lang }: Props) {
  const t = copy[lang];
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL?.replace(/\/$/, "") ?? "";
  const anonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY ?? "";
  const configured = Boolean(supabaseUrl && anonKey);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [session, setSession] = useState<RpcSession | null>(null);
  const [reviewerName, setReviewerName] = useState("");
  const [items, setItems] = useState<Application[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [filter, setFilter] = useState("");
  const [selected, setSelected] = useState<Application | null>(null);
  const [events, setEvents] = useState<ReviewEvent[]>([]);
  const [nextStatus, setNextStatus] = useState("");
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const pageSize = 20;

  async function rpc<T>(name: string, args: Record<string, unknown> = {}): Promise<T> {
    if (!session) throw new Error("Reviewer session missing");
    const response = await fetch(`${supabaseUrl}/rest/v1/rpc/${name}`, {
      method: "POST",
      headers: { apikey: anonKey, Authorization: `Bearer ${session.access_token}`, "Content-Type": "application/json" },
      body: JSON.stringify(args),
      cache: "no-store",
    });
    const result = await response.json().catch(() => null);
    if (!response.ok) throw new Error(typeof result?.message === "string" ? result.message : t.genericError);
    return result as T;
  }

  async function loadApplications(activeSession: RpcSession, nextOffset = offset, nextFilter = filter) {
    const response = await fetch(`${supabaseUrl}/rest/v1/rpc/founding_partner_review_list`, {
      method: "POST",
      headers: { apikey: anonKey, Authorization: `Bearer ${activeSession.access_token}`, "Content-Type": "application/json" },
      body: JSON.stringify({ p_limit: pageSize, p_offset: nextOffset, p_status: nextFilter || null }),
      cache: "no-store",
    });
    const result = await response.json().catch(() => null);
    if (!response.ok) throw new Error(typeof result?.message === "string" ? result.message : t.unavailable);
    setItems(Array.isArray(result?.items) ? result.items : []);
    setTotal(Number(result?.total ?? 0));
    setOffset(nextOffset);
  }

  async function signIn(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!configured || busy) return;
    setBusy(true); setError(""); setNotice("");
    try {
      const response = await fetch(`${supabaseUrl}/auth/v1/token?grant_type=password`, {
        method: "POST",
        headers: { apikey: anonKey, "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
        cache: "no-store",
      });
      const result = await response.json().catch(() => null);
      if (!response.ok || typeof result?.access_token !== "string") throw new Error(t.loginError);
      const candidate: RpcSession = { access_token: result.access_token, user: result.user };
      const contextResponse = await fetch(`${supabaseUrl}/rest/v1/rpc/founding_partner_review_context`, {
        method: "POST",
        headers: { apikey: anonKey, Authorization: `Bearer ${candidate.access_token}`, "Content-Type": "application/json" },
        body: "{}",
        cache: "no-store",
      });
      const context = await contextResponse.json().catch(() => null);
      if (!contextResponse.ok) throw new Error(t.unauthorized);
      setSession(candidate);
      setReviewerName(typeof context?.display_name === "string" && context.display_name ? context.display_name : candidate.user.email ?? "");
      const listResponse = await fetch(`${supabaseUrl}/rest/v1/rpc/founding_partner_review_list`, {
        method: "POST",
        headers: { apikey: anonKey, Authorization: `Bearer ${candidate.access_token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ p_limit: pageSize, p_offset: 0, p_status: null }),
        cache: "no-store",
      });
      const list = await listResponse.json().catch(() => null);
      if (!listResponse.ok) throw new Error(t.unavailable);
      setItems(Array.isArray(list?.items) ? list.items : []);
      setTotal(Number(list?.total ?? 0)); setOffset(0); setPassword("");
    } catch (err) {
      setSession(null); setError(err instanceof Error ? err.message : t.loginError);
    } finally { setBusy(false); }
  }

  async function refresh(nextOffset = offset, nextFilter = filter) {
    setBusy(true); setError(""); setNotice("");
    try { await loadApplications(session!, nextOffset, nextFilter); }
    catch (err) { setError(err instanceof Error ? err.message : t.unavailable); }
    finally { setBusy(false); }
  }

  async function openApplication(application: Application) {
    setBusy(true); setError(""); setNotice("");
    try {
      const result = await rpc<{ application: Application; events: ReviewEvent[] }>("founding_partner_review_detail", { p_application_id: application.id });
      setSelected(result.application); setEvents(result.events ?? []);
      setNextStatus(transitions[result.application.application_status]?.[0] ?? "");
      setNote("");
    } catch (err) { setError(err instanceof Error ? err.message : t.unavailable); }
    finally { setBusy(false); }
  }

  async function recordTransition(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selected || !nextStatus || busy) return;
    setBusy(true); setError(""); setNotice("");
    try {
      await rpc<boolean>("founding_partner_review_transition", {
        p_application_id: selected.id, p_new_status: nextStatus, p_note: note,
      });
      const result = await rpc<{ application: Application; events: ReviewEvent[] }>("founding_partner_review_detail", { p_application_id: selected.id });
      setSelected(result.application); setEvents(result.events ?? []);
      setNextStatus(transitions[result.application.application_status]?.[0] ?? "");
      setNote(""); setNotice(t.saved);
      await loadApplications(session!, offset, filter);
    } catch (err) { setError(err instanceof Error ? err.message : t.genericError); }
    finally { setBusy(false); }
  }

  function signOut() {
    setSession(null); setReviewerName(""); setItems([]); setTotal(0); setSelected(null); setEvents([]);
    setPassword(""); setError(""); setNotice("");
  }

  const field = (label: string, value?: string | null) => value ? (
    <div style={{ marginBottom: "1rem" }}><dt style={{ fontWeight: 700, marginBottom: ".25rem" }}>{label}</dt><dd style={{ margin: 0, whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>{value}</dd></div>
  ) : null;
  const panel: React.CSSProperties = { border: "1px solid var(--color-border)", borderRadius: "12px", padding: "1.25rem", background: "var(--color-bg-surface-alt)" };
  const control: React.CSSProperties = { width: "100%", padding: ".75rem", borderRadius: "8px", border: "1px solid var(--color-border-light)", background: "var(--color-bg-surface)", color: "var(--color-text)", font: "inherit" };
  const button: React.CSSProperties = { padding: ".7rem 1rem", borderRadius: "8px", border: "1px solid var(--color-border)", cursor: busy ? "wait" : "pointer", font: "inherit" };

  return (
    <section className="container" style={{ maxWidth: "1180px", padding: "2rem 1.25rem 5rem" }}>
      <div className="badge badge-green">NAEOS · PRIVATE REVIEW</div>
      <h1 style={{ fontSize: "clamp(2rem, 5vw, 3rem)", marginTop: "1rem" }}>{t.title}</h1>
      <p style={{ color: "var(--color-text-muted)", marginBottom: "2rem" }}>{t.subtitle}</p>
      {!configured ? <div role="alert" style={panel}>{t.config}</div> : !session ? (
        <form onSubmit={signIn} style={{ ...panel, maxWidth: "520px" }}>
          <label style={{ display: "block", fontWeight: 600, marginBottom: ".4rem" }} htmlFor="review-email">{t.email}</label>
          <input id="review-email" type="email" autoComplete="username" required value={email} onChange={(e) => setEmail(e.target.value)} style={{ ...control, marginBottom: "1rem" }}/>
          <label style={{ display: "block", fontWeight: 600, marginBottom: ".4rem" }} htmlFor="review-password">{t.password}</label>
          <input id="review-password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} style={{ ...control, marginBottom: "1rem" }}/>
          {error && <p role="alert" style={{ color: "var(--color-danger)" }}>{error}</p>}
          <button type="submit" className="btn btn-primary" disabled={busy} style={button}>{busy ? t.signingIn : t.signIn}</button>
          <p style={{ color: "var(--color-text-muted)", fontSize: ".85rem", marginTop: "1rem" }}>{t.session}</p>
        </form>
      ) : (
        <>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "1rem", flexWrap: "wrap", marginBottom: "1.25rem" }}>
            <div><strong>{reviewerName}</strong><div style={{ color: "var(--color-text-muted)", fontSize: ".85rem" }}>{session.user.email}</div></div>
            <button type="button" onClick={signOut} style={button}>{t.signOut}</button>
          </div>
          {error && <p role="alert" style={{ color: "var(--color-danger)" }}>{error}</p>}
          {notice && <p role="status" style={{ color: "var(--color-success)" }}>{notice}</p>}
          <div style={{ display: "grid", gridTemplateColumns: "minmax(280px, .9fr) minmax(0, 1.4fr)", gap: "1rem", alignItems: "start" }}>
            <section style={panel}>
              <h2>{t.applications} ({total})</h2>
              <label htmlFor="status-filter" style={{ display: "block", fontWeight: 600, margin: "1rem 0 .4rem" }}>{t.status}</label>
              <select id="status-filter" value={filter} onChange={(e) => { setFilter(e.target.value); void refresh(0, e.target.value); }} style={control}>
                <option value="">{t.all}</option>{statuses.map((status) => <option key={status} value={status}>{status}</option>)}
              </select>
              {busy && <p>{t.loading}</p>}
              {!busy && items.length === 0 && <p>{t.empty}</p>}
              <div style={{ display: "grid", gap: ".65rem", marginTop: "1rem" }}>
                {items.map((application) => (
                  <button key={application.id} type="button" onClick={() => void openApplication(application)} style={{ ...button, textAlign: "left", background: selected?.id === application.id ? "var(--color-bg-surface)" : "transparent", borderColor: selected?.id === application.id ? "var(--color-accent)" : "var(--color-border)" }}>
                    <strong>{application.applicant_name}</strong>
                    <span style={{ display: "block", fontSize: ".86rem", marginTop: ".2rem" }}>{application.primary_role}</span>
                    <span style={{ display: "block", color: "var(--color-text-muted)", fontSize: ".8rem", marginTop: ".25rem" }}>{application.application_status} · {new Date(application.created_at).toLocaleDateString(lang === "id" ? "id-ID" : "en-US")}</span>
                  </button>
                ))}
              </div>
              <div style={{ display: "flex", justifyContent: "space-between", gap: ".5rem", marginTop: "1rem", alignItems: "center" }}>
                <button type="button" disabled={offset === 0 || busy} onClick={() => void refresh(Math.max(0, offset - pageSize))} style={button}>{t.previous}</button>
                <small>{t.page} {Math.floor(offset / pageSize) + 1} {t.of} {Math.max(1, Math.ceil(total / pageSize))}</small>
                <button type="button" disabled={offset + pageSize >= total || busy} onClick={() => void refresh(offset + pageSize)} style={button}>{t.next}</button>
              </div>
            </section>
            <section style={panel}>
              {!selected ? <p>{t.open}</p> : <>
                <h2>{t.details}: {selected.applicant_name}</h2>
                <p><strong>{selected.application_status}</strong> · {selected.primary_role}</p>
                <h3>{t.profile}</h3>
                <dl>
                  {field(t.applicant, selected.applicant_name)}
                  {field(t.email, selected.email)}
                  {field("Location", selected.location)}
                  {field("LinkedIn", selected.linkedin_url)}
                  {field("Portfolio", selected.portfolio_url)}
                  {field("Professional role", selected.professional_role)}
                  {field(t.expertise, selected.expertise_areas?.join(", "))}
                </dl>
                <h3>{t.experience}</h3><p style={{ whiteSpace: "pre-wrap" }}>{selected.relevant_experience}</p>
                {selected.portfolio_summary && <><h3>Portfolio</h3><p style={{ whiteSpace: "pre-wrap" }}>{selected.portfolio_summary}</p></>}
                <h3>{t.proposal}</h3><p style={{ whiteSpace: "pre-wrap" }}>{selected.contribution_proposal}</p>
                <h3>{t.motivation}</h3><p style={{ whiteSpace: "pre-wrap" }}>{selected.motivation}</p>
                <h3>{t.availability}</h3><p>{t.hours}: {selected.weekly_hours} · {t.collaboration}: {selected.collaboration_model}</p>
                {field("Earliest availability", selected.earliest_availability)}
                <h3>{t.expectations}</h3><p style={{ whiteSpace: "pre-wrap" }}>{selected.partnership_expectations}</p>
                {field("Equity expectations", selected.equity_expectations)}
                <h3>{t.history}</h3>
                {events.length === 0 ? <p>{t.noHistory}</p> : <ol>{events.map((item) => <li key={item.id} style={{ marginBottom: ".6rem" }}><strong>{item.previous_status} → {item.new_status}</strong><div style={{ fontSize: ".85rem", color: "var(--color-text-muted)" }}>{new Date(item.created_at).toLocaleString(lang === "id" ? "id-ID" : "en-US")}</div>{item.note && <p>{item.note}</p>}</li>)}</ol>}
                {(transitions[selected.application_status] ?? []).length > 0 && (
                  <form onSubmit={recordTransition} style={{ borderTop: "1px solid var(--color-border)", paddingTop: "1rem", marginTop: "1rem" }}>
                    <h3>{t.nextStatus}</h3>
                    <select required value={nextStatus} onChange={(e) => setNextStatus(e.target.value)} style={{ ...control, marginBottom: ".75rem" }}>
                      {(transitions[selected.application_status] ?? []).map((status) => <option key={status} value={status}>{status}</option>)}
                    </select>
                    <label htmlFor="review-note" style={{ display: "block", fontWeight: 600, marginBottom: ".4rem" }}>{t.note}</label>
                    <textarea id="review-note" maxLength={1000} rows={3} value={note} onChange={(e) => setNote(e.target.value)} style={{ ...control, marginBottom: ".75rem", resize: "vertical" }}/>
                    <button type="submit" disabled={busy || !nextStatus} className="btn btn-primary" style={button}>{busy ? t.saving : t.save}</button>
                  </form>
                )}
              </>}
            </section>
          </div>
        </>
      )}
    </section>
  );
}
