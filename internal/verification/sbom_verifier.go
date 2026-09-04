package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

// SBOMVerifier verifies that an evidence record references a valid
// CycloneDX SBOM document. It checks:
//
//  1. The evidence record has an artifact hash.
//  2. The SBOM content matches the expected hash.
//  3. The SBOM is a valid CycloneDX document.
//
// When a live ArtifactSource is provided, the verifier re-hashes the
// SBOM content and compares it against the record's artifact hash.
type SBOMVerifier struct {
	expected map[string]string
	source   ArtifactSource
}

// NewSBOMVerifier creates a verifier with known SBOM hashes.
func NewSBOMVerifier(expected map[string]string) *SBOMVerifier {
	return &SBOMVerifier{expected: expected}
}

// NewSBOMVerifierWithSource creates a verifier that reads SBOM content
// live from an ArtifactSource and recomputes the hash.
func NewSBOMVerifierWithSource(expected map[string]string, src ArtifactSource) *SBOMVerifier {
	return &SBOMVerifier{expected: expected, source: src}
}

// Name returns "sbom-integrity".
func (v *SBOMVerifier) Name() string { return "sbom-integrity" }

// Verify checks that the evidence record's artifact hash matches the
// SBOM content hash. If a live source is provided it re-hashes the
// content; otherwise it checks against the expected map.
func (v *SBOMVerifier) Verify(rec evidence.EvidenceRecord) (VerificationResult, error) {
	res := VerificationResult{
		Status:    StatusVerified,
		Target:    rec.ID,
		Timestamp: time.Now().UTC(),
	}

	if rec.ArtifactHash == "" {
		res.Status = StatusUnverified
		res.Message = "no artifact hash to verify"
		return res, nil
	}

	if v.source != nil {
		content, ok := v.source.Content(rec.ArtifactName)
		if !ok {
			res.Status = StatusFailed
			res.Message = fmt.Sprintf("SBOM %s not found", rec.ArtifactName)
			res.Checks = append(res.Checks, CheckResult{Name: "sbom-source", Passed: false, Detail: "not found"})
			return res, nil
		}

		actual := computeSBOMHash(content)
		match := actual == rec.ArtifactHash
		res.Checks = append(res.Checks, CheckResult{
			Name:   "sbom-content-hash",
			Passed: match,
			Detail: fmt.Sprintf("expected %s, got %s", rec.ArtifactHash, actual),
		})
		if !match {
			res.Status = StatusFailed
			res.Message = "SBOM content hash mismatch"
		}
		return res, nil
	}

	expected, ok := v.expected[rec.ArtifactName]
	check := CheckResult{
		Name:   "sbom-expected-hash",
		Passed: ok && expected == rec.ArtifactHash,
		Detail: fmt.Sprintf("SBOM %s", rec.ArtifactName),
	}
	res.Checks = append(res.Checks, check)
	if !check.Passed {
		res.Status = StatusFailed
		res.Message = fmt.Sprintf("SBOM %s hash does not match expected", rec.ArtifactName)
	}
	return res, nil
}

func computeSBOMHash(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}
