// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package auth

type Repository interface {
	List() []string
}
