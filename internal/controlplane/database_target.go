// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"errors"
	"net/url"
	"strings"
)

const (
	controlPlaneProductionProjectRef = "oruykkehukoiipldzcvc"
	controlPlaneStagingProjectRef    = "kqmhpsftijskesjdovjm"
)

// validateControlPlaneDatabaseTarget binds each runtime environment to its
// explicitly approved Supabase project. It accepts either the direct database
// hostname or a Supavisor session-pooler hostname whose username carries the
// project ref. It never logs or returns the supplied URL.
func validateControlPlaneDatabaseTarget(raw, environment string) error {
	if err := validateProductionDatabaseURL(raw); err != nil {
		return err
	}

	var expectedProjectRef string
	switch strings.TrimSpace(environment) {
	case "production":
		expectedProjectRef = controlPlaneProductionProjectRef
	case "staging", "preview":
		expectedProjectRef = controlPlaneStagingProjectRef
	default:
		return errors.New("NAEOS_CONTROLPLANE_ENVIRONMENT must be production, staging, or preview")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("NAEOS_CONTROLPLANE_DATABASE_URL must be a valid PostgreSQL URL")
	}
	host := strings.ToLower(u.Hostname())
	directHost := "db." + expectedProjectRef + ".supabase.co"
	if host == directHost {
		return nil
	}

	// Supabase shared session pooler URLs use a regional pooler hostname and
	// encode the project ref in the database username (postgres.<project-ref>).
	username := ""
	if u.User != nil {
		username, _ = u.User.Username()
	}
	if strings.HasSuffix(host, ".pooler.supabase.com") &&
		strings.HasSuffix(username, "."+expectedProjectRef) {
		return nil
	}

	return errors.New("NAEOS_CONTROLPLANE_DATABASE_URL does not match the configured Supabase project for this environment")
}
