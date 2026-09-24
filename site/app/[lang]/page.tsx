// Copyright 2024-2026 NAEOS Foundation
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
  ["Specification", "Define engineering intent."],
  ["NEIR", "Canonical engineering model."],
  ["Policy", "Validate boundaries."],
  ["AI Context", "Prepare relevant context."],
  ["Agent", "Perform bounded work."],
  ["Execution", "Run through engineering workflows."],
  ["Evidence", "Trace results back to intent."],
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
              ? "Kendalikan bagaimana AI membangun software."
              : "Control how AI builds software."}
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
              ? "NAEOS memberi engineering team control plane untuk workflow AI yang tervalidasi, governed, executable, dan dapat dibuktikan."
              : "NAEOS gives engineering teams a control plane for validated, governed, executable and provable AI workflows."}
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
              Get started
            </Link>
            <Link href={base + "/docs/architecture"} className="btn btn-secondary btn-lg">
              View architecture
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
                gridTemplateColumns: "repeat(7, minmax(110px, 1fr))",
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
                      {desc}
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
                <p>Engineering state remains explicit and machine-readable.</p>
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
              ["Policy", "Define what is allowed, required and forbidden."],
              ["Context", "Give agents structured, relevant engineering context."],
              ["Evidence", "Keep decisions, artifacts and results traceable."],
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
            GitHub · CI/CD · repositories ↔ NAEOS ↔ Copilot · Claude · Codex · agents
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
