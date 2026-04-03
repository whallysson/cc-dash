package model

import "time"

// SessionMeta contém os metadados extraídos de um arquivo JSONL de sessão.
type SessionMeta struct {
	SessionID     string                `json:"session_id"`
	Slug          string                `json:"slug"`
	ProjectPath   string                `json:"project_path"`
	SourceFile    string                `json:"-"`
	StartTime     time.Time             `json:"start_time"`
	EndTime       time.Time             `json:"end_time"`
	DurationMin   float64               `json:"duration_minutes"`
	UserMsgCount  int                   `json:"user_message_count"`
	AsstMsgCount  int                   `json:"assistant_message_count"`
	TotalMsgCount int                   `json:"total_message_count"`
	ToolCounts    map[string]int        `json:"tool_counts"`
	ModelTokens   map[string]TokenUsage `json:"model_tokens"`
	TotalTokens   TokenUsage            `json:"total_tokens"`
	EstimatedCost float64               `json:"estimated_cost"`
	FirstPrompt   string                `json:"first_prompt"`
	GitBranch     string                `json:"git_branch"`
	Version       string                `json:"version"`
	Entrypoint    string                `json:"entrypoint"`
	HasCompaction bool                  `json:"has_compaction"`
	HasThinking   bool                  `json:"has_thinking"`
	UsesMCP       bool                  `json:"uses_mcp"`
	UsesTaskAgent bool                  `json:"uses_task_agent"`
	UsesWebSearch bool                  `json:"uses_web_search"`
	UsesWebFetch  bool                  `json:"uses_web_fetch"`
}

// TokenUsage armazena contagem de tokens por tipo.
type TokenUsage struct {
	InputTokens      int64 `json:"input_tokens"`
	OutputTokens     int64 `json:"output_tokens"`
	CacheReadTokens  int64 `json:"cache_read_input_tokens"`
	CacheWriteTokens int64 `json:"cache_creation_input_tokens"`
}

// Total retorna a soma de todos os tokens.
func (t TokenUsage) Total() int64 {
	return t.InputTokens + t.OutputTokens + t.CacheReadTokens + t.CacheWriteTokens
}

// Add soma outro TokenUsage neste.
func (t *TokenUsage) Add(other TokenUsage) {
	t.InputTokens += other.InputTokens
	t.OutputTokens += other.OutputTokens
	t.CacheReadTokens += other.CacheReadTokens
	t.CacheWriteTokens += other.CacheWriteTokens
}

// FileState rastreia o estado de um arquivo para parsing incremental.
type FileState struct {
	Path   string    `json:"path"`
	Mtime  time.Time `json:"mtime"`
	Size   int64     `json:"size"`
	Offset int64     `json:"offset"`
}

// StatsCache representa o formato de ~/.claude/stats-cache.json.
type StatsCache struct {
	Version          int                       `json:"version"`
	LastComputedDate string                    `json:"lastComputedDate"`
	DailyActivity    []DailyActivity           `json:"dailyActivity"`
	DailyModelTokens []DailyModelTokens        `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsageStats `json:"modelUsage"`
	TotalSessions    int                       `json:"totalSessions"`
	TotalMessages    int                       `json:"totalMessages"`
	LongestSession   LongestSession            `json:"longestSession"`
	FirstSessionDate string                    `json:"firstSessionDate"`
	HourCounts       map[string]int            `json:"hourCounts"`
}

type DailyActivity struct {
	Date     string `json:"date"`
	Sessions int    `json:"sessions"`
	Messages int    `json:"messages"`
	Tokens   int64  `json:"tokens"`
}

type DailyModelTokens struct {
	Date   string         `json:"date"`
	Models map[string]int64 `json:"models"`
}

type ModelUsageStats struct {
	InputTokens  int64 `json:"inputTokens"`
	OutputTokens int64 `json:"outputTokens"`
	TotalCost    float64 `json:"totalCost"`
}

type LongestSession struct {
	SessionID    string `json:"sessionId"`
	Duration     int64  `json:"duration"`
	MessageCount int    `json:"messageCount"`
}

// HistoryEntry representa uma linha de ~/.claude/history.jsonl.
type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
	SessionID string `json:"sessionId"`
}

// MemoryEntry representa um arquivo de memória.
type MemoryEntry struct {
	FilePath    string            `json:"file_path"`
	Slug        string            `json:"slug"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	ModTime     time.Time         `json:"mod_time"`
	Frontmatter map[string]string `json:"frontmatter"`
}

