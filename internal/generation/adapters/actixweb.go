package adapters

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

// ActixWebAdapter generates an Actix-Web based Rust project.
type ActixWebAdapter struct{}

func init() {
	Register(ActixWebAdapter{})
}

// Language returns the language this adapter targets.
func (ActixWebAdapter) Language() language.Language {
	return language.LanguageRust
}

func (ActixWebAdapter) Framework() string { return "actix-web" }

// GenerateProject creates a new Actix-Web project skeleton.
func (ActixWebAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated Actix-Web project from NAEOS.\n\n## Quick Start\n\n```bash\ncargo run\n```\n\n## Test\n\n```bash\ncargo test\n```\n", projectName))},
		{Path: "Cargo.toml", Content: []byte(fmt.Sprintf("[package]\nname = \"%s\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n[dependencies]\nactix-web = \"4\"\nserde = { version = \"1\", features = [\"derive\"] }\nserde_json = \"1\"\n\n[dev-dependencies]\ntokio = { version = \"1\", features = [\"macros\"] }\n", slug))},
		{Path: "src/main.rs", Content: []byte(fmt.Sprintf("use actix_web::{get, App, HttpResponse, HttpServer, Responder};\n\n#[get(\"/\")]\nasync fn root() -> impl Responder {\n    HttpResponse::Ok().body(\"Hello from %s!\")\n}\n\n#[actix_web::main]\nasync fn main() -> std::io::Result<()> {\n    HttpServer::new(|| App::new().service(root))\n        .bind((\"127.0.0.1\", 8080))?\n        .run()\n        .await\n}\n", projectName))},
		{Path: "src/lib.rs", Content: []byte("pub mod handler;\npub mod service;\npub mod repository;\n")},
	}
}

// GenerateModule creates a module (handler, service, model) for Actix-Web.
func (ActixWebAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	mod := strutil.Slugify(moduleName)

	return []engine.Artifact{
		{Path: fmt.Sprintf("src/%s/mod.rs", mod), Content: []byte("pub mod handler;\npub mod service;\npub mod models;\n")},
		{Path: fmt.Sprintf("src/%s/handler.rs", mod), Content: []byte("use actix_web::{get, web, HttpResponse, Responder};\n\npub struct Handler;\n\nimpl Handler {\n    pub async fn handle() -> impl Responder {\n        HttpResponse::Ok().body(\"handled\")\n    }\n}\n\n#[get(\"/\")]\npub async fn list() -> impl Responder {\n    HttpResponse::Ok().finish()\n}\n")},
		{Path: fmt.Sprintf("src/%s/service.rs", mod), Content: []byte("pub trait Service: Send + Sync {\n    fn process(&self) -> String;\n}\n\npub struct DefaultService;\n\nimpl Service for DefaultService {\n    fn process(&self) -> String {\n        \"processed\".to_string()\n    }\n}\n")},
		{Path: fmt.Sprintf("src/%s/models.rs", mod), Content: []byte("use serde::{Deserialize, Serialize};\n\n#[derive(Debug, Clone, Serialize, Deserialize)]\npub struct Model {\n    pub name: String,\n}\n")},
		{Path: fmt.Sprintf("tests/%s_test.rs", mod), Content: []byte("use crate::service::DefaultService;\n\n#[test]\nfn test_service() {\n    let svc = DefaultService;\n    assert_eq!(svc.process(), \"processed\");\n}\n")},
	}
}

// GenerateService creates an Actix-Web HTTP service definition.
func (ActixWebAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("src/services/%s", strutil.Slugify(serviceName))

	var artifacts []engine.Artifact
	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/server.rs", dir),
			Content: []byte(fmt.Sprintf("use actix_web::{get, App, HttpResponse, HttpServer, Responder};\n\n#[get(\"/health\")]\nasync fn health() -> impl Responder {\n    HttpResponse::Ok().json(serde_json::json!({\"status\": \"ok\"}))\n}\n\n#[actix_web::main]\nasync fn main() -> std::io::Result<()> {\n    HttpServer::new(|| App::new().service(health))\n        .bind((\"0.0.0.0\", %d))?\n        .run()\n        .await\n}\n", servicePort)),
		})
	}
	return artifacts
}

// GenerateDockerfile returns a Dockerfile for Actix-Web.
func (ActixWebAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    "Dockerfile",
			Content: []byte("FROM rust:1-slim AS builder\nWORKDIR /app\nCOPY . .\nRUN cargo install --path .\n\nFROM debian:bookworm-slim\nCOPY --from=builder /usr/local/cargo/bin/app /usr/local/bin/app\nEXPOSE 8080\nCMD [\"app\"]\n"),
		},
	}
}

// GenerateCI returns a GitHub Actions workflow for Rust.
func (ActixWebAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    ".github/workflows/ci.yml",
			Content: []byte("name: CI\n\non: [push, pull_request]\n\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: dtolnay/rust-toolchain@stable\n      - run: cargo test --verbose\n"),
		},
	}
}

// GenerateDockerCompose returns a docker-compose file for Actix-Web.
func (ActixWebAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    "docker-compose.yml",
			Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8080:8080'\n    environment:\n      - ENV=development\n"),
		},
	}
}

// GenerateArchitectureDoc creates an architecture markdown file.
func (ActixWebAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{
		{
			Path:    "docs/architecture.md",
			Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s (Actix-Web)\n", pattern, projectName)),
		},
	}
}
