// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package clihelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

// PoliciesDir returns the directory where registered policies are persisted.
func PoliciesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrInternal, "cannot determine home directory")
	}
	return filepath.Join(home, ".config", "naeos", "policies"), nil
}

// PolicyPath returns the path for a persisted policy file.
func PolicyPath(id string) (string, error) {
	dir, err := PoliciesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".json"), nil
}

// LoadPolicies loads all registered policies from the policy directory.
func LoadPolicies() (*policy.Registry, error) {
	reg := policy.NewRegistry()
	dir, err := PoliciesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading policies directory")
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading policy file %s", e.Name())
		}
		var versions []policy.Policy
		if err := json.Unmarshal(data, &versions); err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrValidation, "parsing policy file %s", e.Name())
		}
		var activeVersion string
		for i := range versions {
			v := &versions[i]
			if v.Active {
				activeVersion = v.Version
			}
			if err := reg.Register(v); err != nil {
				return nil, err
			}
		}
		if activeVersion != "" && len(versions) > 0 {
			if err := reg.SetActive(versions[0].ID, activeVersion); err != nil {
				return nil, err
			}
		}
	}
	return reg, nil
}
