import { ImageResponse } from "next/og";
import { SITE, type Lang } from "@/lib/site";

// Copyright 2025-2026 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

export const runtime = "edge";
export const alt = "NAEOS — AI Engineering Operating System architecture";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const layers = [
  ["GOVERNANCE", "Vision · Mission · Principles", "#f5b82e"],
  ["CONSTITUTION", "Engineering · AI · Security · Testing", "#20d7bd"],
  ["POLICY", "Rules · Authorization · Guardrails", "#3d8dff"],
  ["KERNEL", "Core Services · Events · State", "#7c5cff"],
  ["RUNTIME", "Execution · Sandbox · Audit", "#2fa8ff"],
  ["COMPILER", "Build · Verify · Validate", "#d7e3f4"],
  ["AI", "Models · Tools · Reasoning · Context", "#a35cff"],
  ["EXTENSIONS", "Plugins · Integrations · Ecosystem", "#20cfa2"],
] as const;

function Layer({
  name,
  detail,
  accent,
  index,
}: {
  name: string;
  detail: string;
  accent: string;
  index: number;
}) {
  return (
    <div
      style={{
        display: "flex",
        position: "relative",
        width: 500,
        height: 43,
        marginLeft: index % 2 === 0 ? 0 : 22,
        marginTop: 4,
        borderRadius: 8,
        border: `1px solid ${accent}`,
        background: `linear-gradient(90deg, ${accent}44, #101a33 72%)`,
        boxShadow: `0 0 18px ${accent}55, 0 8px 0 #071022`,
        alignItems: "center",
        padding: "0 16px",
        color: "#f8fafc",
      }}
    >
      <div style={{ display: "flex", width: 126, fontSize: 16, fontWeight: 800 }}>
        {name}
      </div>
      <div style={{ display: "flex", fontSize: 11, color: "#b8c5d9" }}>{detail}</div>
    </div>
  );
}

export default async function Image({ params }: { params: Promise<{ lang: string }> }) {
  const { lang } = await params;
  const l = lang as Lang;
  const title = SITE.title[l] || SITE.title.en;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          background: "linear-gradient(135deg, #030712 0%, #071329 55%, #020617 100%)",
          color: "#f8fafc",
          fontFamily: "sans-serif",
          position: "relative",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            position: "absolute",
            width: 760,
            height: 760,
            right: -190,
            top: -170,
            borderRadius: 380,
            background: "radial-gradient(circle, #2563eb33 0%, #7c3aed16 42%, transparent 70%)",
          }}
        />

        <div
          style={{
            display: "flex",
            flexDirection: "column",
            width: 390,
            padding: "48px 0 38px 56px",
            justifyContent: "space-between",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
            <div
              style={{
                display: "flex",
                width: 48,
                height: 48,
                borderRadius: 14,
                alignItems: "center",
                justifyContent: "center",
                fontSize: 36,
                fontWeight: 900,
                background: "linear-gradient(135deg,#a855f7,#2563eb)",
                boxShadow: "0 0 28px #6366f166",
              }}
            >
              ∞
            </div>
            <div style={{ display: "flex", flexDirection: "column" }}>
              <div style={{ fontSize: 32, fontWeight: 850, letterSpacing: 3 }}>NAEOS</div>
              <div style={{ fontSize: 10, letterSpacing: 3, color: "#8da1bd" }}>
                AI ENGINEERING OPERATING SYSTEM
              </div>
            </div>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
            <div style={{ display: "flex", fontSize: 49, lineHeight: 1.02, fontWeight: 850 }}>
              Architecture
            </div>
            <div
              style={{
                display: "flex",
                fontSize: 49,
                lineHeight: 1.02,
                fontWeight: 850,
                color: "#4ea1ff",
              }}
            >
              Drives
            </div>
            <div style={{ display: "flex", fontSize: 49, lineHeight: 1.02, fontWeight: 850 }}>
              Engineering.
            </div>
            <div style={{ display: "flex", fontSize: 19, lineHeight: 1.35, color: "#9aaac0", width: 330 }}>
              An open-source, vendor-neutral engineering layer for AI coding agents.
            </div>
          </div>

          <div style={{ display: "flex", gap: 10, flexWrap: "wrap", width: 335 }}>
            {["Governed", "Policy-Driven", "Verified", "Vendor-Neutral", "Extensible"].map((item) => (
              <div
                key={item}
                style={{
                  display: "flex",
                  border: "1px solid #24558d",
                  borderRadius: 999,
                  padding: "7px 12px",
                  fontSize: 11,
                  color: "#b9d7ff",
                  background: "#07182c",
                }}
              >
                {item}
              </div>
            ))}
          </div>
        </div>

        <div
          style={{
            display: "flex",
            flexDirection: "column",
            width: 650,
            padding: "42px 38px 35px 12px",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
            {["Claude", "OpenAI", "Gemini", "Any Agent"].map((agent, i) => (
              <div
                key={agent}
                style={{
                  display: "flex",
                  width: 112,
                  height: 42,
                  alignItems: "center",
                  justifyContent: "center",
                  borderRadius: 9,
                  border: `1px solid ${i === 1 ? "#60a5fa" : "#334e72"}`,
                  background: "#0b1830",
                  color: "#c9d7e9",
                  fontSize: 12,
                  fontWeight: 700,
                  boxShadow: i === 1 ? "0 0 18px #2563eb66" : "none",
                }}
              >
                {agent}
              </div>
            ))}
          </div>

          <div
            style={{
              display: "flex",
              flexDirection: "column",
              position: "relative",
              width: 540,
              padding: "16px 0 25px 0",
              alignItems: "center",
              borderRadius: 18,
              border: "1px solid #2864a4",
              background: "linear-gradient(180deg,#0b1d38cc,#050b18ee)",
              boxShadow: "0 0 50px #2563eb33, 0 22px 0 #020712",
            }}
          >
            <div style={{ display: "flex", fontSize: 12, fontWeight: 800, letterSpacing: 2, color: "#7cc5ff", marginBottom: 8 }}>
              AI CODING AGENTS
            </div>
            {layers.map(([name, detail, accent], index) => (
              <Layer key={name} name={name} detail={detail} accent={accent} index={index} />
            ))}
          </div>

          <div style={{ display: "flex", width: 560, justifyContent: "space-between", alignItems: "center" }}>
            {["Cloud", "On-Premise", "Hybrid", "Edge"].map((item) => (
              <div
                key={item}
                style={{
                  display: "flex",
                  width: 118,
                  height: 36,
                  alignItems: "center",
                  justifyContent: "center",
                  borderRadius: 8,
                  border: "1px solid #21466f",
                  background: "#08172a",
                  color: "#91a9c5",
                  fontSize: 11,
                  fontWeight: 700,
                }}
              >
                {item}
              </div>
            ))}
          </div>

          <div style={{ display: "flex", width: 560, justifyContent: "space-between", alignItems: "center" }}>
            <div style={{ display: "flex", fontSize: 11, color: "#6f87a4" }}>
              GOVERNANCE → POLICY → EXECUTION → VERIFICATION → EVIDENCE
            </div>
            <div style={{ display: "flex", fontSize: 12, color: "#4ea1ff", fontWeight: 800 }}>
              naeos.dev
            </div>
          </div>
        </div>
      </div>
    ),
    { ...size }
  );
}
