import type { Lang } from "@/lib/site";

export const runtime = "edge";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ lang: Lang }> },
) {
  const { lang } = await params;
  const { buildSearchIndex } = await import("@/lib/content");
  const entries = buildSearchIndex(lang);
  return Response.json(entries);
}
