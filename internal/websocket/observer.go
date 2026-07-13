package websocket

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

var _ pipeline.PipelineObserver = (*WSObserver)(nil)

type WSObserver struct {
	broadcaster *EventBroadcaster
}

func NewWSObserver(b *EventBroadcaster) *WSObserver {
	return &WSObserver{broadcaster: b}
}

func (o *WSObserver) OnPipelineStart(pipelineID string) {
	o.broadcaster.PipelineStarted(pipelineID)
}

func (o *WSObserver) OnPipelineComplete(pipelineID string, artifacts int, duration string) {
	o.broadcaster.PipelineCompleted(pipelineID, fmt.Sprintf("%s (%d artifacts)", duration, artifacts))
}

func (o *WSObserver) OnPipelineFailed(pipelineID string, errMsg string) {
	o.broadcaster.PipelineFailed(pipelineID, errMsg)
}

func (o *WSObserver) OnArtifactGenerated(name string, path string) {
	o.broadcaster.ArtifactGenerated(name, path)
}
