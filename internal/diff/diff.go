package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

type ChangeType string

const (
	ChangeAdded     ChangeType = "added"
	ChangeRemoved   ChangeType = "removed"
	ChangeModified  ChangeType = "modified"
	ChangeUnchanged ChangeType = "unchanged"
)

type FileDiff struct {
	Path    string
	Type    ChangeType
	OldSize int
	NewSize int
	Lines   []DiffLine
	Stats   DiffStats
}

type DiffLine struct {
	OldNum  int
	NewNum  int
	Type    ChangeType
	Content string
}

type DiffStats struct {
	Added    int
	Removed  int
	Modified int
	Context  int
}

func ComputeDiff(oldContent, newContent string, path string) *FileDiff {
	if oldContent == "" && newContent != "" {
		return &FileDiff{
			Path:    path,
			Type:    ChangeAdded,
			NewSize: len(newContent),
			Lines:   addedLines(newContent),
			Stats:   DiffStats{Added: countLines(newContent)},
		}
	}
	if oldContent != "" && newContent == "" {
		return &FileDiff{
			Path:    path,
			Type:    ChangeRemoved,
			OldSize: len(oldContent),
			Lines:   removedLines(oldContent),
			Stats:   DiffStats{Removed: countLines(oldContent)},
		}
	}
	if oldContent == newContent {
		return &FileDiff{
			Path:    path,
			Type:    ChangeUnchanged,
			OldSize: len(oldContent),
			NewSize: len(newContent),
		}
	}

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	diffLines := computeLineDiff(oldLines, newLines)

	stats := DiffStats{}
	for _, l := range diffLines {
		switch l.Type {
		case ChangeAdded:
			stats.Added++
		case ChangeRemoved:
			stats.Removed++
		case ChangeUnchanged:
			stats.Context++
		}
	}
	stats.Modified = min(stats.Added, stats.Removed)

	return &FileDiff{
		Path:    path,
		Type:    ChangeModified,
		OldSize: len(oldContent),
		NewSize: len(newContent),
		Lines:   diffLines,
		Stats:   stats,
	}
}

func lcs(a, b []string) [][]int {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	return dp
}

func computeLineDiff(oldLines, newLines []string) []DiffLine {
	n, m := len(oldLines), len(newLines)
	if n == 0 && m == 0 {
		return nil
	}

	dp := lcs(oldLines, newLines)

	var result []DiffLine
	i, j := n, m
	type entry struct {
		oldIdx, newIdx int
		line           string
		typ            ChangeType
	}
	var entries []entry

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			entries = append(entries, entry{i, j, oldLines[i-1], ChangeUnchanged})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			entries = append(entries, entry{0, j, newLines[j-1], ChangeAdded})
			j--
		} else {
			entries = append(entries, entry{i, 0, oldLines[i-1], ChangeRemoved})
			i--
		}
	}

	for k := len(entries) - 1; k >= 0; k-- {
		e := entries[k]
		result = append(result, DiffLine{
			OldNum:  e.oldIdx,
			NewNum:  e.newIdx,
			Type:    e.typ,
			Content: e.line,
		})
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func countLines(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Split(content, "\n"))
}

func addedLines(content string) []DiffLine {
	var lines []DiffLine
	for i, line := range strings.Split(content, "\n") {
		lines = append(lines, DiffLine{
			OldNum:  0,
			NewNum:  i + 1,
			Type:    ChangeAdded,
			Content: line,
		})
	}
	return lines
}

func removedLines(content string) []DiffLine {
	var lines []DiffLine
	for i, line := range strings.Split(content, "\n") {
		lines = append(lines, DiffLine{
			OldNum:  i + 1,
			NewNum:  0,
			Type:    ChangeRemoved,
			Content: line,
		})
	}
	return lines
}

func ComputeDirectoryDiff(oldDir, newDir string, paths []string) []*FileDiff {
	var diffs []*FileDiff
	for _, path := range paths {
		oldContent := readFileIfExists(oldDir, path)
		newContent := readFileIfExists(newDir, path)
		diffs = append(diffs, ComputeDiff(oldContent, newContent, path))
	}
	return diffs
}

func readFileIfExists(dir, path string) string {
	fullPath, err := securityext.ValidateFilePath(filepath.Join(dir, path), dir)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}
	return string(data)
}

