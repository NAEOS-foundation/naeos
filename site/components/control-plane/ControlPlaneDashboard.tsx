// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

"use client";

import { useEffect, useState } from "react";

type Decision = {
  id: string;
  action: string;
  agent: string;
  policy: string;
  status: "ALLOW" | "BLOCK" | "ESCALATE";
  age: string;
  reason: string;
  capability: string;
};

type Props = { lang: "en" | "id" };

type Run = { id: string; decision_id?: string; agent_id: string; capability?: string; status: "ACTIVE" | "EXECUTING" | "VERIFIED" | "BLOCKED"; started_at: string; updated_at: string; execution_id?: string; verification?: string };
type RunsResponse = { runs?: Run[]; total?: number };

type EvidenceResponse = { evidence?: Array<{ DecisionID?: string; decision_id?: string; RequestID?: string; request_id?: string; AgentID?: string; agent_id?: string; Capability?: string; capability?: string; Decision?: string; decision?: string; Reason?: string; reason?: string; PolicyID?: string; policy_id?: string; PolicyVersion?: string | number; policy_version?: string | number; Verification?: { Result?: string; result?: string }; verification?: { Result?: string; result?: string }; ExecutionID?: string; execution_id?: string }>; verified?: boolean; persistence_error?: string };

const decisions: Decision[] = [
  { id: "DEC-91A72", action: "dependency.add", agent: "coding-agent-07", policy: "security/dependency-v3", status: "BLOCK", age: "2 min ago", reason: "Dependency is outside the approved supply-chain policy.", capability: "dependency.write" },
  { id: "DEC-91A68", action: "test.run", agent: "coding-agent-07", policy: "engineering/default-v2", status: "ALLOW", age: "4 min ago", reason: "Capability is granted by the active engineering profile.", capability: "test.execute" },
  { id: "DEC-91A41", action: "production.deploy", agent: "release-agent-02", policy: "production/change-v4", status: "ESCALATE", age: "9 min ago", reason: "Production deployment requires explicit human approval.", capability: "production.deploy" },
];

const CONTROL_PLANE_ENDPOINT =
  process.env.NEXT_PUBLIC_CONTROL_PLANE_API_URL ||
  (process.env.NODE_ENV === "production" ? "https://naeos-control-plane-production.up.railway.app" : "");

