package monitoring

import (
	"time"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

var _ pipeline.PipelineObserver = (*MetricsObserver)(nil)

type MetricsObserver struct {
	metrics *Metrics
}

func NewMetricsObserver(m *Metrics) *MetricsObserver {
	return &MetricsObserver{metrics: m}
}

func (o *MetricsObserver) OnPipelineStart(pipelineID string) {}

func (o *MetricsObserver) OnPipelineComplete(pipelineID string, artifacts int, duration string) {
	o.metrics.IncPipelines("success")
	o.metrics.IncArtifacts()
	if d, err := time.ParseDuration(duration); err == nil {
		o.metrics.ObservePipelineDuration(d.Seconds())
	}
}

func (o *MetricsObserver) OnPipelineFailed(pipelineID string, errMsg string) {
	o.metrics.IncPipelines("failure")
}

func (o *MetricsObserver) OnArtifactGenerated(name string, path string) {
	o.metrics.IncArtifacts()
}
