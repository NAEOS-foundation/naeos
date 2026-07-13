package profiledetect

import (
	"os"
	"path/filepath"
	"strings"
)

type DetectionResult struct {
	Language   string   `json:"language"`
	Framework  string   `json:"framework"`
	Confidence float64  `json:"confidence"`
	Files      []string `json:"files"`
}

type Detector struct {
	rootDir string
}

func NewDetector(rootDir string) *Detector {
	return &Detector{rootDir: rootDir}
}

func (d *Detector) Detect() *DetectionResult {
	signals := map[string]float64{}
	var matchedFiles []string

	checks := []struct {
		file    string
		lang    string
		weight  float64
	}{
		{"go.mod", "go", 1.0},
		{"go.sum", "go", 0.3},
		{"package.json", "javascript", 0.7},
		{"tsconfig.json", "typescript", 0.9},
		{"package-lock.json", "javascript", 0.4},
		{"yarn.lock", "javascript", 0.4},
		{"pnpm-lock.yaml", "javascript", 0.4},
		{"requirements.txt", "python", 0.8},
		{"setup.py", "python", 0.8},
		{"pyproject.toml", "python", 0.9},
		{"Pipfile", "python", 0.7},
		{"Cargo.toml", "rust", 1.0},
		{"Cargo.lock", "rust", 0.3},
		{"pom.xml", "java", 0.9},
		{"build.gradle", "java", 0.9},
		{"build.gradle.kts", "java", 0.9},
		{"Gemfile", "ruby", 0.9},
		{"composer.json", "php", 0.9},
		{"go.sum", "go", 0.2},
		{"main.go", "go", 0.5},
		{"index.ts", "typescript", 0.4},
		{"index.js", "javascript", 0.4},
		{"app.py", "python", 0.4},
		{"main.py", "python", 0.4},
		{"src/main.rs", "rust", 0.5},
		{"src/App.tsx", "typescript", 0.3},
		{"src/App.jsx", "javascript", 0.3},
	}

	for _, c := range checks {
		path := filepath.Join(d.rootDir, c.file)
		if _, err := os.Stat(path); err == nil {
			signals[c.lang] += c.weight
			matchedFiles = append(matchedFiles, c.file)
		}
	}

	framework := d.detectFramework()

	bestLang := ""
	bestScore := 0.0
	for lang, score := range signals {
		if score > bestScore {
			bestScore = score
			bestLang = lang
		}
	}

	confidence := bestScore / 2.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	if bestLang == "" {
		return &DetectionResult{
			Language:   "unknown",
			Framework:  framework,
			Confidence: 0,
			Files:      matchedFiles,
		}
	}

	return &DetectionResult{
		Language:   bestLang,
		Framework:  framework,
		Confidence: confidence,
		Files:      matchedFiles,
	}
}

func (d *Detector) detectFramework() string {
	frameworkChecks := []struct {
		file      string
		framework string
	}{
		{"package.json", ""},
		{"pyproject.toml", ""},
		{"go.mod", ""},
	}

	for _, c := range frameworkChecks {
		path := filepath.Join(d.rootDir, c.file)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)

		switch c.file {
		case "package.json":
			if strings.Contains(content, "next") {
				return "nextjs"
			}
			if strings.Contains(content, "react") {
				return "react"
			}
			if strings.Contains(content, "vue") {
				return "vue"
			}
			if strings.Contains(content, "angular") {
				return "angular"
			}
			if strings.Contains(content, "express") {
				return "express"
			}
			if strings.Contains(content, "fastify") {
				return "fastify"
			}
			if strings.Contains(content, "nestjs") || strings.Contains(content, "@nestjs") {
				return "nestjs"
			}
		case "pyproject.toml":
			if strings.Contains(content, "django") {
				return "django"
			}
			if strings.Contains(content, "fastapi") {
				return "fastapi"
			}
			if strings.Contains(content, "flask") {
				return "flask"
			}
		case "go.mod":
			if strings.Contains(content, "gin-gonic") {
				return "gin"
			}
			if strings.Contains(content, "gorilla/mux") {
				return "gorilla"
			}
			if strings.Contains(content, "echo") {
				return "echo"
			}
			if strings.Contains(content, "fiber") {
				return "fiber"
			}
		}
	}

	return ""
}
