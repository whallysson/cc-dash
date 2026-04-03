package model

// Categorias de ferramentas do Claude Code.
var ToolCategories = map[string]string{
	// File I/O
	"Read":         "file-io",
	"Write":        "file-io",
	"Edit":         "file-io",
	"MultiEdit":    "file-io",
	"NotebookEdit": "file-io",
	"Glob":         "file-io",
	"Grep":         "file-io",

	// Shell
	"Bash":      "shell",
	"Terminal":  "shell",

	// Agent
	"Agent":       "agent",
	"TaskCreate":  "agent",
	"TaskUpdate":  "agent",
	"TaskGet":     "agent",
	"TaskList":    "agent",
	"TaskOutput":  "agent",
	"TaskStop":    "agent",
	"SendMessage": "agent",
	"TeamCreate":  "agent",
	"TeamDelete":  "agent",

	// Web
	"WebSearch": "web",
	"WebFetch":  "web",

	// Planning
	"EnterPlanMode": "planning",
	"ExitPlanMode":  "planning",

	// Skill
	"Skill":      "skill",
	"ToolSearch": "skill",

	// LSP
	"LSP": "code-intelligence",
}

// GetToolCategory retorna a categoria de uma ferramenta.
func GetToolCategory(toolName string) string {
	if cat, ok := ToolCategories[toolName]; ok {
		return cat
	}
	// MCP tools começam com "mcp__"
	if len(toolName) > 5 && toolName[:5] == "mcp__" {
		return "mcp"
	}
	return "other"
}

// IsMCPTool verifica se é uma ferramenta MCP.
func IsMCPTool(name string) bool {
	return len(name) > 5 && name[:5] == "mcp__"
}

// GetMCPServerName extrai o nome do servidor MCP de uma tool name.
// Formato: mcp__<server>__<tool>
func GetMCPServerName(toolName string) string {
	if !IsMCPTool(toolName) {
		return ""
	}
	rest := toolName[5:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == '_' && i+1 < len(rest) && rest[i+1] == '_' {
			return rest[:i]
		}
	}
	return rest
}
