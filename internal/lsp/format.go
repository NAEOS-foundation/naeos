package lsp

import (
	"sort"
	"strings"
	"unicode"
)

type FormatProvider struct{}

func NewFormatProvider() *FormatProvider {
	return &FormatProvider{}
}

var sortOrder = map[string]int{
	"name":        1,
	"project":     2,
	"version":     3,
	"description": 4,
}

func (fp *FormatProvider) Format(text string) ([]TextEdit, error) {
	lines := strings.Split(text, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]

		trimmed := strings.TrimRightFunc(line, unicode.IsSpace)

		result = append(result, trimmed)
		i++
	}

	text = strings.Join(result, "\n")

	lines = strings.Split(text, "\n")
	var blocks []block
	var current block
	inBlock := false

	for _, line := range lines {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if indent == 0 && strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "- ") {
			if inBlock {
				blocks = append(blocks, current)
			}
			current = block{lines: []string{strings.ReplaceAll(line, "\t", "  ")}}
			inBlock = true
			continue
		}

		if inBlock {
			current.lines = append(current.lines, strings.ReplaceAll(line, "\t", "  "))
		}
	}

	if inBlock {
		blocks = append(blocks, current)
	}

	var outLines []string
	sortedBlocks := fp.sortBlocks(blocks)

	for bi, b := range sortedBlocks {
		if bi > 0 {
			outLines = append(outLines, "")
		}
		for _, line := range b.lines {
			fixed := strings.ReplaceAll(line, "\t", "  ")
			outLines = append(outLines, fixed)
		}
	}

	out := strings.Join(outLines, "\n")

	return []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: len(lines), Character: 0},
			},
			NewText: out,
		},
	}, nil
}

type block struct {
	lines []string
	key   string
}

func (fp *FormatProvider) sortBlocks(blocks []block) []block {
	sorted := make([]block, len(blocks))
	copy(sorted, blocks)

	normalize := func(k string) string {
		return strings.TrimSuffix(strings.ToLower(k), ":")
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		ki := normalize(extractKey(sorted[i].lines[0]))
		kj := normalize(extractKey(sorted[j].lines[0]))

		oi, oki := sortOrder[ki]
		oj, okj := sortOrder[kj]

		if oki && okj {
			return oi < oj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return ki < kj
	})

	return sorted
}

func extractKey(line string) string {
	trimmed := strings.TrimSpace(line)
	colonIdx := strings.Index(trimmed, ":")
	if colonIdx > 0 {
		return trimmed[:colonIdx]
	}
	return trimmed
}

func (fp *FormatProvider) FormatRange(text string, r Range) ([]TextEdit, error) {
	fullEdits, err := fp.Format(text)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(text, "\n")
	if len(fullEdits) == 0 {
		return nil, nil
	}

	formattedLines := strings.Split(fullEdits[0].NewText, "\n")

	startLine := r.Start.Line
	endLine := r.End.Line
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if endLine > len(formattedLines) {
		endLine = len(formattedLines)
	}

	var rangeLines []string
	for i := startLine; i < endLine && i < len(formattedLines); i++ {
		rangeLines = append(rangeLines, formattedLines[i])
	}

	return []TextEdit{
		{
			Range: Range{
				Start: Position{Line: startLine, Character: 0},
				End:   Position{Line: endLine, Character: 0},
			},
			NewText: strings.Join(rangeLines, "\n"),
		},
	}, nil
}
