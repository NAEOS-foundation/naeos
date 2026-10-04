// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pluginhost

import (
	"fmt"
	"sort"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// CapabilityBoundary is the authorization boundary for plugin actions.
// Declaration is not authority: privileged actions require an explicit host grant.
type CapabilityBoundary struct {
	grants map[string]map[string]struct{}
}

func NewCapabilityBoundary(grants map[string][]string) (*CapabilityBoundary, error) {
	b := &CapabilityBoundary{grants: make(map[string]map[string]struct{})}
	for plugin, capabilities := range grants {
		name := strings.TrimSpace(plugin)
		if name == "" {
			return nil, naeoserr.New(naeoserr.ErrValidation, "plugin capability grant has empty plugin name")
		}
		set := make(map[string]struct{})
		for _, capability := range capabilities {
			capability, err := normalizeCapability(capability)
			if err != nil {
				return nil, err
			}
			set[capability] = struct{}{}
		}
		b.grants[name] = set
	}
	return b, nil
}

// Authorize uses exact capability matching and fails closed for privileged actions.
func (b *CapabilityBoundary) Authorize(pluginName, action string, required []string) error {
	if len(required) == 0 {
		return nil
	}
	if b == nil {
		return naeoserr.New(naeoserr.ErrPermDenied, "plugin capability boundary is not configured")
	}
	pluginName = strings.TrimSpace(pluginName)
	action = strings.TrimSpace(action)
	if pluginName == "" || action == "" {
		return naeoserr.New(naeoserr.ErrValidation, "plugin name and action are required")
	}
	granted := b.grants[pluginName]
	var missing []string
	for _, capability := range required {
		capability, err := normalizeCapability(capability)
		if err != nil {
			return err
		}
		if _, ok := granted[capability]; !ok {
			missing = append(missing, capability)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return naeoserr.New(naeoserr.ErrPermDenied,
			fmt.Sprintf("plugin %q action %q requires ungranted capabilities: %s",
				pluginName, action, strings.Join(missing, ", ")))
	}
	return nil
}

func (b *CapabilityBoundary) Grants() map[string][]string {
	out := make(map[string][]string, len(b.grants))
	for plugin, set := range b.grants {
		for capability := range set {
			out[plugin] = append(out[plugin], capability)
		}
		sort.Strings(out[plugin])
	}
	return out
}

func normalizeCapability(capability string) (string, error) {
	capability = strings.ToLower(strings.TrimSpace(capability))
	if capability == "" {
		return "", naeoserr.New(naeoserr.ErrValidation, "plugin capability must not be empty")
	}
	if strings.ContainsAny(capability, "*?") {
		return "", naeoserr.New(naeoserr.ErrValidation,
			fmt.Sprintf("plugin capability %q must be exact; wildcards are not allowed", capability))
	}
	return capability, nil
}
