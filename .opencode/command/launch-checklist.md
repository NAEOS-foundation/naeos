---
description: Runs the NAEOS Product Hunt launch readiness checks and prints a status report for the current phase (D-1, D-day, or D+1–D+7).
agent: launch
---

Run the NAEOS Product Hunt launch readiness report. Today is Monday 17 August
2026 (D-1; launch day is Tuesday 18 August 2026).

1. Read `launch/producthunt/launch-checklist.md` and `launch/producthunt/launch-timeline.md`.
2. Run every readiness check in your prompt's "Readiness checks" section
   (build, version, tag, pushed tag, site links, assets, placeholder scan,
   working tree) and record the real result for each.
3. Also check for remaining `[PH LINK]`, `[Name]`, `[N]`, `[date]`, and
   `[maker invite link` placeholders across `launch/`.
4. Print the report as a markdown table with these columns:
   `Phase | Check | Status (OK/WARN/FAIL) | Detail`.
   Group rows by phase: D-7–D-2, D-1, D-day, D+1–D+7.
5. End with a short "Next actions" list of the 3–5 most urgent unchecked items,
   each with the file/line where it lives and who must do it (maker vs. repo).

Do not mark anything OK without running the check. If a check cannot be run
(e.g., PH account, .env), mark it WARN and say why. Keep the report to one
screen if possible. $ARGUMENTS