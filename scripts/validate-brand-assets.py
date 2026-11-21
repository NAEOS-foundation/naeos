#!/usr/bin/env python3
"""Validate NAEOS brand assets and produce a deterministic inventory/checksum manifest."""
from __future__ import annotations
import argparse, hashlib, json, re, sys
from pathlib import Path
from xml.etree import ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
SCAN_ROOTS = ("brand", "site/public")
ASSET_EXTS = {".svg", ".png", ".jpg", ".jpeg", ".webp", ".gif", ".ico"}
PLACEHOLDER_RE = re.compile(r"\b(?:placeholder|replace[-_ ]me|todo|tbd)\b", re.I)
FORBIDDEN_TAGS = {"script", "foreignObject"}
EXTERNAL_REF_RE = re.compile(r"^(?:https?:|//|data:|javascript:)", re.I)

def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(65536), b""):
            digest.update(chunk)
    return digest.hexdigest()

def svg_validate(path: Path) -> list[str]:
    issues = []
    try:
        root = ET.parse(path).getroot()
    except (ET.ParseError, OSError) as exc:
        return [f"SVG parse error: {exc}"]
    if root.tag.split("}")[-1] != "svg":
        issues.append("root element is not <svg>")
    view_box = root.attrib.get("viewBox")
    if not view_box:
        issues.append("missing viewBox")
    else:
        try:
            values = [float(v) for v in re.split(r"[ ,]+", view_box.strip())]
            if len(values) != 4 or values[2] <= 0 or values[3] <= 0:
                issues.append("viewBox must contain four numbers and positive width/height")
        except ValueError:
            issues.append("viewBox contains non-numeric values")
    for element in root.iter():
        tag = element.tag.split("}")[-1]
        if tag in FORBIDDEN_TAGS:
            issues.append(f"forbidden element <{tag}>")
        for attr, value in element.attrib.items():
            local = attr.split("}")[-1].lower()
            if local.startswith("on"):
                issues.append(f"event-handler attribute {attr} is forbidden")
            if local in {"href", "src"} and EXTERNAL_REF_RE.match(value.strip()):
                issues.append(f"external resource reference in {attr}: {value[:80]}")
            if local == "style" and "url(" in value.lower():
                for url in re.findall(r"url\((.*?)\)", value, re.I):
                    cleaned = url.strip(" '").strip('"')
                    if EXTERNAL_REF_RE.match(cleaned):
                        issues.append(f"external CSS resource: {cleaned[:80]}")
    text = path.read_text(encoding="utf-8", errors="replace")
    if PLACEHOLDER_RE.search(text):
        issues.append("placeholder marker found")
    return issues

def inventory():
    entries, errors = [], []
    for root_name in SCAN_ROOTS:
        scan_root = ROOT / root_name
        if not scan_root.exists():
            continue
        for path in sorted(scan_root.rglob("*")):
            if not path.is_file() or path.suffix.lower() not in ASSET_EXTS:
                continue
            rel = path.relative_to(ROOT).as_posix()
            item = {"path": rel, "format": path.suffix.lower().lstrip("."),
                    "size_bytes": path.stat().st_size, "sha256": sha256_file(path)}
            if path.suffix.lower() == ".svg":
                issues = svg_validate(path)
                item["svg_validation"] = "PASS" if not issues else "FAIL"
                errors.extend(f"{rel}: {issue}" for issue in issues)
                try:
                    root = ET.parse(path).getroot()
                    item["viewBox"] = root.attrib.get("viewBox")
                    item["width"] = root.attrib.get("width")
                    item["height"] = root.attrib.get("height")
                except ET.ParseError:
                    pass
            entries.append(item)
    return entries, errors

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", default="brand/exports/asset-inventory.json")
    parser.add_argument("--checksums", default="brand/exports/asset-checksums.sha256")
    parser.add_argument("--write", action="store_true")
    parser.add_argument("--check-generated", action="store_true")
    args = parser.parse_args()
    assets, issues = inventory()
    manifest = json.dumps({"schema_version":"1.0","generator":"scripts/validate-brand-assets.py","assets":assets}, indent=2, sort_keys=True) + "\n"
    checksums = "".join(f"{a['sha256']}  {a['path']}\n" for a in assets)
    if args.write:
        (ROOT / args.manifest).write_text(manifest, encoding="utf-8")
        (ROOT / args.checksums).write_text(checksums, encoding="utf-8")
        print(f"Wrote {len(assets)} inventory entries.")
    if args.check_generated:
        for path, expected in ((args.manifest, manifest), (args.checksums, checksums)):
            target = ROOT / path
            if not target.exists() or target.read_text(encoding="utf-8") != expected:
                issues.append(f"generated file stale or missing: {path}")
    fails = sum(a.get("svg_validation") == "FAIL" for a in assets)
    print(f"Asset inventory: {len(assets)} files; SVG failures: {fails}")
    if issues:
        for issue in issues: print("ERROR:", issue, file=sys.stderr)
        return 1
    print("Brand asset validation passed.")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
