// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package investordemo

// CalculatePayloadDigestForExperiment exposes the production/demo canonical
// payload digest to deterministic experiments without duplicating its algorithm.
// Experiments must use the same digest implementation as the validator.
func CalculatePayloadDigestForExperiment(payload map[string]interface{}) string {
	return calculatePayloadDigest(payload)
}
