const sharp = require('sharp');
const fs = require('fs');
const path = require('path');

const OUT = __dirname;
const BRAND = path.resolve(__dirname, '../../../brand');

const C = {
  bg: '#05050a', surface: '#0a0a14', panel: '#111122', border: '#1a1a33', borderLight: '#2a2a44',
  text: '#f0f0f0', muted: '#9999b0', dim: '#666680',
  cyan: '#08d6ff', blue: '#3a7cf8', violet: '#7c4dff', purple: '#9333ea',
  green: '#27c93f', yellow: '#ffaa00',
};
const SANS = 'Ubuntu Sans, DejaVu Sans, sans-serif';
const MONO = 'DejaVu Sans Mono, monospace';
const GRAD = 'linearGradient(id="ngrad",x1="0.5",y1="0",x2="0.5",y2="1"){stop(offset="0%",stop-color="#08d6ff");stop(offset="25%",stop-color="#3a7cf8");stop(offset="50%",stop-color="#6056f5");stop(offset="75%",stop-color="#7c4dff");stop(offset="100%",stop-color="#9333ea")}';

function esc(s) { return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;'); }
function pathToData(p) { return 'data:image/svg+xml;base64,' + fs.readFileSync(p).toString('base64'); }
function glow(cx, cy, r, color, opacity) {
  return `<radialGradient id="gl${cx}${cy}" cx="0.5" cy="0.5" r="0.5"><stop offset="0%" stop-color="${color}" stop-opacity="${opacity}"/><stop offset="100%" stop-color="${color}" stop-opacity="0"/></radialGradient><circle cx="${cx}" cy="${cy}" r="${r}" fill="url(#gl${cx}${cy})"/>`;
}
function gridSVG(w, h, step = 96) {
  let g = '';
  for (let x = 0; x <= w; x += step) g += `<line x1="${x}" y1="0" x2="${x}" y2="${h}" stroke="#0d0d1c" stroke-width="1"/>`;
  for (let y = 0; y <= h; y += step) g += `<line x1="0" y1="${y}" x2="${w}" y2="${y}" stroke="#0d0d1c" stroke-width="1"/>`;
  return `<g opacity="0.55">${g}</g>`;
}
function chip(x, y, label, fill) {
  const w = label.length * 15 + 34;
  return `<rect x="${x}" y="${y - 26}" width="${w}" height="36" rx="18" fill="${C.panel}" stroke="${C.borderLight}" stroke-width="1.5"/><text x="${x + w / 2}" y="${y + 4}" text-anchor="middle" font-family="${SANS}" font-size="18" font-weight="600" fill="${fill}">${esc(label)}</text>`;
}

async function render(svg, name, w, h) {
  const file = path.join(OUT, name);
  await sharp(Buffer.from(svg)).png().toFile(file);
  console.log(`${name}  ${w}x${h}`);
}

/* ------------------------------------------------------------------ */
/* Server icon 512x512 — brand lockup (Discord rounds it itself)        */
/* ------------------------------------------------------------------ */
async function icon() {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="512" height="512" viewBox="0 0 460 460">${fs.readFileSync(path.join(BRAND, 'logo.svg'), 'utf8').replace(/^<svg[^>]*>/, '').replace(/<\/svg>$/, '')}</svg>`;
  await render(svg, 'server-icon.png', 512, 512);
}

/* ------------------------------------------------------------------ */
/* Server banner 960x540                                               */
/* ------------------------------------------------------------------ */
async function banner() {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="960" height="540" viewBox="0 0 960 540">
  <defs>${GRAD}
    <linearGradient id="txtgrad" x1="0" y1="0" x2="1" y2="0"><stop offset="0%" stop-color="#08d6ff"/><stop offset="100%" stop-color="#9333ea"/></linearGradient>
  </defs>
  <rect width="960" height="540" fill="${C.bg}"/>
  ${gridSVG(960, 540)}
  ${glow(180, 120, 260, '#08d6ff', 0.16)}
  ${glow(480, 260, 240, '#7c4dff', 0.12)}
  <image href="${pathToData(path.join(BRAND, 'logo-mark.svg'))}" x="70" y="150" width="160" height="160"/>
  <text x="250" y="215" font-family="${SANS}" font-size="72" font-weight="800" fill="url(#txtgrad)">NAEOS Community</text>
  <text x="250" y="285" font-family="${SANS}" font-size="30" font-weight="600" fill="${C.text}">Specify Once. Build Anywhere.</text>
  <text x="250" y="330" font-family="${SANS}" font-size="20" fill="${C.muted}">Declarative engineering · Open source · For humans and AI</text>
  <rect x="70" y="450" width="820" height="1" fill="${C.border}"/>
  <text x="70" y="492" font-family="${MONO}" font-size="17" fill="${C.dim}">$ naeos run --config config.yaml --input-file spec.yaml  →  ✓ 124 artifacts</text>
  </svg>`;
  await render(svg, 'server-banner.png', 960, 540);
}

/* ------------------------------------------------------------------ */
/* Invite splash 1920x1080                                             */
/* ------------------------------------------------------------------ */
async function splash() {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1920" height="1080" viewBox="0 0 1920 1080">
  <defs>${GRAD}
    <linearGradient id="txtgrad" x1="0" y1="0" x2="1" y2="0"><stop offset="0%" stop-color="#08d6ff"/><stop offset="100%" stop-color="#9333ea"/></linearGradient>
  </defs>
  <rect width="1920" height="1080" fill="${C.bg}"/>
  ${gridSVG(1920, 1080, 120)}
  ${glow(340, 240, 420, '#08d6ff', 0.15)}
  ${glow(980, 560, 460, '#7c4dff', 0.13)}
  <image href="${pathToData(path.join(BRAND, 'logo-mark.svg'))}" x="140" y="300" width="300" height="300"/>
  <text x="470" y="415" font-family="${SANS}" font-size="150" font-weight="800" fill="url(#txtgrad)">NAEOS Community</text>
  <text x="474" y="520" font-family="${SANS}" font-size="58" font-weight="600" fill="${C.text}">Specify Once. Build Anywhere.</text>
  <text x="474" y="600" font-family="${SANS}" font-size="34" fill="${C.muted}">A declarative engineering platform that turns one specification into validated,</text>
  <text x="474" y="650" font-family="${SANS}" font-size="34" fill="${C.muted}">multi-language software — for humans and AI.</text>
  ${chip(474, 760, 'Website · naeos.dev', C.cyan)}
  ${chip(770, 760, 'GitHub · NAEOS-foundation/naeos', C.green)}
  ${chip(1160, 760, 'Docs · docs.naeos.dev', C.violet)}
  ${chip(1450, 760, 'Open Source · Apache 2.0', C.yellow)}
  <rect x="140" y="880" width="1640" height="2" fill="${C.border}"/>
  <text x="140" y="955" font-family="${MONO}" font-size="30" fill="${C.dim}">$ curl -fsSL https://naeos.dev/install.sh | sh</text>
  <text x="140" y="1010" font-family="${MONO}" font-size="30" fill="${C.dim}">$ naeos run --config config.yaml --input-file spec.yaml   →   ✓ 124 artifacts · 11 tasks</text>
  </svg>`;
  await render(svg, 'invite-splash.png', 1920, 1080);
}

(async () => {
  await icon();
  await banner();
  await splash();
  console.log('done');
})().catch((e) => { console.error(e); process.exit(1); });
