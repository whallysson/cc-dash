package util

import (
	"strings"
)

// ParseFrontmatter extrai campos YAML de frontmatter simples.
// Retorna o frontmatter como map e o conteúdo restante.
func ParseFrontmatter(content string) (map[string]string, string) {
	fm := make(map[string]string)
	if !strings.HasPrefix(content, "---\n") {
		return fm, content
	}

	end := strings.Index(content[4:], "\n---")
	if end == -1 {
		return fm, content
	}

	fmBlock := content[4 : 4+end]
	body := content[4+end+4:]

	for _, line := range strings.Split(fmBlock, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Remove aspas
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		fm[key] = val
	}

	return fm, strings.TrimLeft(body, "\n")
}
