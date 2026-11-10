// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type serviceIdentityContextKey struct{}

const (
	githubRepositoryAPIBase = "https://api.github.com"
	githubRepositoryMaxBody = 1 << 20
)

type repositoryMetadata struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Owner         string    `json:"owner"`
	Description   string    `json:"description"`
	Private       bool      `json:"private"`
	Archived      bool      `json:"archived"`
	DefaultBranch string    `json:"default_branch"`
	OpenIssues    int       `json:"open_issues"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	SizeKB        int       `json:"size_kb"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	URL           string    `json:"url"`
}

type githubRepositoryResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	Private         bool      `json:"private"`
	Archived        bool      `json:"archived"`
	DefaultBranch   string    `json:"default_branch"`
	OpenIssuesCount int       `json:"open_issues_count"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	Size            int       `json:"size"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	HTMLURL         string    `json:"html_url"`
	Owner           struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func isRepositoryMetadataPath(path string) bool {
	const prefix = "/api/v1/repositories/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	return len(parts) == 2 && validRepositorySegment(parts[0]) && validRepositorySegment(parts[1])
}

func validRepositorySegment(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func parseAllowedRepositories(raw string) (map[string]struct{}, error) {
	allowed := make(map[string]struct{})
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("repository allowlist is not configured")
	}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		parts := strings.Split(entry, "/")
		if len(parts) != 2 || !validRepositorySegment(parts[0]) || !validRepositorySegment(parts[1]) {
			return nil, fmt.Errorf("invalid repository allowlist entry")
		}
		allowed[strings.ToLower(parts[0]+"/"+parts[1])] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, errors.New("repository allowlist is empty")
	}
	return allowed, nil
}

func (s *Server) handleRepositoryMetadata(w http.ResponseWriter, r *http.Request) {
	serviceIdentity, ok := r.Context().Value(serviceIdentityContextKey{}).(bool)
	if !ok || !serviceIdentity {
		s.writeError(w, http.StatusUnauthorized, "service identity required")
		return
	}
	apiBase := s.repositoryAPIBase
	if apiBase == "" {
		apiBase = githubRepositoryAPIBase
	}
	client := s.repositoryHTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	s.serveRepositoryMetadata(w, r, os.Getenv("NAEOS_GITHUB_ALLOWED_REPOSITORIES"),
		os.Getenv("GITHUB_TOKEN"), apiBase, client)
}

func (s *Server) serveRepositoryMetadata(w http.ResponseWriter, r *http.Request, allowlist, token, apiBase string, client *http.Client) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	const prefix = "/api/v1/repositories/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		s.writeError(w, http.StatusBadRequest, "repository path is invalid")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if len(parts) != 2 || !validRepositorySegment(parts[0]) || !validRepositorySegment(parts[1]) {
		s.writeError(w, http.StatusBadRequest, "expected repository owner and name")
		return
	}
	owner, name := parts[0], parts[1]
	allowed, err := parseAllowedRepositories(allowlist)
	if err != nil {
		s.writeError(w, http.StatusServiceUnavailable, "repository metadata integration is not configured")
		return
	}
	if _, ok := allowed[strings.ToLower(owner+"/"+name)]; !ok {
		s.writeError(w, http.StatusForbidden, "repository is not allowlisted")
		return
	}
	base, err := url.Parse(strings.TrimSpace(apiBase))
	if err != nil || base.Host == "" || (base.Scheme != "https" && (base.Scheme != "http" || (base.Hostname() != "localhost" && base.Hostname() != "127.0.0.1"))) {
		s.writeError(w, http.StatusServiceUnavailable, "repository metadata provider is not configured")
		return
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	endpoint := strings.TrimRight(base.String(), "/") + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
	// #nosec G704 -- endpoint base is scheme/host validated and path segments are allowlisted.
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "could not create repository metadata request")
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "naeos-core-service-integration")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	// #nosec G704 -- only the validated GitHub API base and restricted repo path are requested.
	resp, err := client.Do(req)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "repository metadata provider unavailable")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		s.writeError(w, http.StatusNotFound, "repository not found")
		return
	}
	if resp.StatusCode != http.StatusOK {
		s.writeError(w, http.StatusBadGateway, "repository metadata provider returned an unsuccessful response")
		return
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, githubRepositoryMaxBody+1))
	if err != nil || len(body) > githubRepositoryMaxBody {
		s.writeError(w, http.StatusBadGateway, "repository metadata response is invalid or too large")
		return
	}
	var source githubRepositoryResponse
	if err := json.Unmarshal(body, &source); err != nil || source.ID <= 0 ||
		!strings.EqualFold(source.FullName, owner+"/"+name) ||
		!strings.EqualFold(source.Owner.Login, owner) ||
		!strings.EqualFold(source.Name, name) {
		s.writeError(w, http.StatusBadGateway, "repository metadata response is invalid")
		return
	}
	result := repositoryMetadata{
		ID:            strconv.FormatInt(source.ID, 10),
		Name:          source.Name,
		FullName:      source.FullName,
		Owner:         source.Owner.Login,
		Description:   source.Description,
		Private:       source.Private,
		Archived:      source.Archived,
		DefaultBranch: source.DefaultBranch,
		OpenIssues:    source.OpenIssuesCount,
		Stars:         source.StargazersCount,
		Forks:         source.ForksCount,
		SizeKB:        source.Size,
		CreatedAt:     source.CreatedAt,
		UpdatedAt:     source.UpdatedAt,
		URL:           source.HTMLURL,
	}
	w.Header().Set("Cache-Control", "no-store")
	s.writeJSON(w, http.StatusOK, result)
}
