package cloud

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mockRunner struct {
	stdout []byte
	err    error
}

func (m *mockRunner) Run(name string, args []string, dir string) ([]byte, error) {
	return m.stdout, m.err
}

type recordingRunner struct {
	calls  [][]string
	stdout []byte
	err    error
}

func (r *recordingRunner) Run(name string, args []string, dir string) ([]byte, error) {
	r.calls = append(r.calls, append([]string{name}, args...))
	return r.stdout, r.err
}

func TestTerraformRunnerCreation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tr := NewTerraformRunner(dir)
	if tr.WorkDir != dir {
		t.Errorf("expected workdir %s, got %s", dir, tr.WorkDir)
	}
	if tr.Runner == nil {
		t.Error("expected non-nil runner")
	}
}

func TestTerraformRunnerWithMock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockRunner{stdout: []byte("{}")}
	tr := NewTerraformRunnerWithRunner(dir, mock)
	if tr.Runner != mock {
		t.Error("expected mock runner")
	}
}

func TestWriteHCL(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "nested", "tf")
	tr := NewTerraformRunner(dir)

	hcl := `resource "null_resource" "test" {}`
	if err := tr.writeHCL(hcl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "main.tf"))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != hcl {
		t.Errorf("HCL content mismatch:\ngot:  %s\nwant: %s", string(data), hcl)
	}
}

func TestInit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockRunner{stdout: []byte("ok")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	if err := tr.Init(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockRunner{err: fmt.Errorf("command failed")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	if err := tr.Init(); err == nil {
		t.Error("expected error from Init")
	}
}

func TestPlanJSON(t *testing.T) {
	t.Parallel()
	planJSON := `{"@level":"info","@message":"Plan: 2 to add, 0 to change, 1 to destroy.","changes":{"add":2,"change":0,"destroy":1}}`
	mock := &mockRunner{stdout: []byte(planJSON)}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)

	output, err := tr.Plan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 2 {
		t.Errorf("expected 2 adds, got %d", output.Changes.Add)
	}
	if output.Changes.Destroy != 1 {
		t.Errorf("expected 1 destroy, got %d", output.Changes.Destroy)
	}
}

func TestPlanMultiLineJSON(t *testing.T) {
	t.Parallel()
	planJSON := strings.Join([]string{
		`{"@level":"info","@message":"Terraform will perform the following actions:"}`,
		`{"@level":"info","@message":"Plan: 1 to add.","changes":{"add":1,"change":0,"destroy":0}}`,
	}, "\n")
	mock := &mockRunner{stdout: []byte(planJSON)}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)

	output, err := tr.Plan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 1 {
		t.Errorf("expected 1 add, got %d", output.Changes.Add)
	}
}

func TestApply(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockRunner{stdout: []byte("ok")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	if err := tr.Apply(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockRunner{err: fmt.Errorf("apply failed")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	if err := tr.Apply(); err == nil {
		t.Error("expected error from Apply")
	}
}

func TestDestroyAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mock := &mockRunner{stdout: []byte("ok")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	if err := tr.DestroyAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutput(t *testing.T) {
	t.Parallel()
	outputJSON := `{"bucket_id":{"value":"my-bucket","sensitive":false,"type":"string"}}`
	mock := &mockRunner{stdout: []byte(outputJSON)}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)

	output, err := tr.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := output["bucket_id"]; !ok {
		t.Error("missing bucket_id in output")
	} else if v.Value != "my-bucket" {
		t.Errorf("expected value 'my-bucket', got %v", v.Value)
	}
}

func TestOutputError(t *testing.T) {
	t.Parallel()
	mock := &mockRunner{err: fmt.Errorf("no outputs")}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)

	_, err := tr.Output()
	if err == nil {
		t.Error("expected error from Output")
	}
}

func TestDeploy(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "deploy-test")
	mock := &mockRunner{stdout: []byte("ok")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	hcl := `resource "null_resource" "test" {}`
	if err := tr.Deploy(hcl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "main.tf"))
	if err != nil {
		t.Fatalf("failed to read main.tf: %v", err)
	}
	if string(data) != hcl {
		t.Errorf("HCL mismatch: got %s", string(data))
	}
}

func TestDeployWritesHCLBeforeInit(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "hcl-order")
	mock := &mockRunner{stdout: []byte("ok")}
	tr := NewTerraformRunnerWithRunner(dir, mock)

	hcl := `resource "null_resource" "test" {}`
	if err := tr.Deploy(hcl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "main.tf"))
	if err != nil {
		t.Fatalf("main.tf not created: %v", err)
	}
	if string(data) != hcl {
		t.Errorf("HCL content mismatch: got %s", string(data))
	}
}

