// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package helm

import "gopkg.in/yaml.v3"

func unmarshalYAML(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
