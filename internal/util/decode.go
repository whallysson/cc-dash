package util

import (
	"fmt"
	"strings"
	"time"
)

// SlugToPath converts a project slug to a path.
// Note: this is lossy (hyphens in the original path become /). Use CWD from JSONL when available.
func SlugToPath(slug string) string {
	return strings.ReplaceAll(slug, "-", "/")
}

// PathToSlug converts a path to a project slug.
func PathToSlug(path string) string {
	return strings.ReplaceAll(path, "/", "-")
}

// ExtractSlugFromPath extracts the slug from the parent directory of a JSONL file.
// Ex: ~/.claude/projects/-Users-foo-bar/abc.jsonl -> -Users-foo-bar
func ExtractSlugFromPath(filePath string) string {
	parts := strings.Split(filePath, "/")
	for i, p := range parts {
		if p == "projects" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// FormatTokens formats a token count for display.
func FormatTokens(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// FormatCost formats a cost in dollars.
func FormatCost(cost float64) string {
	if cost >= 1.0 {
		return fmt.Sprintf("$%.2f", cost)
	}
	if cost >= 0.01 {
		return fmt.Sprintf("$%.3f", cost)
	}
	return fmt.Sprintf("$%.4f", cost)
}

// FormatDuration formats a duration in minutes for display.
func FormatDuration(minutes float64) string {
	if minutes < 1 {
		return "<1m"
	}
	if minutes < 60 {
		return fmt.Sprintf("%.0fm", minutes)
	}
	h := int(minutes) / 60
	m := int(minutes) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

// FormatBytes formats bytes for display.
func FormatBytes(b int64) string {
	if b >= 1<<30 {
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	}
	if b >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	}
	if b >= 1<<10 {
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	}
	return fmt.Sprintf("%d B", b)
}

// ParseTimestamp parses an ISO 8601 timestamp or Unix millis.
func ParseTimestamp(s string) (time.Time, error) {
	// ISO 8601
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unparseable timestamp: %s", s)
}

// TruncateString truncates a string to the specified limit.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
