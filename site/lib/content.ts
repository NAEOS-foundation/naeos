import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";
import GithubSlugger from "github-slugger";
import { DEFAULT_LANG, LANGUAGES, type Lang } from "./site";

const CONTENT_DIR = path.join(process.cwd(), "content");

export interface Heading {
  depth: number;
  text: string;
  id: string;
}

export interface Page {
  lang: Lang;
  /** Source-relative path without extension, e.g. "blog/getting-started-tutorial" */
  relPath: string;
  /** Site URL without trailing slash, e.g. "/blog/getting-started-tutorial" */
  url: string;
  section: string;
  isIndex: boolean;
  title: string;
  description: string;
  date?: string;
  author?: string;
  categories?: string[];
  weight?: number;
  layout?: string;
  slugOverride?: string;
  body: string;
  plain: string;
  summary: string;
  readingTime: number;
  headings: Heading[];
  translations: Partial<Record<Lang, string>>;
}

interface RawDoc {
  lang: Lang;
  relPath: string;
  data: Record<string, unknown>;
  content: string;
}

function walk(dir: string, base = ""): string[] {
  const out: string[] = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name.startsWith(".")) continue;
    const rel = base ? `${base}/${entry.name}` : entry.name;
    if (entry.isDirectory()) out.push(...walk(path.join(dir, entry.name), rel));
    else if (entry.name.endsWith(".md")) out.push(rel);
  }
  return out;
}

function stripMarkdown(md: string): string {
  return md
    .replace(/```[\s\S]*?```/g, " ")
    .replace(/`([^`]*)`/g, "$1")
    .replace(/!\[[^\]]*\]\([^)]*\)/g, " ")
    .replace(/\[([^\]]*)\]\([^)]*\)/g, "$1")
    .replace(/^\s{0,3}#{1,6}\s+/gm, "")
    .replace(/[*_~>]/g, "")
    .replace(/\{\{<[^\s>]+[^>]*>\}\}/g, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function excerpt(plain: string, max = 120): string {
  if (plain.length <= max) return plain;
  const cut = plain.slice(0, max);
  return `${cut.slice(0, Math.min(cut.length, cut.lastIndexOf(" ")))}…`;
}

function extractHeadings(md: string): Heading[] {
  const slugger = new GithubSlugger();
  const headings: Heading[] = [];
  let inFence = false;
  for (const line of md.split("\n")) {
    if (/^\s*(```|~~~)/.test(line)) {
      inFence = !inFence;
      continue;
    }
    if (inFence) continue;
    const m = /^(#{2,3})\s+(.+?)\s*#*$/.exec(line);
    if (!m) continue;
    const text = m[2].replace(/\s*\{[^}]*\}\s*$/, "").trim();
    headings.push({
      depth: m[1].length,
      text,
      id: slugger.slug(text),
    });
  }
  return headings;
}

function docSegments(relPath: string, isIndex: boolean): string[] {
  const parts = relPath.split("/");
  const fileName = parts[parts.length - 1];
  let segments: string[];
  if (isIndex && fileName === "_index") {
    segments = parts.slice(0, -1);
  } else {
    segments = parts;
  }
  return segments;
}

function prefixed(lang: Lang, segments: string[]): string {
  const prefix = lang === DEFAULT_LANG ? "" : "/id";
  const joined = `${prefix}/${segments.join("/")}`.replace(/\/+$/, "");
  return joined === "" ? "/" : joined;
}

/** URL for a raw doc, honoring slug overrides (used before the page cache exists). */
function rawDocUrl(d: RawDoc, byKey: Map<string, RawDoc>): string | undefined {
  if (d.data.draft === true) return undefined;
  const fileName = d.relPath.split("/").pop() ?? "";
  const isIndex = fileName === "_index" || fileName === "index";
  const segments = docSegments(d.relPath, isIndex);
  const slugOverride = str(d.data, "slug");
  if (slugOverride && segments.length > 0) {
    segments[segments.length - 1] = slugOverride;
  }
  void byKey;
  return prefixed(d.lang, segments);
}

