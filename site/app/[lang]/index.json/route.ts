// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

import { buildSearchIndex } from "@/lib/content";
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
  const entries = buildSearchIndex(lang as Lang);
  return Response.json(entries);
}
