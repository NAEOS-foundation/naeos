package adapters

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

type RustAdapter struct{}

func init() {
	Register(RustAdapter{})
}

func (RustAdapter) Language() language.Language {
	return language.LanguageRust
}

func (RustAdapter) Framework() string { return "" }

func (RustAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated from NAEOS pipeline (Rust).\n\n## Quick Start\n\n```bash\ncargo build\ncargo run\n```\n\n## Test\n\n```bash\ncargo test\n```\n", projectName))},
		{Path: "Cargo.toml", Content: []byte(fmt.Sprintf("[package]\nname = \"%s\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n[dependencies]\ntokio = { version = \"1\", features = [\"full\"] }\naxum = \"0.7\"\nserde = { version = \"1\", features = [\"derive\"] }\nserde_json = \"1\"\n\n[dev-dependencies]\ntokio-test = \"0.4\"\n", slug))},
		{Path: "src/main.rs", Content: []byte(fmt.Sprintf("use std::net::SocketAddr;\n\n#[tokio::main]\nasync fn main() {\n    let addr = SocketAddr::from(([0, 0, 0, 0], 8080));\n    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();\n    println!(\"hello from %s listening on {}\", addr);\n}\n", projectName))},
		{Path: "src/lib.rs", Content: []byte("pub mod handler;\npub mod service;\npub mod repository;\n")},
	}
}

func (RustAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	mod := strutil.Slugify(moduleName)

	return []engine.Artifact{
		{Path: fmt.Sprintf("src/%s/mod.rs", mod), Content: []byte("pub mod handler;\npub mod service;\npub mod repository;\npub mod models;\n")},
		{Path: fmt.Sprintf("src/%s/handler.rs", mod), Content: []byte("use crate::service::Service;\n\npub struct Handler {\n    service: Box<dyn Service>,\n}\n\nimpl Handler {\n    pub fn new(service: Box<dyn Service>) -> Self {\n        Self { service }\n    }\n\n    pub fn handle(&self) -> String {\n        self.service.process()\n    }\n}\n")},
		{Path: fmt.Sprintf("src/%s/service.rs", mod), Content: []byte("pub trait Service: Send + Sync {\n    fn process(&self) -> String;\n}\n\npub struct DefaultService;\n\nimpl Service for DefaultService {\n    fn process(&self) -> String {\n        \"processed\".to_string()\n    }\n}\n")},
		{Path: fmt.Sprintf("src/%s/repository.rs", mod), Content: []byte("pub trait Repository: Send + Sync {\n    fn list(&self) -> Vec<String>;\n}\n")},
		{Path: fmt.Sprintf("src/%s/models.rs", mod), Content: []byte("use serde::{Deserialize, Serialize};\n\n#[derive(Debug, Clone, Serialize, Deserialize)]\npub struct Model {\n    pub name: String,\n}\n")},
		{Path: fmt.Sprintf("tests/%s_test.rs", mod), Content: []byte("use crate::service::DefaultService;\n\n#[test]\nfn test_service() {\n    let svc = DefaultService;\n    assert_eq!(svc.process(), \"processed\");\n}\n")},
	}
}

func (RustAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	svc := strutil.Slugify(serviceName)

	var artifacts []engine.Artifact

	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("src/%s/server.rs", svc),
			Content: []byte(fmt.Sprintf("use axum::{routing::get, Router};\n\npub async fn start_server(port: u16) {\n    let app = Router::new().route(\"/\", get(|| async { \"%s\" }));\n    let addr = SocketAddr::from(([0, 0, 0, 0], port));\n    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();\n    println!(\"%s listening on {}\", addr);\n    axum::serve(listener, app).await.unwrap();\n}\n", serviceName, serviceName)),
		})
	}

	return artifacts
}

func (RustAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "Dockerfile",
		Content: []byte("FROM rust:1.78-alpine AS build\nWORKDIR /app\nCOPY . .\nRUN apk add --no-cache musl-dev && cargo build --release\n\nFROM alpine:3.19\nCOPY --from=build /app/target/release/app /app/app\nCMD [\"/app/app\"]\n"),
	}}
}

func (RustAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    ".github/workflows/ci.yml",
		Content: []byte("name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: dtolnay/rust-toolchain@stable\n      - run: cargo test\n"),
	}}
}

func (RustAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docker-compose.yml",
		Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8080:8080'\n"),
	}}
}

func (RustAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docs/architecture.md",
		Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s (Rust)\n", pattern, projectName)),
	}}
}
