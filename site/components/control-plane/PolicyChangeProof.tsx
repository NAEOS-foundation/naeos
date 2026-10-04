"use client";

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import { useState } from "react";

type Step = {
  label: string;
  title: string;
  detail: string;
  status: string;
};

const steps: Step[] = [
  { label: "01", title: "Agent requests capability", detail: "agent-payment-01 requests production.deploy for release-482.", status: "REQUEST" },
  { label: "02", title: "Policy authorizes", detail: "release-policy/v1 grants the capability for this bounded operation.", status: "ALLOW" },
  { label: "03", title: "Authorization is issued", detail: "Decision D-482 binds the capability to policy v1 and artifact sha256:demo-artifact.", status: "AUTHORIZED" },
  { label: "04", title: "Policy changes", detail: "The organization changes the production rule to require human approval.", status: "POLICY CHANGED" },
  { label: "05", title: "Stale authorization is rejected", detail: "The old authorization no longer matches the active policy version. Runtime blocks the side effect.", status: "BLOCKED" },
  { label: "06", title: "Evidence is retained", detail: "The policy transition, stale decision, blocked execution, and verification state remain inspectable.", status: "VERIFIED" },
];

export default function PolicyChangeProof({ lang }: { lang: "en" | "id" }) {
  const id = lang === "id";
  const [step, setStep] = useState(0);
  const current = steps[step];

  return (
    <div style={{ border: "1px solid var(--border-color, #263238)", borderRadius: "16px", padding: "1.25rem", background: "var(--surface, rgba(255,255,255,.02))" }}>
      <div style={{ display: "flex", justifyContent: "space-between", gap: "1rem", alignItems: "flex-start", flexWrap: "wrap" }}>
        <div>
          <div className="eyebrow">POLICY CHANGE PROOF</div>
          <h3 style={{ marginTop: ".5rem" }}>{id ? "Policy berubah di tengah task." : "Policy changes mid-task."}</h3>
          <p style={{ maxWidth: "700px" }}>
            {id
              ? "Skenario deterministik: authorization lama tidak otomatis tetap valid setelah policy berubah."
              : "A deterministic scenario: an existing authorization does not remain valid when the governing policy changes."}
          </p>
        </div>
        <strong style={{ border: "1px solid currentColor", borderRadius: "999px", padding: ".35rem .7rem", fontSize: ".8rem" }}>
          {current.status}
        </strong>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(6, minmax(110px, 1fr))", gap: ".4rem", margin: "1.25rem 0", overflowX: "auto" }}>
        {steps.map((item, index) => (
          <button
            key={item.label}
            type="button"
            onClick={() => setStep(index)}
            aria-current={index === step ? "step" : undefined}
            style={{
              minWidth: "110px",
              textAlign: "left",
              border: index === step ? "2px solid currentColor" : "1px solid var(--border-color, #263238)",
              borderRadius: "10px",
              padding: ".65rem",
              background: index <= step ? "var(--surface-raised, rgba(255,255,255,.06))" : "transparent",
              color: "inherit",
              cursor: "pointer",
            }}
          >
            <small>{item.label}</small>
            <div style={{ fontWeight: 700, marginTop: ".2rem" }}>{item.title}</div>
          </button>
        ))}
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "minmax(0, 1fr) minmax(0, 1fr)", gap: "1rem" }}>
        <div style={{ borderRadius: "12px", padding: "1rem", background: "rgba(127,127,127,.08)" }}>
          <div className="decision-label">{current.status}</div>
          <h4 style={{ margin: ".5rem 0" }}>{current.title}</h4>
          <p>{current.detail}</p>
        </div>
        <pre style={{ margin: 0, borderRadius: "12px", padding: "1rem", overflowX: "auto", background: "#0b0f12", color: "#d7e0e7" }}>{JSON.stringify({
          decision_id: "D-482",
          policy_version: step >= 3 ? 2 : 1,
          authorization: step >= 4 ? "STALE" : step >= 2 ? "ACTIVE" : "NONE",
          execution: step >= 4 ? "BLOCKED" : step >= 2 ? "READY" : "NOT_REQUESTED",
          evidence: step >= 5 ? "RETAINED" : "PENDING",
        }, null, 2)}</pre>
      </div>

      <div className="control-plane-actions" style={{ marginTop: "1rem" }}>
        <button type="button" className="btn btn-secondary" onClick={() => setStep(Math.max(0, step - 1))} disabled={step === 0}>
          {id ? "Sebelumnya" : "Previous"}
        </button>
        <button type="button" className="btn btn-primary" onClick={() => setStep(Math.min(steps.length - 1, step + 1))} disabled={step === steps.length - 1}>
          {step === steps.length - 1 ? (id ? "Selesai" : "Complete") : (id ? "Lanjutkan" : "Next")}
        </button>
        {step === steps.length - 1 && (
          <button type="button" className="btn btn-secondary" onClick={() => setStep(0)}>
            {id ? "Ulangi proof" : "Replay proof"}
          </button>
        )}
      </div>
      <small style={{ display: "block", marginTop: "1rem", opacity: .75 }}>
        {id ? "Proof ini mensimulasikan state transition; tidak menjalankan side effect production." : "This proof simulates the state transition; it does not execute a production side effect."}
      </small>
    </div>
  );
}
