package supabase

import (
	"encoding/json"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type QueryParams struct {
	Query string `json:"query"`
}

type QueryResult struct {
	Rows  []map[string]any `json:"rows"`
	Error string           `json:"error"`
}

type managementQueryResponse []map[string]any

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
	path := "/v1/projects/" + c.config.ProjectRef + "/database/query"
	params := QueryParams{Query: query}
	data, err := c.doManagement("POST", path, nil, params)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrCloud, "execute SQL")
	}
	var rows managementQueryResponse
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "decode response")
	}
	result := &QueryResult{Rows: make([]map[string]any, len(rows))}
	for i, r := range rows {
		result.Rows[i] = r
	}
	return result, nil
}

func (c *Client) GetRoles() ([]Role, error) {
	data, err := c.doManagement("GET", "/v1/projects/"+c.config.ProjectRef+"/database/roles", nil, nil)
	if err != nil {
		return nil, err
	}
	result, err := jsonUnmarshal[[]Role](data)
	if err != nil {
		return nil, err
	}
	return *result, nil
}
