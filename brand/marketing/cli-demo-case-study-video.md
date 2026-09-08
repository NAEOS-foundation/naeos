# CLI Demo Case Study Video

## Format

- Length: 75–90 seconds
- Audience: AI engineers, software architects, and platform engineers
- Primary CTA: Run the checked-in CLI demo
- Evidence source: `examples/demo-cli/`

## Storyboard and narration

| Time | Visual | Narration |
|------|--------|-----------|
| 0–8s | Show `spec.yaml` with the `auth` and `api` modules | “What if one engineering specification could become the starting point for code, documentation, and AI context?” |
| 8–18s | Highlight the gateway, dependency, architecture, and language fields | “This small NAEOS example describes modules, dependencies, an HTTP gateway, an architecture pattern, and two language targets.” |
| 18–30s | Run `./examples/demo-cli/run-demo.sh` | “The runnable demo starts by validating the specification, so errors are found before generation.” |
| 30–43s | Show the validation JSON and `context.md` | “Next, NAEOS creates a context bundle with the project summary and dependency graph.” |
| 43–58s | Show the pipeline log and generated folder | “The same specification then drives the generation pipeline. This run reports 61 artifacts across the generated project.” |
| 58–70s | Open `generated/README.md`, `go.mod`, and `package.json` | “The output includes a project README, Go and TypeScript project files, module code, tests, configuration, and documentation.” |
| 70–82s | Show `summary.md` and the smoke-test message | “The demo verifies key outputs automatically and records a summary, making the result easy to inspect and reproduce.” |
| 82–90s | Show GitHub repository URL | “Try the example from the NAEOS repository, then adapt the specification to your own system.” |

## On-screen commands

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
find examples/demo-cli/.run/generated -maxdepth 2 -type f | sort
```

## Production notes

- Record the terminal at a readable font size.
- Keep the real command output visible; do not replace the artifact count with an animation.
- Label the result as a repository demo, not a production deployment.
- Link to the case study and demo source in the video description.
