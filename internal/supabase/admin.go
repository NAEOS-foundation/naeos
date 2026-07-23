package supabase

import (
	"fmt"
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
	data, err := c.do("POST", path, headers, params)
	if err != nil {
		return nil, fmt.Errorf("execute SQL: %w", err)
	}
	return jsonUnmarshal[QueryResult](data)
}

func (c *Client) GetRoles() ([]Role, error) {
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	data, err := c.do("GET", "/api/v1/projects/"+c.config.ProjectRef+"/database/roles", headers, nil)
	if err != nil {
		return nil, err
	}
	result, err := jsonUnmarshal[[]Role](data)
	if err != nil {
		return nil, err
	}
	return *result, nil
}
