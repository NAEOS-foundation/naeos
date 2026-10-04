// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import Link from "next/link";

export const metadata = {
  title: "AI Engineering Governance Pilot | NAEOS",
  description:
    "A scoped pilot to establish governance, authorization, evidence, and verification controls for AI coding agents.",
};

export default async function EnterprisePage({
  params,
}: {
  params: Promise<{ lang: string }>;
}) {
  const { lang } = await params;
  const id = lang === "id";
  const base = "/" + lang;

  const steps = [
    [
      id ? "01 — Assess" : "01 — Assess",
      id
        ? "Petakan workflow AI coding, capability, risk, dan boundary yang sudah ada."
        : "Map your AI coding workflows, capabilities, risks, and existing execution boundaries.",
    ],
    [
      id ? "02 — Govern" : "02 — Govern",
      id
        ? "Definisikan policy, authorization, approval, dan escalation yang eksplisit."
        : "Define explicit policy, authorization, approval, and escalation rules.",
    ],
    [
      id ? "03 — Prove" : "03 — Prove",
      id
        ? "Jalankan skenario nyata dan kumpulkan evidence yang dapat diverifikasi."
        : "Run real engineering scenarios and capture evidence that can be independently verified.",
    ],
    [
      id ? "04 — Roadmap" : "04 — Roadmap",
      id
        ? "Susun prioritas implementasi menuju AI engineering yang production-ready."
        : "Turn the findings into a prioritized path toward production-ready AI engineering.",
    ],
  ];

  return (
    <main>
      <section className="section">
        <div className="container">
          <span className="problem-tag problem-tag-green">
            AI Engineering Governance Pilot
          </span>
          <h1 className="section-title" style={{ marginTop: "1rem", maxWidth: "900px" }}>
            {id
              ? "Buktikan control plane NAEOS pada workflow engineering Anda."
              : "Prove the NAEOS control plane on your engineering workflow."}
          </h1>
          <p className="section-subtitle" style={{ maxWidth: "820px" }}>
            {id
              ? "Pilot terfokus untuk organisasi yang sudah menggunakan atau mengevaluasi AI coding agents dan membutuhkan kontrol yang eksplisit atas policy, authorization, execution, evidence, dan verification."
              : "A focused pilot for organizations already using or evaluating AI coding agents and needing explicit control over policy, authorization, execution, evidence, and verification."}
          </p>
          <div className="cta-band-actions" style={{ marginTop: "1.5rem" }}>
            <a
              className="btn btn-primary btn-lg"
              href="mailto:hello@naeos.dev?subject=NAEOS%20AI%20Engineering%20Governance%20Pilot"
            >
              {id ? "Diskusikan pilot" : "Discuss the pilot"}
            </a>
            <Link href={base + "/control-plane"} className="btn btn-secondary btn-lg">
              {id ? "Lihat control plane" : "View the control plane"}
            </Link>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <h2 className="section-title">
            {id ? "Apa yang diuji" : "What the pilot tests"}
          </h2>
          <div className="features-grid">
            {steps.map(([title, description]) => (
              <div className="feature-card" key={title}>
                <h3>{title}</h3>
                <p>{description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="problem-grid">
            <div className="problem-card">
              <span className="problem-tag problem-tag-green">Govern</span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id ? "Policy sebelum capability." : "Policy before capability."}
              </h2>
              <p>
                {id
                  ? "Pisahkan intent agent dari keputusan authorization. Capability tidak tersedia hanya karena agent memintanya."
                  : "Separate agent intent from authorization. A capability should not become available merely because an agent requests it."}
              </p>
            </div>
            <div className="problem-card">
              <span className="problem-tag">Verify</span>
              <h2 style={{ marginTop: ".75rem" }}>
                {id ? "Evidence sebelum trust." : "Evidence before trust."}
              </h2>
              <p>
                {id
                  ? "Ukur apa yang benar-benar terjadi, bukan hanya apa yang agent klaim telah dilakukan."
                  : "Measure what actually happened, not only what the agent claims it did."}
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="cta-band">
            <h2>
              {id
                ? "Target akhir: keputusan engineering yang dapat dipertanggungjawabkan."
                : "The outcome: engineering decisions you can defend."}
            </h2>
            <p>
              {id
                ? "Pilot menghasilkan assessment, governance map, proof scenarios, evidence findings, dan roadmap implementasi."
                : "The pilot produces an assessment, governance map, proof scenarios, evidence findings, and an implementation roadmap."}
            </p>
            <div className="cta-band-actions">
              <a
                className="btn btn-primary btn-lg"
                href="mailto:hello@naeos.dev?subject=NAEOS%20AI%20Engineering%20Governance%20Pilot"
              >
                {id ? "Mulai percakapan" : "Start a conversation"}
              </a>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
