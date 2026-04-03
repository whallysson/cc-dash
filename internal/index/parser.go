package index

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/whallysson/cc-dash/internal/model"
	"github.com/whallysson/cc-dash/internal/util"
)

// peekType extracts the top-level type from a JSON line.
// The top-level "type" field has distinct values from nested ones ("message" is nested, never top-level).
// We use bytes.Contains with specific values to avoid confusion with nested fields.
func peekType(line []byte) string {
	switch {
	case bytes.Contains(line, []byte(`"type":"user"`)):
		return "user"
	case bytes.Contains(line, []byte(`"type":"assistant"`)):
		return "assistant"
	case bytes.Contains(line, []byte(`"type":"system"`)):
		return "system"
	case bytes.Contains(line, []byte(`"type":"progress"`)):
		return "progress"
	case bytes.Contains(line, []byte(`"type":"file-history-snapshot"`)):
		return "file-history-snapshot"
	case bytes.Contains(line, []byte(`"type":"attachment"`)):
		return "attachment"
	case bytes.Contains(line, []byte(`"type":"permission-mode"`)):
		return "permission-mode"
	case bytes.Contains(line, []byte(`"type":"summary"`)):
		return "summary"
	default:
		return ""
	}
}

// peekStringField extracts a simple string field using byte scanning.
func peekStringField(line []byte, field string) string {
	needle := []byte(`"` + field + `":"`)
	idx := bytes.Index(line, needle)
	if idx == -1 {
		return ""
	}
	start := idx + len(needle)
	end := bytes.IndexByte(line[start:], '"')
	if end == -1 {
		return ""
	}
	return string(line[start : start+end])
}

// jsonlLine represents the relevant fields of a JSONL line.
// We use a partial struct to avoid parsing irrelevant fields.
type jsonlLine struct {
	Type       string    `json:"type"`
	Timestamp  string    `json:"timestamp"`
	SessionID  string    `json:"sessionId"`
	CWD        string    `json:"cwd"`
	Version    string    `json:"version"`
	GitBranch  string    `json:"gitBranch"`
	Entrypoint string    `json:"entrypoint"`
	Message    *jsonlMsg `json:"message,omitempty"`
}

type jsonlMsg struct {
	Role    string          `json:"role"`
	Model   string          `json:"model"`
	Content json.RawMessage `json:"content"`
	Usage   *jsonlUsage     `json:"usage,omitempty"`
}

