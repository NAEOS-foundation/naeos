package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Resource describes an entity exposed via the MCP resources interface.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContents is the payload returned by resources/read.
type ResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

const (
	mimeMarkdown = "text/markdown"
	mimePlain    = "text/plain"
	mimeJSON     = "application/json"
)

const (
	uriSchemeDocs      = "naeos://docs/"
	uriSchemeArtifacts = "naeos://artifacts/"
	uriSchemeJobs      = "naeos://jobs/"
)

// conceptNames returns the documentation concepts available as resources,
// in deterministic order.
func conceptNames() []string {
	return []string{
		"pipeline", "neir", "spec", "kernel", "policy",
		"profile", "compiler", "context", "module", "service",
	}
}

// listResources returns all resources exposed by the server: static
// documentation concepts, artifacts from the artifact store (if configured),
// and tracked pipeline jobs.
func (s *Server) listResources() []Resource {
	resources := make([]Resource, 0, len(conceptNames()))
	for _, name := range conceptNames() {
		resources = append(resources, Resource{
			URI:         uriSchemeDocs + name,
			Name:        "NAEOS Concept: " + name,
			Description: "Documentation for the NAEOS concept \"" + name + "\"",
			MimeType:    mimePlain,
		})
	}

	if s.store != nil {
		for _, a := range s.store.List() {
			resources = append(resources, Resource{
				URI:         uriSchemeArtifacts + a.Path,
				Name:        a.Path,
				Description: fmt.Sprintf("Artifact (%s, %d bytes)", a.Kind, a.Size),
				MimeType:    artifactMimeType(a.Kind),
			})
		}
	}

	s.mu.RLock()
	for id, job := range s.pipelineJobs {
		status := job.Status
		if status == "" {
			status = "unknown"
		}
		resources = append(resources, Resource{
			URI:         uriSchemeJobs + id,
			Name:        "Pipeline Job " + id,
			Description: fmt.Sprintf("Pipeline job status: %s", status),
			MimeType:    mimeJSON,
		})
	}
	s.mu.RUnlock()

	sort.Slice(resources, func(i, j int) bool { return resources[i].URI < resources[j].URI })
	return resources
}

// readResource returns the contents of a single resource identified by URI.
func (s *Server) readResource(uri string) ([]ResourceContents, error) {
	switch {
	case strings.HasPrefix(uri, uriSchemeDocs):
		concept := strings.TrimPrefix(uri, uriSchemeDocs)
		explanation := s.explainConcept(concept)
		if strings.HasPrefix(explanation, "Unknown concept:") {
			return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("resource %q not found; available docs: %s", uri, strings.Join(conceptNames(), ", ")))
		}
		return []ResourceContents{{URI: uri, MimeType: mimePlain, Text: explanation}}, nil

	case strings.HasPrefix(uri, uriSchemeArtifacts):
		if s.store == nil {
			return nil, naeoserr.New(naeoserr.ErrNotFound, "no artifact store configured")
		}
		path := strings.TrimPrefix(uri, uriSchemeArtifacts)
		artifact, ok := s.store.Get(path)
		if !ok {
			return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("artifact %q not found; use resources/list to discover artifact URIs", path))
		}
		return []ResourceContents{{
			URI:      uri,
			MimeType: artifactMimeType(artifact.Kind),
			Text:     string(artifact.Content),
		}}, nil

	case strings.HasPrefix(uri, uriSchemeJobs):
		jobID := strings.TrimPrefix(uri, uriSchemeJobs)
		s.mu.RLock()
		job, ok := s.pipelineJobs[jobID]
		s.mu.RUnlock()
		if !ok {
			return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("pipeline job %q not found", jobID))
		}
		data, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshal pipeline job")
		}
		return []ResourceContents{{URI: uri, MimeType: mimeJSON, Text: string(data)}}, nil

	default:
		return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("unsupported resource URI scheme: %q; expected naeos://docs/, naeos://artifacts/, or naeos://jobs/", uri))
	}
}

func artifactMimeType(kind artifacts.ArtifactKind) string {
	switch kind {
	case artifacts.KindDocs:
		return mimeMarkdown
	case artifacts.KindCode, artifacts.KindConfig, artifacts.KindTest, artifacts.KindMigration:
		return mimePlain
	default:
		return mimePlain
	}
}
