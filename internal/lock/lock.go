package lock

import (
	"crypto/md5" //nolint:gosec // G501: MD5 is a user-selectable algorithm, not the default
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type HashAlgorithm string

const (
	HashSHA256 HashAlgorithm = "sha256"
	HashSHA512 HashAlgorithm = "sha512"
	HashMD5    HashAlgorithm = "md5"
)

type LockFile struct {
	Version   string         `json:"version"`
	Generated string         `json:"generated"`
	Artifacts []LockArtifact `json:"artifacts"`
	Checksum  string         `json:"checksum"`
	Algorithm HashAlgorithm  `json:"algorithm,omitempty"`
}

type LockArtifact struct {
	Path     string `json:"path"`
	Size     int    `json:"size"`
	Checksum string `json:"checksum"`
}

type ArtifactInfo struct {
	Path    string
	Content []byte
}

type VerifyResult struct {
	Changes  []string
	Errors   []error
	Duration time.Duration
}

type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Path      string    `json:"path"`
	Details   string    `json:"details,omitempty"`
}

type AuditLog struct {
	Entries []AuditEntry `json:"entries"`
	mu      sync.Mutex
}

func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

func (a *AuditLog) Add(action, path, details string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Entries = append(a.Entries, AuditEntry{
		Timestamp: time.Now(),
		Action:    action,
		Path:      path,
		Details:   details,
	})
}

func (a *AuditLog) GetEntries() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]AuditEntry, len(a.Entries))
	copy(result, a.Entries)
	return result
}

func (a *AuditLog) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Entries = nil
}

func hashContent(content []byte, algo HashAlgorithm) string {
	switch algo {
	case HashSHA512:
		h := sha512.Sum512(content)
		return hex.EncodeToString(h[:])
	case HashMD5:
		h := md5.Sum(content) //nolint:gosec // G401/G501: MD5 is a user-selectable algorithm, not the default
		return hex.EncodeToString(h[:])
	default:
		h := sha256.Sum256(content)
		return hex.EncodeToString(h[:])
	}
}

func Generate(artifacts []ArtifactInfo) (*LockFile, error) {
	return GenerateWithAlgorithm(artifacts, HashSHA256)
}

func GenerateWithAlgorithm(artifacts []ArtifactInfo, algo HashAlgorithm) (*LockFile, error) {
	lock := &LockFile{
		Version:   "1",
		Generated: time.Now().UTC().Format(time.RFC3339),
		Algorithm: algo,
	}

	for _, a := range artifacts {
		lock.Artifacts = append(lock.Artifacts, LockArtifact{
			Path:     a.Path,
			Size:     len(a.Content),
			Checksum: hashContent(a.Content, algo),
		})
	}

	sort.Slice(lock.Artifacts, func(i, j int) bool {
		return lock.Artifacts[i].Path < lock.Artifacts[j].Path
	})

	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshal lock file")
	}

	lock.Checksum = hashContent(data, algo)

	return lock, nil
}

func WriteToFile(lock *LockFile, path string) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshal lock file")
	}
	return os.WriteFile(path, data, 0o600)
}

func ReadFromFile(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read lock file")
	}
	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse lock file")
	}
	return &lock, nil
}

func Verify(lock *LockFile, current []ArtifactInfo) ([]string, error) {
	result := VerifyConcurrent(lock, current, 4)
	return result.Changes, mergeErrors(result.Errors)
}

func VerifyConcurrent(lock *LockFile, current []ArtifactInfo, workers int) *VerifyResult {
	start := time.Now()
	result := &VerifyResult{}

	if lock == nil {
		result.Errors = append(result.Errors, naeoserr.New(naeoserr.ErrInternal, "lock file is nil"))
		result.Duration = time.Since(start)
		return result
	}

	algo := lock.Algorithm
	if algo == "" {
		algo = HashSHA256
	}

	type hashResult struct {
		path string
		hash string
	}

	results := make([]hashResult, len(current))

	var wg sync.WaitGroup
	for i, a := range current {
		wg.Add(1)
		go func(idx int, art ArtifactInfo) {
			defer wg.Done()
			hash := hashContent(art.Content, algo)
			results[idx] = hashResult{path: art.Path, hash: hash}
		}(i, a)
	}
	wg.Wait()

	existing := make(map[string]LockArtifact)
	for _, a := range lock.Artifacts {
		existing[a.Path] = a
	}

	currentMap := make(map[string]bool)
	for _, r := range results {
		currentMap[r.path] = true
		if old, ok := existing[r.path]; ok {
			if old.Checksum != r.hash {
				result.Changes = append(result.Changes, fmt.Sprintf("modified: %s", r.path))
			}
		} else {
			result.Changes = append(result.Changes, fmt.Sprintf("added: %s", r.path))
		}
	}

	for _, a := range lock.Artifacts {
		if !currentMap[a.Path] {
			result.Changes = append(result.Changes, fmt.Sprintf("removed: %s", a.Path))
		}
	}

	result.Duration = time.Since(start)
	return result
}

func VerifyDryRun(lock *LockFile, current []ArtifactInfo) *VerifyResult {
	return VerifyConcurrent(lock, current, 1)
}

func mergeErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var msg string
	for i, e := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += e.Error()
	}
	return naeoserr.New(naeoserr.ErrInternal, msg)
}
