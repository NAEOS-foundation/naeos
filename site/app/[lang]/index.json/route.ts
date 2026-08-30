import { buildSearchIndex } from "@/lib/content";
import { LANGUAGES, type Lang } from "@/lib/site";

export const dynamic = "force-static";

export function generateStaticParams() {
  return LANGUAGES.map((lang) => ({ lang }));
}

export const dynamic = "force-static";

export function generateStaticParams() {
  return [{ lang: "en" }, { lang: "id" }];
}

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ lang: string }> },
) {
  const { lang } = await params;
  const entries = buildSearchIndex(lang as Lang);
  return Response.json(entries);
}