package supabase

import (
	"fmt"
	"os"
	"path/filepath"
)

type EdgeFunction struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	Entrypoint string `json:"entrypoint_path"`
	ImportMap  bool   `json:"import_map"`
	VerifyJWT  bool   `json:"verify_jwt"`
}

type DeployFunctionParams struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Entrypoint string `json:"entrypoint_path"`
	VerifyJWT  bool   `json:"verify_jwt"`
	ImportMap  bool   `json:"import_map"`
}

func (c *Client) ListFunctions() ([]EdgeFunction, error) {
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	data, err := c.do("GET", "/api/v1/projects/"+c.config.ProjectRef+"/functions", headers, nil)
	if err != nil {
		return nil, err
	}
	result, err := jsonUnmarshal[[]EdgeFunction](data)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) DeployFunction(slug, name, entrypoint, body string, verifyJWT, importMap bool) (*EdgeFunction, error) {
	params := DeployFunctionParams{
		Slug:       slug,
		Name:       name,
		Body:       body,
		Entrypoint: entrypoint,
		VerifyJWT:  verifyJWT,
		ImportMap:  importMap,
	}
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	data, err := c.do("POST", "/api/v1/projects/"+c.config.ProjectRef+"/functions", headers, params)
	if err != nil {
		return nil, err
	}
	return jsonUnmarshal[EdgeFunction](data)
}

func (c *Client) DeleteFunction(slug string) error {
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	_, err := c.do("DELETE", "/api/v1/projects/"+c.config.ProjectRef+"/functions/"+slug, headers, nil)
	if err != nil {
		return fmt.Errorf("delete function: %w", err)
	}
	return nil
}

func (c *Client) DeployFunctionFromFile(slug string, filePath string) (*EdgeFunction, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	name := slug
	entrypoint := "index.ts"
	return c.DeployFunction(slug, name, entrypoint, string(data), true, false)
}

func (c *Client) GetFunctionURL(slug string) string {
	return c.config.URL + "/functions/v1/" + slug
}

func (c *Client) InvokeFunction(slug string, body map[string]any) ([]byte, error) {
	headers := map[string]string{
		"apikey": c.config.AnonKey,
	}
	if token := c.AuthToken(); token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return c.do("POST", "/functions/v1/"+slug, headers, body)
}

func DefaultFunctionsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "supabase/functions"
	}
	return filepath.Join(home, ".naeos", "supabase", "functions")
}