/** Resolve Hugo {{< relref "path.md" >}} to site URLs against the page language. */
function resolveRelRefs(md: string, lang: Lang, byKey: Map<string, RawDoc>): string {
  return md.replace(
    /\{\{<\s*relref\s+"([^"]+)"\s*>\}\}/g,
    (_match, target: string) => {
      const [file, anchor] = target.split("#");
      const normalized = file.replace(/^\//, "").replace(/\.md$/, "");
      const doc =
        byKey.get(`${lang}:${normalized}`) ??
        byKey.get(`${DEFAULT_LANG}:${normalized}`);
      const url = doc ? rawDocUrl(doc, byKey) : undefined;
      if (!url) return target;
      return `${url}${anchor ? `#${anchor}` : ""}`;
    },
  );
}

function loadRaw(): RawDoc[] {
  const docs: RawDoc[] = [];
  for (const rel of walk(CONTENT_DIR)) {
    const segments = rel.split("/");
    if (segments.some((s) => s.startsWith("_") && s !== "_index.md")) continue;
    const lang = segments[0] === "id" && segments.length > 1 ? "id" : "en";
    const relNoLang =
      lang === "id" ? segments.slice(1).join("/").replace(/\.md$/, "") : rel.replace(/\.md$/, "");
    const raw = fs.readFileSync(path.join(CONTENT_DIR, rel), "utf8");
    const parsed = matter(raw);
    docs.push({
      lang,
      relPath: relNoLang,
      data: parsed.data as Record<string, unknown>,
      content: parsed.content,
    });
  }
  return docs;
}

let cache: Map<string, Page> | null = null;

function str(data: Record<string, unknown>, key: string): string | undefined {
  const v = data[key];
  if (typeof v === "string" && v.length > 0) return v;
  return undefined;
}

function dateString(data: Record<string, unknown>, key: string): string | undefined {
  const v = data[key];
  if (v instanceof Date) return v.toISOString();
  if (typeof v === "string" && v.length > 0) return v;
  return undefined;
}

function buildPages(): Map<string, Page> {
  const raw = loadRaw();
  const byKey = new Map<string, RawDoc>();
  for (const d of raw) byKey.set(`${d.lang}:${d.relPath}`, d);

  const pages = new Map<string, Page>();
  for (const d of raw) {
    if (d.data.draft === true) continue;
    const fileName = d.relPath.split("/").pop() ?? "";
    const isIndex = fileName === "_index" || fileName === "index";
    const section = d.relPath.includes("/")
      ? d.relPath.split("/")[0]
      : "";
    const slugOverride = str(d.data, "slug");
    const urlPath = rawDocUrl(d, byKey)!;

    const translations: Partial<Record<Lang, string>> = {};
    for (const other of LANGUAGES) {
      if (other === d.lang) continue;
      const twin = byKey.get(`${other}:${d.relPath}`);
      if (twin && twin.data.draft !== true) {
        const twinSlug = str(twin.data, "slug");
        const twinSegments = docSegments(d.relPath, isIndex);
        if (twinSlug && twinSegments.length > 0) {
          twinSegments[twinSegments.length - 1] = twinSlug;
        }
        translations[other] = prefixed(other, twinSegments);
      }
    }

    const plain = stripMarkdown(d.content);
    const words = plain.split(" ").filter(Boolean).length;

    const page: Page = {
      lang: d.lang,
      relPath: d.relPath,
      url: urlPath,
      section,
      isIndex,
      title: str(d.data, "title") ?? "Untitled",
      description: str(d.data, "description") ?? "",
      date: dateString(d.data, "date"),
      author: str(d.data, "author"),
      categories: Array.isArray(d.data.categories)
        ? (d.data.categories as string[])
        : undefined,
      weight: typeof d.data.weight === "number" ? d.data.weight : undefined,
      layout: str(d.data, "layout"),
      slugOverride,
      body: resolveRelRefs(d.content, d.lang, byKey),
      plain,
      summary: excerpt(str(d.data, "description") || plain),
      readingTime: Math.max(1, Math.round(words / 200)),
      headings: extractHeadings(d.content),
      translations,
    };
    pages.set(`${d.lang}:${urlPath}`, page);
  }
  return pages;
}

export function getAllPages(lang?: Lang): Page[] {
  if (!cache) cache = buildPages();
  const all = [...cache.values()];
  return lang ? all.filter((p) => p.lang === lang) : all;
}

export function getPage(url: string, lang: Lang): Page | undefined {
  if (!cache) cache = buildPages();
  return cache.get(`${lang}:${url}`);
}

export function findPageByRelPath(relPath: string, lang: Lang): Page | undefined {
  return (
    getAllPages(lang).find((p) => p.relPath === relPath) ??
    getAllPages(DEFAULT_LANG).find((p) => p.relPath === relPath)
  );
}

export function getRegularPages(section: string, lang: Lang): Page[] {
  return getAllPages(lang)
    .filter(
      (p) =>
        p.section === section &&
        !p.isIndex &&
        !p.url.startsWith(`/${section}/_`),
    )
    .sort(sortPages);
}

function sortPages(a: Page, b: Page): number {
  const wa = a.weight ?? 0;
  const wb = b.weight ?? 0;
  if (wa !== wb) return wa - wb;
  if (a.date && b.date) return b.date.localeCompare(a.date);
  return a.title.localeCompare(b.title);
}

export function getBlogPosts(lang: Lang): Page[] {
  return getRegularPages("blog", lang).sort((a, b) =>
    (b.date ?? "").localeCompare(a.date ?? ""),
  );
}

export function getSectionIndex(section: string, lang: Lang): Page | undefined {
  return getAllPages(lang).find(
    (p) => p.section === section && p.isIndex,
  );
}

/** Curated docs navigation order (mirrors the Hugo sidebar). */
export const DOCS_ORDER: { title: string; url: string }[] = [
  { title: "Getting Started", url: "/docs/getting-started" },
  { title: "Installation", url: "/docs/installation" },
  { title: "Quick Reference", url: "/docs/quick-reference" },
  { title: "Core Principles", url: "/docs/core-principles" },
  { title: "Architecture", url: "/docs/architecture" },
  { title: "Project Structure", url: "/docs/project-structure" },
  { title: "Spec Language", url: "/docs/spec-language" },
  { title: "NEIR Model", url: "/docs/neir-model" },
  { title: "Pipeline Engine", url: "/docs/pipeline-engine" },
  { title: "AI Compiler", url: "/docs/ai-compiler" },
  { title: "Context Bundles", url: "/docs/context-bundles" },
  { title: "Profiles", url: "/docs/profiles" },
  { title: "Plugin SDK", url: "/docs/plugin-sdk" },
  { title: "Validation", url: "/docs/validation" },
  { title: "Governance", url: "/docs/governance" },
  { title: "Cloud Deployment", url: "/docs/cloud-deployment" },
  { title: "CLI Reference", url: "/docs/cli-reference" },
  { title: "API Reference", url: "/docs/api" },
  { title: "Troubleshooting", url: "/docs/troubleshooting" },
  { title: "Glossary", url: "/docs/glossary" },
  { title: "Contributing", url: "/docs/contributing" },
];

export function localizedDocsOrder(lang: Lang): { title: string; url: string }[] {
  return DOCS_ORDER.map((entry) => {
    const page = getPage(entry.url, lang);
    return page
      ? { title: entry.title, url: page.url, localTitle: page.title }
      : null;
  }).filter((e): e is { title: string; url: string; localTitle: string } => e !== null)
    .map(({ url, localTitle }) => ({ url, title: localTitle }));
}

export function docsNeighbors(url: string, lang: Lang) {
  const list = localizedDocsOrder(lang);
  const idx = list.findIndex((d) => d.url === url);
  return {
    prev: idx > 0 ? list[idx - 1] : undefined,
    next: idx >= 0 && idx < list.length - 1 ? list[idx + 1] : undefined,
  };
}

export interface SearchEntry {
  title: string;
  permalink: string;
  section: string;
  content: string;
  sections: string[];
}

export function buildSearchIndex(lang: Lang): SearchEntry[] {
  return getAllPages(lang)
    .filter((p) => !(p.section === "" && p.isIndex))
    .map((p) => ({
      title: p.title,
      permalink: `${p.url}/`,
      section: p.section || "page",
      content: p.plain.slice(0, 5000),
      sections: p.headings.map((h) => h.text),
    }));
}
