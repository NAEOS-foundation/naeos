// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { generateRssFeed } from "@/lib/rss";
import { LANGUAGES, type Lang } from "@/lib/site";

export const dynamic = "force-static";

export function generateStaticParams() {
  return LANGUAGES.map((lang) => ({ lang }));
}

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ lang: string }> },
) {
  const { lang } = await params;
  const xml = generateRssFeed(lang as Lang);
  return new Response(xml, {
    headers: { "Content-Type": "application/rss+xml; charset=utf-8" },
  });
}
