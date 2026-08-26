import { getBlogPosts } from "@/lib/content";
import { SITE } from "@/lib/site";
import type { Lang } from "@/lib/site";

function escapeXml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function generateRssFeed(lang: Lang): string {
  const posts = getBlogPosts(lang).slice(0, 20);
  const prefix = lang === "id" ? `${SITE.baseUrl}/id` : SITE.baseUrl;

  const items = posts
    .map(
      (p) => `    <item>
      <title>${escapeXml(p.title)}</title>
      <link>${prefix}${p.url}/</link>
      <guid>${prefix}${p.url}/</guid>
      <pubDate>${p.date ? new Date(p.date).toUTCString() : ""}</pubDate>
      <description>${escapeXml(p.description)}</description>
    </item>`,
    )
    .join("\n");

  return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>${escapeXml(SITE.title[lang])}</title>
    <link>${prefix}/blog/</link>
    <description>${escapeXml(SITE.description[lang])}</description>
    <language>${lang === "id" ? "id" : "en"}</language>
${items}
  </channel>
</rss>
`;
}
