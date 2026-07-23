package supabase

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Bucket struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	Public    bool   `json:"public"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type FileObject struct {
	Name           string `json:"name"`
	BucketID       string `json:"bucket_id"`
	Owner          string `json:"owner"`
	ID             string `json:"id"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	LastAccessedAt string `json:"last_accessed_at"`
	Metadata       struct {
		Size       int    `json:"size"`
		Mimetype   string `json:"mimetype"`
		CacheControl string `json:"cacheControl"`
	} `json:"metadata"`
}

type CreateBucketParams struct {
	Name          string `json:"name"`
	Public        bool   `json:"public"`
}

func (c *Client) ListBuckets() ([]Bucket, error) {
	resp, err := c.doAuth("GET", "/storage/v1/bucket", nil)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]Bucket](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) CreateBucket(name string, public bool) (*Bucket, error) {
	params := CreateBucketParams{Name: name, Public: public}
	resp, err := c.doAuth("POST", "/storage/v1/bucket", params)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Bucket](resp)
}

func (c *Client) GetBucket(id string) (*Bucket, error) {
	resp, err := c.doAuth("GET", "/storage/v1/bucket/"+id, nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse[Bucket](resp)
}

func (c *Client) DeleteBucket(id string) error {
	resp, err := c.doAuth("DELETE", "/storage/v1/bucket/"+id, nil)
	if err != nil {
		return fmt.Errorf("delete bucket: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("delete bucket: %d %s", resp.StatusCode, strings.TrimSpace(string(body[:n])))
	}
	return nil
}

func (c *Client) ListFiles(bucket, prefix string) ([]FileObject, error) {
	params := map[string]any{
		"prefix": prefix,
	}
	resp, err := c.doAuth("POST", "/storage/v1/object/list/"+bucket, params)
	if err != nil {
		return nil, err
	}
	result, err := decodeResponse[[]FileObject](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) UploadFile(bucket, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	writer.Close()

	url := c.config.URL + "/storage/v1/object/" + bucket + "/" + remotePath
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.AuthToken())
	req.Header.Set("apikey", c.config.AnonKey)
	req.ContentLength = stat.Size()

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(body))
	}
	return nil
}

func (c *Client) DownloadFile(bucket, remotePath, localPath string) error {
	url := c.config.URL + "/storage/v1/object/" + bucket + "/" + remotePath
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AuthToken())
	req.Header.Set("apikey", c.config.AnonKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (c *Client) DeleteFile(bucket, path string) error {
	params := map[string]any{
		"prefixes": []string{path},
	}
	resp, err := c.doAuth("DELETE", "/storage/v1/object/"+bucket, params)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("delete file: %d %s", resp.StatusCode, strings.TrimSpace(string(body[:n])))
	}
	return nil
}
