// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package config

type Config struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}
