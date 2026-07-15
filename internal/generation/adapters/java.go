package adapters

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

type JavaAdapter struct{}

func init() {
	Register(JavaAdapter{})
}

func (JavaAdapter) Language() language.Language {
	return language.LanguageJava
}

func (JavaAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)
	javaPkg := pkgName(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated from NAEOS pipeline (Java).\n\n## Quick Start\n\n```bash\nmvn compile\nmvn exec:java\n```\n\n## Test\n\n```bash\nmvn test\n```\n", projectName))},
		{Path: "pom.xml", Content: []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>%s</artifactId>
    <version>0.1.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
    </properties>

    <dependencies>
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>5.10.2</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>
`, slug))},
		{Path: fmt.Sprintf("src/main/java/com/example/%s/App.java", javaPkg), Content: []byte(fmt.Sprintf("package com.example.%s;\n\npublic class App {\n    public static void main(String[] args) {\n        System.out.println(\"hello from %s\");\n    }\n}\n", javaPkg, projectName))},
	}
}

func (JavaAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	javaPkg := pkgName(projectName)
	javaMod := pkgName(moduleName)
	dir := fmt.Sprintf("src/main/java/com/example/%s/%s", javaPkg, javaMod)

	return []engine.Artifact{
		{Path: fmt.Sprintf("%s/Handler.java", dir), Content: []byte(fmt.Sprintf("package com.example.%s.%s;\n\npublic class Handler {\n    private final Service service;\n\n    public Handler(Service service) {\n        this.service = service;\n    }\n\n    public String handle() {\n        return service.process();\n    }\n}\n", javaPkg, javaMod))},
		{Path: fmt.Sprintf("%s/Service.java", dir), Content: []byte(fmt.Sprintf("package com.example.%s.%s;\n\npublic interface Service {\n    String process();\n}\n", javaPkg, javaMod))},
		{Path: fmt.Sprintf("%s/Repository.java", dir), Content: []byte(fmt.Sprintf("package com.example.%s.%s;\n\nimport java.util.List;\n\npublic interface Repository {\n    List<String> list();\n}\n", javaPkg, javaMod))},
		{Path: fmt.Sprintf("%s/Model.java", dir), Content: []byte(fmt.Sprintf("package com.example.%s.%s;\n\npublic class Model {\n    private String name;\n\n    public String getName() { return name; }\n    public void setName(String name) { this.name = name; }\n}\n", javaPkg, javaMod))},
		{Path: fmt.Sprintf("src/test/java/com/example/%s/%s/HandlerTest.java", javaPkg, javaMod), Content: []byte(fmt.Sprintf("package com.example.%s.%s;\n\nimport org.junit.jupiter.api.Test;\nimport static org.junit.jupiter.api.Assertions.*;\n\npublic class HandlerTest {\n    @Test\n    public void testHandle() {\n        Service service = new DefaultService();\n        Handler handler = new Handler(service);\n        assertNotNull(handler);\n        assertEquals(\"processed\", handler.handle());\n    }\n}\n", javaPkg, javaMod))},
	}
}

func (JavaAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	javaPkg := pkgName(projectName)
	javaSvc := pkgName(serviceName)
	dir := fmt.Sprintf("src/main/java/com/example/%s/%s", javaPkg, javaSvc)

	var artifacts []engine.Artifact

	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/Server.java", dir),
			Content: []byte(fmt.Sprintf("package com.example.%s.%s;\n\npublic class Server {\n    public static void start(int port) {\n        System.out.printf(\"%%s listening on port %%d%%n\", %q, port);\n    }\n}\n", javaPkg, javaSvc, serviceName)),
		})
	}

	return artifacts
}

func (JavaAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "Dockerfile",
		Content: []byte("FROM eclipse-temurin:21-jdk-alpine AS build\nWORKDIR /app\nCOPY pom.xml .\nCOPY src ./src\nRUN apk add --no-cache maven && mvn package -DskipTests\n\nFROM eclipse-temurin:21-jre-alpine\nWORKDIR /app\nCOPY --from=build /app/target/*.jar app.jar\nCMD [\"java\", \"-jar\", \"app.jar\"]\n"),
	}}
}

func (JavaAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    ".github/workflows/ci.yml",
		Content: []byte("name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-java@v4\n        with:\n          java-version: '21'\n          distribution: 'temurin'\n      - run: mvn test\n"),
	}}
}

func (JavaAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docker-compose.yml",
		Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8080:8080'\n"),
	}}
}

func (JavaAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docs/architecture.md",
		Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s (Java)\n", pattern, projectName)),
	}}
}
