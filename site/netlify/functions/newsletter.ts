import type { Config } from "@netlify/functions";

import { db } from "../../db/index.js";
import { newsletterSubscribers } from "../../db/schema.js";

const MAX_EMAIL_LENGTH = 254;
const MAX_REQUEST_LENGTH = 4096;
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const SUPPORTED_LOCALES = new Set(["en", "id"]);

type NewsletterRequest = {
  email?: unknown;
  locale?: unknown;
  website?: unknown;
};

function json(body: object, status: number): Response {
  return Response.json(body, {
    status,
    headers: {
      "Cache-Control": "no-store",
    },
  });
}

function normalizeLocale(locale: unknown): string {
  if (typeof locale !== "string") {
    return "en";
  }

  const normalized = locale.trim().toLowerCase().split("-")[0];
  return SUPPORTED_LOCALES.has(normalized) ? normalized : "en";
}

export default async (request: Request): Promise<Response> => {
  if (request.method !== "POST") {
    return json({ ok: false, error: "method_not_allowed" }, 405);
  }

  const contentLength = Number(request.headers.get("content-length") || "0");
  if (contentLength > MAX_REQUEST_LENGTH) {
    return json({ ok: false, error: "invalid_request" }, 413);
  }

  let rawBody: string;
  try {
    rawBody = await request.text();
  } catch {
    return json({ ok: false, error: "invalid_request" }, 400);
  }

  if (rawBody.length > MAX_REQUEST_LENGTH) {
    return json({ ok: false, error: "invalid_request" }, 413);
  }

  let payload: NewsletterRequest;
  try {
    payload = JSON.parse(rawBody) as NewsletterRequest;
  } catch {
    return json({ ok: false, error: "invalid_request" }, 400);
  }

  if (typeof payload.website === "string" && payload.website.trim() !== "") {
    return json({ ok: true }, 200);
  }

  if (typeof payload.email !== "string") {
    return json({ ok: false, error: "invalid_email" }, 400);
  }

  const email = payload.email.trim().toLowerCase();
  if (email.length === 0 || email.length > MAX_EMAIL_LENGTH || !EMAIL_PATTERN.test(email)) {
    return json({ ok: false, error: "invalid_email" }, 400);
  }

  try {
    const inserted = await db
      .insert(newsletterSubscribers)
      .values({
        email,
        locale: normalizeLocale(payload.locale),
        source: "footer",
      })
      .onConflictDoNothing({ target: newsletterSubscribers.email })
      .returning({ id: newsletterSubscribers.id });

    return json({ ok: true }, inserted.length > 0 ? 201 : 200);
  } catch {
    return json({ ok: false, error: "database_unavailable" }, 503);
  }
};

export const config: Config = {
  path: "/api/newsletter",
};
