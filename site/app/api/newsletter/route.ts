import { NextResponse } from "next/server";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const dynamic = "force-dynamic";

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

  const locale = body.locale === "id" ? "id" : "en";

  try {
    const [{ db }, { newsletterSubscribers }] = await Promise.all([
      import("@/db"),
      import("@/db/schema"),
    ]);
    const result = await db
      .insert(newsletterSubscribers)
      .values({ email, locale })
      .onConflictDoNothing({ target: newsletterSubscribers.email })
      .returning();
    return NextResponse.json(
      { ok: true },
      { status: result.length > 0 ? 201 : 200 },
    );
  } catch {
    return NextResponse.json(
      { ok: false, error: "database_unavailable" },
      { status: 503 },
    );
  }
}
