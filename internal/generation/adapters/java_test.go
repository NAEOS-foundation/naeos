package adapters

import (
	"strings"
	"testing"
)

func TestJavaAdapter_GenerateProject(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateProject("MyProject")
	if len(artifacts) < 3 {
		t.Fatalf("expected at least 3 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{"README.md", "pom.xml"} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
	for _, a := range artifacts {
		if a.Path == "pom.xml" {
			content := string(a.Content)
			if !strings.Contains(content, "maven.apache.org/POM") {
				t.Errorf("pom.xml should contain maven POM header")
			}
			if !strings.Contains(content, "junit") {
				t.Errorf("pom.xml should contain junit dependency")
			}
		}
	}
}

func TestJavaAdapter_GenerateModule(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateModule("users", "./internal/users", "MyProject")
	if len(artifacts) < 5 {
		t.Fatalf("expected at least 5 artifacts, got %d", len(artifacts))
	}
	paths := make(map[string]bool)
	for _, a := range artifacts {
		paths[a.Path] = true
	}
	for _, p := range []string{
		"src/main/java/com/example/myproject/users/Handler.java",
		"src/main/java/com/example/myproject/users/Service.java",
		"src/main/java/com/example/myproject/users/Repository.java",
		"src/main/java/com/example/myproject/users/Model.java",
		"src/test/java/com/example/myproject/users/HandlerTest.java",
	} {
		if !paths[p] {
			t.Errorf("missing expected file: %s", p)
		}
	}
}

func TestJavaAdapter_GenerateService(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateService("api-gateway", "http", 8080, "MyProject")
	if len(artifacts) < 1 {
		t.Fatalf("expected at least 1 artifact, got %d", len(artifacts))
	}
	found := false
	for _, art := range artifacts {
		if strings.Contains(art.Path, "Server.java") {
			found = true
			content := string(art.Content)
			if !strings.Contains(content, "listening on port") {
				t.Errorf("Server.java should contain 'listening on port'")
			}
		}
	}
	if !found {
		t.Error("expected Server.java for http service")
	}
}

func TestJavaAdapter_GenerateServiceNonHTTP(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateService("worker", "grpc", 9090, "MyProject")
	if len(artifacts) != 0 {
		t.Errorf("expected 0 artifacts for non-http service, got %d", len(artifacts))
	}
}

func TestJavaAdapter_GenerateDockerfile(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateDockerfile("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "eclipse-temurin:21") {
		t.Errorf("Dockerfile should use eclipse-temurin:21")
	}
	if !strings.Contains(content, "mvn package") {
		t.Errorf("Dockerfile should use mvn package")
	}
}

func TestJavaAdapter_GenerateCI(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateCI("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "actions/setup-java@v4") {
		t.Errorf("CI should use actions/setup-java@v4")
	}
	if !strings.Contains(content, "mvn test") {
		t.Errorf("CI should run mvn test")
	}
}

func TestJavaAdapter_GenerateDockerCompose(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateDockerCompose("MyProject")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "8080:8080") {
		t.Errorf("docker-compose should map port 8080")
	}
}

func TestJavaAdapter_GenerateArchitectureDoc(t *testing.T) {
	t.Parallel()
	a := JavaAdapter{}
	artifacts := a.GenerateArchitectureDoc("MyProject", "hexagonal")
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	content := string(artifacts[0].Content)
	if !strings.Contains(content, "hexagonal") {
		t.Errorf("architecture doc should contain pattern")
	}
	if !strings.Contains(content, "Java") {
		t.Errorf("architecture doc should contain Java")
	}
}
