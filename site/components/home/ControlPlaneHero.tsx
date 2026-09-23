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
      "NAEOS turns engineering intent into a governed, traceable execution system — from specification and NEIR to policy, agents, execution, and evidence.",
    primary: "Explore the control plane",
    secondary: "Get started",
    caption: "Specification → NEIR → Policy → AI Context → Agent → Execution → Evidence",
    stages: [
      ["Specification", "human intent"],
      ["Policy", "constraints"],
      ["AI Context", "bounded context"],
      ["Agent", "execution"],
      ["Execution", "verified change"],
      ["Evidence", "audit trail"],
    ],
  },
  id: {
    eyebrow: "OPEN-SOURCE ENGINEERING CONTROL PLANE",
    title: "Kendalikan bagaimana AI membangun software.",
    subtitle:
      "NAEOS mengubah intent engineering menjadi sistem eksekusi yang teratur dan dapat dilacak — dari specification dan NEIR hingga policy, agent, execution, dan evidence.",
    primary: "Lihat control plane",
    secondary: "Mulai sekarang",
    caption: "Specification → NEIR → Policy → AI Context → Agent → Execution → Evidence",
    stages: [
      ["Specification", "human intent"],
      ["Policy", "constraints"],
      ["AI Context", "bounded context"],
      ["Agent", "execution"],
      ["Execution", "verified change"],
      ["Evidence", "audit trail"],
    ],
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
          <h1 id="control-plane-title">{copy.title}</h1>
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

            <div className="control-plane-graph">
              <div className="control-plane-inputs">
                <div className="control-plane-node control-plane-node-input">
                  <span className="control-plane-index">01</span>
                  <strong>{copy.stages[0][0]}</strong>
                  <small>{copy.stages[0][1]}</small>
                </div>
              </div>

              <div className="control-plane-core">
                <span className="control-plane-core-label">CANONICAL ENGINEERING MODEL</span>
                <div className="control-plane-neir">
                  <span className="control-plane-neir-mark">NEIR</span>
                  <strong>Normalized Engineering Intermediate Representation</strong>
                  <small>One machine-readable model connects intent, constraints, context, execution, and evidence.</small>
                </div>
                <div className="control-plane-policy">
                  <span>POLICY</span>
                  <i>validated</i>
                </div>
              </div>

              <div className="control-plane-outputs">
                {copy.stages.slice(1).map(([stage, detail], index) => (
                  <div className="control-plane-node-wrap" key={stage}>
                    <span className="control-plane-connector" aria-hidden="true">→</span>
                    <div className="control-plane-node">
                      <span className="control-plane-index">{String(index + 2).padStart(2, "0")}</span>
                      <strong>{stage}</strong>
                      <small>{detail}</small>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="control-plane-evidence">
              <span className="evidence-dot" />
              <span>Specification-linked trace across policy decisions, agent actions, execution results, and artifacts.</span>
              <em>evidence</em>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