// PlanFile representa um arquivo de plano.
type PlanFile struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Content string    `json:"content"`
	ModTime time.Time `json:"mod_time"`
}

// TodoFile representa um arquivo de todo.
type TodoFile struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Items   []TodoItem `json:"items"`
	ModTime time.Time  `json:"mod_time"`
}

type TodoItem struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Status      string `json:"status"`
	Priority    string `json:"priority,omitempty"`
	SessionID   string `json:"sessionId,omitempty"`
}

// Settings representa ~/.claude/settings.json.
type Settings struct {
	Env             map[string]interface{} `json:"env,omitempty"`
	Permissions     map[string]interface{} `json:"permissions,omitempty"`
	Hooks           map[string]interface{} `json:"hooks,omitempty"`
	McpServers      map[string]interface{} `json:"mcpServers,omitempty"`
	EnabledPlugins  map[string]interface{} `json:"enabledPlugins,omitempty"`
	StatusLine      map[string]interface{} `json:"statusLine,omitempty"`
	TeammateMode    string                 `json:"teammateMode,omitempty"`
	EffortLevel     string                 `json:"effortLevel,omitempty"`
}

// SkillInfo representa uma skill instalada.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
}

// PluginInfo representa um plugin instalado.
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path"`
}

// ProjectSummary agrupa informações de um projeto.
type ProjectSummary struct {
	Slug          string    `json:"slug"`
	ProjectPath   string    `json:"project_path"`
	SessionCount  int       `json:"session_count"`
	TotalMessages int       `json:"total_messages"`
	TotalTokens   int64     `json:"total_tokens"`
	TotalCost     float64   `json:"total_cost"`
	TotalDuration float64   `json:"total_duration_minutes"`
	LastActive    time.Time `json:"last_active"`
	GitBranches   []string  `json:"git_branches"`
	TopTools      []ToolCount `json:"top_tools"`
}

type ToolCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ReplayData contém dados completos de replay de sessão.
type ReplayData struct {
	Turns       []ReplayTurn      `json:"turns"`
	Compactions []CompactionEvent `json:"compactions"`
	HasMore     bool              `json:"has_more"`
	NextOffset  int               `json:"next_offset"`
	TotalCost   float64           `json:"total_cost"`
	Slug        string            `json:"slug"`
	Version     string            `json:"version"`
	GitBranch   string            `json:"git_branch"`
}

type ReplayTurn struct {
	Index       int         `json:"index"`
	Role        string      `json:"role"`
	Text        string      `json:"text"`
	Model       string      `json:"model,omitempty"`
	Tokens      TokenUsage  `json:"tokens"`
	ToolCalls   []ToolCall  `json:"tool_calls,omitempty"`
	HasThinking bool        `json:"has_thinking"`
	ThinkingText string     `json:"thinking_text,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
	UUID        string      `json:"uuid"`
	Cost        float64     `json:"cost"`
	DurationMs  int64       `json:"duration_ms,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Input    string `json:"input,omitempty"`
	Result   string `json:"result,omitempty"`
	IsError  bool   `json:"is_error"`
}

type CompactionEvent struct {
	TurnIndex int    `json:"turn_index"`
	PreTokens int64  `json:"pre_tokens,omitempty"`
	Trigger   string `json:"trigger,omitempty"`
	Summary   string `json:"summary,omitempty"`
}

