// Copyright 2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package dependencyrisk

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const RegistrySchemaVersion = "1.0.0"

// ReviewRegistry must be loaded from a trusted base revision by its caller.
// A valid record is not trustworthy merely because its fields are valid.
type ReviewRegistry struct {
	SchemaVersion string         `json:"schema_version"`
	PolicyVersion string         `json:"policy_version"`
	Records       []ReviewRecord `json:"records"`
}

// ReviewRecord binds one exact dependency version to reviewable provenance.
type ReviewRecord struct {
	Ecosystem          string      `json:"ecosystem"`
	Name               string      `json:"name"`
	Version            string      `json:"version"`
	UpstreamURL        string      `json:"upstream_url"`
	SPDXLicense        string      `json:"spdx_license"`
	LicenseEvidenceURL string      `json:"license_evidence_url"`
	Checksum           string      `json:"checksum"`
	ProvenanceURL      string      `json:"provenance_url"`
	Purpose            string      `json:"purpose"`
	Domains            []Domain    `json:"domains"`
	Criticality        Criticality `json:"criticality"`
	Reviewer           string      `json:"reviewer"`
	ReviewReference    string      `json:"review_reference"`
	ReviewedAt         string      `json:"reviewed_at"`
	ReviewExpiresAt    string      `json:"review_expires_at"`
}

var (
	sha256Checksum = regexp.MustCompile(`^sha256:[0-9a-fA-F]{64}$`)
	goChecksum     = regexp.MustCompile(`^h1:[A-Za-z0-9+/]{43}=$`)
)

// ValidateReviewRegistry rejects incomplete, duplicate, malformed, or
// version-mismatched schema records. Trust in the source revision is a separate
// responsibility and must be enforced by the caller.
func ValidateReviewRegistry(reg ReviewRegistry, policyVersion string, now time.Time) error {
	if reg.SchemaVersion != RegistrySchemaVersion {
		return fmt.Errorf("unsupported dependency review registry schema %q", reg.SchemaVersion)
	}
	if policyVersion == "" || reg.PolicyVersion != policyVersion {
		return errors.New("dependency review registry policy version mismatch")
	}
	seen := make(map[string]struct{}, len(reg.Records))
	for i, rec := range reg.Records {
		if err := validateReviewRecord(rec, now); err != nil {
			return fmt.Errorf("registry record %d: %w", i, err)
		}
		key := strings.ToLower(strings.TrimSpace(rec.Ecosystem)) + "\x00" + strings.TrimSpace(rec.Name) + "\x00" + strings.TrimSpace(rec.Version)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("registry contains duplicate record for %s %s@%s", rec.Ecosystem, rec.Name, rec.Version)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateReviewRecord(rec ReviewRecord, now time.Time) error {
	required := map[string]string{
		"ecosystem":        rec.Ecosystem,
		"name":             rec.Name,
		"version":          rec.Version,
		"SPDX license":     rec.SPDXLicense,
		"purpose":          rec.Purpose,
		"reviewer":         rec.Reviewer,
		"review reference": rec.ReviewReference,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("missing %s", field)
		}
	}
	if !validHTTPSURL(rec.UpstreamURL) {
		return errors.New("upstream_url must be an absolute HTTPS URL")
	}
	if !validHTTPSURL(rec.LicenseEvidenceURL) {
		return errors.New("license_evidence_url must be an absolute HTTPS URL")
	}
	if !validHTTPSURL(rec.ProvenanceURL) {
		return errors.New("provenance_url must be an absolute HTTPS URL")
	}
	if !validChecksum(rec.Checksum) {
		return errors.New("checksum must be sha256:<64 hex chars> or a Go h1:<base64> checksum")
	}
	switch rec.Criticality {
	case Low, Medium, High, Critical:
	default:
		return fmt.Errorf("invalid criticality %q", rec.Criticality)
	}
	if len(rec.Domains) == 0 {
		return errors.New("domains must contain at least one recognized domain")
	}
	seenDomains := map[Domain]struct{}{}
	for _, domain := range rec.Domains {
		switch domain {
		case Policy, Security, AuditEvidence, Runtime, DeploymentCI, General:
		default:
			return fmt.Errorf("unrecognized domain %q", domain)
		}
		if _, exists := seenDomains[domain]; exists {
			return fmt.Errorf("duplicate domain %q", domain)
		}
		seenDomains[domain] = struct{}{}
	}
	reviewedAt, err := time.Parse(time.RFC3339, rec.ReviewedAt)
	if err != nil {
		return errors.New("reviewed_at must be an RFC3339 timestamp")
	}
	if reviewedAt.After(now) {
		return errors.New("reviewed_at cannot be in the future")
	}
	expiresAt, err := time.Parse(time.RFC3339, rec.ReviewExpiresAt)
	if err != nil {
		return errors.New("review_expires_at must be an RFC3339 timestamp")
	}
	if !expiresAt.After(reviewedAt) {
		return errors.New("review_expires_at must be after reviewed_at")
	}
	if !now.Before(expiresAt) {
		return errors.New("dependency review record is stale")
	}
	return nil
}

// FindExactReview returns only an exact ecosystem/name/version match.
// The caller must validate the registry and establish its trusted provenance.
func (reg ReviewRegistry) FindExactReview(ecosystem, name, version string) (ReviewRecord, bool) {
	for _, rec := range reg.Records {
		if strings.EqualFold(strings.TrimSpace(rec.Ecosystem), strings.TrimSpace(ecosystem)) &&
			strings.TrimSpace(rec.Name) == strings.TrimSpace(name) &&
			strings.TrimSpace(rec.Version) == strings.TrimSpace(version) {
			return rec, true
		}
	}
	return ReviewRecord{}, false
}

// SortedRecordKeys returns stable record identifiers for evidence and diagnostics.
func (reg ReviewRegistry) SortedRecordKeys() []string {
	keys := make([]string, 0, len(reg.Records))
	for _, rec := range reg.Records {
		keys = append(keys, strings.ToLower(strings.TrimSpace(rec.Ecosystem))+":"+strings.TrimSpace(rec.Name)+"@"+strings.TrimSpace(rec.Version))
	}
	sort.Strings(keys)
	return keys
}

func validHTTPSURL(value string) bool {
	u, err := url.Parse(strings.TrimSpace(value))
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil
}

func validChecksum(value string) bool {
	if sha256Checksum.MatchString(value) {
		_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
		return err == nil
	}
	if goChecksum.MatchString(value) {
		_, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, "h1:"))
		return err == nil
	}
	return false
}
