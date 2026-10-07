// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

const (
	v2IdempotencyCompleted = "completed"
	v2IdempotencyPending   = "pending"
	v2IdempotencyTTL       = 24 * time.Hour
)

func (s *Server) ensureV2IdempotencyTable() error {
	if s.db == nil {
		return nil
	}
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS naeos_v2_idempotency (
		idempotency_key VARCHAR(255) PRIMARY KEY,
		fingerprint VARCHAR(64) NOT NULL,
		status VARCHAR(16) NOT NULL,
		headers TEXT NOT NULL,
		body TEXT NOT NULL,
		response_status INTEGER NOT NULL,
		expires_at BIGINT NOT NULL
	)`)
	return err
}

func encodeIdempotencyBody(body []byte) string {
	return base64.RawStdEncoding.EncodeToString(body)
}

func decodeIdempotencyBody(body string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(body)
}

func headerJSON(header http.Header) string {
	raw, _ := json.Marshal(header)
	return string(raw)
}

func parseHeaderJSON(raw string) http.Header {
	var header http.Header
	if err := json.Unmarshal([]byte(raw), &header); err != nil || header == nil {
		return make(http.Header)
	}
	return header
}

func (s *Server) loadDurableIdempotency(key, _ string) (idempotencyEntry, bool, error) {
	if err := s.ensureV2IdempotencyTable(); err != nil {
		return idempotencyEntry{}, false, err
	}
	now := time.Now().Unix()
	_, _ = s.db.Exec("DELETE FROM naeos_v2_idempotency WHERE expires_at <= ?", now)

	rows, err := s.db.Query(
		"SELECT fingerprint, status, headers, body, response_status, expires_at FROM naeos_v2_idempotency WHERE idempotency_key = ?",
		key,
	)
	if err != nil {
		return idempotencyEntry{}, false, err
	}
	if len(rows) == 0 {
		return idempotencyEntry{}, false, nil
	}
	row := rows[0]
	storedFingerprint, _ := row["fingerprint"].(string)
	status, _ := row["status"].(string)
	headers, _ := row["headers"].(string)
	body, _ := row["body"].(string)
	responseStatus := 0
	switch value := row["response_status"].(type) {
	case int64:
		responseStatus = int(value)
	case int:
		responseStatus = value
	}
	expires := int64(0)
	switch value := row["expires_at"].(type) {
	case int64:
		expires = value
	case int:
		expires = int64(value)
	}
	entry := idempotencyEntry{
		fingerprint: storedFingerprint,
		status:      responseStatus,
		header:      parseHeaderJSON(headers),
		expiresAt:   time.Unix(expires, 0),
	}
	if decoded, err := decodeIdempotencyBody(body); err == nil {
		entry.body = decoded
	}
	if status == v2IdempotencyPending {
		entry.status = http.StatusConflict
	}
	return entry, true, nil
}

func (s *Server) claimDurableIdempotency(key, fingerprint string) (bool, error) {
	if err := s.ensureV2IdempotencyTable(); err != nil {
		return false, err
	}
	_, err := s.db.Exec(
		"INSERT INTO naeos_v2_idempotency (idempotency_key, fingerprint, status, headers, body, response_status, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		key, fingerprint, v2IdempotencyPending, "{}", "", 0, time.Now().Add(v2IdempotencyTTL).Unix(),
	)
	if err == nil {
		return true, nil
	}
	return false, nil
}

func (s *Server) completeDurableIdempotency(key, fingerprint string, status int, header http.Header, body []byte) error {
	_, err := s.db.Exec(
		"UPDATE naeos_v2_idempotency SET fingerprint = ?, status = ?, headers = ?, body = ?, response_status = ?, expires_at = ? WHERE idempotency_key = ?",
		fingerprint, v2IdempotencyCompleted, headerJSON(header), encodeIdempotencyBody(body), status, time.Now().Add(v2IdempotencyTTL).Unix(), key,
	)
	return err
}

func (s *Server) deleteDurableIdempotency(key string) {
	if s.db != nil {
		_, _ = s.db.Exec("DELETE FROM naeos_v2_idempotency WHERE idempotency_key = ?", key)
	}
}
