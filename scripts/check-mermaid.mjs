#!/usr/bin/env node
// Validate all ```mermaid blocks in WHITEPAPER*.md (and any files passed as args).
// Usage: node scripts/check-mermaid.mjs [file...]
// Requires: mermaid and jsdom installed in the current working directory
//           (e.g. `npm install --prefix . mermaid jsdom`), or importable
//           relative to process.cwd().

import { readFileSync, existsSync } from 'fs';
import { createRequire } from 'module';

const require = createRequire(process.cwd() + '/package.json');

const { JSDOM } = require('jsdom');
const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', {
  pretendToBeVisual: true,
});
globalThis.window = dom.window;
globalThis.document = dom.window.document;
Object.defineProperty(globalThis, 'navigator', {
  value: dom.window.navigator,
  configurable: true,
});

const mermaidModule = require('mermaid');
const mermaid = mermaidModule.default || mermaidModule;
await mermaid.initialize({ startOnLoad: false, securityLevel: 'loose' });

const defaults = ['WHITEPAPER.md', 'WHITEPAPER-EN.md'];
const files = process.argv.length > 2
  ? process.argv.slice(2)
  : defaults.filter((f) => existsSync(f));

if (files.length === 0) {
  console.error('No whitepaper files found in the current directory.');
  process.exit(2);
}

const errors = [];
for (const file of files) {
  const src = readFileSync(file, 'utf-8');
  const blocks = [...src.matchAll(/```mermaid\s*\n(.*?)```/gs)];
  console.log(file, blocks.length, 'diagram(s)');
  let i = 0;
  for (const b of blocks) {
    i++;
    try {
      await mermaid.parse(b[1]);
      console.log('  ok', i);
    } catch (e) {
      const msg = String(e.message || e).split('\n')[0].slice(0, 140);
      errors.push(`${file}#${i}: ${msg}`);
      console.log('  FAIL', i, msg);
    }
  }
}
if (errors.length) {
  console.error('\nERRORS:\n' + errors.join('\n'));
  process.exit(1);
}
console.log('\nALL DIAGRAMS VALID');
