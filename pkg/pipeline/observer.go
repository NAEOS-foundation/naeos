package pipeline

var _ PipelineObserver = (*multiObserver)(nil)

type multiObserver struct {
	observers []PipelineObserver
}

func ChainObservers(observers ...PipelineObserver) PipelineObserver {
	return &multiObserver{observers: observers}
}

func (m *multiObserver) OnPipelineStart(pipelineID string) {
	for _, o := range m.observers {
		o.OnPipelineStart(pipelineID)
	}
}

func (m *multiObserver) OnPipelineComplete(pipelineID string, artifacts int, duration string) {
	for _, o := range m.observers {
		o.OnPipelineComplete(pipelineID, artifacts, duration)
	}
}

func (m *multiObserver) OnPipelineFailed(pipelineID string, errMsg string) {
	for _, o := range m.observers {
		o.OnPipelineFailed(pipelineID, errMsg)
	}
}

func (m *multiObserver) OnArtifactGenerated(name string, path string) {
	for _, o := range m.observers {
		o.OnArtifactGenerated(name, path)
	}
}
