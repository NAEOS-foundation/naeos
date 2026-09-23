// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import Link from "next/link";

type Props = {
  base: string;
  lang: "en" | "id";
};

const labels = {
  en: {
    eyebrow: "OPEN-SOURCE ENGINEERING CONTROL PLANE",
    title: "Control how AI builds software.",
    subtitle:
      "NAEOS turns engineering intent into a governed, traceable execution pipeline — from specification to AI agents, artifacts, and evidence.",
    primary: "Explore the control plane",
    secondary: "Get started",
    caption: "Intent → Model → Policy → Agent → Execution → Evidence",
    problem: "AI can generate code. The engineering system still needs control.",
    problemBody:
      "Keep your AI coding tools. Add a deterministic layer for specification, policy, context, execution, and auditability.",
    stages: ["Specification", "NEIR", "Policy", "AI Context", "Agent", "Execution", "Evidence"],
  },
  id: {
    eyebrow: "OPEN-SOURCE ENGINEERING CONTROL PLANE",
    title: "Kendalikan bagaimana AI membangun software.",
    subtitle:
      "NAEOS mengubah intent engineering menjadi pipeline eksekusi yang teratur dan dapat dilacak — dari spesifikasi hingga agent AI, artifact, dan evidence.",
    primary: "Lihat control plane",
    secondary: "Mulai sekarang",
    caption: "Intent → Model → Policy → Agent → Execution → Evidence",
    problem: "AI dapat menghasilkan kode. Sistem engineering tetap membutuhkan kontrol.",
    problemBody:
      "Tetap gunakan AI coding tools Anda. Tambahkan lapisan deterministik untuk specification, policy, context, execution, dan auditability.",
    stages: ["Specification", "NEIR", "Policy", "AI Context", "Agent", "Execution", "Evidence"],
  },
} as const;

export default function ControlPlaneHero({ base, lang }: Props) {
  const copy = labels[lang];
  return (
    <section className="control-plane-hero" aria-labelledby="control-plane-title">
      <div className="control-plane-grid" aria-hidden="true" />
      <div className="container control-plane-inner">
        <div className="control-plane-copy">
          <div className="control-plane-eyebrow">
            <span className="control-plane-status" />
            {copy.eyebrow}
          </div>
          <h2 id="control-plane-title">{copy.title}</h2>
          <p className="control-plane-subtitle">{copy.subtitle}</p>
          <div className="control-plane-actions">
            <Link href={`${base}/docs/getting-started`} className="btn btn-primary btn-lg">
              {copy.secondary}
            </Link>
            <a href="#control-plane" className="btn btn-secondary btn-lg">
              {copy.primary}
            </a>
          </div>
          <p className="control-plane-caption">{copy.caption}</p>
        </div>

        <div className="control-plane-visual" id="control-plane">
          <div className="control-plane-window">
            <div className="control-plane-windowbar">
              <span /><span /><span />
              <code>naeos / control-plane</code>
              <b>TRACEABLE</b>
            </div>
            <div className="control-plane-flow">
              {copy.stages.map((stage, index) => (
                <div className="control-plane-node-wrap" key={stage}>
                  <div className={`control-plane-node ${index === 1 ? "is-core" : ""}`}>
                    <span className="control-plane-index">{String(index + 1).padStart(2, "0")}</span>
                    <strong>{stage}</strong>
                    <small>{index === 0 ? "human intent" : index === 1 ? "canonical model" : index === 2 ? "decision" : index === 6 ? "audit trail" : "execution layer"}</small>
                  </div>
                  {index < copy.stages.length - 1 && <span className="control-plane-connector" aria-hidden="true">→</span>}
                </div>
              ))}
            </div>
            <div className="control-plane-evidence">
              <span className="evidence-dot" />
              <span>{copy.problem}</span>
              <em>evidence</em>
            </div>
          </div>
        </div>

        <div className="control-plane-problem">
          <div className="control-plane-problem-mark">01</div>
          <div>
            <h3>{copy.problem}</h3>
            <p>{copy.problemBody}</p>
          </div>
        </div>
      </div>
    </section>
  );
}
