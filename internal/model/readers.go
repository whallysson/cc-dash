package model

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ReadStatsCache reads ~/.claude/stats-cache.json.
func ReadStatsCache(claudeDir string) (*StatsCache, error) {
	path := filepath.Join(claudeDir, "stats-cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stats StatsCache
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// ReadHistory reads ~/.claude/history.jsonl with limit and search.
func ReadHistory(claudeDir string, limit int, query string) ([]HistoryEntry, error) {
	path := filepath.Join(claudeDir, "history.jsonl")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1*1024*1024)

	for scanner.Scan() {
		var entry HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if query != "" && !containsCI(entry.Display, query) {
			continue
		}
		entries = append(entries, entry)
	}

	// Sort by timestamp descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp > entries[j].Timestamp
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, scanner.Err()
}

// ReadMemories reads all memory files.
func ReadMemories(claudeDir string) ([]MemoryEntry, error) {
	var memories []MemoryEntry

	// Global memory
	globalDir := filepath.Join(claudeDir, "memory")
	readMemDir(globalDir, "global", &memories)

	// Per-project memory
	projectsDir := filepath.Join(claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			memDir := filepath.Join(projectsDir, entry.Name(), "memory")
			readMemDir(memDir, entry.Name(), &memories)
		}
	}

	sort.Slice(memories, func(i, j int) bool {
		return memories[i].ModTime.After(memories[j].ModTime)
	})

	return memories, nil
}

func readMemDir(dir, slug string, out *[]MemoryEntry) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}

		fm, body := parseFrontmatter(string(content))
		mem := MemoryEntry{
			FilePath:    path,
			Slug:        slug,
			Name:        fm["name"],
			Description: fm["description"],
			Type:        fm["type"],
			Content:     body,
			ModTime:     info.ModTime(),
			Frontmatter: fm,
		}
		if mem.Name == "" {
			mem.Name = strings.TrimSuffix(e.Name(), ".md")
		}
		*out = append(*out, mem)
	}
}

// ReadPlans reads all plans from ~/.claude/plans/.
func ReadPlans(claudeDir string) ([]PlanFile, error) {
	dir := filepath.Join(claudeDir, "plans")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var plans []PlanFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		plans = append(plans, PlanFile{
			Name:    e.Name(),
			Path:    path,
			Content: string(content),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(plans, func(i, j int) bool {
		return plans[i].ModTime.After(plans[j].ModTime)
	})

	return plans, nil
}

// ReadTodos reads all todo files from ~/.claude/todos/.
func ReadTodos(claudeDir string) ([]TodoFile, error) {
	dir := filepath.Join(claudeDir, "todos")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var todos []TodoFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}

		var items []TodoItem
		if err := json.Unmarshal(data, &items); err != nil {
			// Try as single object
			var single TodoItem
			if err := json.Unmarshal(data, &single); err == nil {
				items = []TodoItem{single}
			} else {
				continue
			}
		}

		todos = append(todos, TodoFile{
			Name:    e.Name(),
			Path:    path,
			Items:   items,
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(todos, func(i, j int) bool {
		return todos[i].ModTime.After(todos[j].ModTime)
	})

	return todos, nil
}

// ReadSettings reads ~/.claude/settings.json.
func ReadSettings(claudeDir string) (*Settings, error) {
	path := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// ReadInstalledPlugins reads ~/.claude/plugins/installed_plugins.json.
func ReadInstalledPlugins(claudeDir string) ([]PluginInfo, error) {
	path := filepath.Join(claudeDir, "plugins", "installed_plugins.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var plugins []PluginInfo
	for name, info := range raw {
		p := PluginInfo{Name: name, Path: path}
		if m, ok := info.(map[string]interface{}); ok {
			if v, ok := m["version"].(string); ok {
				p.Version = v
			}
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

// GetClaudeStorageBytes returns the total size of the ~/.claude/ directory.
func GetClaudeStorageBytes(claudeDir string) int64 {
	var total int64
	filepath.Walk(claudeDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// parseFrontmatter extracts simple YAML frontmatter fields.
func parseFrontmatter(content string) (map[string]string, string) {
	fm := make(map[string]string)
	if !strings.HasPrefix(content, "---\n") {
		return fm, content
	}
	end := strings.Index(content[4:], "\n---")
	if end == -1 {
		return fm, content
	}
	block := content[4 : 4+end]
	body := content[4+end+4:]

	for _, line := range strings.Split(block, "\n") {
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
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		fm[key] = val
	}
	return fm, strings.TrimLeft(body, "\n")
}

// containsCI performs a case-insensitive search.
func containsCI(haystack, needle string) bool {
	return strings.Contains(
		strings.ToLower(haystack),
		strings.ToLower(needle),
	)
}

// IsStaleMemory checks if a memory file is stale (>30 days old).
func IsStaleMemory(modTime time.Time) bool {
	return time.Since(modTime) > 30*24*time.Hour
}
