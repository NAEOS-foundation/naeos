// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package api

import "testing"

func TestHandler(t *testing.T) {
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("handler should not be nil")
	}
}
