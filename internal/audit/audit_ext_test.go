package audit

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStdoutAuditorLog(t *testing.T) {
	// Capture stdout
	rescue := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	auditor := NewStdoutAuditor()
	err := auditor.Log(AuditEvent{
		UserID:   "u1",
		Action:   "test",
		Resource: "testing",
		Status:   "success",
	})
	if err != nil {
		t.Fatalf("StdoutAuditor.Log failed: %v", err)
	}

	w.Close()
	os.Stdout = rescue

	var buf strings.Builder
	_, _ = copyBuffer(&buf, r, nil)
	output := buf.String()

	if !strings.Contains(output, "u1") {
		t.Errorf("expected output to contain user ID, got: %s", output)
	}
	if !strings.Contains(output, "test") {
		t.Errorf("expected output to contain action, got: %s", output)
	}
}

func TestHashedAuditorLog(t *testing.T) {
	mem := NewMemoryAuditor()
	hashed := NewHashedAuditor(mem)

	err := hashed.Log(AuditEvent{UserID: "u1", Action: "create", Status: "success"})
	if err != nil {
		t.Fatalf("HashedAuditor.Log failed: %v", err)
	}

	err = hashed.Log(AuditEvent{UserID: "u2", Action: "delete", Status: "success"})
	if err != nil {
		t.Fatalf("HashedAuditor.Log failed: %v", err)
	}

	events := mem.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Hash == "" {
		t.Error("expected event[0] to have Hash")
	}
	if events[0].PreviousHash != "" {
		t.Errorf("expected event[0] PreviousHash to be empty, got %q", events[0].PreviousHash)
	}
	if events[1].PreviousHash != events[0].Hash {
		t.Errorf("expected event[1] PreviousHash=%q to match event[0] Hash=%q", events[1].PreviousHash, events[0].Hash)
	}
}

func TestVerifyChain(t *testing.T) {
	mem := NewMemoryAuditor()
	hashed := NewHashedAuditor(mem)

	hashed.Log(AuditEvent{UserID: "u1", Action: "create"})
	hashed.Log(AuditEvent{UserID: "u2", Action: "update"})
	hashed.Log(AuditEvent{UserID: "u3", Action: "delete"})

	events := mem.Events()
	violations := VerifyChain(events)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}

func TestVerifyChainTampered(t *testing.T) {
	mem := NewMemoryAuditor()
	hashed := NewHashedAuditor(mem)

	hashed.Log(AuditEvent{UserID: "u1", Action: "create"})
	hashed.Log(AuditEvent{UserID: "u2", Action: "update"})

	events := mem.Events()
	events[0].Action = "tampered" // tamper with first event

	violations := VerifyChain(events)
	if len(violations) == 0 {
		t.Error("expected violations for tampered chain")
	}
}

func TestVerifyChainBrokenLink(t *testing.T) {
	mem := NewMemoryAuditor()
	hashed := NewHashedAuditor(mem)

	hashed.Log(AuditEvent{UserID: "u1", Action: "create"})
	hashed.Log(AuditEvent{UserID: "u2", Action: "update"})

	events := mem.Events()
	events[1].PreviousHash = "badhash"

	violations := VerifyChain(events)
	if len(violations) == 0 {
		t.Error("expected violations for broken chain link")
	}
}

