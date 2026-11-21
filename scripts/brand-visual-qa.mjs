#!/usr/bin/env node
import fs from 'node:fs/promises';
import path from 'node:path';
import { chromium } from 'playwright';

const root = process.cwd();
const brandDir = path.join(root, 'brand');
const outputDir = path.join(root, 'brand-visual-qa-output');
const pngDir = path.join(outputDir, 'png');
const reportPath = path.join(outputDir, 'report.json');

function dimensions(svg) {
  const viewBox = svg.match(/viewBox=["']([^"']+)["']/i)?.[1];
  const vb = viewBox?.trim().split(/[ ,]+/).map(Number);
  const width = svg.match(/<svg[^>]*\bwidth=["']([^"']+)["']/i)?.[1];
  const height = svg.match(/<svg[^>]*\bheight=["']([^"']+)["']/i)?.[1];
  const parse = v => v && Number.parseFloat(v);
  return { width: parse(width) || (vb?.length === 4 ? vb[2] : 0), height: parse(height) || (vb?.length === 4 ? vb[3] : 0), viewBox };
}
async function walk(dir) {
  const out=[];
  for (const ent of await fs.readdir(dir,{withFileTypes:true})) {
    const p=path.join(dir,ent.name);
    if(ent.isDirectory()) out.push(...await walk(p));
    else if(ent.isFile() && ent.name.toLowerCase().endsWith('.svg')) out.push(p);
  }
  return out;
}
await fs.rm(outputDir,{recursive:true,force:true});
await fs.mkdir(pngDir,{recursive:true});
const files=(await walk(brandDir)).sort();
if(!files.length) throw new Error('No SVG files found under brand/');
const browser=await chromium.launch({headless:true});
const results=[];
try {
  for(const file of files) {
    const relative=path.relative(root,file);
    const svg=await fs.readFile(file,'utf8');
    const dim=dimensions(svg);
    const issues=[];
    if(!dim.width || !dim.height || !Number.isFinite(dim.width) || !Number.isFinite(dim.height)) issues.push('missing/invalid canvas dimensions');
    if(/<script\b/i.test(svg)) issues.push('script element present');
    if(/(?:href|xlink:href)=["']https?:\/\//i.test(svg)) issues.push('external linked resource (may be unavailable in isolated render)');
    const output=path.join(pngDir,relative.replace(/\.svg$/i,'.png'));
    await fs.mkdir(path.dirname(output),{recursive:true});
    if(!issues.includes('missing/invalid canvas dimensions')) {
      const page=await browser.newPage({viewport:{width:Math.ceil(dim.width),height:Math.ceil(dim.height)},deviceScaleFactor:1});
      await page.setContent('<!doctype html><html><head><style>html,body{margin:0;padding:0;width:'+dim.width+'px;height:'+dim.height+'px;overflow:hidden;background:transparent}</style></head><body>'+svg+'</body></html>',{waitUntil:'load'});
      await page.screenshot({path:output,omitBackground:true});
      await page.close();
    }
    results.push({file:relative,canvas:{width:dim.width,height:dim.height,viewBox:dim.viewBox},png:issues.length?null:path.relative(root,output),warnings:issues});
  }
} finally { await browser.close(); }
const report={schemaVersion:1,generatedAt:new Date().toISOString(),renderer:'Playwright Chromium',assetsScanned:results.length,assetsRendered:results.filter(x=>x.png).length,assetsWithWarnings:results.filter(x=>x.warnings.length).length,results};
await fs.writeFile(reportPath,JSON.stringify(report,null,2)+'\n');
const galleryItems=results.filter(x=>x.png).map(x=>'<figure><img src="png/'+path.relative(brandDir,path.join(root,x.png)).replaceAll(path.sep,'/')+'"><figcaption>'+x.file+' — '+x.canvas.width+'×'+x.canvas.height+'</figcaption></figure>').join('\n');
await fs.writeFile(path.join(outputDir,'gallery.html'),'<!doctype html><meta charset="utf-8"><title>NAEOS Brand Visual QA Gallery</title><style>body{font-family:Inter,Arial,sans-serif;background:#0b0f19;color:#f7faff;padding:24px}main{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:20px}figure{margin:0;padding:12px;background:#151b2b;border-radius:12px}img{width:100%;height:auto;background:repeating-conic-gradient(#ddd 0% 25%,#fff 0% 50%) 50%/20px 20px}figcaption{font-size:12px;overflow-wrap:anywhere;padding-top:8px}</style><h1>NAEOS Brand Visual QA</h1><p>Automated render preview. Human visual review is still required.</p><main>'+galleryItems+'</main>');
console.log(JSON.stringify({assetsScanned:report.assetsScanned,assetsRendered:report.assetsRendered,assetsWithWarnings:report.assetsWithWarnings,report:reportPath,gallery:path.join(outputDir,'gallery.html')},null,2));
if(results.some(x=>!x.png)) process.exitCode=1;
