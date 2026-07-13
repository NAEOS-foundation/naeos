package websocket

import (
	"testing"
)

func TestWSObserver(t *testing.T) {
	server := NewServer()
	go server.Run()

	broadcaster := NewEventBroadcaster(server)
	observer := NewWSObserver(broadcaster)

	observer.OnPipelineStart("test-pipeline-1")
	observer.OnPipelineComplete("test-pipeline-1", 5, "1.5s")
	observer.OnPipelineFailed("test-pipeline-2", "validation failed")
	observer.OnArtifactGenerated("main.go", "cmd/main.go")
}
