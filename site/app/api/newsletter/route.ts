import { NextResponse } from "next/server";
import { getCloudflareContext } from "@opennextjs/cloudflare";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const KV_KEY_PREFIX = "subscriber:";

interface NewsletterKv {
  put(key: string, value: string): Promise<void>;
}

export const dynamic = "force-dynamic";

async function newsletterKv(): Promise<NewsletterKv | null> {
  try {
    const ctx = await getCloudflareContext({ async: true });
    const env = ctx.env as CloudflareEnv & { NEWSLETTER_KV?: NewsletterKv };
    if (!env.NEWSLETTER_KV) return null;
    return env.NEWSLETTER_KV;
  } catch {
    return null;
  }
}

export async function POST(request: Request) {
  let body: { email?: unknown; locale?: unknown; website?: unknown };
  const contentLength = request.headers.get("content-length");
  if (contentLength && Number(contentLength) > 4096) {
    return NextResponse.json({ ok: false, error: "payload_too_large" }, { status: 413 });
  }
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ ok: false, error: "invalid_json" }, { status: 400 });
  }

  if (
    typeof body.website === "string" &&
    body.website.trim().length > 0
  ) {
    return NextResponse.json({ ok: true });
  }

  const email = typeof body.email === "string" ? body.email.trim().toLowerCase() : "";
  if (!email || email.length > 254 || !EMAIL_RE.test(email)) {
    return NextResponse.json({ ok: false, error: "invalid_email" }, { status: 400 });
  }

  const kv = await newsletterKv();
  if (kv) {
    const record = {
      email,
      locale: typeof body.locale === "string" ? body.locale : "en",
      subscribed_at: new Date().toISOString(),
    };
    await kv.put(`${KV_KEY_PREFIX}${email}`, JSON.stringify(record));
  }

  return NextResponse.json({ ok: true });
}