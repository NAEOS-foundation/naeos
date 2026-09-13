// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package config

type Config struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}
