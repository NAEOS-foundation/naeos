// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

import "jsr:@supabase/functions-js/edge-runtime.d.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2.49.1";

const allowedOrigins = new Set(["https://naeos.dev", "https://www.naeos.dev"]);
const roles = ["Technical Founding Partner / CTO", "AI Systems & Engineering Partner", "Product & Engineering Strategy Partner", "Commercial & GTM Partner", "Governance & Security Specialist"];
const expertiseAreas = ["AI agents & orchestration", "Backend / distributed systems", "Frontend & UX", "DevOps / platform", "Security & governance", "Product strategy", "Sales / GTM", "Partnerships", "Research / evaluation"];
const weeklyHoursChoices = ["2–5", "5–10", "10–15", "15–20", "20+"];
const collaborationChoices = ["Part-time", "Flexible", "Other"];
const cors = (origin: string) => ({
  "Access-Control-Allow-Origin": allowedOrigins.has(origin) ? origin : "https://naeos.dev",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
  "Access-Control-Allow-Headers": "authorization, apikey, content-type, x-client-info",
  "Access-Control-Max-Age": "86400",
  "Vary": "Origin",
});
const response = (body: Record<string, unknown>, status: number, origin: string) =>
  new Response(JSON.stringify(body), { status, headers: { ...cors(origin), "Content-Type": "application/json" } });
const clean = (v: unknown, max = 4000) => typeof v === "string" ? v.trim().slice(0, max) : "";
const validUrl = (v: string) => !v || /^https:\/\//i.test(v);
async function hash(value: string) {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(value));
  return Array.from(new Uint8Array(digest)).map((b) => b.toString(16).padStart(2, "0")).join("");
}

Deno.serve(async (req) => {
  const origin = req.headers.get("origin") ?? "";
  if (req.method === "OPTIONS") return new Response(null, { status: 204, headers: cors(origin) });
  if (req.method !== "POST") return response({ error: "Method not allowed" }, 405, origin);
  if (!allowedOrigins.has(origin)) return response({ error: "Origin not allowed" }, 403, origin);

  let body: Record<string, unknown>;
  try { body = await req.json(); } catch { return response({ error: "Invalid JSON" }, 400, origin); }
  if (clean(body.website, 100)) return response({ ok: true }, 202, origin);

  const name = clean(body.applicant_name, 120);
  const email = clean(body.email, 254).toLowerCase();
  const role = clean(body.primary_role, 120);
  const experience = clean(body.relevant_experience, 4000);
  const contribution = clean(body.contribution_proposal, 4000);
  const motivation = clean(body.motivation, 4000);
  const weeklyHours = clean(body.weekly_hours, 40);
  const collaboration = clean(body.collaboration_model, 100);
  const expectations = clean(body.partnership_expectations, 4000);
  const alternateRole = clean(body.alternate_role, 120);
  const availabilityDate = clean(body.earliest_availability, 10);
  const submittedAreas = Array.isArray(body.expertise_areas) ? body.expertise_areas : [];
  const parsedDate = availabilityDate ? Date.parse(availabilityDate + "T00:00:00.000Z") : 0;
  const validDate = !availabilityDate || (/^\d{4}-\d{2}-\d{2}$/.test(availabilityDate) &&
    Number.isFinite(parsedDate) && new Date(parsedDate).toISOString().slice(0, 10) === availabilityDate);
  if (name.length < 2 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) || !roles.includes(role) ||
      (alternateRole && !roles.includes(alternateRole)) ||
      experience.length < 20 || contribution.length < 20 || motivation.length < 20 || expectations.length < 10 ||
      !weeklyHoursChoices.includes(weeklyHours) || !collaborationChoices.includes(collaboration) || body.trial_available !== true || body.consent !== true ||
      !validDate || submittedAreas.length > 20 || submittedAreas.some((v) => typeof v !== "string" || !expertiseAreas.includes(v))) {
    return response({ error: "Please complete all required fields and consent." }, 400, origin);
  }

  const linkedin = clean(body.linkedin_url, 500);
  const portfolio = clean(body.portfolio_url, 500);
  if (!validUrl(linkedin) || !validUrl(portfolio)) return response({ error: "URLs must start with https://." }, 400, origin);

  const url = Deno.env.get("SUPABASE_URL");
  const key = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");
  if (!url || !key) return response({ error: "Application service unavailable." }, 503, origin);
  const db = createClient(url, key, { auth: { persistSession: false } });
  // Trust only platform-provided Cloudflare client IP; forwarded headers can be spoofed.
  const ip = req.headers.get("cf-connecting-ip") ?? "unknown";
  const ipHash = await hash(key + ":" + ip);
  const application = {
    applicant_name: name, email, location: clean(body.location, 150) || null,
    linkedin_url: linkedin || null, portfolio_url: portfolio || null,
    professional_role: clean(body.professional_role, 200) || null, primary_role: role,
    alternate_role: alternateRole || null,
    expertise_areas: submittedAreas,
    relevant_experience: experience, portfolio_summary: clean(body.portfolio_summary, 3000) || null,
    contribution_proposal: contribution, motivation, weekly_hours: weeklyHours, collaboration_model: collaboration,
    earliest_availability: availabilityDate || null, trial_available: true,
    partnership_expectations: expectations, equity_expectations: clean(body.equity_expectations, 1000) || null,
    consent: true,
  };
  const { data: submitted, error: submitError } = await db.rpc("founding_partner_intake_submit", { p_ip_hash: ipHash, p_application: application });
  if (submitError) return response({ error: "Submission service temporarily unavailable." }, 503, origin);
  if (submitted !== true) return response({ error: "Too many submissions. Please try again later." }, 429, origin);
  await db.schema("recruitment").from("intake_rate_limits").delete().lt("created_at", new Date(Date.now() - 7 * 86400000).toISOString());
  return response({ ok: true }, 201, origin);
});
