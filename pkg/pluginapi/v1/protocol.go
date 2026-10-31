// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package pluginapi defines public, versioned plugin wire types. The types do
// not imply that a host/runtime implements or supports this protocol yet.
package pluginapi

import (
	"encoding/json"
	"errors"
	"strings"
)

const ProtocolVersion = "1.0"

// Error codes defined by protocol v1. Plugins should use one of these codes
// for failed responses; hosts may reject or remap unknown codes.
const (
	ErrorCodeInvalidRequest     = "invalid_request"
	ErrorCodeUnsupportedVersion = "unsupported_version"
	ErrorCodeInvalidParams      = "invalid_params"
	ErrorCodeDenied             = "denied"
	ErrorCodeTimeout            = "timeout"
	ErrorCodeResourceLimit      = "resource_limit"
	ErrorCodePluginError        = "plugin_error"
	ErrorCodeInvalidResponse    = "invalid_response"
)

var knownErrorCodes = map[string]struct{}{
	ErrorCodeInvalidRequest:     {},
	ErrorCodeUnsupportedVersion: {},
	ErrorCodeInvalidParams:      {},
	ErrorCodeDenied:             {},
	ErrorCodeTimeout:            {},
	ErrorCodeResourceLimit:      {},
	ErrorCodePluginError:        {},
	ErrorCodeInvalidResponse:    {},
}

// ValidErrorCode reports whether the code is one of the documented v1 error
// codes. The list is finite; plugins must not invent transient codes.
func ValidErrorCode(code string) bool {
	_, ok := knownErrorCodes[code]
	return ok
}

// Manifest declares plugin metadata and the exact protocol versions supported.
type Manifest struct {
	Name             string           `json:"name"`
	Version          string           `json:"version"`
	Description      string           `json:"description"`
	Author           string           `json:"author,omitempty"`
	License          string           `json:"license,omitempty"`
	ProtocolVersions []string         `json:"protocol_versions"`
	Actions          []ActionManifest `json:"actions,omitempty"`
}

type ActionManifest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Params       map[string]string `json:"params,omitempty"`
	Returns      string            `json:"returns,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
}

// Request is one host-to-plugin invocation. Method is currently "execute";
// Action identifies the plugin operation.
type Request struct {
	ProtocolVersion string                     `json:"protocol_version"`
	RequestID       string                     `json:"request_id"`
	Method          string                     `json:"method"`
	Action          string                     `json:"action"`
	Params          map[string]json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	ProtocolVersion string          `json:"protocol_version"`
	RequestID       string          `json:"request_id"`
	OK              bool            `json:"ok"`
	Result          json.RawMessage `json:"result,omitempty"`
	Error           *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SupportsProtocol reports whether the manifest declares an exact protocol
// version supported by the host. Exact versions avoid ambiguous range syntax.
func (m Manifest) SupportsProtocol(version string) bool {
	for _, supported := range m.ProtocolVersions {
		if strings.TrimSpace(supported) == version {
			return true
		}
	}
	return false
}

// Validate checks required envelope fields before a host dispatches a request.
func (r Request) Validate() error {
	if r.ProtocolVersion != ProtocolVersion {
		return errors.New("unsupported plugin protocol version")
	}
	if strings.TrimSpace(r.RequestID) == "" {
		return errors.New("plugin request ID is required")
	}
	if r.Method != "execute" {
		return errors.New("plugin request method must be execute")
	}
	if strings.TrimSpace(r.Action) == "" {
		return errors.New("plugin action is required")
	}
	return nil
}

// Validate checks response correlation and success/error consistency.
func (r Response) Validate(request Request) error {
	if r.ProtocolVersion != request.ProtocolVersion {
		return errors.New("plugin response protocol version does not match request")
	}
	if r.RequestID != request.RequestID {
		return errors.New("plugin response request ID does not match request")
	}
	if r.OK && r.Error != nil {
		return errors.New("successful plugin response must not include an error")
	}
	if !r.OK && (r.Error == nil || strings.TrimSpace(r.Error.Code) == "") {
		return errors.New("failed plugin response must include an error code")
	}
	if !r.OK && r.Error != nil && !ValidErrorCode(r.Error.Code) {
		return errors.New("failed plugin response uses an undocumented error code")
	}
	return nil
}
