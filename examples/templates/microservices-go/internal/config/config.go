// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package config

import "os"

type Config struct {
	Port string
}

func Load() Config {
	return Config{Port: getenv("PORT", "8080")}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
