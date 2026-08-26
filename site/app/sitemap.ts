import type { MetadataRoute } from "next";
import { getAllPages } from "@/lib/content";
import { SITE } from "@/lib/site";

export default function sitemap(): MetadataRoute.Sitemap {
  const pages = getAllPages();
  return pages.map((p) => ({
    url: `${SITE.baseUrl}${p.url}/`,
    lastModified: p.date ? new Date(p.date) : undefined,
    changeFrequency: p.section === "blog" ? "weekly" : "monthly",
    priority: p.url === "/" || (p.section === "" && p.isIndex) ? 1.0 : 0.5,
  }));
}
