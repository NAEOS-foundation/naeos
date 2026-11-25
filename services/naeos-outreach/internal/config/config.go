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

// Load validates an address with the secure loopback-only default.
func Load(addr string) (Config, error) {
	return LoadWithOptions(addr, false)
}

// LoadWithOptions permits non-loopback binding only when the caller explicitly
// opts in. Applications should keep this disabled unless container/network
// exposure is intentional and protected by an appropriate deployment boundary.
func LoadWithOptions(addr string, allowNonLoopback bool) (Config, error) {
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
	if !isLoopback(host) && !allowNonLoopback {
		return Config{}, fmt.Errorf("NAEOS_OUTREACH_ADDR host %q must be a loopback address unless NAEOS_OUTREACH_ALLOW_NON_LOOPBACK is explicitly enabled", host)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("NAEOS_OUTREACH_ADDR port must be a number between 1 and 65535")
	}
	return Config{Addr: net.JoinHostPort(host, strconv.Itoa(port))}, nil
}

// isLoopback reports whether host is a loopback address (IPv4 127/8, IPv6 ::1,
// or the hostname "localhost").
func isLoopback(host string) bool {
	if host == "localhost" || host == "localhost.localdomain" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}