func TestTempWorkDir(t *testing.T) {
	t.Parallel()
	dir, err := TempWorkDir("naeos-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(dir)

	if !strings.HasPrefix(filepath.Base(dir), "naeos-test-terraform-") {
		t.Errorf("unexpected dir name: %s", filepath.Base(dir))
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestParsePlanJSONNoChanges(t *testing.T) {
	t.Parallel()
	mock := &mockRunner{stdout: []byte(`{"@message":"no changes"}`)}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)

	output, err := tr.Plan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 0 || output.Changes.Destroy != 0 {
		t.Errorf("expected zero changes, got add=%d destroy=%d", output.Changes.Add, output.Changes.Destroy)
	}
}

func TestDeployError(t *testing.T) {
	t.Parallel()
	mock := &mockRunner{err: fmt.Errorf("init failed")}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)

	err := tr.Deploy(`resource "null_resource" "test" {}`)
	if err == nil {
		t.Error("expected error from Deploy")
	}
}

func TestNewRunnerPoolDefaults(t *testing.T) {
	pool := NewRunnerPool(0, 0)
	if pool.maxSize != 16 {
		t.Errorf("expected default maxSize 16, got %d", pool.maxSize)
	}
	if pool.idleTTL != 30*time.Minute {
		t.Errorf("expected default idleTTL 30m, got %v", pool.idleTTL)
	}
}

func TestPlanError(t *testing.T) {
	t.Parallel()
	mock := &mockRunner{err: fmt.Errorf("plan failed")}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)
	_, err := tr.Plan()
	if err == nil {
		t.Error("expected error from Plan")
	}
}

func TestApplyDestroyError(t *testing.T) {
	t.Parallel()
	mock := &mockRunner{err: fmt.Errorf("destroy failed")}
	tr := NewTerraformRunnerWithRunner(t.TempDir(), mock)
	if err := tr.ApplyDestroy(); err == nil {
		t.Error("expected error from ApplyDestroy")
	}
}

func TestRunnerPoolGetPutRemove(t *testing.T) {
	t.Parallel()
	pool := NewRunnerPool(10, time.Minute)
	tr := NewTerraformRunner(t.TempDir())

	pool.Put("proj1", AWS, tr, true)

	got, ok := pool.Get("proj1", AWS)
	if !ok {
		t.Fatal("expected to find runner")
	}
	if got != tr {
		t.Error("expected same runner instance")
	}

	pool.Remove("proj1", AWS)
	_, ok = pool.Get("proj1", AWS)
	if ok {
		t.Error("expected runner to be removed")
	}
}

func TestRunnerPoolEviction(t *testing.T) {
	t.Parallel()
	pool := NewRunnerPool(2, time.Minute)

	tr1 := NewTerraformRunner(t.TempDir())
	tr2 := NewTerraformRunner(t.TempDir())
	tr3 := NewTerraformRunner(t.TempDir())

	pool.Put("p1", AWS, tr1, true)
	pool.Put("p2", GCP, tr2, true)
	pool.Put("p3", Azure, tr3, true)

	if pool.Size() > 2 {
		t.Errorf("expected pool size <= 2, got %d", pool.Size())
	}
}

func TestRunnerPoolNotFound(t *testing.T) {
	t.Parallel()
	pool := NewRunnerPool(10, time.Minute)
	_, ok := pool.Get("nonexistent", AWS)
	if ok {
		t.Error("expected not found")
	}
}

