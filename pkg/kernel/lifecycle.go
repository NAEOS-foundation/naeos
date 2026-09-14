// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package kernel

type Lifecycle interface {
	Initialize() error
	Start() error
	Stop() error
}
