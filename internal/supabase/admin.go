package supabase

import (
	"fmt"
	"strings"
)

type QueryParams struct {
	Query string `json:"query"`
}

type QueryResult struct {
	Rows  []map[string]any `json:"rows"`
	Error string           `json:"error"`
}

type ProjectInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Region       string `json:"region"`
	CreatedAt    string `json:"created_at"`
}

type Role struct {
	Name string `json:"role"`
}

type APIInfo struct {
	ProjectRef string `json:"project_ref"`
	AnonKey    string `json:"anon_key"`
	URL        string `json:"url"`
}

func (c *Client) ExecuteSQL(query string) (*QueryResult, error) {
	path := "/api/v1/projects/" + c.config.ProjectRef + "/database/query"
	params := QueryParams{Query: query}
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	resp, err := c.do("POST", path, headers, params)
	if err != nil {
		return nil, fmt.Errorf("execute SQL: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body := make([]byte, 4096)
		n, _ := resp.Body.Read(body)
		return nil, fmt.Errorf("execute SQL: %d %s", resp.StatusCode, strings.TrimSpace(string(body[:n])))
	}

	return decodeResponse[QueryResult](resp)
}

func (c *Client) GetRoles() ([]Role, error) {
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	resp, err := c.do("GET", "/api/v1/projects/"+c.config.ProjectRef+"/database/roles", headers, nil)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]Role](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}
