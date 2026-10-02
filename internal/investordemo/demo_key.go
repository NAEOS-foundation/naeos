// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package investordemo

// demoSigningKey returns deterministic, non-secret material used only by the
// local investor demo. Production deployments must inject signing material
// through NewHandoffValidatorWithSigningKey.
func demoSigningKey() []byte {
	return []byte("NAEOS_LOCAL_DEMO_ONLY_HMAC_KEY_V1")
}
