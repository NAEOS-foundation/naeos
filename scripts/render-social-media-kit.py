#!/usr/bin/env python3
"""Render NAEOS Social Media Kit SVG templates to PNG and build a contact sheet.
Requires cairosvg, Pillow. Outputs are generated artifacts, not release assets.
"""
from __future__ import annotations
import hashlib, json
from pathlib import Path
import cairosvg
from PIL import Image, ImageDraw, ImageFont
ROOT=Path(__file__).resolve().parents[1]
KIT=ROOT/"brand"/"marketing"/"social-media"
TEMPLATES=KIT/"templates"
PREVIEWS=KIT/"previews"
SHEET=KIT/"SOCIAL-MEDIA-KIT-PREVIEW.png"
NAMES=["linkedin-post-1x1.svg","linkedin-post-4x5.svg","carousel-cover-4x5.svg","carousel-content-4x5.svg","quote-insight-1x1.svg","announcement-4x5.svg","linkedin-company-banner.svg"]
def sha256(path):
    h=hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda:f.read(65536),b""): h.update(block)
    return h.hexdigest()
def create_sheet(items):
    cw,ch=520,390; tw,th=480,330; cols=2; rows=(len(items)+cols-1)//cols
    sheet=Image.new("RGB",(cols*cw,rows*ch),"#0B0F19"); draw=ImageDraw.Draw(sheet)
    try: font=ImageFont.truetype("DejaVuSans.ttf",18)
    except OSError: font=ImageFont.load_default()
    for i,item in enumerate(items):
        x,y=(i%cols)*cw,(i//cols)*ch
        draw.text((x+20,y+12),Path(item["source"]).name,fill="#F7FAFF",font=font)
        with Image.open(ROOT/item["preview"]) as im:
            im=im.convert("RGB"); im.thumbnail((tw,th),Image.Resampling.LANCZOS)
            sheet.paste(im,(x+(cw-im.width)//2,y+48+(th-im.height)//2))
    sheet.save(SHEET,format="PNG",optimize=True)
def main():
    PREVIEWS.mkdir(parents=True,exist_ok=True); items=[]
    for name in NAMES:
        source=TEMPLATES/name
        if not source.is_file(): raise FileNotFoundError(f"Missing template: {source}")
        output=PREVIEWS/(source.stem+".png")
        cairosvg.svg2png(url=str(source),write_to=str(output))
        with Image.open(output) as im: im.verify()
        with Image.open(output) as im:
            items.append({"source":source.relative_to(ROOT).as_posix(),"preview":output.relative_to(ROOT).as_posix(),"width":im.width,"height":im.height,"sha256":sha256(output)})
    create_sheet(items)
    (PREVIEWS/"render-manifest.json").write_text(json.dumps(items,indent=2)+"\n",encoding="utf-8")
    print(f"Rendered {len(items)} templates; contact sheet: {SHEET.relative_to(ROOT)}")
    for item in items: print(f'{item["preview"]} | {item["width"]}x{item["height"]} | SHA256 {item["sha256"]}')
if __name__=="__main__": main()
