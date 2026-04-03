package util

import (
	"fmt"
	"strings"
	"time"
)

// SlugToPath converte slug de projeto para path.
// Nota: isso é lossy (hyphens no path original viram /). Usar CWD do JSONL quando disponível.
func SlugToPath(slug string) string {
	return strings.ReplaceAll(slug, "-", "/")
}

// PathToSlug converte path para slug de projeto.
func PathToSlug(path string) string {
	return strings.ReplaceAll(path, "/", "-")
}

// ExtractSlugFromPath extrai o slug do diretório pai de um arquivo JSONL.
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

// FormatTokens formata contagem de tokens para exibição.
func FormatTokens(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// FormatCost formata custo em dólares.
func FormatCost(cost float64) string {
	if cost >= 1.0 {
		return fmt.Sprintf("$%.2f", cost)
	}
	if cost >= 0.01 {
		return fmt.Sprintf("$%.3f", cost)
	}
	return fmt.Sprintf("$%.4f", cost)
}

// FormatDuration formata duração em minutos para exibição.
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

// FormatBytes formata bytes para exibição.
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

// ParseTimestamp parseia timestamp ISO 8601 ou Unix millis.
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

// TruncateString trunca uma string no limite especificado.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
