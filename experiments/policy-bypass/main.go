// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Command policy-bypass runs the NAEOS governance bypass experiment and
// writes the findings report. Run with: go run ./experiments/policy-bypass
package main

import (
	"fmt"
	"os"
)

func main() {
	results := runAll()
	printResults(results)
	if err := writeReport(results); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
		os.Exit(1)
	}
	fmt.Printf("\nreport written to %s\n", reportPath)
}
