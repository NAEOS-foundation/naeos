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
	if host == directHost && (u.Port() == "" || u.Port() == "5432") {
		return nil
	}

	// Supavisor shared-pooler URLs encode the project ref after the database
	// role name: postgres.<project-ref> or <custom-role>.<project-ref>.
	username := ""
	if u.User != nil {
		username = u.User.Username()
	}
	projectSuffix := "." + expectedProjectRef
	if strings.HasSuffix(host, ".pooler.supabase.com") && (u.Port() == "" || u.Port() == "5432") && strings.HasSuffix(username, projectSuffix) {
		roleName := strings.TrimSuffix(username, projectSuffix)
		if validControlPlanePoolerRoleName(roleName) {
			return nil
		}
	}

	return errors.New("NAEOS_CONTROLPLANE_DATABASE_URL does not match the configured Supabase project for this environment")
}

func validControlPlanePoolerRoleName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || r == '_') {
			if (r < 'a' || r > 'z') && r != '_' {
				return false
			}
			continue
		}
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {

		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}
