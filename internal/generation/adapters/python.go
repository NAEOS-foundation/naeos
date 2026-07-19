package adapters

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

type PythonAdapter struct{}

func init() {
	Register(PythonAdapter{})
}

func (PythonAdapter) Language() language.Language {
	return language.LanguagePython
}

func (PythonAdapter) Framework() string { return "" }

func (PythonAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)
	pkg := pkgName(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated from NAEOS pipeline (Python).\n\n## Quick Start\n\n```bash\npip install -e .\npython -m %s\n```\n\n## Test\n\n```bash\npytest\n```\n", projectName, pkg))},
		{Path: "pyproject.toml", Content: []byte(fmt.Sprintf("[build-system]\nrequires = [\"setuptools>=68.0\", \"wheel\"]\nbuild-backend = \"setuptools.build_meta\"\n\n[project]\nname = \"%s\"\nversion = \"0.1.0\"\nrequires-python = \">=3.11\"\ndependencies = []\n\n[project.scripts]\n%s = \"%s.__main__:main\"\n\n[tool.pytest.ini_options]\ntestpaths = [\"tests\"]\n", slug, pkg, pkg))},
		{Path: fmt.Sprintf("%s/__init__.py", pkg), Content: []byte(fmt.Sprintf("\"\"\"%s package.\"\"\"\n\n__version__ = \"0.1.0\"\n", projectName))},
		{Path: fmt.Sprintf("%s/__main__.py", pkg), Content: []byte(fmt.Sprintf("def main():\n    print(\"hello from %s\")\n\n\nif __name__ == \"__main__\":\n    main()\n", projectName))},
	}
}

func (PythonAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/%s", strutil.Slugify(moduleName))

	return []engine.Artifact{
		{Path: fmt.Sprintf("%s/__init__.py", dir), Content: []byte(fmt.Sprintf("\"\"\"%s module.\"\"\"\n", moduleName))},
		{Path: fmt.Sprintf("%s/handler.py", dir), Content: []byte("from .service import Service\n\n\nclass Handler:\n    def __init__(self, service: Service):\n        self.service = service\n\n    def handle(self) -> str:\n        return self.service.process()\n")},
		{Path: fmt.Sprintf("%s/service.py", dir), Content: []byte("from abc import ABC, abstractmethod\n\n\nclass Service(ABC):\n    @abstractmethod\n    def process(self) -> str:\n        ...\n\n\nclass DefaultService(Service):\n    def process(self) -> str:\n        return \"processed\"\n")},
		{Path: fmt.Sprintf("%s/repository.py", dir), Content: []byte("from abc import ABC, abstractmethod\n\n\nclass Repository(ABC):\n    @abstractmethod\n    def list(self) -> list[str]:\n        ...\n")},
		{Path: fmt.Sprintf("%s/models.py", dir), Content: []byte("from dataclasses import dataclass\n\n\n@dataclass\nclass Model:\n    name: str\n")},
		{Path: fmt.Sprintf("tests/test_%s.py", strutil.Slugify(moduleName)), Content: []byte(fmt.Sprintf("from %s.handler import Handler\nfrom %s.service import DefaultService\n\ndef test_handler():\n    svc = DefaultService()\n    handler = Handler(svc)\n    assert handler.handle() == \"processed\"\n", pkgName(moduleName), pkgName(moduleName)))},
	}
}

func (PythonAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/services/%s", strutil.Slugify(serviceName))

	var artifacts []engine.Artifact
	artifacts = append(artifacts, engine.Artifact{
		Path:    fmt.Sprintf("%s/__init__.py", dir),
		Content: []byte(fmt.Sprintf("def start(port: int) -> None:\n    print(f\"%s listening on port {port}\")\n", serviceName)),
	})

	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/server.py", dir),
			Content: []byte(fmt.Sprintf("from http.server import HTTPServer, SimpleHTTPRequestHandler\n\n\ndef create_server(port: int) -> HTTPServer:\n    server = HTTPServer((\"\", port), SimpleHTTPRequestHandler)\n    print(f\"%s listening on port {port}\")\n    return server\n", serviceName)),
		})
	}

	return artifacts
}

func (PythonAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "Dockerfile",
		Content: []byte("FROM python:3.12-slim AS build\nWORKDIR /app\nCOPY pyproject.toml .\nRUN pip install --no-cache-dir -e .\nCOPY . .\n\nFROM python:3.12-slim\nWORKDIR /app\nCOPY --from=build /app .\nCMD [\"python\", \"-m\", \"src\"]\n"),
	}}
}

func (PythonAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    ".github/workflows/ci.yml",
		Content: []byte("name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-python@v5\n        with:\n          python-version: '3.12'\n      - run: pip install -e .\n      - run: pytest\n"),
	}}
}

func (PythonAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docker-compose.yml",
		Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8000:8000'\n"),
	}}
}

func (PythonAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docs/architecture.md",
		Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s (Python)\n", pattern, projectName)),
	}}
}