// OverviewStats agrega todos os dados da overview.
type OverviewStats struct {
	TotalSessions    int              `json:"total_sessions"`
	TotalMessages    int              `json:"total_messages"`
	TotalTokens      int64            `json:"total_tokens"`
	TotalCost        float64          `json:"total_cost"`
	TotalProjects    int              `json:"total_projects"`
	StorageBytes     int64            `json:"storage_bytes"`
	DailyActivity    []DailyActivity  `json:"daily_activity"`
	HourCounts       map[string]int   `json:"hour_counts"`
	ModelBreakdown   map[string]int64 `json:"model_breakdown"`
	RecentSessions   []SessionMeta    `json:"recent_sessions"`
}

// CostAnalytics agrega dados de custo.
type CostAnalytics struct {
	TotalCost      float64                     `json:"total_cost"`
	CostByDate     []DailyCost                 `json:"cost_by_date"`
	CostByProject  []ProjectCost               `json:"cost_by_project"`
	CostByModel    map[string]ModelCostDetail  `json:"cost_by_model"`
	CacheEfficiency CacheEfficiency            `json:"cache_efficiency"`
}

type DailyCost struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
}

type ProjectCost struct {
	Slug string  `json:"slug"`
	Cost float64 `json:"cost"`
}

type ModelCostDetail struct {
	Model        string     `json:"model"`
	Tokens       TokenUsage `json:"tokens"`
	Cost         float64    `json:"cost"`
	SessionCount int        `json:"session_count"`
}

type CacheEfficiency struct {
	TotalCacheReads  int64   `json:"total_cache_reads"`
	TotalCacheWrites int64   `json:"total_cache_writes"`
	TotalInputTokens int64   `json:"total_input_tokens"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	EstimatedSavings float64 `json:"estimated_savings"`
}

// ToolAnalytics agrega dados de ferramentas.
type ToolAnalytics struct {
	ToolRanking     []ToolRankEntry          `json:"tool_ranking"`
	ToolsByCategory map[string][]ToolRankEntry `json:"tools_by_category"`
	McpServers      []McpServerInfo          `json:"mcp_servers,omitempty"`
	VersionHistory  []VersionEntry           `json:"version_history"`
	FeatureAdoption []FeatureEntry           `json:"feature_adoption"`
}

type ToolRankEntry struct {
	Name     string `json:"name"`
	Count    int    `json:"count"`
	Category string `json:"category"`
	Sessions int    `json:"sessions"`
}

type McpServerInfo struct {
	Name      string   `json:"name"`
	Tools     []string `json:"tools,omitempty"`
	UsageCount int     `json:"usage_count"`
}

type VersionEntry struct {
	Version      string    `json:"version"`
	FirstSeen    time.Time `json:"first_seen"`
	SessionCount int       `json:"session_count"`
}

type FeatureEntry struct {
	Feature      string  `json:"feature"`
	SessionCount int     `json:"session_count"`
	Percentage   float64 `json:"percentage"`
}

// ActivityData agrega dados de atividade.
type ActivityData struct {
	Heatmap       []HeatmapEntry  `json:"heatmap"`
	CurrentStreak int             `json:"current_streak"`
	LongestStreak int             `json:"longest_streak"`
	DayOfWeek     map[string]int  `json:"day_of_week"`
	PeakHours     map[string]int  `json:"peak_hours"`
	TotalDays     int             `json:"total_days"`
	ActiveDays    int             `json:"active_days"`
}

type HeatmapEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// ExportPayload para download de dados.
type ExportPayload struct {
	Version   string        `json:"version"`
	ExportedAt string       `json:"exported_at"`
	Sessions  []SessionMeta `json:"sessions"`
	Stats     *StatsCache   `json:"stats,omitempty"`
}

// WSMessage mensagem WebSocket.
type WSMessage struct {
	Type     string      `json:"type"`
	Resource string      `json:"resource"`
	Data     interface{} `json:"data"`
}