func FormatDiff(diff *FileDiff) string {
	if diff == nil {
		return ""
	}
	var sb strings.Builder

	switch diff.Type {
	case ChangeAdded:
		fmt.Fprintf(&sb, "\033[32m+++ %s (added)\033[0m\n", diff.Path)
	case ChangeRemoved:
		fmt.Fprintf(&sb, "\033[31m--- %s (removed)\033[0m\n", diff.Path)
	case ChangeUnchanged:
		fmt.Fprintf(&sb, "    %s (unchanged)\n", diff.Path)
		return sb.String()
	case ChangeModified:
		fmt.Fprintf(&sb, "\033[33m~~~ %s (modified: %d -> %d bytes)\033[0m\n", diff.Path, diff.OldSize, diff.NewSize)
	}

	for _, line := range diff.Lines {
		switch line.Type {
		case ChangeAdded:
			fmt.Fprintf(&sb, "\033[32m+ %s\033[0m\n", line.Content)
		case ChangeRemoved:
			fmt.Fprintf(&sb, "\033[31m- %s\033[0m\n", line.Content)
		default:
			fmt.Fprintf(&sb, "  %s\n", line.Content)
		}
	}
	return sb.String()
}

func FormatUnified(diff *FileDiff, contextLines int) string {
	if diff == nil || diff.Type == ChangeUnchanged {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- a/%s\n", diff.Path)
	fmt.Fprintf(&sb, "+++ b/%s\n", diff.Path)

	if diff.Type == ChangeAdded {
		for i, line := range diff.Lines {
			fmt.Fprintf(&sb, "+%s\n", line.Content)
			_ = i
		}
		return sb.String()
	}
	if diff.Type == ChangeRemoved {
		for _, line := range diff.Lines {
			fmt.Fprintf(&sb, "-%s\n", line.Content)
		}
		return sb.String()
	}

	var hunks []hunk
	var current *hunk
	for i, line := range diff.Lines {
		if line.Type == ChangeUnchanged {
			if current != nil {
				current.End = i
				hunks = append(hunks, *current)
				current = nil
			}
			continue
		}
		if current == nil {
			start := i - contextLines
			if start < 0 {
				start = 0
			}
			current = &hunk{Start: start, End: i}
		}
		current.End = i + 1
	}
	if current != nil {
		hunks = append(hunks, *current)
	}

	for _, h := range hunks {
		oldStart := 0
		oldCount := 0
		newStart := 0
		newCount := 0
		for _, line := range diff.Lines[h.Start:h.End] {
			switch line.Type {
			case ChangeRemoved:
				if oldCount == 0 {
					oldStart = line.OldNum
				}
				oldCount++
			case ChangeAdded:
				if newCount == 0 {
					newStart = line.NewNum
				}
				newCount++
			case ChangeUnchanged:
				if oldCount == 0 && newCount == 0 {
					oldStart = line.OldNum
					newStart = line.NewNum
				}
				oldCount++
				newCount++
			}
		}
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount)
		for _, line := range diff.Lines[h.Start:h.End] {
			switch line.Type {
			case ChangeAdded:
				fmt.Fprintf(&sb, "+%s\n", line.Content)
			case ChangeRemoved:
				fmt.Fprintf(&sb, "-%s\n", line.Content)
			default:
				fmt.Fprintf(&sb, " %s\n", line.Content)
			}
		}
	}
	return sb.String()
}

type hunk struct {
	Start int
	End   int
}

func Summary(diffs []*FileDiff) (added, removed, modified, unchanged int) {
	for _, d := range diffs {
		switch d.Type {
		case ChangeAdded:
			added++
		case ChangeRemoved:
			removed++
		case ChangeModified:
			modified++
		case ChangeUnchanged:
			unchanged++
		}
	}
	return
}

func ApplyPatch(original string, diff *FileDiff) string {
	if diff == nil {
		return original
	}
	switch diff.Type {
	case ChangeAdded:
		return patchAdded(original, diff)
	case ChangeRemoved:
		return ""
	case ChangeUnchanged:
		return original
	case ChangeModified:
		return patchModified(original, diff)
	}
	return original
}

func patchAdded(original string, diff *FileDiff) string {
	var sb strings.Builder
	sb.WriteString(original)
	if original != "" && !strings.HasSuffix(original, "\n") {
		sb.WriteString("\n")
	}
	for _, line := range diff.Lines {
		sb.WriteString(line.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func patchModified(original string, diff *FileDiff) string {
	var result []string
	removed := map[int]bool{}
	for _, line := range diff.Lines {
		if line.Type == ChangeRemoved && line.OldNum > 0 {
			removed[line.OldNum] = true
		}
	}

	oldLines := strings.Split(original, "\n")
	for i, line := range oldLines {
		if !removed[i+1] {
			result = append(result, line)
		}
	}

	for _, line := range diff.Lines {
		if line.Type == ChangeAdded {
			insertIdx := len(result)
			if line.NewNum > 0 {
				insertIdx = line.NewNum - 1
				if insertIdx > len(result) {
					insertIdx = len(result)
				}
			}
			newResult := make([]string, 0, len(result)+1)
			newResult = append(newResult, result[:insertIdx]...)
			newResult = append(newResult, line.Content)
			newResult = append(newResult, result[insertIdx:]...)
			result = newResult
		}
	}

	return strings.Join(result, "\n")
}
