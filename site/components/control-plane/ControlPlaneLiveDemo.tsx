// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

"use client";

import { useMemo, useState } from "react";

type DecisionResponse = {
  status?: string;
  decision_id?: string;
  request_id?: string;
  allowed?: boolean;
  needs_approval?: boolean;
  reason?: string;
  message?: string;
  policy_id?: string;
  policy_version?: number;
  requested?: string;
};

type Props = { lang: "en" | "id" };

export default function ControlPlaneLiveDemo({ lang }: Props) {
  const id = lang === "id";
  const configuredEndpoint = process.env.NEXT_PUBLIC_CONTROL_PLANE_API_URL ?? "";
  const [endpoint, setEndpoint] = useState(configuredEndpoint);
  const [agentId, setAgentId] = useState("agent-demo");
  const [capability, setCapability] = useState("database.delete");
  const [artifactHash, setArtifactHash] = useState("sha256:demo-artifact");
  const [result, setResult] = useState<DecisionResponse | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const endpointLabel = useMemo(() => endpoint || (id ? "Belum dikonfigurasi" : "Not configured"), [endpoint, id]);

  async function evaluate() {
    setLoading(true);
    setError("");
    setResult(null);

    if (!endpoint.trim()) {
      setError(
        id
          ? "Konfigurasikan endpoint Control Plane terlebih dahulu. Untuk development, gunakan server demo NAEOS di http://localhost:8080."
          : "Configure the Control Plane endpoint first. For development, use the NAEOS demo server at http://localhost:8080.",
      );
      setLoading(false);
      return;
    }

    try {
      const response = await fetch(endpoint.replace(/\/$/, "") + "/api/control-plane/decision", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          agent_id: agentId,
          capability,
          artifact_hash: artifactHash,
        }),
      });

      const payload = (await response.json()) as DecisionResponse;
      if (!response.ok) {
        throw new Error(payload.message || `HTTP ${response.status}`);
      }
      setResult(payload);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : id
            ? "Evaluasi gagal."
            : "Evaluation failed.",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="control-plane-live">
      <div className="control-plane-live-header">
        <div>
          <div className="eyebrow">LIVE POLICY EVALUATION</div>
          <h3>{id ? "Uji decision dari Control Plane yang sebenarnya." : "Evaluate against the real Control Plane."}</h3>
          <p>
            {id
              ? "Demo ini hanya mengevaluasi authorization request. Tidak ada side effect yang dieksekusi."
              : "This demo evaluates an authorization request only. No side effect is executed."}
          </p>
        </div>
        <span className={endpoint ? "live-status connected" : "live-status"}>
          {endpoint ? (id ? "ENDPOINT READY" : "ENDPOINT READY") : (id ? "BELUM TERHUBUNG" : "NOT CONNECTED")}
        </span>
      </div>

      <label>
        <span>{id ? "Control Plane endpoint" : "Control Plane endpoint"}</span>
        <input value={endpoint} onChange={(event) => setEndpoint(event.target.value)} placeholder="http://localhost:8080" />
      </label>

      <div className="control-plane-live-fields">
        <label>
          <span>Agent</span>
          <input value={agentId} onChange={(event) => setAgentId(event.target.value)} />
        </label>
        <label>
          <span>Capability</span>
          <input value={capability} onChange={(event) => setCapability(event.target.value)} />
        </label>
        <label>
          <span>Artifact hash</span>
          <input value={artifactHash} onChange={(event) => setArtifactHash(event.target.value)} />
        </label>
      </div>

      <button type="button" className="btn btn-primary btn-lg" onClick={evaluate} disabled={loading}>
        {loading ? (id ? "Mengevaluasi…" : "Evaluating…") : (id ? "Evaluate Request" : "Evaluate Request")}
      </button>

      {error && <div className="control-plane-live-error" role="alert">{error}</div>}

      {result && (
        <div className="control-plane-live-result" aria-live="polite">
          <div>
            <span className="decision-label">DECISION</span>
            <strong className={result.status === "ALLOW" ? "decision-allow" : "decision-deny"}>{result.status ?? "UNKNOWN"}</strong>
          </div>
          <dl>
            <div><dt>decision_id</dt><dd>{result.decision_id || "—"}</dd></div>
            <div><dt>policy</dt><dd>{result.policy_id ? `${result.policy_id} v${result.policy_version ?? "?"}` : "—"}</dd></div>
            <div><dt>reason</dt><dd>{result.reason || "—"}</dd></div>
            <div><dt>execution</dt><dd>{result.status === "ALLOW" ? "not executed" : "blocked"}</dd></div>
          </dl>
        </div>
      )}

      <small className="control-plane-live-endpoint">
        {id ? "Endpoint aktif: " : "Active endpoint: "}{endpointLabel}
      </small>
    </div>
  );
}
