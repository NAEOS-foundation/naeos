import { generateRssFeed } from "@/lib/rss";

export async function GET() {
  const xml = generateRssFeed("en");
  return new Response(xml, {
    headers: { "Content-Type": "application/rss+xml; charset=utf-8" },
  });
}

export const dynamic = "force-static";
