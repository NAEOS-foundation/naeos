package cicd

import (
	"fmt"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type DockerComposeGenerator struct{}

func (g *DockerComposeGenerator) Name() string {
	return "Docker Compose"
}

func (g *DockerComposeGenerator) Generate(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("version: '3.8'\n\n")
	sb.WriteString("services:\n")

	serviceName := config.Project
	if serviceName == "" {
		serviceName = "app"
	}

	fmt.Fprintf(&sb, "  %s:\n", serviceName)
	sb.WriteString("    build:\n")
	sb.WriteString("      context: .\n")
	sb.WriteString("      dockerfile: Dockerfile\n")

	if len(config.Languages) > 0 {
		switch config.Languages[0] {
		case "go":
			sb.WriteString("    image: golang:1.22\n")
			sb.WriteString("    command: go run main.go\n")
		case "node", "typescript":
			sb.WriteString("    image: node:20\n")
			sb.WriteString("    command: npm start\n")
		case "python":
			sb.WriteString("    image: python:3.12\n")
			sb.WriteString("    command: python main.py\n")
		case "java":
			sb.WriteString("    image: eclipse-temurin:21\n")
			sb.WriteString("    command: java -jar app.jar\n")
		case "rust":
			sb.WriteString("    image: rust:latest\n")
			sb.WriteString("    command: ./target/release/app\n")
		}
	}

	sb.WriteString("    volumes:\n")
	sb.WriteString("      - .:/app\n")

	if len(config.Secrets) > 0 {
		sb.WriteString("    env_file:\n")
		sb.WriteString("      - .env\n")
	}

	sb.WriteString("    ports:\n")
	sb.WriteString("      - '8080:8080'\n")
	sb.WriteString("    networks:\n")
	sb.WriteString("      - app-network\n")

	// Add database service for common patterns
	if len(config.Languages) > 0 {
		switch config.Languages[0] {
		case "go", "node", "typescript", "java", "python":
			sb.WriteString("\n  db:\n")
			sb.WriteString("    image: postgres:16\n")
			sb.WriteString("    environment:\n")
			sb.WriteString("      POSTGRES_DB: app\n")
			sb.WriteString("      POSTGRES_USER: postgres\n")
			sb.WriteString("      POSTGRES_PASSWORD: postgres\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - db-data:/var/lib/postgresql/data\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - '5432:5432'\n")
			sb.WriteString("    networks:\n")
			sb.WriteString("      - app-network\n")
		}
	}

	// Redis cache service for node/java/go
	if len(config.Languages) > 0 {
		switch config.Languages[0] {
		case "go", "node", "typescript", "java":
			sb.WriteString("\n  redis:\n")
			sb.WriteString("    image: redis:7-alpine\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - '6379:6379'\n")
			sb.WriteString("    networks:\n")
			sb.WriteString("      - app-network\n")
		}
	}

	sb.WriteString("\nvolumes:\n")
	sb.WriteString("  db-data:\n\n")

	sb.WriteString("networks:\n")
	sb.WriteString("  app-network:\n")
	sb.WriteString("    driver: bridge\n")

	// Custom steps as comments
	if len(config.Steps) > 0 {
		sb.WriteString("\n# Custom steps:\n")
		for _, step := range config.Steps {
			fmt.Fprintf(&sb, "#   %s: %s\n", step.Name, step.Command)
		}
	}

	return sb.String(), nil
}

func (g *DockerComposeGenerator) GenerateDockerfile(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	if len(config.Languages) == 0 {
		return "", naeoserr.New(naeoserr.ErrPipeline, "no languages specified for Dockerfile generation")
	}

	switch config.Languages[0] {
	case "go":
		sb.WriteString("FROM golang:1.22 AS builder\n")
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY go.mod go.sum ./\n")
		sb.WriteString("RUN go mod download\n")
		sb.WriteString("COPY . .\n")
		sb.WriteString("RUN CGO_ENABLED=0 go build -o /app/main .\n\n")
		sb.WriteString("FROM gcr.io/distroless/static-debian12\n")
		sb.WriteString("COPY --from=builder /app/main /\n")
		sb.WriteString("CMD [\"/main\"]\n")
	case "node", "typescript":
		sb.WriteString("FROM node:20 AS builder\n")
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY package*.json ./\n")
		sb.WriteString("RUN npm ci\n")
		sb.WriteString("COPY . .\n")
		sb.WriteString("RUN npm run build\n\n")
		sb.WriteString("FROM node:20-slim\n")
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY --from=builder /app/dist ./dist\n")
		sb.WriteString("COPY --from=builder /app/node_modules ./node_modules\n")
		sb.WriteString("EXPOSE 3000\n")
		sb.WriteString("CMD [\"node\", \"dist/index.js\"]\n")
	case "python":
		sb.WriteString("FROM python:3.12-slim\n")
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY requirements.txt .\n")
		sb.WriteString("RUN pip install --no-cache-dir -r requirements.txt\n")
		sb.WriteString("COPY . .\n")
		sb.WriteString("EXPOSE 8000\n")
		sb.WriteString("CMD [\"python\", \"main.py\"]\n")
	case "java":
		sb.WriteString("FROM eclipse-temurin:21 AS builder\n")
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY pom.xml .\n")
		sb.WriteString("RUN mvn dependency:go-offline\n")
		sb.WriteString("COPY src ./src\n")
		sb.WriteString("RUN mvn clean package -DskipTests\n\n")
		sb.WriteString("FROM eclipse-temurin:21-jre\n")
		sb.WriteString("COPY --from=builder /app/target/*.jar /app/app.jar\n")
		sb.WriteString("EXPOSE 8080\n")
		sb.WriteString("CMD [\"java\", \"-jar\", \"/app/app.jar\"]\n")
	case "rust":
		sb.WriteString("FROM rust:latest AS builder\n")
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY Cargo.toml Cargo.lock ./\n")
		sb.WriteString("RUN mkdir src && echo 'fn main() {}' > src/main.rs\n")
		sb.WriteString("RUN cargo build --release\n")
		sb.WriteString("RUN rm -rf src\n")
		sb.WriteString("COPY src ./src\n")
		sb.WriteString("RUN cargo build --release\n\n")
		sb.WriteString("FROM debian:bookworm-slim\n")
		sb.WriteString("COPY --from=builder /app/target/release/app /usr/local/bin/app\n")
		sb.WriteString("CMD [\"app\"]\n")
	default:
		return "", naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("unsupported language for Dockerfile: %s", config.Languages[0]))
	}

	return sb.String(), nil
}
