import { generateRssFeed } from "@/lib/rss";
import type { Lang } from "@/lib/site";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ lang: Lang }> },
) {
  const { lang } = await params;
  const xml = generateRssFeed(lang);
  return new Response(xml, {
    headers: { "Content-Type": "application/rss+xml; charset=utf-8" },
  });
}
