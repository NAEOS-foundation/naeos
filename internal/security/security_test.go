package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuditHardcodedSecret(t *testing.T) {
	auditor := NewAuditor()
	content := `password=secret123
api_key=abc123
`
	findings := auditor.Audit("config.go", content)
	if len(findings) == 0 {
		t.Fatal("expected findings for hardcoded secrets")
	}
	found := false
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			found = true
		}
	}
	if !found {
		t.Error("expected critical severity finding")
	}
}

func TestAuditSQLInjection(t *testing.T) {
	auditor := NewAuditor()
	content := `query := fmt.Sprintf("SELECT * FROM users WHERE id=%s", userID)
`
	findings := auditor.Audit("db.go", content)
	found := false
	for _, f := range findings {
		if f.ID == "sql-injection" {
			found = true
		}
	}
	if !found {
		t.Error("expected SQL injection finding")
	}
}

func TestAuditInsecureListen(t *testing.T) {
	auditor := NewAuditor()
	content := `http.ListenAndServe("0.0.0.0:8080", nil)
`
	findings := auditor.Audit("main.go", content)
	found := false
	for _, f := range findings {
		if f.ID == "insecure-listen" {
			found = true
		}
	}
	if !found {
		t.Error("expected insecure listen finding")
	}
}

func TestAuditClean(t *testing.T) {
	auditor := NewAuditor()
	content := `package main

import "os"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
}
`
	findings := auditor.Audit("main.go", content)
	for _, f := range findings {
		if f.Severity == SeverityCritical || f.Severity == SeverityHigh {
			t.Errorf("unexpected high/critical finding: %s", f.Title)
		}
	}
}

func TestAuditFiles(t *testing.T) {
	auditor := NewAuditor()
	files := map[string]string{
		"main.go":  `password=secret123`,
		"other.go": `package main`,
	}
	result := auditor.AuditFiles(files)
	if result.Summary.Total == 0 {
		t.Error("expected findings")
	}
}

func TestSQLInjectionPython(t *testing.T) {
	auditor := NewAuditor()
	content := `query = f"SELECT * FROM users WHERE id={user_id}"`
	findings := auditor.Audit("db.py", content)
	found := false
	for _, f := range findings {
		if f.ID == "sql-injection" {
			found = true
		}
	}
	if !found {
		t.Error("expected SQL injection finding for Python")
	}
}

func TestSQLInjectionTypeScript(t *testing.T) {
	auditor := NewAuditor()
	content := "const query = `SELECT * FROM users WHERE id=${userId}`"
	findings := auditor.Audit("db.ts", content)
	found := false
	for _, f := range findings {
		if f.ID == "sql-injection" {
			found = true
		}
	}
	if !found {
		t.Error("expected SQL injection finding for TypeScript")
	}
}

func TestSQLInjectionJava(t *testing.T) {
	auditor := NewAuditor()
	content := `String query = "SELECT * FROM users WHERE id=" + userId;`
	findings := auditor.Audit("Db.java", content)
	found := false
	for _, f := range findings {
		if f.ID == "sql-injection" {
			found = true
		}
	}
	if !found {
		t.Error("expected SQL injection finding for Java")
	}
}

func TestSQLInjectionRust(t *testing.T) {
	auditor := NewAuditor()
	content := `let query = format!("SELECT * FROM users WHERE id={}", user_id);`
	findings := auditor.Audit("db.rs", content)
	found := false
	for _, f := range findings {
		if f.ID == "sql-injection" {
			found = true
		}
	}
	if !found {
		t.Error("expected SQL injection finding for Rust")
	}
}

func TestEvalPython(t *testing.T) {
	auditor := NewAuditor()
	content := `result = eval(user_input)`
	findings := auditor.Audit("app.py", content)
	found := false
	for _, f := range findings {
		if f.ID == "eval-usage" {
			found = true
		}
	}
	if !found {
		t.Error("expected eval finding for Python")
	}
}

func TestUnsafeDeserializationPython(t *testing.T) {
	auditor := NewAuditor()
	content := `data = pickle.loads(raw_data)`
	findings := auditor.Audit("app.py", content)
	found := false
	for _, f := range findings {
		if f.ID == "unsafe-deserialization" {
			found = true
		}
	}
	if !found {
		t.Error("expected unsafe deserialization finding for Python")
	}
}

func TestUnsafeDeserializationJava(t *testing.T) {
	auditor := NewAuditor()
	content := `ObjectInputStream ois = new ObjectInputStream(inputStream);`
	findings := auditor.Audit("App.java", content)
	found := false
	for _, f := range findings {
		if f.ID == "unsafe-deserialization" {
			found = true
		}
	}
	if !found {
		t.Error("expected unsafe deserialization finding for Java")
	}
}

func TestHardcodedSecretPython(t *testing.T) {
	auditor := NewAuditor()
	content := `api_key = "sk-1234567890abcdef"`
	findings := auditor.Audit("config.py", content)
	found := false
	for _, f := range findings {
		if f.ID == "hardcoded-secret" {
			found = true
		}
	}
	if !found {
		t.Error("expected hardcoded secret finding for Python")
	}
}

func TestHardcodedSecretEnvExempt(t *testing.T) {
	auditor := NewAuditor()
	content := `api_key = os.environ.get("API_KEY")`
	findings := auditor.Audit("config.py", content)
	for _, f := range findings {
		if f.ID == "hardcoded-secret" {
			t.Error("should not flag env var usage as hardcoded secret")
		}
	}
}

func TestScanDir(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("print('hello')\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not scanned\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, "vendor"), 0o755)
	os.WriteFile(filepath.Join(dir, "vendor", "dep.go"), []byte("package dep\n"), 0o644)

	files, err := ScanDir(dir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}

	if _, ok := files["main.go"]; !ok {
		t.Error("expected main.go to be scanned")
	}
	if _, ok := files["app.py"]; !ok {
		t.Error("expected app.py to be scanned")
	}
	if _, ok := files["readme.txt"]; ok {
		t.Error("readme.txt should not be scanned")
	}
	if _, ok := files["vendor/dep.go"]; ok {
		t.Error("vendor files should be skipped")
	}
}

func TestScanDir_Empty(t *testing.T) {
	dir := t.TempDir()
	files, err := ScanDir(dir)
	if err != nil {
		t.Fatalf("ScanDir failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestAuditFiles_Summary(t *testing.T) {
	auditor := NewAuditor()
	files := map[string]string{
		"main.go":  `password=secret123`,
		"clean.go": `package main`,
	}
	result := auditor.AuditFiles(files)
	if result.Summary.Critical == 0 {
		t.Error("expected at least 1 critical finding")
	}
}