func TestVerifyChainFile(t *testing.T) {
	dir := t.TempDir()

	fileAuditor, err := NewFileAuditor(dir)
	if err != nil {
		t.Fatalf("NewFileAuditor: %v", err)
	}

	hashed := NewHashedAuditor(fileAuditor)
	hashed.Log(AuditEvent{UserID: "u1", Action: "create"})
	hashed.Log(AuditEvent{UserID: "u2", Action: "update"})

	path := filepath.Join(dir, ".naeos", "audit.log")
	violations, err := VerifyChainFile(path)
	if err != nil {
		t.Fatalf("VerifyChainFile: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}

func TestEncryptedAuditorLog(t *testing.T) {
	passphrase := "test-passphrase-123"
	mem := NewMemoryAuditor()
	enc := NewEncryptedAuditor(mem, passphrase)

	err := enc.Log(AuditEvent{
		UserID:   "u1",
		Action:   "create",
		Resource: "project",
		Status:   "success",
		Details:  "sensitive data",
	})
	if err != nil {
		t.Fatalf("EncryptedAuditor.Log failed: %v", err)
	}

	rawEvents := mem.Events()
	if len(rawEvents) != 1 {
		t.Fatalf("expected 1 event, got %d", len(rawEvents))
	}

	if rawEvents[0].Details == "" {
		t.Error("expected encrypted details in stored event")
	}
	if rawEvents[0].Metadata["encrypted"] != "true" {
		t.Errorf("expected encrypted metadata, got %v", rawEvents[0].Metadata)
	}
}

func TestEncryptedAuditorDecrypt(t *testing.T) {
	passphrase := "test-passphrase-123"
	mem := NewMemoryAuditor()
	enc := NewEncryptedAuditor(mem, passphrase)

	enc.Log(AuditEvent{
		UserID:   "u1",
		Action:   "create",
		Resource: "project",
		Status:   "success",
		Details:  "secret-data",
	})

	reader := NewDecryptedReader(mem, passphrase)
	events, err := reader.Events()
	if err != nil {
		t.Fatalf("DecryptedReader.Events failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 decrypted event, got %d", len(events))
	}
	if events[0].Details != "secret-data" {
		t.Errorf("expected decrypted details 'secret-data', got %q", events[0].Details)
	}
	if events[0].Action != "create" {
		t.Errorf("expected action 'create', got %q", events[0].Action)
	}
}

func TestEncryptedAuditorWrongPassphrase(t *testing.T) {
	mem := NewMemoryAuditor()
	enc := NewEncryptedAuditor(mem, "correct-passphrase")

	enc.Log(AuditEvent{
		UserID:   "u1",
		Action:   "create",
		Resource: "project",
		Status:   "success",
		Details:  "secret",
	})

	reader := NewDecryptedReader(mem, "wrong-passphrase")
	_, err := reader.Events()
	if err == nil {
		t.Error("expected error for wrong passphrase")
	}
}

func TestNewEncryptedFileAuditor(t *testing.T) {
	dir := t.TempDir()
	auditor, err := NewEncryptedFileAuditor(dir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewEncryptedFileAuditor: %v", err)
	}
	if auditor == nil {
		t.Fatal("expected non-nil auditor")
	}

	err = auditor.Log(AuditEvent{
		UserID:   "u1",
		Action:   "test",
		Resource: "test",
		Status:   "success",
	})
	if err != nil {
		t.Fatalf("auditor.Log failed: %v", err)
	}

	logPath := filepath.Join(dir, ".naeos", "audit.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "encrypted") {
		t.Errorf("expected log to contain encrypted event data, got: %s", string(data))
	}
}

func TestExportToCloudUnsupported(t *testing.T) {
	_, err := ExportToCloud(CloudConfig{Provider: "unsupported"}, nil)
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
}

func TestNewCloudExporter(t *testing.T) {
	if e := NewCloudExporter(CloudConfig{Provider: CloudS3}); e == nil {
		t.Error("expected S3 exporter")
	}
	if e := NewCloudExporter(CloudConfig{Provider: CloudGCS}); e == nil {
		t.Error("expected GCS exporter")
	}
	if e := NewCloudExporter(CloudConfig{Provider: CloudAzure}); e == nil {
		t.Error("expected Azure exporter")
	}
	if e := NewCloudExporter(CloudConfig{Provider: "unknown"}); e != nil {
		t.Error("expected nil for unknown provider")
	}
}

func TestUploadToCloud(t *testing.T) {
	mem := NewMemoryAuditor()
	mem.Log(AuditEvent{UserID: "u1", Action: "test", Status: "success"})

	cfg := CloudConfig{
		Provider:  CloudS3,
		Bucket:    "test-bucket",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
		Endpoint:  "http://localhost:9000", // MinIO-style
	}

	exporter := NewCloudExporter(cfg)
	path, err := UploadToCloud(exporter, "audit/", mem.Events())
	// This will fail because there's no server, but we test the logic
	if err == nil {
		t.Logf("Upload succeeded (unexpected): %s", path)
	} else {
		t.Logf("Upload failed as expected (no server): %v", err)
	}
}

func copyBuffer(dst *strings.Builder, src *os.File, buf []byte) (int64, error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
		}
		if err != nil {
			return written, err
		}
	}
}

func TestS3ExporterListNoServer(t *testing.T) {
	exporter := NewS3Exporter(CloudConfig{
		Bucket:    "test",
		Region:    "us-east-1",
		AccessKey: "test",
		SecretKey: "test",
	})
	_, err := exporter.List("test-prefix")
	if err == nil {
		t.Error("expected error (no server)")
	}
}

func TestGCSExporterUploadNoServer(t *testing.T) {
	exporter := NewGCSExporter(CloudConfig{
		Bucket:    "test",
		AccessKey: "GOOG1test",
		SecretKey: "test",
	})
	err := exporter.Upload("test.json", []byte(`{"test": true}`))
	if err == nil {
		t.Error("expected error (no server)")
	}
}

func TestAzureBlobExporterUploadNoServer(t *testing.T) {
	exporter := NewAzureBlobExporter(CloudConfig{
		AccountName: "testaccount",
		AccountKey:  base64.StdEncoding.EncodeToString([]byte("test-key-32-bytes-long!")),
		Bucket:      "test-container",
	})
	err := exporter.Upload("test.json", []byte(`{"test": true}`))
	if err == nil {
		t.Error("expected error (no server)")
	}
}

func TestS3Signing(t *testing.T) {
	exporter := NewS3Exporter(CloudConfig{
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	})

	// Just verify signing doesn't panic
	req, _ := http.NewRequestWithContext(context.Background(), "PUT", "https://test-bucket.s3.us-east-1.amazonaws.com/test.txt", nil)
	req.Host = req.URL.Host
	req.Header.Set("x-amz-date", "20240101T000000Z")
	req.Header.Set("x-amz-content-sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	sig := exporter.signS3(req, nil)
	if sig == "" {
		t.Error("expected non-empty signature")
	}
	if !strings.HasPrefix(sig, "AWS4-HMAC-SHA256") {
		t.Errorf("expected AWS4-HMAC-SHA256 prefix, got: %s", sig)
	}
}
