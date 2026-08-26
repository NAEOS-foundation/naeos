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
  version: "3.1.0",
  accentColor: "#00ff88",
  twitter: "https://twitter.com/naeos_dev",
  twitterHandle: "@naeos_dev",
  discord: "https://discord.gg/naeos",
  stats: { cli: 67, languages: 5, ai_platforms: 6, specs: 56 },
  websocketUrl: "wss://ws.naeos.dev/ws",
  umamiWebsiteId: process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID ?? "",
} as const;

export function langPath(lang: Lang, path: string): string {
  const clean = path.startsWith("/") ? path : `/${path}`;
  return lang === DEFAULT_LANG ? clean : `/id${clean}`;
}
