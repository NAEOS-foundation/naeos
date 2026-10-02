// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package helm

import "gopkg.in/yaml.v3"

func unmarshalYAML(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
