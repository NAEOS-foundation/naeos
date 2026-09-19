# NAEOS Site — Third-Party Notices

This file records runtime third-party assets loaded by the NAEOS website
(`site/`). Build-time npm dependencies are recorded separately in
`site/package-lock.json` (see the root `NOTICE` for the lockfile-based
attribution, including the `khroma` / `sharp` / `@img/sharp-libvips-*` notes).

All items below are loaded at runtime from a CDN in the visitor's browser;
they are not vendored into this repository. The allowed CDN origins are
enforced by the Content-Security-Policy in `site/next.config.ts`
(`script-src`, `style-src`, `font-src`) and mirrored in
`site/public/_headers`.

| Asset | Source / Pin | License | Evidence |
|---|---|---|---|
| Inter (font) | Google Fonts (`fonts.googleapis.com/css2`, unversioned API) | OFL-1.1 | `site/app/[lang]/layout.tsx:74-86`, `site/styles/globals.css:23` |
| JetBrains Mono (font) | Google Fonts (`fonts.googleapis.com/css2`, unversioned API) | OFL-1.1 | `site/app/[lang]/layout.tsx:74-86`, `site/styles/globals.css:24` |
| Umami analytics script | `https://cloud.umami.is/script.js` (hosted, loads only when `SITE.umamiWebsiteId` is set) | MIT (Umami) | `site/app/[lang]/layout.tsx:103-108` |
| xterm.js + xterm.css | jsDelivr, pinned `xterm@5.3.0` | MIT | `site/components/home/terminalBridge.ts:50-58` |
| swagger-ui bundle + css | unpkg, pinned `swagger-ui-dist@5.32.15` | Apache-2.0 | `site/components/views/SwaggerApi.tsx:35-39` |

Notes:

- Google Fonts and Umami are served unversioned by their providers; the
  xterm.js and swagger-ui pins above are exact and should be bumped
  deliberately (see `.github/dependabot.yml`; CDN pins are reviewed by hand).
- None of the items above are copyleft (no GPL/AGPL/LGPL). Font embedding via
  the Google Fonts stylesheet is permitted by the OFL-1.1.
- `img-src 'self' data: blob: https:` also permits documentation diagrams and
  OpenGraph images; those first-party images live in `site/public/images/`
  and `brand/`.
