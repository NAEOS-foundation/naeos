// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import type { Metadata } from "next";
import Link from "next/link";
import { LANGUAGES, DEFAULT_LANG, SITE, type Lang } from "@/lib/site";
import ControlPlaneLiveDemo from "@/components/control-plane/ControlPlaneLiveDemo";

export function generateStaticParams() {
  return LANGUAGES.map((lang) => ({ lang }));
}

export async function generateMetadata(props: { params: Promise<{ lang: string }> }): Promise<Metadata> {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw) ? (raw as Lang) : DEFAULT_LANG;
  const id = lang === "id";
  return {
    title: "Control Plane — NAEOS",
    description: id
      ? "Bagaimana NAEOS menghubungkan intent, specification, policy, runtime, execution, dan evidence."
      : "How NAEOS connects intent, specification, policy, runtime, execution, and evidence.",
  };
}

type Step = { id: string; title: string; description: string };

export default async function ControlPlanePage(props: { params: Promise<{ lang: string }> }) {
  const { lang: raw } = await props.params;
  const lang = (LANGUAGES as readonly string[]).includes(raw) ? (raw as Lang) : DEFAULT_LANG;
  const id = lang === "id";
  const base = id ? "/id" : "";

  const steps: Step[] = id
    ? [
        { id: "01", title: "Intent", description: "Engineer mendefinisikan tujuan dan batas perubahan." },
        { id: "02", title: "Specification", description: "Intent menjadi spesifikasi yang dapat divalidasi." },
        { id: "03", title: "NEIR", description: "NAEOS membangun representasi engineering yang terstruktur." },
        { id: "04", title: "Policy", description: "Policy menentukan apakah aksi diizinkan, ditolak, atau membutuhkan approval." },
        { id: "05", title: "Runtime", description: "Runtime mengeksekusi keputusan melalui capability yang terotorisasi." },
        { id: "06", title: "Execution", description: "Side effect terjadi pada sistem target." },
        { id: "07", title: "Evidence", description: "Observation dan receipt menjadi bukti yang dapat diaudit." },
      ]
    : [
        { id: "01", title: "Intent", description: "The engineer defines the goal and change boundaries." },
        { id: "02", title: "Specification", description: "Intent becomes a specification that can be validated." },
        { id: "03", title: "NEIR", description: "NAEOS builds a structured engineering representation." },
        { id: "04", title: "Policy", description: "Policy decides whether an action is allowed, denied, or requires approval." },
        { id: "05", title: "Runtime", description: "Runtime executes the decision through authorized capabilities." },
        { id: "06", title: "Execution", description: "The side effect occurs on the target system." },
        { id: "07", title: "Evidence", description: "Observation and receipts become auditable evidence." },
      ];

  return (
    <div className="control-plane-page">
      <section className="control-plane-hero">
        <div className="eyebrow">ENGINEERING CONTROL PLANE</div>
        <h1>{id ? "AI boleh mengusulkan." : "AI can propose."}<br />{id ? "Engineering tetap dikendalikan." : "Engineering stays governed."}</h1>
        <p>
          {id
            ? "NAEOS memisahkan intent, keputusan policy, eksekusi runtime, dan observasi sehingga setiap aksi memiliki batas dan evidence."
            : "NAEOS separates intent, policy decisions, runtime execution, and observation so every action has boundaries and evidence."}
        </p>
        <div className="control-plane-actions">
          <Link href={`${base}/docs/architecture`} className="btn btn-primary btn-lg">
            {id ? "Lihat Arsitektur" : "Explore Architecture"}
          </Link>
          <a href={SITE.repo} className="btn btn-secondary btn-lg" target="_blank" rel="noopener">
            {id ? "Lihat GitHub" : "View GitHub"}
          </a>
        </div>
      </section>

      <section className="control-plane-flow" aria-labelledby="control-plane-flow-title">
        <div className="section-heading">
          <div className="eyebrow">THE CONTROL LOOP</div>
          <h2 id="control-plane-flow-title">{id ? "Dari intent sampai evidence." : "From intent to evidence."}</h2>
        </div>
        <div className="control-plane-grid">
          {steps.map((step, index) => (
            <article className="control-plane-step" key={step.id}>
              <span className="control-plane-step-id">{step.id}</span>
              <h3>{step.title}</h3>
              <p>{step.description}</p>
              {index < steps.length - 1 && <span className="control-plane-arrow" aria-hidden="true">→</span>}
            </article>
          ))}
        </div>
      </section>

      <section className="control-plane-demo" aria-labelledby="control-plane-demo-title">
        <div className="section-heading">
          <div className="eyebrow">30-SECOND DEMO</div>
          <h2 id="control-plane-demo-title">{id ? "Model mengusulkan. Policy yang menentukan." : "The model proposes. Policy decides."}</h2>
          <p>{id ? "Contoh sederhana bagaimana keputusan dipisahkan dari eksekusi." : "A simple example of separating authorization from execution."}</p>
        </div>
        <div className="control-plane-demo-grid">
          <div className="control-plane-terminal">
            <div className="terminal-label">REQUEST</div>
            <code>agent.prod.delete_database</code>
            <pre>{'{\n  "environment": "production",\n  "resource": "database",\n  "action": "delete"\n}'}</pre>
          </div>
          <div className="control-plane-decision">
            <span className="decision-label">POLICY DECISION</span>
            <strong>DENY</strong>
            <p>{id ? "Runtime tidak menerima capability untuk melakukan side effect ini." : "Runtime receives no capability to perform this side effect."}</p>
          </div>
          <div className="control-plane-terminal">
            <div className="terminal-label">EVIDENCE</div>
            <code>decision_id → receipt</code>
            <pre>{'decision: DENY\npolicy: production-safety/v2\nexecution: blocked\nobservation: recorded'}</pre>
          </div>
        </div>
      </section>

      <section className="control-plane-live-section" aria-labelledby="control-plane-live-title">
        <div className="section-heading">
          <div className="eyebrow">LIVE PROOF</div>
          <h2 id="control-plane-live-title">{id ? "Sekarang evaluasi request melalui engine." : "Now evaluate a request through the engine."}</h2>
          <p>
            {id
              ? "Hubungkan halaman ini ke API demo NAEOS untuk melihat decision ID, policy, reason, dan status secara langsung."
              : "Connect this page to the NAEOS demo API to see the decision ID, policy, reason, and status directly from the engine."}
          </p>
        </div>
        <ControlPlaneLiveDemo lang={lang} />
      </section>

      <section className="control-plane-principles">
        <div>
          <div className="eyebrow">{id ? "PRINSIP" : "PRINCIPLES"}</div>
          <h2>{id ? "Decision bukan execution." : "Decision is not execution."}</h2>
        </div>
        <div className="principle-list">
          <div><strong>01</strong><span>{id ? "Model mengusulkan, bukan memberi dirinya sendiri authority." : "The model proposes; it does not grant itself authority."}</span></div>
          <div><strong>02</strong><span>{id ? "Policy menjadi sumber keputusan yang dapat diverifikasi." : "Policy becomes a verifiable source of authorization."}</span></div>
          <div><strong>03</strong><span>{id ? "Observation mengonfirmasi side effect, bukan sekadar log agent." : "Observation confirms side effects, not merely agent history."}</span></div>
          <div><strong>04</strong><span>{id ? "Evidence harus tetap berguna setelah agent selesai." : "Evidence must remain useful after the agent is gone."}</span></div>
        </div>
      </section>
    </div>
  );
}
