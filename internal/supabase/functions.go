package supabase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type EdgeFunction struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Status       string `json:"status"`
	Version      int    `json:"version"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Entrypoint   string `json:"entrypoint_path"`
	ImportMap    bool   `json:"import_map"`
	VerifyJWT    bool   `json:"verify_jwt"`
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
	resp, err := c.do("GET", "/api/v1/projects/"+c.config.ProjectRef+"/functions", headers, nil)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]EdgeFunction](resp)
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
	resp, err := c.do("POST", "/api/v1/projects/"+c.config.ProjectRef+"/functions", headers, params)
	if err != nil {
		return nil, err
	}
	return decodeResponse[EdgeFunction](resp)
}

func (c *Client) DeleteFunction(slug string) error {
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	resp, err := c.do("DELETE", "/api/v1/projects/"+c.config.ProjectRef+"/functions/"+slug, headers, nil)
	if err != nil {
		return fmt.Errorf("delete function: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("delete function: %d %s", resp.StatusCode, strings.TrimSpace(string(body[:n])))
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
	resp, err := c.do("POST", "/functions/v1/"+slug, headers, body)
	if err != nil {
		return nil, err
	}
	return decodeRaw(resp)
}

func DefaultFunctionsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "supabase/functions"
	}
	return filepath.Join(home, ".naeos", "supabase", "functions")
}