func TestParsePlanJSONEmpty(t *testing.T) {
	t.Parallel()
	output, err := parsePlanJSON([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 0 || output.Changes.Change != 0 || output.Changes.Destroy != 0 {
		t.Errorf("expected zero changes, got add=%d change=%d destroy=%d", output.Changes.Add, output.Changes.Change, output.Changes.Destroy)
	}
}

func TestParsePlanJSONNoSummary(t *testing.T) {
	t.Parallel()
	input := []byte(`{"@message":"Creating..."}
{"@message":"Still creating..."}`)
	output, err := parsePlanJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 0 || output.Changes.Change != 0 || output.Changes.Destroy != 0 {
		t.Errorf("expected zero changes, got add=%d change=%d destroy=%d", output.Changes.Add, output.Changes.Change, output.Changes.Destroy)
	}
}

func TestParsePlanJSONMalformed(t *testing.T) {
	t.Parallel()
	input := []byte(`not json
{broken json
{"@message":"test"}`)
	output, err := parsePlanJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 0 || output.Changes.Change != 0 || output.Changes.Destroy != 0 {
		t.Errorf("expected zero changes, got add=%d change=%d destroy=%d", output.Changes.Add, output.Changes.Change, output.Changes.Destroy)
	}
}

func TestParsePlanJSONValid(t *testing.T) {
	t.Parallel()
	input := []byte(`{"@message":"Planning..."}
{"@message":"Plan: 2 to add, 0 to change, 1 to destroy.","changes":{"add":2,"change":0,"destroy":1}}`)
	output, err := parsePlanJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Changes.Add != 2 {
		t.Errorf("expected 2 adds, got %d", output.Changes.Add)
	}
	if output.Changes.Destroy != 1 {
		t.Errorf("expected 1 destroy, got %d", output.Changes.Destroy)
	}
}

func TestAWSDestroy(t *testing.T) {
	t.Parallel()
	runner := &recordingRunner{stdout: []byte("ok")}
	adapter := &AWSAdapter{Runner: runner}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "test-destroy-aws",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
		},
	}

	if err := adapter.Destroy(config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundDestroy := false
	for _, call := range runner.calls {
		if call[0] == "terraform" && len(call) > 1 && call[1] == "destroy" {
			foundDestroy = true
			break
		}
	}
	if !foundDestroy {
		t.Errorf("expected terraform destroy call, got calls: %v", runner.calls)
	}
}

func TestGCPDestroy(t *testing.T) {
	t.Parallel()
	runner := &recordingRunner{stdout: []byte("ok")}
	adapter := &GCPAdapter{Runner: runner}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "test-destroy-gcp",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
		},
	}

	if err := adapter.Destroy(config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundDestroy := false
	for _, call := range runner.calls {
		if call[0] == "terraform" && len(call) > 1 && call[1] == "destroy" {
			foundDestroy = true
			break
		}
	}
	if !foundDestroy {
		t.Errorf("expected terraform destroy call, got calls: %v", runner.calls)
	}
}

func TestAzureDestroy(t *testing.T) {
	t.Parallel()
	runner := &recordingRunner{stdout: []byte("ok")}
	adapter := &AzureAdapter{Runner: runner}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "test-destroy-azure",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
		},
	}

	if err := adapter.Destroy(config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundDestroy := false
	for _, call := range runner.calls {
		if call[0] == "terraform" && len(call) > 1 && call[1] == "destroy" {
			foundDestroy = true
			break
		}
	}
	if !foundDestroy {
		t.Errorf("expected terraform destroy call, got calls: %v", runner.calls)
	}
}

func TestConcurrentStateAccess(t *testing.T) {
	t.Parallel()
	sm := NewStateManager()
	record := &DeploymentRecord{
		Project:      "test",
		Provider:     AWS,
		Environment:  "dev",
		Region:       "us-east-1",
		Status:       "deployed",
		TerraformDir: t.TempDir(),
	}

	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func() {
			sm.Save(record)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		go func() {
			sm.Load("test", AWS)
			done <- true
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}
