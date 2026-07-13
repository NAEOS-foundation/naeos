package container

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(neir *model.NEIR) []engine.Artifact {
	var artifacts []engine.Artifact
	artifacts = append(artifacts, g.generateDockerfile(neir)...)
	artifacts = append(artifacts, g.generateDockerCompose(neir)...)
	artifacts = append(artifacts, g.generateK8sManifests(neir)...)
	return artifacts
}

func (g *Generator) generateDockerfile(neir *model.NEIR) []engine.Artifact {
	var artifacts []engine.Artifact

	lang := "go"
	if neir.Generation != nil && len(neir.Generation.Languages) > 0 {
		lang = string(neir.Generation.Languages[0])
	}

	switch lang {
	case "go":
		artifacts = append(artifacts, g.dockerfileGo(neir))
	case "typescript", "javascript":
		artifacts = append(artifacts, g.dockerfileNode(neir))
	case "python":
		artifacts = append(artifacts, g.dockerfilePython(neir))
	case "java":
		artifacts = append(artifacts, g.dockerfileJava(neir))
	case "rust":
		artifacts = append(artifacts, g.dockerfileRust(neir))
	default:
		artifacts = append(artifacts, g.dockerfileGo(neir))
	}

	return artifacts
}

func (g *Generator) dockerfileGo(neir *model.NEIR) engine.Artifact {
	name := "app"
	if neir.Project != nil && neir.Project.Name != "" {
		name = neir.Project.Name
	}
	port := 8080
	if len(neir.Services) > 0 && neir.Services[0].Port > 0 {
		port = neir.Services[0].Port
	}

	dockerfile := fmt.Sprintf(`FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/%s ./cmd/%s

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/%s .
EXPOSE %d
CMD ["./%s"]
`, name, name, name, port, name)

	return engine.Artifact{
		Path:    "Dockerfile",
		Content: []byte(dockerfile),
	}
}

func (g *Generator) dockerfileNode(neir *model.NEIR) engine.Artifact {
	port := 3000
	if len(neir.Services) > 0 && neir.Services[0].Port > 0 {
		port = neir.Services[0].Port
	}

	dockerfile := fmt.Sprintf(`FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./
EXPOSE %d
CMD ["node", "dist/index.js"]
`, port)

	return engine.Artifact{
		Path:    "Dockerfile",
		Content: []byte(dockerfile),
	}
}

func (g *Generator) dockerfilePython(neir *model.NEIR) engine.Artifact {
	port := 8000
	if len(neir.Services) > 0 && neir.Services[0].Port > 0 {
		port = neir.Services[0].Port
	}

	dockerfile := fmt.Sprintf(`FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE %d
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "%d"]
`, port, port)

	return engine.Artifact{
		Path:    "Dockerfile",
		Content: []byte(dockerfile),
	}
}

func (g *Generator) dockerfileJava(_ *model.NEIR) engine.Artifact {
	dockerfile := `FROM eclipse-temurin:21-jdk AS builder
WORKDIR /app
COPY . .
RUN ./gradlew bootJar --no-daemon

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]
`

	return engine.Artifact{
		Path:    "Dockerfile",
		Content: []byte(dockerfile),
	}
}

func (g *Generator) dockerfileRust(neir *model.NEIR) engine.Artifact {
	name := "app"
	if neir.Project != nil && neir.Project.Name != "" {
		name = neir.Project.Name
	}
	port := 8080
	if len(neir.Services) > 0 && neir.Services[0].Port > 0 {
		port = neir.Services[0].Port
	}

	dockerfile := fmt.Sprintf(`FROM rust:1.77 AS builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /app/target/release/%s .
EXPOSE %d
CMD ["./%s"]
`, name, port, name)

	return engine.Artifact{
		Path:    "Dockerfile",
		Content: []byte(dockerfile),
	}
}

func (g *Generator) generateDockerCompose(neir *model.NEIR) []engine.Artifact {
	if len(neir.Services) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\nservices:\n")

	for _, svc := range neir.Services {
		port := 8080
		if svc.Port > 0 {
			port = svc.Port
		}

		sb.WriteString(fmt.Sprintf("  %s:\n", svc.Name))
		sb.WriteString("    build: .\n")
		sb.WriteString("    ports:\n")
		sb.WriteString(fmt.Sprintf("      - \"%d:%d\"\n", port, port))
		sb.WriteString("    environment:\n")
		sb.WriteString(fmt.Sprintf("      - SERVICE_NAME=%s\n", svc.Name))
		sb.WriteString("\n")
	}

	return []engine.Artifact{{
		Path:    "docker-compose.yaml",
		Content: []byte(sb.String()),
	}}
}

func (g *Generator) generateK8sManifests(neir *model.NEIR) []engine.Artifact {
	if len(neir.Services) == 0 {
		return nil
	}

	name := "app"
	if neir.Project != nil && neir.Project.Name != "" {
		name = neir.Project.Name
	}

	var artifacts []engine.Artifact

	artifacts = append(artifacts, engine.Artifact{
		Path:    "k8s/namespace.yaml",
		Content: []byte(fmt.Sprintf("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: %s\n", name)),
	})

	for _, svc := range neir.Services {
		port := 8080
		if svc.Port > 0 {
			port = svc.Port
		}

		deployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
        - name: %s
          image: %s:latest
          ports:
            - containerPort: %d
          readinessProbe:
            httpGet:
              path: /healthz
              port: %d
            initialDelaySeconds: 5
            periodSeconds: 10
`, svc.Name, svc.Name, svc.Name, svc.Name, svc.Name, svc.Name, port, port)

		svcManifest := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  selector:
    app: %s
  ports:
    - port: %d
      targetPort: %d
  type: ClusterIP
`, svc.Name, svc.Name, port, port)

		artifacts = append(artifacts,
			engine.Artifact{Path: fmt.Sprintf("k8s/deployment-%s.yaml", svc.Name), Content: []byte(deployment)},
			engine.Artifact{Path: fmt.Sprintf("k8s/service-%s.yaml", svc.Name), Content: []byte(svcManifest)},
		)
	}

	return artifacts
}
