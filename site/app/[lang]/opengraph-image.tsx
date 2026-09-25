import { ImageResponse } from "next/og";
import { SITE, type Lang } from "@/lib/site";

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export const runtime = "edge";
export const alt = "NAEOS — Enterprise AI Engineering Framework";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image({ params }: { params: Promise<{ lang: string }> }) {
  const { lang } = await params;
  const l = lang as Lang;
  const title = SITE.title[l] || SITE.title.en;
  const description = SITE.description[l] || SITE.description.en;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          padding: "72px",
          background: "#0B1220",
          color: "#F8FAFC",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 22 }}>
          <div style={{ fontSize: 54, fontWeight: 800, letterSpacing: "-2px" }}>∞</div>
          <div style={{ fontSize: 42, fontWeight: 800, letterSpacing: "2px" }}>NAEOS</div>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 22, maxWidth: 1000 }}>
          <div style={{ fontSize: 64, lineHeight: 1.05, fontWeight: 800 }}>{title}</div>
          <div style={{ fontSize: 28, lineHeight: 1.35, color: "#94A3B8" }}>{description}</div>
        </div>
        <div style={{ display: "flex", fontSize: 24, color: "#22D3EE" }}>naeos.dev</div>
      </div>
    ),
    { ...size }
  );
}