type jsonlUsage struct {
	InputTokens      int64 `json:"input_tokens"`
	OutputTokens     int64 `json:"output_tokens"`
	CacheReadTokens  int64 `json:"cache_read_input_tokens"`
	CacheWriteTokens int64 `json:"cache_creation_input_tokens"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
}

// ParseSessionFile parses a complete JSONL file and returns SessionMeta.
func ParseSessionFile(path string) (*model.SessionMeta, int64, error) {
	return ParseSessionFileFrom(path, 0)
}

// ParseSessionFileFrom parses a JSONL file from a byte offset.
// Returns SessionMeta and the final offset (for incremental parsing).
func ParseSessionFileFrom(path string, offset int64) (*model.SessionMeta, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return nil, 0, err
		}
	}

	slug := util.ExtractSlugFromPath(path)

	meta := &model.SessionMeta{
		Slug:        slug,
		SourceFile:  path,
		ToolCounts:  make(map[string]int),
		ModelTokens: make(map[string]model.TokenUsage),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024) // 256KB buffer, 10MB max

	var (
		lineCount    int
		bytesRead    int64
		firstTS      time.Time
		lastTS       time.Time
		cwdResolved  bool
	)

	for scanner.Scan() {
		line := scanner.Bytes()
		bytesRead += int64(len(line)) + 1 // +1 para newline
		lineCount++

		lineType := peekType(line)

		switch lineType {
		case "progress", "file-history-snapshot", "attachment", "permission-mode":
			// Skip - irrelevant lines for the index (50-60% of volume)
			continue

		case "user":
			meta.UserMsgCount++
			if meta.FirstPrompt == "" {
				parseUserPrompt(line, meta)
			}

		case "assistant":
			meta.AsstMsgCount++
			parseAssistantLine(line, meta)

		case "system":
			// Detect compaction
			if bytes.Contains(line, []byte(`compact_boundary`)) || bytes.Contains(line, []byte(`"subtype":"compact"`)) {
				meta.HasCompaction = true
			}
		}

		// Extract metadata from the first lines
		if lineCount <= 50 && !cwdResolved {
			if cwd := peekStringField(line, "cwd"); cwd != "" {
				meta.ProjectPath = cwd
				cwdResolved = true
			}
		}

		// Extract timestamp, version, gitBranch, sessionID, entrypoint
		if ts := peekStringField(line, "timestamp"); ts != "" {
			t, err := util.ParseTimestamp(ts)
			if err == nil {
				if firstTS.IsZero() {
					firstTS = t
				}
				lastTS = t
			}
		}

		if meta.SessionID == "" {
			if sid := peekStringField(line, "sessionId"); sid != "" {
				meta.SessionID = sid
			}
		}
		if meta.Version == "" {
			if v := peekStringField(line, "version"); v != "" {
				meta.Version = v
			}
		}
		if meta.GitBranch == "" {
			if gb := peekStringField(line, "gitBranch"); gb != "" {
				meta.GitBranch = gb
			}
		}
		if meta.Entrypoint == "" {
			if ep := peekStringField(line, "entrypoint"); ep != "" {
				meta.Entrypoint = ep
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	// Calculate derived fields
	meta.StartTime = firstTS
	meta.EndTime = lastTS
	if !firstTS.IsZero() && !lastTS.IsZero() {
		meta.DurationMin = lastTS.Sub(firstTS).Minutes()
	}
	meta.TotalMsgCount = meta.UserMsgCount + meta.AsstMsgCount

	// Calculate total tokens and cost
	for mdl, tokens := range meta.ModelTokens {
		meta.TotalTokens.Add(tokens)
		meta.EstimatedCost += model.CalculateCost(mdl, tokens)
	}

	// Fallback: if ProjectPath was not resolved via cwd, use slug
	if meta.ProjectPath == "" {
		meta.ProjectPath = util.SlugToPath(slug)
	}

	return meta, offset + bytesRead, nil
}

// parseUserPrompt extracts the first user prompt.
func parseUserPrompt(line []byte, meta *model.SessionMeta) {
	var parsed jsonlLine
	if err := json.Unmarshal(line, &parsed); err != nil {
		return
	}
	if parsed.Message == nil {
		return
	}

	// Content can be string or array
	content := parsed.Message.Content
	if len(content) == 0 {
		return
	}

	// Try as string
	var textStr string
	if err := json.Unmarshal(content, &textStr); err == nil {
		meta.FirstPrompt = util.TruncateString(cleanPrompt(textStr), 500)
		return
	}

	// Try as array of content items
	var items []contentItem
	if err := json.Unmarshal(content, &items); err == nil {
		for _, item := range items {
			if item.Type == "text" && item.Text != "" {
				meta.FirstPrompt = util.TruncateString(cleanPrompt(item.Text), 500)
				return
			}
		}
	}
}

// cleanPrompt removes XML tags and excessive whitespace from the prompt.
func cleanPrompt(s string) string {
	// Remove tags <system-reminder>...</system-reminder>
	for {
		start := strings.Index(s, "<system-reminder>")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "</system-reminder>")
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+18:]
	}

	s = strings.TrimSpace(s)
	return s
}

// parseAssistantLine extracts usage, model, and tool calls from an assistant line.
func parseAssistantLine(line []byte, meta *model.SessionMeta) {
	var parsed jsonlLine
	if err := json.Unmarshal(line, &parsed); err != nil {
		return
	}
	if parsed.Message == nil {
		return
	}

	mdl := parsed.Message.Model
	if mdl == "" || mdl == "<synthetic>" {
		return
	}

	// Accumulate tokens per model
	if parsed.Message.Usage != nil {
		u := parsed.Message.Usage
		existing := meta.ModelTokens[mdl]
		existing.InputTokens += u.InputTokens
		existing.OutputTokens += u.OutputTokens
		existing.CacheReadTokens += u.CacheReadTokens
		existing.CacheWriteTokens += u.CacheWriteTokens
		meta.ModelTokens[mdl] = existing
	}

	// Extract tool calls and features from content
	if parsed.Message.Content != nil {
		var items []contentItem
		if err := json.Unmarshal(parsed.Message.Content, &items); err == nil {
			for _, item := range items {
				switch item.Type {
				case "tool_use":
					if item.Name != "" {
						meta.ToolCounts[item.Name]++
						detectFeature(item.Name, meta)
					}
				case "thinking":
					meta.HasThinking = true
				}
			}
		}
	}
}

// detectFeature sets feature flags based on tool name.
func detectFeature(toolName string, meta *model.SessionMeta) {
	switch {
	case toolName == "Agent" || toolName == "TaskCreate" || toolName == "SendMessage" || toolName == "TeamCreate":
		meta.UsesTaskAgent = true
	case toolName == "WebSearch":
		meta.UsesWebSearch = true
	case toolName == "WebFetch":
		meta.UsesWebFetch = true
	case model.IsMCPTool(toolName):
		meta.UsesMCP = true
	}
}
