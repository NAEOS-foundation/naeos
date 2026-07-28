package supabase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
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
		Size         int    `json:"size"`
		Mimetype     string `json:"mimetype"`
		CacheControl string `json:"cacheControl"`
	} `json:"metadata"`
}

type CreateBucketParams struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

func (c *Client) ListBuckets() ([]Bucket, error) {
	data, err := c.doAuth("GET", "/storage/v1/bucket", nil)
	if err != nil {
		return nil, err
	}
	result, err := jsonUnmarshal[[]Bucket](data)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) CreateBucket(name string, public bool) (*Bucket, error) {
	params := CreateBucketParams{Name: name, Public: public}
	data, err := c.doAuth("POST", "/storage/v1/bucket", params)
	if err != nil {
		return nil, err
	}
	return jsonUnmarshal[Bucket](data)
}

func (c *Client) GetBucket(id string) (*Bucket, error) {
	data, err := c.doAuth("GET", "/storage/v1/bucket/"+id, nil)
	if err != nil {
		return nil, err
	}
	return jsonUnmarshal[Bucket](data)
}

func (c *Client) DeleteBucket(id string) error {
	_, err := c.doAuth("DELETE", "/storage/v1/bucket/"+id, nil)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrCloud, "delete bucket")
	}
	return nil
}

func (c *Client) ListFiles(bucket, prefix string) ([]FileObject, error) {
	params := map[string]any{
		"prefix": prefix,
	}
	data, err := c.doAuth("POST", "/storage/v1/object/list/"+bucket, params)
	if err != nil {
		return nil, err
	}
	result, err := jsonUnmarshal[[]FileObject](data)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) UploadFile(bucket, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "open file")
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create form file")
	}
	if _, err := io.Copy(part, file); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "copy file")
	}
	writer.Close()

	url := c.config.URL + "/storage/v1/object/" + bucket + "/" + remotePath
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, &buf)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create request")
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.AuthToken())
	req.Header.Set("apikey", c.config.AnonKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrCloud, "upload request")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return naeoserr.New(naeoserr.ErrCloud, fmt.Sprintf("upload failed: %s", string(body)))
	}
	return nil
}

func (c *Client) DownloadFile(bucket, remotePath, localPath string) error {
	url := c.config.URL + "/storage/v1/object/" + bucket + "/" + remotePath
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create request")
	}

	req.Header.Set("Authorization", "Bearer "+c.AuthToken())
	req.Header.Set("apikey", c.config.AnonKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrCloud, "download request")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return naeoserr.New(naeoserr.ErrCloud, fmt.Sprintf("download failed: status %d", resp.StatusCode))
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create parent dir")
	}

	out, err := os.Create(localPath)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create local file")
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "write file")
	}
	return nil
}

func (c *Client) DeleteFile(bucket, path string) error {
	params := map[string]any{
		"prefixes": []string{path},
	}
	_, err := c.doAuth("DELETE", "/storage/v1/object/"+bucket, params)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrCloud, "delete file")
	}
	return nil
}
