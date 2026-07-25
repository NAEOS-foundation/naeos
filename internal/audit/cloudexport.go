package audit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CloudProvider string

const (
	CloudS3    CloudProvider = "s3"
	CloudGCS   CloudProvider = "gcs"
	CloudAzure CloudProvider = "azure"
)

type CloudConfig struct {
	Provider   CloudProvider
	Bucket     string
	Prefix     string
	Region     string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	AccountName string
	AccountKey  string
}

type CloudExporter interface {
	Upload(path string, data []byte) error
	List(path string) ([]string, error)
}

func NewCloudExporter(cfg CloudConfig) CloudExporter {
	switch cfg.Provider {
	case CloudS3:
		return NewS3Exporter(cfg)
	case CloudGCS:
		return NewGCSExporter(cfg)
	case CloudAzure:
		return NewAzureBlobExporter(cfg)
	default:
		return nil
	}
}

type S3Exporter struct {
	cfg CloudConfig
	client *http.Client
}

func NewS3Exporter(cfg CloudConfig) *S3Exporter {
	return &S3Exporter{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *S3Exporter) Upload(path string, data []byte) error {
	endpoint := e.cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", e.cfg.Bucket, e.cfg.Region)
	}

	url := fmt.Sprintf("%s/%s", endpoint, path)
	body := bytes.NewReader(data)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-amz-content-sha256", fmt.Sprintf("%x", sha256.Sum256(data)))
	req.Header.Set("x-amz-date", time.Now().UTC().Format("20060102T150405Z"))

	sig := e.signS3(req, data)
	req.Header.Set("Authorization", sig)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("s3 upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("s3 upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (e *S3Exporter) List(path string) ([]string, error) {
	endpoint := e.cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", e.cfg.Bucket, e.cfg.Region)
	}

	url := fmt.Sprintf("%s/?prefix=%s", endpoint, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create list request: %w", err)
	}

	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("x-amz-content-sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	req.Header.Set("x-amz-date", time.Now().UTC().Format("20060102T150405Z"))

	sig := e.signS3(req, nil)
	req.Header.Set("Authorization", sig)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("s3 list: %w", err)
	}
	defer resp.Body.Close()

	var listResult struct {
		XMLName xml.Name `xml:"ListBucketResult"`
		Contents []struct {
			Key string `xml:"Key"`
		} `xml:"Contents"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}

	var keys []string
	for _, c := range listResult.Contents {
		keys = append(keys, c.Key)
	}
	return keys, nil
}

func (e *S3Exporter) signS3(req *http.Request, data []byte) string {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	hashedPayload := fmt.Sprintf("%x", sha256.Sum256(data))

	// Canonical request
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		req.Host, hashedPayload, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalReq := fmt.Sprintf("%s\n/%s\n\n%s\n%s\n%s",
		req.Method, strings.TrimPrefix(req.URL.Path, "/"),
		canonicalHeaders, signedHeaders, hashedPayload)

	canonicalReqHash := fmt.Sprintf("%x", sha256.Sum256([]byte(canonicalReq)))

	// String to sign
	scope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, e.cfg.Region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, scope, canonicalReqHash)

	// Signing key
	signingKey := e.signingKey(dateStamp)
	signature := fmt.Sprintf("%x", hmacSHA256(signingKey, []byte(stringToSign)))

	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		e.cfg.AccessKey, scope, signedHeaders, signature)
}

func (e *S3Exporter) signingKey(dateStamp string) []byte {
	kSecret := []byte("AWS4" + e.cfg.SecretKey)
	kDate := hmacSHA256(kSecret, []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(e.cfg.Region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

type GCSExporter struct {
	cfg    CloudConfig
	client *http.Client
}

func NewGCSExporter(cfg CloudConfig) *GCSExporter {
	return &GCSExporter{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *GCSExporter) Upload(path string, data []byte) error {
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", e.cfg.Bucket, path)
	body := bytes.NewReader(data)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-date", time.Now().UTC().Format(time.RFC1123))
	req.Header.Set("x-goog-content-length", fmt.Sprintf("%d", len(data)))

	sig := e.signGCS(req, data)
	req.Header.Set("Authorization", sig)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("gcs upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gcs upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (e *GCSExporter) List(path string) ([]string, error) {
	url := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o?prefix=%s", e.cfg.Bucket, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create list request: %w", err)
	}

	req.Header.Set("x-goog-date", time.Now().UTC().Format(time.RFC1123))

	sig := e.signGCS(req, nil)
	req.Header.Set("Authorization", sig)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gcs list: %w", err)
	}
	defer resp.Body.Close()

	var listResult struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}

	var names []string
	for _, item := range listResult.Items {
		names = append(names, item.Name)
	}
	return names, nil
}

func (e *GCSExporter) signGCS(req *http.Request, data []byte) string {
	if e.cfg.AccessKey == "" || e.cfg.SecretKey == "" {
		return ""
	}

	payload := sha256Hash(data)
	date := time.Now().UTC().Format("20060102T150405Z")

	// GCS HMAC signing (S3-compatible style)
	stringToSign := fmt.Sprintf("%s\n\n\n%s\n%s", payload, date, req.URL.Path)

	mac := hmac.New(sha256.New, []byte(e.cfg.SecretKey))
	mac.Write([]byte(stringToSign))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("GOOG1 %s:%s", e.cfg.AccessKey, sig)
}

func sha256Hash(data []byte) string {
	if len(data) == 0 {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

type AzureBlobExporter struct {
	cfg    CloudConfig
	client *http.Client
}

func NewAzureBlobExporter(cfg CloudConfig) *AzureBlobExporter {
	return &AzureBlobExporter{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *AzureBlobExporter) Upload(path string, data []byte) error {
	url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", e.cfg.AccountName, e.cfg.Bucket, path)
	body := bytes.NewReader(data)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("x-ms-date", time.Now().UTC().Format(time.RFC1123))
	req.Header.Set("x-ms-version", "2021-08-06")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	sig := e.signAzure(req, data)
	req.Header.Set("Authorization", sig)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("azure upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("azure upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (e *AzureBlobExporter) List(path string) ([]string, error) {
	url := fmt.Sprintf("https://%s.blob.core.windows.net/%s?restype=container&comp=list&prefix=%s",
		e.cfg.AccountName, e.cfg.Bucket, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create list request: %w", err)
	}

	req.Header.Set("x-ms-date", time.Now().UTC().Format(time.RFC1123))
	req.Header.Set("x-ms-version", "2021-08-06")

	sig := e.signAzure(req, nil)
	req.Header.Set("Authorization", sig)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("azure list: %w", err)
	}
	defer resp.Body.Close()

	var blobList struct {
		XMLName xml.Name `xml:"EnumerationResults"`
		Blobs   struct {
			Blob []struct {
				Name string `xml:"Name"`
			} `xml:"Blob"`
		} `xml:"Blobs"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&blobList); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}

	var names []string
	for _, b := range blobList.Blobs.Blob {
		names = append(names, b.Name)
	}
	return names, nil
}

