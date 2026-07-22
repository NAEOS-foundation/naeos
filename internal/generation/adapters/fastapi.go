package adapters

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

// FastAPIAdapter generates a FastAPI-based Python project.
type FastAPIAdapter struct{}

func init() {
	Register(FastAPIAdapter{})
}

// Language returns the language this adapter targets.
func (FastAPIAdapter) Language() language.Language {
	return language.LanguagePython
}

func (FastAPIAdapter) Framework() string { return "fastapi" }

// GenerateProject creates the base project layout for a FastAPI app.
func (FastAPIAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)
	pkg := pkgName(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated FastAPI project from NAEOS.\n\n## Quick Start\n\n```bash\npip install -e .\nuvicorn %s.__main__:app --reload\n```\n\n## Test\n\n```bash\npytest\n```\n", projectName, pkg))},
		{Path: "pyproject.toml", Content: []byte(fmt.Sprintf("[build-system]\nrequires = [\"setuptools>=68.0\", \"wheel\"]\nbuild-backend = \"setuptools.build_meta\"\n\n[project]\nname = \"%s\"\nversion = \"0.1.0\"\nrequires-python = \">=3.11\"\ndependencies = [\"fastapi\", \"uvicorn[standard]\"]\n\n[project.scripts]\n%s = \"%s.__main__:main\"\n\n[tool.pytest.ini_options]\ntestpaths = [\"tests\"]\n", slug, pkg, pkg))},
		{Path: fmt.Sprintf("%s/__init__.py", pkg), Content: []byte(fmt.Sprintf("\"\"\"%s package.\"\"\"\n\n__version__ = \"0.1.0\"\n", projectName))},
		{Path: fmt.Sprintf("%s/__main__.py", pkg), Content: []byte("import uvicorn\nfrom .app import app\n\ndef main() -> None:\n    uvicorn.run(app, host=\"0.0.0.0\", port=8000)\n\nif __name__ == \"__main__\":\n    main()\n")},
		{Path: fmt.Sprintf("%s/app.py", pkg), Content: []byte(fmt.Sprintf("from fastapi import FastAPI\n\napp = FastAPI(title=\"%s\")\n\n@app.get(\"/\")\nasync def root():\n    return {\"message\": \"Hello World\"}\n", projectName))},
	}
}

// GenerateModule creates a module (router + service + model) for FastAPI.
func (FastAPIAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/%s", strutil.Slugify(moduleName))
	slug := strutil.Slugify(moduleName)

	return []engine.Artifact{
		{Path: fmt.Sprintf("%s/__init__.py", dir), Content: []byte(fmt.Sprintf("\"\"\"%s module.\"\"\"\n", moduleName))},
		{Path: fmt.Sprintf("%s/router.py", dir), Content: []byte(fmt.Sprintf("from fastapi import APIRouter, Depends\n\nrouter = APIRouter(prefix=\"/%s\", tags=[\"%s\"])\n\n@router.get(\"/\")\nasync def read_items():\n    return {\"items\": []}\n", slug, slug))},
		{Path: fmt.Sprintf("%s/service.py", dir), Content: []byte("from abc import ABC, abstractmethod\nfrom typing import Any\n\nclass Service(ABC):\n    @abstractmethod\n    async def process(self) -> Any:\n        ...\n\nclass DefaultService(Service):\n    async def process(self) -> Any:\n        return {\"result\": \"processed\"}\n")},
		{Path: fmt.Sprintf("%s/models.py", dir), Content: []byte("from pydantic import BaseModel\n\nclass Model(BaseModel):\n    name: str\n\n")},
		{Path: fmt.Sprintf("tests/test_%s.py", slug), Content: []byte(fmt.Sprintf("from src.%s.app import app\n\ndef test_app_exists():\n    assert app is not None\n", slug))},
	}
}

// GenerateService creates a service (HTTP server) definition for FastAPI.
func (FastAPIAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/services/%s", strutil.Slugify(serviceName))

	var artifacts []engine.Artifact
	artifacts = append(artifacts, engine.Artifact{
		Path:    fmt.Sprintf("%s/__init__.py", dir),
		Content: []byte(fmt.Sprintf("def start(port: int) -> None:\n    print(\"%s listening on port {port}\")\n", serviceName)),
	})

	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/server.py", dir),
			Content: []byte(fmt.Sprintf("from fastapi import FastAPI\nimport uvicorn\n\napp = FastAPI(title=\"%s\")\n\n@app.get(\"/health\")\nasync def health():\n    return {\"status\": \"ok\"}\n\nif __name__ == \"__main__\":\n    uvicorn.run(app, host=\"0.0.0.0\", port=%d)\n", serviceName, servicePort)),
		})
	}
	return artifacts
}

// GenerateDockerfile returns a Dockerfile suitable for a FastAPI app.
func (FastAPIAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    "Dockerfile",
			Content: []byte("FROM python:3.12-slim AS builder\nWORKDIR /app\nCOPY pyproject.toml .\nRUN pip install --no-cache-dir -e .\nCOPY . .\n\nFROM python:3.12-slim\nWORKDIR /app\nCOPY --from=builder /app .\nEXPOSE 8000\nCMD [\"uvicorn\", \"src.__main__:app\", \"--host\", \"0.0.0.0\", \"--port\", \"8000\"]\n"),
		},
	}
}

// GenerateCI returns a basic GitHub Actions workflow for FastAPI.
func (FastAPIAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    ".github/workflows/ci.yml",
			Content: []byte("name: CI\n\non: [push, pull_request]\n\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-python@v5\n        with:\n          python-version: '3.12'\n      - run: pip install -e .[test]\n      - run: pytest\n"),
		},
	}
}

// GenerateDockerCompose returns a docker-compose file exposing the API.
func (FastAPIAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    "docker-compose.yml",
			Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8000:8000'\n    environment:\n      - ENV=development\n"),
		},
	}
}

// GenerateArchitectureDoc creates a simple architecture markdown file.
func (FastAPIAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    "docs/architecture.md",
			Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s (FastAPI)\n", pattern, projectName)),
		},
	}
}
