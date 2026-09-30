// Copyright 2025-2026 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package kernel

type Lifecycle interface {
	Initialize() error
	Start() error
	Stop() error
}
