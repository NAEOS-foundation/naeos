// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package supabase

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const sessionFile = "session.json"

func sessionFilePath() string {
	return filepath.Join(configDir(), sessionFile)
}

// SaveSession stores user tokens separately from Supabase project credentials.
func SaveSession(session *Session) error {
	if session == nil || session.AccessToken == "" || session.RefreshToken == "" {
		return errors.New("valid Supabase session is required")
	}
	dir := configDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	//nolint:gosec // user auth tokens are intentionally persisted with mode 0600.
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".session-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, sessionFilePath()); err != nil {
		return err
	}
	return nil
}

// LoadSession loads the locally persisted user session. A missing session is
// represented by (nil, nil).
func LoadSession() (*Session, error) {
	data, err := os.ReadFile(sessionFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		return nil, errors.New("stored Supabase session is incomplete")
	}
	return &session, nil
}

// ClearSession removes the locally persisted user session.
func ClearSession() error {
	err := os.Remove(sessionFilePath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
