import { buildSearchIndex } from "@/lib/content";

export const dynamic = "force-static";

export async function GET() {
  const entries = buildSearchIndex("en");
  return Response.json(entries);
}
