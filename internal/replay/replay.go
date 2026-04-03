package replay

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/whallysson/cc-dash/internal/model"
	"github.com/whallysson/cc-dash/internal/util"
)

// ParseReplay parses a session JSONL file for replay, with pagination.
// offset: number of relevant turns to skip
// limit: maximum number of turns to return (0 = all)
func ParseReplay(path string, offset, limit int) (*model.ReplayData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024)

	data := &model.ReplayData{
		Turns:       make([]model.ReplayTurn, 0),
		Compactions: make([]model.CompactionEvent, 0),
	}

	var (
		turnIndex    int
		collected    int
		skipped      int
		totalCost    float64
	)

	for scanner.Scan() {
		line := scanner.Bytes()
		lineType := peekType(line)

		switch lineType {
		case "progress", "file-history-snapshot", "attachment", "permission-mode":
			continue

		case "user":
			turn := parseUserTurn(line, turnIndex)
			if turn == nil {
				continue
			}
			turnIndex++

			if skipped < offset {
				skipped++
				continue
			}
			if limit > 0 && collected >= limit {
				data.HasMore = true
				continue
			}
			data.Turns = append(data.Turns, *turn)
			collected++

		case "assistant":
			turn := parseAssistantTurn(line, turnIndex)
			if turn == nil {
				continue
			}
			totalCost += turn.Cost
			turnIndex++

			if skipped < offset {
				skipped++
				continue
			}
			if limit > 0 && collected >= limit {
				data.HasMore = true
				continue
			}
			data.Turns = append(data.Turns, *turn)
			collected++

		case "system":
			if compaction := parseCompaction(line, turnIndex); compaction != nil {
				data.Compactions = append(data.Compactions, *compaction)
			}
		}

		// Extract metadata
		if data.Slug == "" {
			if slug := peekStringField(line, "slug"); slug != "" {
				data.Slug = slug
			}
		}
		if data.Version == "" {
			if v := peekStringField(line, "version"); v != "" {
				data.Version = v
			}
		}
		if data.GitBranch == "" {
			if gb := peekStringField(line, "gitBranch"); gb != "" {
				data.GitBranch = gb
			}
		}
	}

	data.TotalCost = totalCost
	data.NextOffset = offset + collected

	return data, scanner.Err()
}

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
	case bytes.Contains(line, []byte(`"type":"summary"`)):
		return "summary"
	default:
		return ""
	}
}

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

// Partial structs for parsing
type replayLine struct {
	Type      string     `json:"type"`
	Timestamp string     `json:"timestamp"`
	UUID      string     `json:"uuid"`
	Message   *replayMsg `json:"message,omitempty"`
}

type replayMsg struct {
	Role    string          `json:"role"`
	Model   string          `json:"model"`
	Content json.RawMessage `json:"content"`
	Usage   *replayUsage    `json:"usage,omitempty"`
}

type replayUsage struct {
	InputTokens      int64 `json:"input_tokens"`
	OutputTokens     int64 `json:"output_tokens"`
	CacheReadTokens  int64 `json:"cache_read_input_tokens"`
	CacheWriteTokens int64 `json:"cache_creation_input_tokens"`
}

type contentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
}

func parseUserTurn(line []byte, index int) *model.ReplayTurn {
	var parsed replayLine
	if err := json.Unmarshal(line, &parsed); err != nil {
		return nil
	}
	if parsed.Message == nil {
		return nil
	}

	turn := &model.ReplayTurn{
		Index: index,
		Role:  "user",
		UUID:  parsed.UUID,
	}

	if ts, err := util.ParseTimestamp(parsed.Timestamp); err == nil {
		turn.Timestamp = ts
	}

	// Extract text from content
	turn.Text = extractText(parsed.Message.Content)

	return turn
}

func parseAssistantTurn(line []byte, index int) *model.ReplayTurn {
	var parsed replayLine
	if err := json.Unmarshal(line, &parsed); err != nil {
		return nil
	}
	if parsed.Message == nil {
		return nil
	}

	turn := &model.ReplayTurn{
		Index: index,
		Role:  "assistant",
		Model: parsed.Message.Model,
		UUID:  parsed.UUID,
	}

	if ts, err := util.ParseTimestamp(parsed.Timestamp); err == nil {
		turn.Timestamp = ts
	}

	// Tokens
	if parsed.Message.Usage != nil {
		u := parsed.Message.Usage
		turn.Tokens = model.TokenUsage{
			InputTokens:      u.InputTokens,
			OutputTokens:     u.OutputTokens,
			CacheReadTokens:  u.CacheReadTokens,
			CacheWriteTokens: u.CacheWriteTokens,
		}
		turn.Cost = model.CalculateCost(parsed.Message.Model, turn.Tokens)
	}

	// Content: text, thinking, tool calls
	if parsed.Message.Content != nil {
		var blocks []contentBlock
		if err := json.Unmarshal(parsed.Message.Content, &blocks); err == nil {
			var textParts []string
			for _, b := range blocks {
				switch b.Type {
				case "text":
					textParts = append(textParts, b.Text)
				case "thinking":
					turn.HasThinking = true
					turn.ThinkingText = util.TruncateString(b.Thinking, 4096)
				case "tool_use":
					tc := model.ToolCall{
						ID:   b.ID,
						Name: b.Name,
					}
					if b.Input != nil {
						tc.Input = util.TruncateString(string(b.Input), 2048)
					}
					turn.ToolCalls = append(turn.ToolCalls, tc)
				}
			}
			turn.Text = strings.Join(textParts, "\n")
		}
	}

	return turn
}

func parseCompaction(line []byte, turnIndex int) *model.CompactionEvent {
	// Check if this is a compaction event
	var raw map[string]interface{}
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil
	}

	msg, ok := raw["message"].(map[string]interface{})
	if !ok {
		return nil
	}

	content, ok := msg["content"].(string)
	if !ok {
		return nil
	}

	if !strings.Contains(content, "compact") {
		return nil
	}

	return &model.CompactionEvent{
		TurnIndex: turnIndex,
		Summary:   util.TruncateString(content, 1024),
	}
}

func extractText(content json.RawMessage) string {
	if content == nil {
		return ""
	}

	// Try as string
	var textStr string
	if err := json.Unmarshal(content, &textStr); err == nil {
		return cleanText(textStr)
	}

	// Try as array
	var blocks []contentBlock
	if err := json.Unmarshal(content, &blocks); err == nil {
		var parts []string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				parts = append(parts, b.Text)
			}
		}
		return cleanText(strings.Join(parts, "\n"))
	}

	return ""
}

func cleanText(s string) string {
	// Remove tags system-reminder
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
	return strings.TrimSpace(s)
}
