// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const DefaultAddr = "127.0.0.1:8088"

type Config struct {
	Addr string
}

func Load(addr string) (Config, error) {
	if addr == "" {
		addr = DefaultAddr
	}
	if strings.TrimSpace(addr) != addr {
		return Config{}, fmt.Errorf("NAEOS_OUTREACH_ADDR must not contain surrounding whitespace")
	}
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid NAEOS_OUTREACH_ADDR %q: %w", addr, err)
	}
	if host == "" {
		return Config{}, fmt.Errorf("NAEOS_OUTREACH_ADDR host must not be empty")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("NAEOS_OUTREACH_ADDR port must be a number between 1 and 65535")
	}
	return Config{Addr: net.JoinHostPort(host, strconv.Itoa(port))}, nil
}
