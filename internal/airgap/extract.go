package airgap

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// ExtractOptions configures bundle extraction.
type ExtractOptions struct {
	// Destination is the directory where bundle contents are extracted.
	Destination string
	// VerifyHashes verifies each file's SHA-256 hash against the manifest.
	VerifyHashes bool
}

// Extract unpacks an airgap bundle into the destination directory.
func Extract(path string, opts ExtractOptions) (*Bundle, error) {
	bundle, err := ReadBundle(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "open %s", path)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "gzip open")
	}
	defer gz.Close()

	if opts.Destination == "" {
		opts.Destination = "."
	}
	if err := os.MkdirAll(opts.Destination, 0o755); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "create destination")
	}

	expectedHashes := make(map[string]string)
	for _, g := range [][]FileRef{bundle.Charts, bundle.SBOMManifests, bundle.Signatures} {
		for _, ref := range g {
			expectedHashes[strings.TrimPrefix(prefixTarPath(ref.Path), "bundled/")] = ref.Hash
		}
	}

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read tar")
		}
		if hdr.Name == "manifest.json" {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		rel := strings.TrimPrefix(hdr.Name, "bundled/")
		if rel == hdr.Name {
			rel = filepath.Base(hdr.Name)
		}
		if filepath.IsAbs(rel) || strings.Contains(rel, "..") {
			return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("unsafe path in bundle: %s", hdr.Name))
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read %s", hdr.Name)
		}

		if opts.VerifyHashes {
			expected, ok := expectedHashes[rel]
			if !ok {
				return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("no expected hash for %s", rel))
			}
			h := sha256.Sum256(content)
			actual := hex.EncodeToString(h[:])
			if actual != expected {
				return nil, naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("hash mismatch for %s", rel))
			}
		}

		dst := filepath.Join(opts.Destination, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "create dir")
		}
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "write %s", rel)
		}
	}

	return bundle, nil
}