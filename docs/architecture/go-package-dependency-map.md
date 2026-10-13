# Go package dependency map

Use the repository's Go toolchain to generate a current package-import graph before proposing package moves or consolidations.

## Generate

From the repository root:

```bash
./scripts/go-package-graph.sh > /tmp/naeos-go-packages.dot
```

If Graphviz is installed, render an SVG for inspection:

```bash
dot -Tsvg /tmp/naeos-go-packages.dot -o /tmp/naeos-go-packages.svg
```

The script can also write directly to a supplied output path:

```bash
./scripts/go-package-graph.sh /tmp/naeos-go-packages.dot
```

## Semantics and limitations

- Each edge points from the importing package to the imported package.
- Only direct imports between packages inside the NAEOS Go module are emitted. Third-party and standard-library dependencies are omitted to keep the graph focused on internal coupling.
- The graph reflects packages discovered by `go list ./...`; test-only import edges are not represented.
- The script requires the Go toolchain and a successful `go list ./...`. Graphviz is optional and is not needed to generate DOT.
- Output is generated on demand and is intentionally not checked in: a committed snapshot could become stale.
- A graph is evidence for review, not an automatic refactoring recommendation. Inspect package responsibilities, interfaces, tests, and runtime behavior before changing boundaries.

## Safe review sequence

1. Generate the graph at the commit under review and retain the commit SHA with any findings.
2. Identify high-fan-in packages and verify their callers with targeted searches.
3. Check whether any proposed dependency inversion or package move changes authorization, runtime enforcement, evidence, persistence, CLI, or public documentation contracts.
4. Make one behavior-preserving change per PR and run the relevant package tests plus repository CI.
