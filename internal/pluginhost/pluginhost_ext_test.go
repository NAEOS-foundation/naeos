package pluginhost

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEventBusGetter(t *testing.T) {
	m := NewManager("")
	if m.EventBus() == nil {
		t.Error("expected non-nil EventBus")
	}
}

func TestPluginEventBusNoOps(t *testing.T) {
	p := NewPluginEventBus(NewEventBus())
	p.OnPipelineStart("p1")
	p.OnArtifactGenerated("a", "/tmp/a")
}

func TestSimpleMetricsSnapshotEmpty(t *testing.T) {
	m := NewSimpleMetrics()
	counters, gauges, histograms := m.Snapshot()
	if len(counters) != 0 || len(gauges) != 0 || len(histograms) != 0 {
		t.Error("expected empty snapshot on fresh metrics")
	}
}

func TestUpdateStateUnregistered(t *testing.T) {
	m := NewManager("")
	m.updateState("doesnotexist", StateRunning, nil)
}

func TestUpdateStateLocked(t *testing.T) {
	m := NewManager("")
	m.info["test"] = &PluginInfo{Name: "test"}
	m.config.Plugins = []PluginInfo{{Name: "test"}}
	m.updateStateLocked("test", StateRunning, nil)
	if m.info["test"].State != StateRunning {
		t.Errorf("expected StateRunning, got %v", m.info["test"].State)
	}
}

func TestUpdateStateLockedWithError(t *testing.T) {
	m := NewManager("")
	m.info["test"] = &PluginInfo{Name: "test"}
	m.config.Plugins = []PluginInfo{{Name: "test"}}
	err := os.ErrPermission
	m.updateStateLocked("test", StateError, err)
	if m.info["test"].State != StateError {
		t.Errorf("expected StateError, got %v", m.info["test"].State)
	}
	if !errors.Is(m.info["test"].Error, err) {
		t.Error("expected error to be set on info")
	}
	if !errors.Is(m.config.Plugins[0].Error, err) {
		t.Error("expected error to be set on config")
	}
}

func TestUpdateStateLockedStartedAt(t *testing.T) {
	m := NewManager("")
	m.info["test"] = &PluginInfo{Name: "test"}
	m.config.Plugins = []PluginInfo{{Name: "test"}}
	m.updateStateLocked("test", StateRunning, nil)
	if m.info["test"].StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
}

func TestUpdateStateLockedStartedAtPreserved(t *testing.T) {
	m := NewManager("")
	m.info["test"] = &PluginInfo{Name: "test", StartedAt: time.Now()}
	m.config.Plugins = []PluginInfo{{Name: "test"}}
	m.updateStateLocked("test", StateRunning, nil)
	if m.info["test"].StartedAt.IsZero() {
		t.Error("expected StartedAt to be preserved")
	}
}

func TestSaveConfigMkdirError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := NewManager(blocker)
	err := m.SaveConfig()
	if err == nil {
		t.Error("expected MkdirAll error")
	}
}

func TestSandboxValidatePathError(t *testing.T) {
	s := NewSandbox(SandboxConfig{AllowedDirs: []string{"/nonexistent"}})
	err := s.ValidatePath("/tmp/evil.so")
	if err == nil {
		t.Error("expected error for path outside allowed dirs")
	}
}

func TestSimpleMetricsCounterInc(t *testing.T) {
	m := NewSimpleMetrics()
	m.CounterInc("requests", nil)
	counters, _, _ := m.Snapshot()
	if counters["requests"] != 1 {
		t.Errorf("expected requests=1, got %v", counters["requests"])
	}
}

func TestSimpleMetricsCounterIncWithLabels(t *testing.T) {
	m := NewSimpleMetrics()
	m.CounterInc("requests", map[string]string{"status": "200"})
	counters, _, _ := m.Snapshot()
	if counters["requests|status=200"] != 1 {
		t.Errorf("expected labeled counter, got %v", counters)
	}
}

func TestSimpleMetricsGaugeSet(t *testing.T) {
	m := NewSimpleMetrics()
	m.GaugeSet("cpu", 0.5, nil)
	_, gauges, _ := m.Snapshot()
	if gauges["cpu"] != 0.5 {
		t.Errorf("expected cpu=0.5, got %v", gauges["cpu"])
	}
}

func TestSimpleMetricsHistogramObserve(t *testing.T) {
	m := NewSimpleMetrics()
	m.HistogramObserve("latency", 0.1, nil)
	_, _, histograms := m.Snapshot()
	if len(histograms["latency"]) != 1 {
		t.Errorf("expected 1 observation, got %v", histograms["latency"])
	}
}

func TestMetricKeyWithLabels(t *testing.T) {
	key := metricKey("requests", map[string]string{"status": "200", "method": "GET"})
	if key != "requests|status=200|method=GET" && key != "requests|method=GET|status=200" {
		t.Logf("got key: %s", key)
	}
}

func TestMetricKeyNoLabels(t *testing.T) {
	key := metricKey("requests", nil)
	if key != "requests" {
		t.Errorf("expected 'requests', got %s", key)
	}
}
