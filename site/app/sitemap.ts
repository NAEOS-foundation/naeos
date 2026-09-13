// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import type { MetadataRoute } from "next";
import { getAllPages } from "@/lib/content";
import { SITE } from "@/lib/site";

export default function sitemap(): MetadataRoute.Sitemap {
  const pages = getAllPages();
  return pages.map((p) => ({
    url: `${SITE.baseUrl}${p.url}${p.url.endsWith("/") ? "" : "/"}`,
    lastModified: p.date ? new Date(p.date) : undefined,
    changeFrequency: p.section === "blog" ? "weekly" : "monthly",
    priority: p.url === "/" || (p.section === "" && p.isIndex) ? 1.0 : 0.5,
  }));
}