func (e *AzureBlobExporter) signAzure(req *http.Request, data []byte) string {
	now := time.Now().UTC().Format(time.RFC1123)

	var contentLength int
	if data != nil {
		contentLength = len(data)
	}

	canonicalizedHeaders := fmt.Sprintf("x-ms-blob-type:%s\nx-ms-date:%s\nx-ms-version:%s\n",
		req.Header.Get("x-ms-blob-type"), now, "2021-08-06")

	canonicalizedResource := fmt.Sprintf("/%s/%s\n", e.cfg.AccountName, e.cfg.Bucket)
	if path := strings.TrimPrefix(req.URL.Path, "/"+e.cfg.AccountName+"/"+e.cfg.Bucket+"/"); path != "" {
		canonicalizedResource += fmt.Sprintf("comp:%s\n", req.URL.Query().Get("comp"))
	}

	stringToSign := fmt.Sprintf("%s\n\n\n%d\n\n%s\n%s\n%s",
		req.Method,
		contentLength,
		"application/json",
		canonicalizedHeaders,
		canonicalizedResource,
	)

	decodedKey, _ := base64.StdEncoding.DecodeString(e.cfg.AccountKey)
	mac := hmac.New(sha256.New, decodedKey)
	mac.Write([]byte(stringToSign))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("SharedKey %s:%s", e.cfg.AccountName, sig)
}

func UploadToCloud(exporter CloudExporter, prefix string, events []AuditEvent) (string, error) {
	data, err := json.Marshal(events)
	if err != nil {
		return "", fmt.Errorf("marshal events: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05")
	path := fmt.Sprintf("%saudit-export-%s.json", prefix, timestamp)

	if err := exporter.Upload(path, data); err != nil {
		return "", err
	}

	return path, nil
}

func ExportToCloud(cfg CloudConfig, events []AuditEvent) (string, error) {
	exporter := NewCloudExporter(cfg)
	if exporter == nil {
		return "", fmt.Errorf("unsupported cloud provider: %s", cfg.Provider)
	}
	return UploadToCloud(exporter, cfg.Prefix, events)
}


