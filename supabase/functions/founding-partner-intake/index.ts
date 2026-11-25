import "jsr:@supabase/functions-js/edge-runtime.d.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2.49.1";

const allowedOrigins = new Set(["https://naeos.dev", "https://www.naeos.dev"]);
const roles = ["Technical Founding Partner / CTO", "AI Systems & Engineering Partner", "Product & Engineering Strategy Partner", "Commercial & GTM Partner", "Governance & Security Specialist"];
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
  if (name.length < 2 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) || !roles.includes(role) ||
      experience.length < 20 || contribution.length < 20 || motivation.length < 20 || expectations.length < 10 ||
      !weeklyHours || !collaboration || body.trial_available !== true || body.consent !== true) {
    return response({ error: "Please complete all required fields and consent." }, 400, origin);
  }

  const linkedin = clean(body.linkedin_url, 500);
  const portfolio = clean(body.portfolio_url, 500);
  if (!validUrl(linkedin) || !validUrl(portfolio)) return response({ error: "URLs must start with https://." }, 400, origin);

  const url = Deno.env.get("SUPABASE_URL");
  const key = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");
  if (!url || !key) return response({ error: "Application service unavailable." }, 503, origin);
  const db = createClient(url, key, { auth: { persistSession: false } });
  const ip = req.headers.get("cf-connecting-ip") ?? req.headers.get("x-forwarded-for")?.split(",")[0]?.trim() ?? "unknown";
  const ipHash = await hash(key + ":" + ip);
  const cutoff = new Date(Date.now() - 3600000).toISOString();
  const { count, error: countError } = await db.schema("recruitment").from("intake_rate_limits")
    .select("id", { count: "exact", head: true }).eq("ip_hash", ipHash).gte("created_at", cutoff);
  if (countError) return response({ error: "Submission service temporarily unavailable." }, 503, origin);
  if ((count ?? 0) >= 5) return response({ error: "Too many submissions. Please try again later." }, 429, origin);
  const { error: limitError } = await db.schema("recruitment").from("intake_rate_limits").insert({ ip_hash: ipHash });
  if (limitError) return response({ error: "Submission service temporarily unavailable." }, 503, origin);

  const { error } = await db.schema("recruitment").from("founding_partner_applications").insert({
    applicant_name: name, email, location: clean(body.location, 150) || null,
    linkedin_url: linkedin || null, portfolio_url: portfolio || null,
    professional_role: clean(body.professional_role, 200) || null, primary_role: role,
    alternate_role: clean(body.alternate_role, 120) || null,
    expertise_areas: Array.isArray(body.expertise_areas) ? body.expertise_areas.filter((v) => typeof v === "string").map((v) => clean(v, 80)).slice(0, 20) : [],
    relevant_experience: experience, portfolio_summary: clean(body.portfolio_summary, 3000) || null,
    contribution_proposal: contribution, motivation, weekly_hours: weeklyHours, collaboration_model: collaboration,
    earliest_availability: clean(body.earliest_availability, 10) || null, trial_available: true,
    partnership_expectations: expectations, equity_expectations: clean(body.equity_expectations, 1000) || null,
    consent: true, consent_at: new Date().toISOString(),
  });
  if (error) return response({ error: "We could not save your application. Please try again." }, 503, origin);
  await db.schema("recruitment").from("intake_rate_limits").delete().lt("created_at", new Date(Date.now() - 7 * 86400000).toISOString());
  return response({ ok: true }, 201, origin);
});
