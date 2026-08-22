# NAEOS — Conference Paper (Industry/Tool Track)

Distilled 2-page-core + evaluation paper for international SE venues, generated from the full monograph (`thesis/`).

## Files
- `main.tex` — ACM `sigconf` (`nonacm`), pdfLaTeX
- `refs.bib` — bibliography

## Compile

**Overleaf (easiest):** upload both files → compiler pdfLaTeX → done.

**Local:**
```bash
pdflatex main.tex && bibtex main && pdflatex main.tex && pdflatex main.tex
```

**No LaTeX installed?**
```bash
# tectonic (single binary)
tectonic main.tex
```
or `docker run --rm -v "$PWD":/work -w /work ghcr.io/xu-cheng/texlive-full pdflatex main.tex`.

## Venue mapping

| Venue | Track | Fit | Page limit* |
|---|---|---|---|
| ICSE-SEIP | Software Engineering in Practice | Best fit — evaluated industrial tool | ~10 pp |
| FSE-Industry | Industry track | Strong fit — open-source runtime | ~8–12 pp |
| ASE-Industry | Industry track | Strong fit | ~8 pp |
| SPLC | Tools/demos | Product-line angle (conditional modules, profiles) | varies |

*Check the current CFP; adjust `\documentclass[...]` and trim §5 if needed.

## Pre-submission checklist
- [ ] Add co-authors if applicable; ORCID in `\author{}` block
- [ ] Anonymization: industry tracks usually **single-blind** — keep name; double-check CFP
- [ ] Re-run experiments from `thesis/98-appendix.md`, refresh Table 2 numbers
- [ ] Complete remaining RQ1 protocol steps (warm-cache hit rates, mutation granularity) to strengthen claims
- [ ] Artifact badge: link Zenodo DOI 10.5281/zenodo.22060578 in camera-ready
- [ ] Page-budget check: current draft ≈ 4–5 pp; expand Evaluation with scaling plots before submission to reach limits comfortably
