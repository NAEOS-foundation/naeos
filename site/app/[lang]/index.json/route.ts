import { buildSearchIndex } from "@/lib/content";
import type { Lang } from "@/lib/site";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ lang: string }> },
) {
  const { lang } = await params;
  const entries = buildSearchIndex(lang as Lang);
  return Response.json(entries);
}
