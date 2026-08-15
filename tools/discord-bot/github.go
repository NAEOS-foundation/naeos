package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubClient is a small standard-library client for the GitHub REST API.
type GitHubClient struct {
	repo      string
	token     string
	http      *http.Client
	baseURL   string
	userAgent string
}

// NewGitHubClient constructs a client for the given owner/repo.
func NewGitHubClient(repo, token string) *GitHubClient {
	return &GitHubClient{
		repo:      repo,
		token:     token,
		http:      &http.Client{Timeout: 15 * time.Second},
		baseURL:   "https://api.github.com",
		userAgent: "naeos-discord-bot/1.0",
	}
}

type repoStatus struct {
	Stars       int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	OpenIssues  int       `json:"open_issues_count"`
	License     string    `json:"-"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
	Archived    bool      `json:"archived"`
	HTMLURL     string    `json:"html_url"`
}

type repoLicense struct {
	License struct {
		SPDXID string `json:"spdx_id"`
	} `json:"license"`
}

type release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

// Repo returns the configured repository in owner/repo form.
func (c *GitHubClient) Repo() string {
	return c.repo
}

// Status fetches repository summary information.
func (c *GitHubClient) Status(ctx context.Context) (*repoStatus, error) {
	var st repoStatus
	if err := c.getJSON(ctx, "/repos/"+c.repo, &st); err != nil {
		return nil, err
	}
	var lic repoLicense
	if err := c.getJSON(ctx, "/repos/"+c.repo+"/license", &lic); err == nil && lic.License.SPDXID != "" {
		st.License = lic.License.SPDXID
	}
	return &st, nil
}

// LatestRelease fetches the newest published (non-draft) release.
func (c *GitHubClient) LatestRelease(ctx context.Context) (*release, error) {
	rels, err := c.releases(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return nil, fmt.Errorf("no releases found for %s", c.repo)
	}
	return &rels[0], nil
}

// releases fetches up to n stable releases, skipping drafts and prereleases.
func (c *GitHubClient) releases(ctx context.Context, n int) ([]release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases?per_page=%d", c.baseURL, c.repo, n)
	var rels []release
	if err := c.getJSON(ctx, url, &rels); err != nil {
		return nil, err
	}
	out := make([]release, 0, len(rels))
	for _, r := range rels {
		if r.Draft || r.Prerelease {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *GitHubClient) getJSON(ctx context.Context, url string, v any) error {
	if !strings.HasPrefix(url, "http") {
		url = c.baseURL + url
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("github api %s: %s: %s", resp.Status, url, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
