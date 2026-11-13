// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

const naeosModule = "github.com/NAEOS-foundation/naeos"

func TestParseFindingsKeepsOnlyReachableVulnerabilities(t *testing.T) {
	input := `{"config":{}}` +
		`{"finding":{"osv":"GO-2026-6865","trace":[{"module":"github.com/klauspost/compress","version":"v1.18.5","package":"github.com/klauspost/compress/fse"}]}}` +
		`{"finding":{"osv":"GO-2026-6865","trace":[{"module":"github.com/klauspost/compress","version":"v1.18.5","package":"github.com/klauspost/compress/fse","function":"Decompress","position":{"line":10}},{"module":"` + naeosModule + `","package":"` + naeosModule + `/internal/telemetry","function":"ExportSpans","position":{"line":20}}]}}` +
		`{"finding":{"osv":"GO-2026-6865","trace":[{"module":"github.com/klauspost/compress","version":"v1.18.5","package":"github.com/klauspost/compress/fse","function":"Decompress","position":{"line":110}},{"module":"` + naeosModule + `","package":"` + naeosModule + `/internal/telemetry","function":"ExportSpans","position":{"line":220}}]}}`

	findings, err := parseFindings(strings.NewReader(input), naeosModule)
	if err != nil {
		t.Fatalf("parseFindings() error = %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("parseFindings() returned %d findings, want 1 reachable finding", len(findings))
	}
}

func TestParseFindingsRequiresConfigEvent(t *testing.T) {
	_, err := parseFindings(strings.NewReader(`{"progress":{}}`), naeosModule)
	if err == nil {
		t.Fatal("parseFindings() error = nil, want missing config error")
	}
}

func TestNewFindingsFlagsNewTraceForExistingAdvisory(t *testing.T) {
	baselineFinding := govulnFinding{
		OSV: "GO-2026-6865",
		Trace: []traceFrame{
			{Module: "github.com/klauspost/compress", Version: "v1.18.5", Package: "github.com/klauspost/compress/fse", Function: "Decompress"},
			{Module: naeosModule, Package: naeosModule + "/internal/telemetry", Function: "ExportSpans"},
		},
	}
	newTrace := govulnFinding{
		OSV: "GO-2026-6865",
		Trace: []traceFrame{
			{Module: "github.com/klauspost/compress", Version: "v1.18.5", Package: "github.com/klauspost/compress/fse", Function: "Decompress"},
			{Module: naeosModule, Package: naeosModule + "/internal/api", Function: "Handle"},
		},
	}
	baseline := findingSet{findingKey(baselineFinding): baselineFinding}
	head := findingSet{
		findingKey(baselineFinding): baselineFinding,
		findingKey(newTrace):        newTrace,
	}

	added := newFindings(baseline, head)
	if len(added) != 1 || added[0].Trace[len(added[0].Trace)-1].Function != "Handle" {
		t.Fatalf("newFindings() = %+v, want the new call trace only", added)
	}
}
