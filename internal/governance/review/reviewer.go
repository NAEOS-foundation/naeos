package review

import (
	"fmt"
	"strings"
)

type ReviewStatus string

const (
	StatusApproved  ReviewStatus = "approved"
	StatusRejected  ReviewStatus = "rejected"
	StatusPending   ReviewStatus = "pending"
	StatusChanges   ReviewStatus = "changes_requested"
)

type ReviewComment struct {
	RuleID  string
	Message string
}

type ReviewResult struct {
	Status   ReviewStatus
	Comments []ReviewComment
	Summary  string
}

type Reviewer interface {
	Review(input any) error
	ReviewArtifact(name, content string, rules []string) (*ReviewResult, error)
}

type DefaultReviewer struct{}

func NewReviewer() Reviewer {
	return DefaultReviewer{}
}

func (DefaultReviewer) Review(input any) error {
	if input == nil {
		return fmt.Errorf("review input is nil")
	}
	return nil
}

func (DefaultReviewer) ReviewArtifact(name, content string, rules []string) (*ReviewResult, error) {
	if name == "" {
		return nil, fmt.Errorf("artifact name must not be empty")
	}

	result := &ReviewResult{
		Status: StatusApproved,
	}

	if content == "" {
		result.Status = StatusRejected
		result.Comments = append(result.Comments, ReviewComment{
			Message: "artifact content is empty",
		})
		return result, nil
	}

	for _, rule := range rules {
		switch rule {
		case "no-todo":
			if containsTODO(content) {
				result.Status = StatusChanges
				result.Comments = append(result.Comments, ReviewComment{
					RuleID:  rule,
					Message: fmt.Sprintf("artifact %s contains TODO comments", name),
				})
			}
		case "no-placeholder":
			if containsPlaceholder(content) {
				result.Status = StatusChanges
				result.Comments = append(result.Comments, ReviewComment{
					RuleID:  rule,
					Message: fmt.Sprintf("artifact %s contains placeholder text", name),
				})
			}
		case "has-package-declaration":
			if !containsPackageDecl(content) {
				result.Status = StatusRejected
				result.Comments = append(result.Comments, ReviewComment{
					RuleID:  rule,
					Message: fmt.Sprintf("Go file %s missing package declaration", name),
				})
			}
		case "has-license-header":
			if !containsLicense(content) {
				result.Status = StatusChanges
				result.Comments = append(result.Comments, ReviewComment{
					RuleID:  rule,
					Message: fmt.Sprintf("file %s missing license header", name),
				})
			}
		}
	}

	if len(result.Comments) == 0 {
		result.Summary = fmt.Sprintf("artifact %s passed all review rules", name)
	} else {
		result.Summary = fmt.Sprintf("artifact %s has %d review comments", name, len(result.Comments))
	}

	return result, nil
}

func containsTODO(content string) bool {
	for _, line := range splitLines(content) {
		trimmed := trimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "todo") {
			return true
		}
		if strings.Contains(lower, "todo") {
			return true
		}
	}
	return false
}

func containsPlaceholder(content string) bool {
	placeholders := []string{"TODO", "FIXME", "XXX", "PLACEHOLDER", "CHANGEME", "REPLACE_ME"}
	lowerContent := strings.ToLower(content)
	for _, p := range placeholders {
		if strings.Contains(lowerContent, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func containsPackageDecl(content string) bool {
	return containsStr(content, "package ")
}

func containsLicense(content string) bool {
	lines := splitLines(content)
	maxLines := 20
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	header := strings.Join(lines[:maxLines], "\n")
	licenseMarkers := []string{"license", "apache", "mit", "copyright", "licensed under"}
	lowerHeader := strings.ToLower(header)

	for _, marker := range licenseMarkers {
		if strings.Contains(lowerHeader, marker) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func containsStr(s, substr string) bool {
	lowerS := strings.ToLower(s)
	lowerSub := strings.ToLower(substr)
	for i := 0; i <= len(lowerS)-len(lowerSub); i++ {
		if lowerS[i:i+len(lowerSub)] == lowerSub {
			return true
		}
	}
	return false
}
