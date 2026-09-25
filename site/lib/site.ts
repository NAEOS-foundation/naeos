// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export const LANGUAGES = ["en", "id"] as const;
export type Lang = (typeof LANGUAGES)[number];

export const DEFAULT_LANG: Lang = "en";

export const SITE = {
  baseUrl: "https://naeos.dev",
  title: {
    en: "NAEOS — Declarative Platform Engineering System",
    id: "NAEOS — Sistem Rekayasa Platform Deklaratif",
  } as Record<Lang, string>,
  description: {
    en: "Transform YAML/JSON specifications into validated, multi-language project structures with full traceability from intent to implementation.",
    id: "Ubah spesifikasi YAML/JSON menjadi struktur proyek multi-bahasa yang tervalidasi dengan ketelusuran penuh dari niat hingga implementasi.",
  } as Record<Lang, string>,
  copyright: "Copyright © 2026 NAEOS Foundation. All rights reserved.",
  repo: "https://github.com/NAEOS-foundation/naeos",
  repoOwner: "NAEOS-foundation",
  repoName: "naeos",
  version: "3.6.0",
  accentColor: "#08d6ff",
  twitter: "https://twitter.com/naeos_dev",
  instagram: "https://www.instagram.com/naeos_dev/",
  instagramHandle: "@naeos_dev",
  linkedin: "https://www.linkedin.com/company/naeos/",
  discord: "https://discord.com/invite/WnUWmm7XMv",
  slack: "https://join.slack.com/t/naeos/shared_invite/zt-4audirbp0-piCfWuubxo8wDh_XGpkxAQ",
  stats: { cli: 200, languages: 5, ai_platforms: 7, specs: 57 },
  websocketUrl: "wss://ws.naeos.dev/ws",
  umamiWebsiteId: process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID ?? "",
} as const;

export function langPath(lang: Lang, path: string): string {
  const clean = path.startsWith("/") ? path : `/${path}`;
  return lang === DEFAULT_LANG ? clean : `/id${clean}`;
}
