package cicd

type langCommands struct {
	Build string
	Test  string
	Image string
}

var langMap = map[string]langCommands{
	"go": {
		Build: "go build ./...",
		Test:  "go test ./...",
		Image: "golang:1.22",
	},
	"node": {
		Build: "npm ci && npm run build",
		Test:  "npm test",
		Image: "node:20",
	},
	"typescript": {
		Build: "npm ci && npm run build",
		Test:  "npm test",
		Image: "node:20",
	},
	"python": {
		Build: "pip install -r requirements.txt",
		Test:  "pytest",
		Image: "python:3.12",
	},
	"java": {
		Build: "mvn clean compile",
		Test:  "mvn test",
		Image: "maven:3.9-eclipse-temurin-21",
	},
	"rust": {
		Build: "cargo build --release",
		Test:  "cargo test",
		Image: "rust:latest",
	},
}

func buildCommand(lang string) string {
	if c, ok := langMap[lang]; ok {
		return c.Build
	}
	return ""
}

func testCommand(lang string) string {
	if c, ok := langMap[lang]; ok {
		return c.Test
	}
	return ""
}

func langImage(lang string) string {
	if c, ok := langMap[lang]; ok {
		return c.Image
	}
	return ""
}

func supportedLangs() []string {
	keys := make([]string, 0, len(langMap))
	for k := range langMap {
		keys = append(keys, k)
	}
	return keys
}
