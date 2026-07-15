package adapters

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

type TypeScriptAdapter struct{}

func init() {
	Register(TypeScriptAdapter{})
}

func (TypeScriptAdapter) Language() language.Language {
	return language.LanguageTypeScript
}

func (TypeScriptAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated from NAEOS pipeline (TypeScript).\n\n## Quick Start\n\n```bash\nnpm install\nnpm run dev\n```\n\n## Test\n\n```bash\nnpm test\n```\n", projectName))},
		{Path: "package.json", Content: []byte(fmt.Sprintf(`{
  "name": "%s",
  "version": "0.1.0",
  "scripts": {
    "dev": "tsx src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js",
    "test": "vitest"
  },
  "dependencies": {},
  "devDependencies": {
    "typescript": "^5.4.0",
    "tsx": "^4.7.0",
    "vitest": "^1.6.0",
    "@types/node": "^20.0.0"
  }
}
`, slug))},
		{Path: "tsconfig.json", Content: []byte(`{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"]
}
`)},
		{Path: "src/index.ts", Content: []byte(fmt.Sprintf("export function main(): void {\n  console.log(\"hello from %s\");\n}\n\nmain();\n", projectName))},
	}
}

func (TypeScriptAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/%s", strutil.Slugify(moduleName))

	return []engine.Artifact{
		{Path: fmt.Sprintf("%s/index.ts", dir), Content: []byte(fmt.Sprintf("export * from \"./handler\";\nexport * from \"./service\";\nexport * from \"./repository\";\n"))},
		{Path: fmt.Sprintf("%s/handler.ts", dir), Content: []byte(fmt.Sprintf("import { Service } from \"./service\";\n\nexport class Handler {\n  constructor(private service: Service) {}\n\n  handle(): string {\n    return this.service.process();\n  }\n}\n"))},
		{Path: fmt.Sprintf("%s/service.ts", dir), Content: []byte("export interface Service {\n  process(): string;\n}\n\nexport class DefaultService implements Service {\n  process(): string {\n    return \"processed\";\n  }\n}\n")},
		{Path: fmt.Sprintf("%s/repository.ts", dir), Content: []byte("export interface Repository {\n  list(): string[];\n}\n")},
		{Path: fmt.Sprintf("%s/types.ts", dir), Content: []byte("export interface Model {\n  name: string;\n}\n")},
		{Path: fmt.Sprintf("%s/handler.test.ts", dir), Content: []byte(fmt.Sprintf("import { describe, it, expect } from \"vitest\";\nimport { Handler } from \"./handler\";\nimport { DefaultService } from \"./service\";\n\ndescribe(\"Handler\", () => {\n  it(\"should handle request\", () => {\n    const service = new DefaultService();\n    const handler = new Handler(service);\n    expect(handler.handle()).toBe(\"processed\");\n  });\n});\n"))},
	}
}

func (TypeScriptAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/services/%s", strutil.Slugify(serviceName))

	var artifacts []engine.Artifact
	artifacts = append(artifacts, engine.Artifact{
		Path:    fmt.Sprintf("%s/index.ts", dir),
		Content: []byte(fmt.Sprintf("export function start(port: number): void {\n  console.log(\"%s listening on port\", port);\n}\n", serviceName)),
	})

	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/server.ts", dir),
			Content: []byte(fmt.Sprintf("import http from \"node:http\";\n\nexport function createServer(port: number) {\n  const server = http.createServer((req, res) => {\n    res.writeHead(200, { \"Content-Type\": \"application/json\" });\n    res.end(JSON.stringify({ service: \"%s\", status: \"ok\" }));\n  });\n\n  server.listen(port, () => {\n    console.log(\"%s listening on port\", port);\n  });\n\n  return server;\n}\n", serviceName, serviceName)),
		})
	}

	return artifacts
}

func (TypeScriptAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "Dockerfile",
		Content: []byte("FROM node:22-alpine AS build\nWORKDIR /app\nCOPY package*.json .\nRUN npm ci\nCOPY . .\nRUN npm run build\n\nFROM node:22-alpine\nWORKDIR /app\nCOPY --from=build /app/dist ./dist\nCOPY --from=build /app/node_modules ./node_modules\nCMD [\"node\", \"dist/index.js\"]\n"),
	}}
}

func (TypeScriptAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    ".github/workflows/ci.yml",
		Content: []byte("name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-node@v4\n        with:\n          node-version: '22'\n      - run: npm ci\n      - run: npm test\n"),
	}}
}

func (TypeScriptAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docker-compose.yml",
		Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '3000:3000'\n"),
	}}
}

func (TypeScriptAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docs/architecture.md",
		Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s (TypeScript)\n", pattern, projectName)),
	}}
}
