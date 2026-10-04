// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import type { Metadata } from "next";
import Link from "next/link";
import { getPage } from "@/lib/content";
import { LANGUAGES, DEFAULT_LANG, SITE, type Lang } from "@/lib/site";
import { pageMetadata } from "@/lib/metadata";

export function generateStaticParams() {
  return LANGUAGES.map((lang) => ({ lang }));
}

export async function generateMetadata(
  props: { params: Promise<{ lang: string }> },
): Promise<Metadata> {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw)
    ? (raw as Lang)
    : DEFAULT_LANG;
  return pageMetadata(getPage("/", lang), lang);
}

const stages = [
  ["Specification", "Define engineering intent.", "Tentukan intent engineering."],
  ["Policy", "Decide what is allowed.", "Tentukan apa yang diizinkan."],
  ["Authorization", "Bind execution to capabilities.", "Ikat execution pada capability."],
  ["Runtime", "Execute inside the boundary.", "Jalankan di dalam boundary."],
  ["Evidence", "Capture durable observations.", "Catat evidence yang tahan lama."],
  ["Verification", "Independently verify outcomes.", "Verifikasi hasil secara independen."],
] as const;

export default async function HomePage(
  props: { params: Promise<{ lang: string }> },
) {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw)
    ? (raw as Lang)
    : DEFAULT_LANG;
  const base = lang === "en" ? "" : "/id";
  const id = lang === "id";
  const stageDescriptions = stages.map(([, en, idDesc]) => (id ? idDesc : en));

  return (
    <>
      <section className="hero">
        <div className="hero-bg" />
        <div className="hero-grid" />
        <div
          className="container"
          style={{
            paddingTop: "5rem",
            paddingBottom: "5rem",
            position: "relative",
          }}
        >
          <span className="badge badge-green">AI ENGINEERING CONTROL PLANE</span>
          <h1
            style={{
              maxWidth: "900px",
              fontSize: "clamp(3rem, 7vw, 5.5rem)",
              letterSpacing: "-0.04em",
              marginTop: "1.25rem",
            }}
          >
            {id
              ? "Control plane untuk AI coding agents."
              : "The Engineering Control Plane for AI Coding Agents."}
          </h1>
          <p
            style={{
              maxWidth: "760px",
              marginTop: "1.5rem",
              fontSize: "1.25rem",
              color: "var(--color-text-muted)",
            }}
          >
            {id
              ? "Bangun software dengan AI di bawah kontrol architecture, policy, execution, evidence, dan verification yang eksplisit."
              : "Build software with AI under explicit architecture, policy, execution, evidence, and verification controls."}
          </p>
          <div
            style={{
              display: "flex",
              gap: ".75rem",
              flexWrap: "wrap",
              marginTop: "2rem",
            }}
          >
            <Link href={base + "/docs/getting-started"} className="btn btn-primary btn-lg">
              {id ? "Mulai dengan NAEOS" : "Explore NAEOS"}
            </Link>
            <Link href={base + "/enterprise"} className="btn btn-secondary btn-lg">
              {id ? "Mulai pilot" : "Start a governance pilot"}
            </Link>
            <a
              href={SITE.repo}
              className="btn btn-secondary btn-lg"
              target="_blank"
              rel="noopener"
            >
              GitHub
            </a>
          </div>
          <div
            className="card"
            style={{
              marginTop: "4rem",
              padding: "2rem",
              background: "rgba(10,10,20,.82)",
              boxShadow: "var(--shadow-glow)",
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: ".75rem",
                color: "var(--color-text-dim)",
                marginBottom: "1.5rem",
              }}
            >
              NAEOS CONTROL PLANE
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(6, minmax(125px, 1fr))",
                gap: ".5rem",
                overflowX: "auto",
              }}
            >
              {stages.map(([name, desc], i) => (
                <div key={name} style={{ minWidth: "110px" }}>
                  <div
                    style={{
                      minHeight: "120px",
                      padding: "1rem",
                      border:
                        i === 1
                          ? "1px solid var(--color-accent)"
                          : "1px solid var(--color-border-light)",
                      borderRadius: "var(--radius-md)",
                      background:
                        i === 1
                          ? "var(--color-accent-glow)"
                          : "var(--color-bg-surface)",
                    }}
                  >
                    <small
                      style={{
                        color: "var(--color-accent)",
                        fontFamily: "var(--font-mono)",
                      }}
                    >
                      0{i + 1}
                    </small>
                    <strong style={{ display: "block", marginTop: ".6rem" }}>
                      {name}
                    </strong>
                    <span
                      style={{
                        display: "block",
                        marginTop: ".4rem",
                        color: "var(--color-text-muted)",
                        fontSize: ".72rem",
                      }}
                    >
                      {stageDescriptions[i]}
                    </span>
                  </div>
                  {i < stages.length - 1 && (
                    <div
                      style={{
                        textAlign: "center",
                        color: "var(--color-text-dim)",
                        paddingTop: ".4rem",
                      }}
                    >
                      →
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="problem-grid">
            <div className="problem-card">
              <span className="problem-tag problem-tag-green">{id ? "Govern" : "Govern"}</span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id ? "Tetapkan boundary sebelum agent bertindak." : "Set the boundary before the agent acts."}
              </h2>
              <p>
                {id
                  ? "Policy dan authorization menentukan capability yang tersedia dan kondisi eksekusinya."
                  : "Policy and authorization determine which capabilities are available and under what conditions they may execute."}
              </p>
            </div>
            <div className="problem-card">
              <span className="problem-tag">{id ? "Verify" : "Verify"}</span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id ? "Jangan samakan execution dengan keberhasilan." : "Do not confuse execution with success."}
              </h2>
              <p>
                {id
                  ? "NAEOS memisahkan runtime execution dari evidence dan independent verification."
                  : "NAEOS separates runtime execution from durable evidence and independent verification."}
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <h2 className="section-title">
            {id ? "AI harus di-govern, bukan hanya di-prompt." : "AI must be governed, not just prompted."}
          </h2>
          <p className="section-subtitle">
            {id
              ? "NAEOS mengubah AI coding dari assistant tanpa batas menjadi workflow engineering yang dapat dikontrol."
              : "NAEOS turns AI coding from an unbounded assistant into a controlled engineering workflow."}
          </p>
          <div className="features-grid">
            {[
              ["Architecture", id ? "Tetapkan constraint dan boundary engineering." : "Define engineering constraints and system boundaries."],
              ["Policy", id ? "Tentukan action yang allowed, restricted, atau prohibited." : "Determine which actions are allowed, restricted, or prohibited."],
              ["Evidence", id ? "Simpan keputusan, observasi, artifact, dan receipt." : "Preserve decisions, observations, artifacts, and receipts."],
              ["Verification", id ? "Evaluasi outcome secara independen dari execution." : "Evaluate outcomes independently from execution."],
            ].map(([title, desc]) => (
              <div className="feature-card" key={title}>
                <h3>{title}</h3>
                <p>{desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="cta-band">
            <h2>{id ? "Uji control plane pada workflow engineering Anda." : "Test the control plane on your engineering workflow."}</h2>
            <p>
              {id
                ? "Mulai dengan 30-Day AI Engineering Governance Pilot."
                : "Start with a scoped 30-Day AI Engineering Governance Pilot."}
            </p>
            <div className="cta-band-actions">
              <Link href={base + "/enterprise"} className="btn btn-primary btn-lg">
                {id ? "Mulai pilot" : "Start a governance pilot"}
              </Link>
              <Link href={base + "/control-plane"} className="btn btn-secondary btn-lg">
                {id ? "Lihat control plane" : "View the control plane"}
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="problem-grid">
            <div className="problem-card problem-card-before">
              <span className="problem-tag problem-tag-red">
                {id ? "Tanpa control plane" : "Without a control plane"}
              </span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id
                  ? "AI dapat menghasilkan kode, tetapi workflow engineering tetap sulit dikendalikan."
                  : "AI can generate code, but the engineering workflow remains hard to control."}
              </h2>
              <ul className="problem-list">
                <li>Context drift across agents and repositories</li>
                <li>Inconsistent validation and policy enforcement</li>
                <li>Weak traceability from intent to artifact</li>
              </ul>
            </div>
            <div className="problem-card problem-card-after">
              <span className="problem-tag problem-tag-green">
                {id ? "Dengan NAEOS" : "With NAEOS"}
              </span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id
                  ? "Intent, execution, governance, dan evidence menjadi satu sistem."
                  : "Intent, execution, governance and evidence become one system."}
              </h2>
              <p>
                {id
                  ? "NAEOS berada di antara engineering system dan AI agents, memberi struktur dan batasan pada setiap tahap."
                  : "NAEOS sits between the engineering system and AI agents, adding structure and boundaries at every stage."}
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="problem-grid">
            <div className="problem-card">
              <span className="problem-tag problem-tag-green">{id ? "Bukti" : "Proof"}</span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id ? "Authorization dipisahkan dari execution." : "Authorization is separated from execution."}
              </h2>
              <p>
                {id
                  ? "Aksi production dapat diusulkan agent, dievaluasi policy, diblok sebelum side effect, lalu dihubungkan ke evidence."
                  : "A production action can be proposed by an agent, evaluated by policy, blocked before side effect, and linked to evidence."}
              </p>
            </div>
            <div className="problem-card">
              <span className="problem-tag">{id ? "Contoh" : "Example"}</span>
              <pre style={{ marginTop: "1rem", overflowX: "auto" }}>
                <code>{"request: agent.prod.delete_database\npolicy: production-safety/v2\ndecision: DENY\nexecution: blocked\nevidence: decision_id → receipt"}</code>
              </pre>
            </div>
          </div>
          <div style={{ marginTop: "1.5rem", textAlign: "center" }}>
            <Link href={base + "/control-plane"} className="btn btn-secondary">
              {id ? "Buka live demo control plane" : "Open the live control-plane demo"}
            </Link>
          </div>
        </div>
      </section>

      <section className="section section-pipeline">
        <div className="container">
          <h2 className="section-title">
            {id
              ? "NEIR adalah pusat model engineering"
              : "NEIR is the engineering model at the center"}
          </h2>
          <p className="section-subtitle">
            {id
              ? "Satu representasi canonical menghubungkan spesifikasi dengan validation, context, agents, dan execution."
              : "One canonical representation connects specifications to validation, context, agents and execution."}
          </p>
          <div className="how-grid">
            {[
              ["01", "Specification", "intent → architecture → constraints"],
              ["02", "NEIR", "modules → services → dependencies → graph"],
              ["03", "Policy", "validate → allow / block"],
            ].map(([step, title, code]) => (
              <div className="how-card" key={title}>
                <div className="how-step">{step}</div>
                <h3>{title}</h3>
                <p>{id ? "State engineering tetap eksplisit dan machine-readable." : "Engineering state remains explicit and machine-readable."}</p>
                <pre style={{ marginTop: "1rem" }}>
                  <code>{code}</code>
                </pre>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <h2 className="section-title">
            {id ? "Governance sebelum execution" : "Governance before execution"}
          </h2>
          <p className="section-subtitle">
            {id
              ? "AI tidak mendapat jalur tanpa batas ke engineering workflow."
              : "AI does not get an unbounded path into the engineering workflow."}
          </p>
          <div className="features-grid">
            {[
              ["Policy", id ? "Tentukan apa yang diizinkan, diwajibkan, dan dilarang." : "Define what is allowed, required and forbidden."],
              ["Context", id ? "Berikan context engineering yang terstruktur dan relevan kepada agent." : "Give agents structured, relevant engineering context."],
              ["Evidence", id ? "Jaga agar keputusan, artifact, dan hasil tetap dapat ditelusuri." : "Keep decisions, artifacts and results traceable."],
            ].map(([title, desc]) => (
              <div className="feature-card" key={title}>
                <h3>{title}</h3>
                <p>{desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <h2 className="section-title">
            {id
              ? "NAEOS di antara engineering stack dan AI"
              : "NAEOS between your engineering stack and AI"}
          </h2>
          <p className="section-subtitle">
            {id ? "GitHub · CI/CD · repository ↔ NAEOS ↔ Copilot · Claude · Codex · agents" : "GitHub · CI/CD · repositories ↔ NAEOS ↔ Copilot · Claude · Codex · agents"}
          </p>
          <div
            className="card"
            style={{
              padding: "2rem",
              textAlign: "center",
              fontFamily: "var(--font-mono)",
            }}
          >
            <strong>Engineering</strong>
            <span style={{ margin: "0 1rem" }}>↔</span>
            <strong style={{ color: "var(--color-accent)" }}>
              NAEOS / NEIR / Policy / Evidence
            </strong>
            <span style={{ margin: "0 1rem" }}>↔</span>
            <strong>AI Agents</strong>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <h2 className="section-title">
            {id ? "Dari intent ke evidence" : "From intent to evidence"}
          </h2>
          <p className="section-subtitle">
            {id
              ? "Setiap tahap dapat ditelusuri kembali ke intent awal."
              : "Every stage remains traceable back to the original engineering intent."}
          </p>
          <div className="pipeline-strip">
            {["Intent", "Decision", "Execution", "Artifact", "Evidence"].map(
              (item, i) => (
                <div key={item} style={{ display: "contents" }}>
                  <div className="pipeline-stage">
                    <div className="pipeline-stage-index">{i + 1}</div>
                    <div className="pipeline-stage-name">{item}</div>
                  </div>
                  {i < 4 && <div className="pipeline-arrow">→</div>}
                </div>
              ),
            )}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="cta-band">
            <h2>{id ? "Mulai dari specification." : "Start with a specification."}</h2>
            <p>
              {id
                ? "Build dengan control, policy, dan evidence."
                : "Build with control, policy and evidence."}
            </p>
            <div className="code-block">
              <div className="code-block-header">
                <span>bash</span>
              </div>
              <pre>
                <code>
                  {"naeos init\nnaeos validate --input-file spec.yaml\nnaeos run --config naeos.yaml --input-file spec.yaml"}
                </code>
              </pre>
            </div>
            <div className="cta-band-actions">
              <Link href={base + "/docs/getting-started"} className="btn btn-primary btn-lg">
                Get started
              </Link>
              <Link href={base + "/docs/architecture"} className="btn btn-secondary btn-lg">
                Architecture
              </Link>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
