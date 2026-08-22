# NAEOS Thesis

[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.22060578.svg)](https://doi.org/10.5281/zenodo.22060578)
[![DOI v1.1.0](https://zenodo.org/badge/DOI/10.5281/zenodo.22061988.svg)](https://doi.org/10.5281/zenodo.22061988)

Technical monograph on the NAEOS declarative engineering runtime — prepared to international publication standards (IEEE-style citations, artifact-evaluation appendix, explicit novelty claims), framed for **independent publication** (Zenodo DOI / arXiv preprint / industry-track papers) rather than an academic degree.

> **Title:** *NAEOS: A Declarative Engineering Runtime for Specification-Driven, AI-Native Software Construction*

## Document structure

| File | Part | Content |
|---|---|---|
| `FRONTMATTER.md` | — | Title page, declaration, extended abstract, keywords/MSC |
| `00-abstract.md` | — | Abstract & keywords |
| `01-introduction.md` | 1 | Background, problem statement, research questions, novelty-tagged contributions |
| `02-background.md` | 2 | Related work: MDE/MDA, IaC, compiler IRs, AI-assisted development |
| `03-architecture.md` | 3 | Layered architecture, compilation pipeline, kernel, data flow |
| `04-core-design.md` | 4 | Spec Language v2, NEIR, executable governance, brokers, WASM sandbox |
| `05-ai-integration.md` | 5 | AI compiler, MCP server, LSP, prompt library, nondeterminism containment |
| `06-evaluation.md` | 6 | Executed determinism experiment, replicated benchmarks, quality assurance, threats to validity |
| `07-conclusion.md` | 7 | Conclusions per RQ, limitations, future work |
| `98-appendix.md` | App. A | Artifact evaluation & full reproduction protocol |
| `99-references.md` | — | IEEE-style references (real canonical citations; one flagged entry to complete) |

## Status / TODO markers

Executed experiments (results already written into Chapter 6):

- Determinism: repeated `naeos run` executions over `examples/spec-full.yaml`
  produced 79 byte-identical artifacts (SHA-256 pairwise comparison), including
  warm-cache runs at 2.7× lower wall time (151 ms → 56 ms).
- Mutation granularity: non-emitted field change → 0 artifacts; project-name change →
  exactly its downstream set (10 + 1 rename); surfaced a template-propagation
  limitation (service port) as an honest finding.
- Benchmarks: full `pkg/pipeline` suite replicated on AMD EPYC 7763, including
  Small/Medium/Large scaling and parallel variants.

Remaining `[TODO]` item in Chapter 6:

- Per-stage hit-rate breakdown via `--profile` (requires profiling instrumentation).

Before publication: the Zenodo DOI is minted and linked (10.5281/zenodo.22060578).
Optionally add an arXiv preprint identifier later. All 38 references are complete
canonical citations ([R27] = Hou et al., ACM TOSEM 2024, the standard LLM4SE survey).
Author details, biography, declaration signature, and acknowledgments are already
filled in for Bayu Priatno.
