const sharp = require('sharp');
const fs = require('fs');
const path = require('path');

const OUT = __dirname;
const BRAND = path.resolve(__dirname, '../../../brand');

const C = {
  bg: '#05050a', surface: '#0a0a14', panel: '#111122', border: '#1a1a33', borderLight: '#2a2a44',
  text: '#f0f0f0', muted: '#9999b0', dim: '#666680',
  cyan: '#08d6ff', blue: '#3a7cf8', violet: '#7c4dff', purple: '#9333ea',
  green: '#27c93f', yellow: '#ffaa00', red: '#ff4444', info: '#60a5fa',
  code: '#bd93f9', string: '#ffb86c', comment: '#5a5a72', keyword: '#ff79c6',
};
const GRAD = 'linearGradient(id="ngrad",x1="0.5",y1="0",x2="0.5",y2="1"){stop(offset="0%",stop-color="#08d6ff");stop(offset="25%",stop-color="#3a7cf8");stop(offset="50%",stop-color="#6056f5");stop(offset="75%",stop-color="#7c4dff");stop(offset="100%",stop-color="#9333ea")}';
const SANS = 'Ubuntu Sans, DejaVu Sans, sans-serif';
const MONO = 'DejaVu Sans Mono, monospace';

function esc(s) { return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;'); }

async function render(svg, name, w, h) {
  const file = path.join(OUT, name);
  await sharp(Buffer.from(svg)).png().toFile(file);
  const meta = await sharp(file).metadata();
  console.log(`${name}  ${meta.width}x${meta.height}`);
}

function gridSVG(w, h, step = 80) {
  let g = '';
  for (let x = 0; x <= w; x += step) g += `<line x1="${x}" y1="0" x2="${x}" y2="${h}" stroke="#0d0d1c" stroke-width="1"/>`;
  for (let y = 0; y <= h; y += step) g += `<line x1="0" y1="${y}" x2="${w}" y2="${y}" stroke="#0d0d1c" stroke-width="1"/>`;
  return `<g opacity="0.6">${g}</g>`;
}

function glow(cx, cy, r, color, opacity = 0.35) {
  return `<radialGradient id="gl${cx}${cy}" cx="0.5" cy="0.5" r="0.5"><stop offset="0%" stop-color="${color}" stop-opacity="${opacity}"/><stop offset="100%" stop-color="${color}" stop-opacity="0"/></radialGradient>
  <circle cx="${cx}" cy="${cy}" r="${r}" fill="url(#gl${cx}${cy})"/>`;
}

function window(title, w, h) {
  return `<rect x="0" y="0" width="${w}" height="${h}" rx="14" fill="${C.surface}" stroke="${C.border}" stroke-width="1"/>
  <rect x="0" y="0" width="${w}" height="44" rx="14" fill="${C.panel}"/>
  <rect x="0" y="30" width="${w}" height="14" fill="${C.panel}"/>
  <circle cx="20" cy="22" r="6" fill="${C.red}"/><circle cx="42" cy="22" r="6" fill="${C.yellow}"/><circle cx="64" cy="22" r="6" fill="${C.green}"/>
  <text x="${w / 2}" y="29" text-anchor="middle" font-family="${MONO}" font-size="14" fill="${C.dim}">${esc(title)}</text>`;
}

function t(x, y, s, size, fill, family = SANS, weight = 400, anchor = 'start') {
  return `<text x="${x}" y="${y}" font-family="${family}" font-size="${size}" fill="${fill}" font-weight="${weight}" text-anchor="${anchor}">${esc(s)}</text>`;
}

function chip(x, y, label, fill, bg = 'none', border = C.borderLight) {
  const w = label.length * 14 + 30;
  return `<rect x="${x}" y="${y - 24}" width="${w}" height="34" rx="17" fill="${bg}" stroke="${border}" stroke-width="1"/>
  <text x="${x + w / 2}" y="${y + 4}" text-anchor="middle" font-family="${SANS}" font-size="17" font-weight="600" fill="${fill}">${esc(label)}</text>`;
}

function termLine(x, y, parts, size = 21, lh = 34) {
  let out = `<text x="${x}" y="${y}" font-family="${MONO}" font-size="${size}">`;
  for (const p of parts) {
    out += `<tspan fill="${p.c}">${esc(p.s)}</tspan>`;
  }
  out += '</text>';
  return { svg: out, next: y + lh };
}

/* ------------------------------------------------------------------ */
/* 1. Logo 240x240                                                      */
/* ------------------------------------------------------------------ */
async function logo() {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="240" height="240" viewBox="0 0 460 460">${fs.readFileSync(path.join(BRAND, 'logo.svg'), 'utf8').replace(/^<svg[^>]*>/, '').replace(/<\/svg>$/, '')}</svg>`;
  await render(svg, 'logo-240.png', 240, 240);
}

/* ------------------------------------------------------------------ */
/* 2. Hero cover 1600x900                                              */
/* ------------------------------------------------------------------ */
async function hero() {
  const run = [
    [{ s: '$ naeos run --config config.yaml --input-file spec-full.yaml', c: C.text }],
    [{ s: 'INFO ', c: C.blue }, { s: 'parsing specification (2407 bytes)', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'resolving cross-references', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'building NEIR model', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'scheduling 9 tasks', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'generating artifacts', c: C.muted }],
    [{ s: '✓ pipeline complete — 124 artifacts · 11 tasks · 124 reviews', c: C.green }],
  ];
  let y = 100;
  let content = '';
  for (const parts of run) {
    const r = termLine(24, y, parts, 18, 32);
    content += r.svg;
    y = r.next;
  }
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="900" viewBox="0 0 1600 900">
  <defs>${GRAD}
    <linearGradient id="txtgrad" x1="0" y1="0" x2="1" y2="0"><stop offset="0%" stop-color="#08d6ff"/><stop offset="100%" stop-color="#9333ea"/></linearGradient>
  </defs>
  <rect width="1600" height="900" fill="${C.bg}"/>
  ${gridSVG(1600, 900)}
  ${glow(500, 420, 420, '#08d6ff', 0.14)}
  ${glow(820, 300, 300, '#7c4dff', 0.12)}
  <image href="${pathToData(path.join(BRAND, 'logo-mark.svg'))}" x="70" y="60" width="190" height="190"/>
  <text x="280" y="175" font-family="${SANS}" font-size="100" font-weight="800" fill="url(#txtgrad)">NAEOS</text>
  <text x="90" y="300" font-family="${SANS}" font-size="44" font-weight="700" fill="${C.text}">Specify Once. Build Anywhere.</text>
  <text x="90" y="360" font-family="${SANS}" font-size="24" fill="${C.muted}">A declarative engineering platform that turns one specification into validated,</text>
  <text x="90" y="396" font-family="${SANS}" font-size="24" fill="${C.muted}">multi-language software — for humans and AI.</text>
  ${chip(90, 480, 'v3.0.0', C.cyan, C.panel)}
  ${chip(210, 480, 'Open Source · Apache 2.0', C.green, C.panel)}
  ${chip(480, 480, 'Go · Single Binary', C.violet, C.panel)}
  <g transform="translate(880,90)">
    ${window('naeos run — pipeline', 640, 430)}
    ${content}
  </g>
  <rect x="880" y="556" width="640" height="1" fill="${C.border}"/>
  <text x="880" y="600" font-family="${SANS}" font-size="22" fill="${C.dim}">1 spec in → 124 artifacts · code · docs · AI context</text>
  </svg>`;
  await render(svg, '01-hero-cover.png', 1600, 900);
}

/* ------------------------------------------------------------------ */
/* 3. Pipeline diagram 1600x900                                        */
/* ------------------------------------------------------------------ */
async function pipeline() {
  const cols = [
    { title: 'Input', sub: 'What you describe', color: C.cyan, items: ['Spec YAML / JSON', 'CLI Commands', 'Industry Profiles', 'AI Context', 'Schema Registry'] },
    { title: 'Core Layer', sub: 'What NAEOS validates', color: C.blue, items: ['Parser & Normalizer', 'Resolver', 'Validator', 'Scheduler (DAG)', 'Kernel & Event Bus', 'Policy & Review'] },
    { title: 'Generation', sub: 'What NAEOS produces', color: C.violet, items: ['Generator', 'Language Adapters', 'Renderers', 'AI Compiler', 'Test Runner'] },
    { title: 'Output', sub: 'What you ship', color: C.purple, items: ['Code · 5 languages', 'Configs & Docs', 'AI Instruction Sets', 'Artifacts & Audit Trail', 'MCP Server'] },
  ];
  let body = '';
  const cw = 340, gap = 60, x0 = 60, y0 = 300, ch = 500;
  cols.forEach((col, i) => {
    const x = x0 + i * (cw + gap);
    body += `<rect x="${x}" y="${y0}" width="${cw}" height="${ch}" rx="12" fill="${C.surface}" stroke="${C.border}" stroke-width="1"/>
    <rect x="${x}" y="${y0}" width="${cw}" height="8" rx="4" fill="${col.color}"/>
    <text x="${x + 24}" y="${y0 + 62}" font-family="${SANS}" font-size="30" font-weight="700" fill="${col.color}">${col.title}</text>
    <text x="${x + 24}" y="${y0 + 92}" font-family="${SANS}" font-size="18" fill="${C.dim}">${col.sub}</text>
    <line x1="${x + 24}" y1="${y0 + 112}" x2="${x + cw - 24}" y2="${y0 + 112}" stroke="${C.borderLight}"/>`;
    col.items.forEach((item, j) => {
      const iy = y0 + 148 + j * 56;
      body += `<circle cx="${x + 30}" cy="${iy - 6}" r="4" fill="${col.color}"/>
      <text x="${x + 46}" y="${iy}" font-family="${SANS}" font-size="21" fill="${C.text}">${esc(item)}</text>`;
    });
    if (i < 3) {
      body += `<text x="${x + cw + gap / 2}" y="${y0 + ch / 2 + 12}" font-family="${SANS}" font-size="40" font-weight="700" fill="${C.cyan}" text-anchor="middle">→</text>`;
    }
  });
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="900" viewBox="0 0 1600 900">
  <defs>${GRAD}</defs>
  <rect width="1600" height="900" fill="${C.bg}"/>
  ${gridSVG(1600, 900)}
  <text x="60" y="110" font-family="${SANS}" font-size="54" font-weight="800" fill="${C.text}">One specification. A complete engineering pipeline.</text>
  <text x="60" y="170" font-family="${SANS}" font-size="24" fill="${C.muted}">Parse → validate → build the model → schedule → generate → review — every step traceable.</text>
  ${body}
  <rect x="60" y="830" width="1480" height="56" rx="10" fill="${C.panel}" stroke="${C.border}" stroke-width="1"/>
  <circle cx="96" cy="858" r="6" fill="${C.green}"/>
  <text x="120" y="866" font-family="${SANS}" font-size="22" fill="${C.text}">Intent → Implementation</text>
  <text x="420" y="866" font-family="${SANS}" font-size="22" fill="${C.muted}">— fully traceable and auditable, from spec to shipped code and AI context.</text>
  </svg>`;
  await render(svg, '02-pipeline-diagram.png', 1600, 900);
}

/* ------------------------------------------------------------------ */
/* 4. CLI pipeline screenshot 1600x900                                 */
/* ------------------------------------------------------------------ */
async function cli() {
  const lines = [
    [{ s: '$ naeos run --config config.yaml --input-file spec-full.yaml', c: C.text }],
    [{ s: 'INFO ', c: C.blue }, { s: 'parsing specification (2407 bytes)', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'normalizing specification', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'resolving cross-references', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'building NEIR model', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'building execution graph', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'scheduling 9 tasks', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'generating artifacts', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'running language adapters', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'reviewing 124 artifacts', c: C.muted }],
    [{ s: 'INFO ', c: C.blue }, { s: 'writing 124 artifacts to ./out', c: C.muted }],
    [{ s: '✓ pipeline complete — 124 artifacts · 11 tasks · 124 reviews', c: C.green }],
    [{ s: 'mode=development  verbose=true  output_dir=./out', c: C.dim }],
  ];
  let y = 108;
  let content = '';
  for (const parts of lines) {
    const r = termLine(28, y, parts, 22, 38);
    content += r.svg;
    y = r.next;
  }
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="900" viewBox="0 0 1600 900">
  <defs>${GRAD}</defs>
  <rect width="1600" height="900" fill="${C.bg}"/>
  ${gridSVG(1600, 900)}
  <text x="180" y="90" font-family="${SANS}" font-size="42" font-weight="700" fill="${C.text}">Run the pipeline</text>
  <text x="180" y="135" font-family="${SANS}" font-size="22" fill="${C.muted}">One spec in — validated artifacts, code, docs, and AI context out.</text>
  <g transform="translate(180,180)">
    ${window('naeos run — pipeline.log', 1240, 640)}
    ${content}
    <rect x="28" y="586" width="1184" height="1" fill="${C.border}"/>
    <text x="28" y="622" font-family="${MONO}" font-size="18" fill="${C.dim}">naeos run --watch  →  re-runs on every spec change</text>
  </g>
  </svg>`;
  await render(svg, '03-cli-pipeline.png', 1600, 900);
}

/* ------------------------------------------------------------------ */
/* 5. AI compiler screenshot 1600x900                                  */
/* ------------------------------------------------------------------ */
async function ai() {
  const targets = [
    ['GitHub Copilot', '.github/copilot-instructions.md'],
    ['Claude Code', 'CLAUDE.md'],
    ['Cursor', '.cursorrules'],
    ['Gemini CLI', '.gemini/CONFIG.md'],
    ['Codex', 'AGENTS.md'],
    ['OpenCode', 'AGENTS.md'],
  ];
  let content = '';
  let y = 108;
  const P = (parts) => { const r = termLine(28, y, parts, 22, 38); content += r.svg; y = r.next; };
  P([{ s: '$ naeos context --input-file spec-full.yaml', c: C.text }]);
  P([{ s: '# e-commerce-platform — AI Context Bundle', c: C.cyan }]);
  P([{ s: '## Summary', c: C.purple }]);
  P([{ s: 'Project: e-commerce-platform · Modules: 5 · Services: 2 · Languages: go, typescript', c: C.muted }]);
  P([{ s: '## Modules', c: C.purple }]);
  P([{ s: '- ', c: C.dim }, { s: 'auth', c: C.cyan }, { s: '  Authentication and authorization module  (deps: core, user)', c: C.muted }]);
  P([{ s: '- ', c: C.dim }, { s: 'order', c: C.cyan }, { s: '  Order processing module  (deps: core, user, payment)', c: C.muted }]);
  P([{ s: '## AI Targets — compiled instruction sets', c: C.purple }]);
  for (const [name, file] of targets) {
    P([{ s: '✓ ', c: C.green }, { s: name.padEnd(14), c: C.text }, { s: '→ ', c: C.dim }, { s: file, c: C.cyan }]);
  }
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="900" viewBox="0 0 1600 900">
  <defs>${GRAD}</defs>
  <rect width="1600" height="900" fill="${C.bg}"/>
  ${gridSVG(1600, 900)}
  <text x="180" y="90" font-family="${SANS}" font-size="42" font-weight="700" fill="${C.text}">Compile for AI</text>
  <text x="180" y="135" font-family="${SANS}" font-size="22" fill="${C.muted}">The engineering model becomes instruction sets — every AI tool works from the truth.</text>
  <g transform="translate(180,180)">
    ${window('naeos context — AI context bundle', 1240, 640)}
    ${content}
    <rect x="28" y="586" width="1184" height="1" fill="${C.border}"/>
    <text x="28" y="622" font-family="${MONO}" font-size="18" fill="${C.dim}">6 adapters · 1 model · MCP server for AI agents</text>
  </g>
  </svg>`;
  await render(svg, '04-ai-compiler.png', 1600, 900);
}

/* ------------------------------------------------------------------ */
/* 6. LSP editor screenshot 1600x900                                   */
/* ------------------------------------------------------------------ */
async function lsp() {
  const gutter = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17];
  const code = [
    [{ s: '# spec.yaml — e-commerce-platform', c: C.comment }],
    [{ s: 'project: ', c: C.cyan }, { s: 'e-commerce-platform', c: C.string }],
    [{ s: '', c: '' }],
    [{ s: 'modules:', c: C.keyword }],
    [{ s: '  - name: ', c: C.cyan }, { s: 'auth', c: C.string }, { s: '        ', c: C.comment }, { s: '# ← hover for LSP info', c: C.comment }],
    [{ s: '    path: ', c: C.cyan }, { s: './internal/auth', c: C.string }],
    [{ s: '    dependencies: ', c: C.cyan }, { s: '[core, user]', c: C.string }],
    [{ s: '  - name: ', c: C.cyan }, { s: 'order', c: C.string }],
    [{ s: '    path: ', c: C.cyan }, { s: './internal/order', c: C.string }],
    [{ s: '', c: '' }],
    [{ s: 'services:', c: C.keyword }],
    [{ s: '  - name: ', c: C.cyan }, { s: 'api-gateway', c: C.string }],
    [{ s: '    kind: ', c: C.cyan }, { s: 'http', c: C.string }, { s: '          ', c: C.comment }, { s: '# ✓ valid kind', c: C.comment }],
    [{ s: '    port: ', c: C.cyan }, { s: '8080', c: C.code }],
    [{ s: '', c: '' }],
    [{ s: 'architecture:', c: C.keyword }],
    [{ s: '  pattern: ', c: C.cyan }, { s: 'hexagonal', c: C.string }],
  ];
  let codeSVG = '';
  const gx = 28, gy = 106, lh = 30;
  gutter.forEach((n, i) => {
    codeSVG += t(gx, gy + i * lh + 14, String(n), 13, C.dim, MONO);
  });
  code.forEach((segments, i) => {
    let s = `<text x="${gx + 46}" y="${gy + i * lh + 14}" font-family="${MONO}" font-size="15">`;
    for (const p of segments) s += `<tspan fill="${p.c}">${esc(p.s)}</tspan>`;
    s += '</text>';
    codeSVG += s;
  });
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="900" viewBox="0 0 1600 900">
  <defs>${GRAD}</defs>
  <rect width="1600" height="900" fill="${C.bg}"/>
  ${gridSVG(1600, 900)}
  <text x="400" y="80" font-family="${SANS}" font-size="42" font-weight="700" fill="${C.text}">A spec editor that understands your system</text>
  <text x="400" y="125" font-family="${SANS}" font-size="22" fill="${C.muted}">NEIR-aware LSP — autocomplete, diagnostics, hover docs, go-to-definition for .naeos.yaml</text>

  <!-- activity bar -->
  <rect x="40" y="170" width="56" height="640" fill="${C.surface}" stroke="${C.border}" stroke-width="1"/>
  <rect x="52" y="186" width="32" height="32" rx="6" fill="${C.cyan}" opacity="0.9"/>
  <rect x="52" y="240" width="32" height="32" rx="6" fill="none" stroke="${C.borderLight}" stroke-width="2"/>
  <rect x="52" y="294" width="32" height="32" rx="6" fill="none" stroke="${C.borderLight}" stroke-width="2"/>
  <rect x="52" y="348" width="32" height="32" rx="6" fill="none" stroke="${C.borderLight}" stroke-width="2"/>
  <rect x="52" y="402" width="32" height="32" rx="6" fill="none" stroke="${C.borderLight}" stroke-width="2"/>

  <!-- sidebar -->
  <rect x="96" y="170" width="300" height="640" fill="${C.surface}" stroke="${C.border}" stroke-width="1"/>
  <text x="116" y="208" font-family="${SANS}" font-size="13" font-weight="700" fill="${C.dim}">EXPLORER</text>
  ${t(116, 248, '◉ e-commerce-platform', 15, C.text)}
  ${t(136, 282, '▸ spec-full.yaml', 15, C.cyan)}
  ${t(136, 316, '▸ modules/', 15, C.muted)}
  ${t(156, 350, '▸ auth', 14, C.dim)}
  ${t(156, 384, '▸ core', 14, C.dim)}
  ${t(156, 418, '▸ order', 14, C.dim)}
  ${t(136, 452, '▸ services/', 15, C.muted)}
  ${t(136, 486, '▸ docs/', 15, C.muted)}

  <!-- editor -->
  <rect x="396" y="170" width="1164" height="640" fill="#0e0e1a" stroke="${C.border}" stroke-width="1"/>
  <rect x="396" y="170" width="1164" height="40" fill="${C.panel}"/>
  <rect x="396" y="170" width="260" height="40" fill="#141424"/>
  <rect x="396" y="209" width="260" height="2" fill="${C.cyan}"/>
  <text x="416" y="196" font-family="${SANS}" font-size="14" fill="${C.text}">spec-full.yaml</text>
  <circle cx="646" cy="190" r="9" fill="${C.borderLight}"/><text x="643" y="194" font-family="${SANS}" font-size="13" fill="${C.muted}">×</text>
  <text x="676" y="196" font-family="${SANS}" font-size="14" fill="${C.dim}">+</text>

  <!-- code -->
  <g transform="translate(396,170)">
    ${codeSVG}
    <!-- gutter underline: 'modules' line 4 + hover on line 5 -->
    <rect x="0" y="${gy + 4 * lh}" width="1164" height="${lh}" fill="#08d6ff" opacity="0.05"/>
    <rect x="0" y="${gy + 5 * lh}" width="1164" height="${lh}" fill="#0f0f22"/>
    <text x="28" y="${gy + 5 * lh + 14}" font-family="${MONO}" font-size="13" fill="${C.cyan}">5</text>

    <!-- hover tooltip -->
    <rect x="600" y="${gy + 5 * lh - 66}" width="430" height="52" rx="8" fill="#141428" stroke="${C.borderLight}" stroke-width="1"/>
    <text x="620" y="${gy + 5 * lh - 44}" font-family="${SANS}" font-size="14" font-weight="600" fill="${C.cyan}">module auth</text>
    <text x="620" y="${gy + 5 * lh - 24}" font-family="${SANS}" font-size="13" fill="${C.muted}">dependencies: core, user · 2 services reference it</text>

    <!-- diagnostic ✓ -->
    <line x1="640" y1="${gy + 13 * lh + 26}" x2="712" y2="${gy + 13 * lh + 26}" stroke="${C.green}" stroke-width="2"/>
    <text x="640" y="${gy + 13 * lh + 45}" font-family="${SANS}" font-size="12" fill="${C.green}">✓ valid kind · no diagnostics</text>
    <line x1="640" y1="${gy + 17 * lh + 26}" x2="760" y2="${gy + 17 * lh + 26}" stroke="${C.green}" stroke-width="2"/>
    <text x="640" y="${gy + 17 * lh + 45}" font-family="${SANS}" font-size="12" fill="${C.green}">✓ pattern hexagonal — known architecture</text>
  </g>

  <!-- status bar -->
  <rect x="40" y="824" width="1520" height="40" fill="${C.surface}" stroke="${C.border}" stroke-width="1"/>
  <circle cx="64" cy="844" r="5" fill="${C.green}"/>
  <text x="80" y="849" font-family="${SANS}" font-size="14" fill="${C.muted}">LSP connected · NEIR model loaded</text>
  <text x="1360" y="849" font-family="${SANS}" font-size="14" fill="${C.muted}">0 errors · 0 warnings</text>
  </svg>`;
  await render(svg, '05-lsp-editor.png', 1600, 900);
}

/* ------------------------------------------------------------------ */
function pathToData(p) {
  return 'data:image/svg+xml;base64,' + fs.readFileSync(p).toString('base64');
}

(async () => {
  await logo();
  await hero();
  await pipeline();
  await cli();
  await ai();
  await lsp();
  console.log('done');
})().catch((e) => { console.error(e); process.exit(1); });