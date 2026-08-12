# Purpose

- Standard documentation conventions for NAEOS: language, structure, naming and templates.

## Language

- Primary language: English. Provide Bahasa Indonesia translations in files with suffix `-id.md` or under `docs/id/`.
- Use short, active sentences. Keep the overview concise.

## File & section structure

- Each major topic should include these sections: Overview, Concepts, Architecture, Workflow, Configuration, Examples, Troubleshooting, Validation.
- Use H2 (`##`) for main sections and H3 (`###`) for subsections.

## Front-matter & summary

- Start with a 1–2 line "Summary:" that states the page purpose.
- Add a "Last updated: YYYY-MM-DD" line at the bottom of each page.

## Code & CLI examples

- Always tag fenced code blocks with the language: ```bash, ```go, etc.
- Provide minimal runnable examples for CLI commands.

## Linking

- Prefer relative links for internal references.
- Keep link paths stable; update links when moving files.

## Formatting

- Use sentence case for headings.
- Use tables sparingly; prefer short lists and examples.

## Review & CI

- Every docs PR should run markdownlint and a link checker.
- Include a short changelog entry in the PR body when adding or changing docs that affect users.