export default function ControlPlaneDashboard({ lang }: Props) {
  const id = lang === "id";
  const [selected, setSelected] = useState(decisions[0]);
  const [liveDecisions, setLiveDecisions] = useState<Decision[]>([]);
  const [evidenceState, setEvidenceState] = useState<"idle" | "loading" | "ready" | "error">("idle");
  const [verification, setVerification] = useState<boolean | null>(null);
  const [runs, setRuns] = useState<Run[]>([]);

  useEffect(() => {
    if (!CONTROL_PLANE_ENDPOINT) return;
    let cancelled = false;
    const endpoint = CONTROL_PLANE_ENDPOINT.endsWith("/") ? CONTROL_PLANE_ENDPOINT.slice(0, -1) : CONTROL_PLANE_ENDPOINT;
    fetch(endpoint + "/api/control-plane/runs", { cache: "no-store" })
      .then(async (response) => { if (!response.ok) throw new Error("HTTP " + response.status); return (await response.json()) as RunsResponse; })
      .then((payload) => { if (!cancelled) setRuns(payload.runs ?? []); })
      .catch(() => { if (!cancelled) setRuns([]); });
    return () => { cancelled = true; };
  }, []);

  useEffect(() => {
    if (!CONTROL_PLANE_ENDPOINT) return;
    let cancelled = false;
    setEvidenceState("loading");
    fetch(`${CONTROL_PLANE_ENDPOINT.replace(/\/$/, "")}/api/control-plane/evidence`, { cache: "no-store" })
      .then(async (response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return (await response.json()) as EvidenceResponse;
      })
      .then((payload) => {
        if (cancelled) return;
        const mapped = (payload.evidence ?? []).map((bundle, index): Decision => ({
          id: bundle.DecisionID ?? bundle.decision_id ?? `LIVE-${index + 1}`,
          action: bundle.Capability ?? bundle.capability ?? "control-plane.action",
          agent: bundle.AgentID ?? bundle.agent_id ?? "unknown-agent",
          policy: `${bundle.PolicyID ?? bundle.policy_id ?? "unknown-policy"} v${bundle.PolicyVersion ?? bundle.policy_version ?? "?"}`,
          status: ((bundle.Decision ?? bundle.decision ?? "BLOCK").toUpperCase() === "ALLOW" ? "ALLOW" : (bundle.Decision ?? bundle.decision ?? "").toUpperCase() === "ESCALATE" ? "ESCALATE" : "BLOCK"),
          age: "live",
          reason: bundle.Reason ?? bundle.reason ?? "No reason recorded.",
          capability: bundle.Capability ?? bundle.capability ?? "unknown",
        }));
        const next = mapped.slice(0, 8);
        setLiveDecisions(next);
        if (next.length > 0) setSelected(next[0]);
        setVerification(payload.verified ?? null);
        setEvidenceState("ready");
      })
      .catch(() => {
        if (!cancelled) setEvidenceState("error");
      });
    return () => { cancelled = true; };
  }, []);

  const hasLiveDecisions = liveDecisions.length > 0;
  const visibleDecisions = hasLiveDecisions ? liveDecisions : decisions;
  const activeRuns = runs.filter((run) => run.status === "ACTIVE" || run.status === "EXECUTING").length;
  const blockedDecisions = liveDecisions.filter((decision) => decision.status === "BLOCK").length;
  const verifiedRuns = runs.filter((run) => run.status === "VERIFIED").length;
  const knownAgents = new Set([
    ...runs.map((run) => run.agent_id),
    ...liveDecisions.map((decision) => decision.agent),
  ].filter(Boolean)).size;
  const verificationRate = runs.length ? Math.round((verifiedRuns / runs.length) * 100) : null;
  const selectedHasLiveEvidence = hasLiveDecisions && liveDecisions.some((decision) => decision.id === selected.id);

  return (
    <section className="control-plane-dashboard" aria-labelledby="control-plane-dashboard-title">
      <div className="section-heading">
        <div className="eyebrow">CONTROL PLANE CONSOLE</div>
        <h2 id="control-plane-dashboard-title">{id ? "Operasikan governance, bukan sekadar melihat log." : "Operate governance, not just logs."}</h2>
        <p>{id ? "Satu permukaan untuk melihat agent, policy decision, execution state, dan evidence." : "One surface for agents, policy decisions, execution state, and evidence."}</p>
      </div>
      <div className="control-plane-shell">
        <aside className="control-plane-sidebar" aria-label="Control Plane navigation">
          <div className="control-plane-brand"><span>∞</span><strong>NAEOS</strong></div>
          <nav>
            <div className="control-plane-nav-group"><span>CONTROL</span><button className="active" type="button">Dashboard</button><button type="button" disabled title={id ? "Belum tersedia" : "Not available yet"}>Agents</button><button type="button" onClick={() => document.getElementById("control-plane-runs")?.scrollIntoView({ behavior: "smooth", block: "start" })}>Runs</button><button type="button" onClick={() => document.getElementById("control-plane-decisions")?.scrollIntoView({ behavior: "smooth", block: "start" })}>Decisions</button></div>
            <div className="control-plane-nav-group"><span>GOVERNANCE</span><button type="button" disabled title={id ? "Belum tersedia" : "Not available yet"}>Policies</button><button type="button" disabled title={id ? "Belum tersedia" : "Not available yet"}>Exceptions</button></div>
            <div className="control-plane-nav-group"><span>TRUST</span><button type="button" onClick={() => document.getElementById("control-plane-evidence")?.scrollIntoView({ behavior: "smooth", block: "start" })}>Verification</button><button type="button" onClick={() => document.getElementById("control-plane-evidence")?.scrollIntoView({ behavior: "smooth", block: "start" })}>Evidence</button></div>
          </nav>
        </aside>
        <div className="control-plane-main">
          <header className="control-plane-console-header"><div><span className="console-kicker">CONTROL PLANE</span><h3>{id ? "Ringkasan operasional" : "Operational overview"}</h3></div><span className={"console-status " + (evidenceState === "ready" && verification ? "is-healthy" : evidenceState === "error" ? "is-error" : "is-neutral")}><i /> {evidenceState === "ready" ? (verification ? (id ? "Evidence terverifikasi" : "Evidence verified") : (id ? "Evidence belum terverifikasi" : "Evidence unverified")) : evidenceState === "loading" ? (id ? "Menyinkronkan evidence" : "Syncing evidence") : evidenceState === "error" ? (id ? "Evidence tidak tersedia" : "Evidence unavailable") : (id ? "API belum terhubung" : "API not connected")}</span></header>
          <div className="control-plane-metrics">
            <article><span>Agents</span><strong>12</strong><small>2 active now</small></article>
            <article><span>Active Runs</span><strong>4</strong><small>1 awaiting verification</small></article>
            <article><span>Blocked</span><strong>3</strong><small>last 24 hours</small></article>
            <article><span>Verified</span><strong>98.7%</strong><small>decision → evidence</small></article>
          </div>
          <div className="control-plane-console-grid" id="control-plane-decisions">
            <div className="control-plane-panel">
              <div className="panel-header"><div><span className="console-kicker">RECENT DECISIONS</span><h4>{hasLiveDecisions ? (id ? "Keputusan policy live" : "Live policy decisions") : (id ? "Contoh keputusan" : "Sample decisions")}</h4></div><span className="panel-count">{visibleDecisions.length}</span></div>{!hasLiveDecisions && <p className="control-plane-data-note">{evidenceState === "error" ? (id ? "Data live gagal dimuat. Daftar di bawah adalah contoh, bukan data produksi." : "Live data failed to load. The list below is sample data, not production data.") : (id ? "Data contoh — hubungkan API untuk melihat keputusan nyata." : "Sample data — connect the API to view real decisions.")}</p>}
              <div className="decision-list">
                {visibleDecisions.map((decision) => (
                  <button key={decision.id} type="button" className={selected.id === decision.id ? "decision-row selected" : "decision-row"} onClick={() => setSelected(decision)}>
                    <span className={"decision-dot " + decision.status.toLowerCase()} />
                    <span className="decision-copy"><strong>{decision.action}</strong><small>{decision.agent} · {decision.age}</small></span>
                    <span className={"decision-badge " + decision.status.toLowerCase()}>{decision.status}</span>
                  </button>
                ))}
              </div>
            </div>
            <div className="control-plane-panel decision-detail-panel" id="control-plane-evidence">
              <div className="panel-header"><div><span className="console-kicker">DECISION DETAIL</span><h4>{selected.id}</h4></div><span className={"decision-badge " + selected.status.toLowerCase()}>{selected.status}</span></div>
              <div className="decision-detail-action">{selected.action}</div>
              <dl className="decision-facts">
                <div><dt>Agent</dt><dd>{selected.agent}</dd></div><div><dt>Policy</dt><dd>{selected.policy}</dd></div><div><dt>Capability</dt><dd>{selected.capability}</dd></div><div><dt>Reason</dt><dd>{selected.reason}</dd></div>
              </dl>
              <div className="decision-chain"><span>Request</span><b>→</b><span>Policy</span><b>→</b><span>Decision</span><b>→</b><span>Evidence</span></div>
              <div className="decision-detail-actions"><button type="button" className="btn btn-secondary btn-sm" disabled={!selectedHasLiveEvidence || !CONTROL_PLANE_ENDPOINT} onClick={() => window.open(`${CONTROL_PLANE_ENDPOINT.replace(/\/$/, "")}/api/control-plane/evidence?decision_id=${encodeURIComponent(selected.id)}`, "_blank", "noopener,noreferrer")}>{id ? "Lihat Evidence" : "View Evidence"}</button><button type="button" className="btn btn-secondary btn-sm" disabled title={id ? "Navigasi policy belum tersedia" : "Policy navigation is not available yet"}>{id ? "Lihat Policy" : "View Policy"}</button></div>
            </div>
          </div>
          <div className="control-plane-run" id="control-plane-runs">
            <div className="run-heading"><div><span className="console-kicker">REAL RUNS</span><h4>{runs.length ? runs.length + (id ? " jejak eksekusi" : " execution traces") : (id ? "Menunggu jejak eksekusi" : "Waiting for execution traces")}</h4></div><span className="run-progress-label">{runs.length ? "LIVE" : "—"}</span></div>
            <div className="run-list">
              {runs.map((run) => (
                <div className="run-item" key={run.id}>
                  <div><strong>{run.id}</strong><small>{run.agent_id} · {run.capability ?? "execution"}</small></div>
                  <span className={"run-state " + run.status.toLowerCase()}>{run.status}</span>
                </div>
              ))}
              {!runs.length && <div className="run-empty">{evidenceState === "error" ? (id ? "Run belum dapat dimuat. Periksa koneksi API." : "Runs could not be loaded. Check the API connection.") : (id ? "Belum ada run berbasis eksekusi di ledger control plane." : "No execution-backed runs are present in the control-plane ledger.")}</div>}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
