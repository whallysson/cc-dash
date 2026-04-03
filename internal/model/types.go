package model

import "time"

// SessionMeta contains metadata extracted from a session JSONL file.
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

// TokenUsage stores token counts by type.
type TokenUsage struct {
	InputTokens      int64 `json:"input_tokens"`
	OutputTokens     int64 `json:"output_tokens"`
	CacheReadTokens  int64 `json:"cache_read_input_tokens"`
	CacheWriteTokens int64 `json:"cache_creation_input_tokens"`
}

// Total returns the sum of all tokens.
func (t TokenUsage) Total() int64 {
	return t.InputTokens + t.OutputTokens + t.CacheReadTokens + t.CacheWriteTokens
}

// Add adds another TokenUsage to this one.
func (t *TokenUsage) Add(other TokenUsage) {
	t.InputTokens += other.InputTokens
	t.OutputTokens += other.OutputTokens
	t.CacheReadTokens += other.CacheReadTokens
	t.CacheWriteTokens += other.CacheWriteTokens
}

// FileState tracks the state of a file for incremental parsing.
type FileState struct {
	Path      string    `json:"path"`
	Mtime     time.Time `json:"mtime"`
	Size      int64     `json:"size"`
	Offset    int64     `json:"offset"`
	SessionID string    `json:"session_id,omitempty"`
}

// StatsCache represents the format of ~/.claude/stats-cache.json.
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

// HistoryEntry represents a line from ~/.claude/history.jsonl.
type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
	SessionID string `json:"sessionId"`
}

// MemoryEntry represents a memory file.
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

// PlanFile represents a plan file.
type PlanFile struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Content string    `json:"content"`
	ModTime time.Time `json:"mod_time"`
}

// TodoFile represents a todo file.
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

// Settings represents ~/.claude/settings.json.
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

// SkillInfo represents an installed skill.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
}

// PluginInfo represents an installed plugin.
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path"`
}

// ProjectSummary groups information about a project.
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

// ReplayData contains complete session replay data.
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

// OverviewStats aggregates all overview data.
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

// CostAnalytics aggregates cost data.
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

// ToolAnalytics aggregates tool usage data.
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

// ActivityData aggregates activity data.
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

// ExportPayload for data download.
type ExportPayload struct {
	Version   string        `json:"version"`
	ExportedAt string       `json:"exported_at"`
	Sessions  []SessionMeta `json:"sessions"`
	Stats     *StatsCache   `json:"stats,omitempty"`
}

// WSMessage is a WebSocket message.
type WSMessage struct {
	Type     string      `json:"type"`
	Resource string      `json:"resource"`
	Data     interface{} `json:"data"`
}

// EfficiencyData aggregates token health and efficiency metrics.
type EfficiencyData struct {
	CostPerMessage   CostPerMessageStats `json:"cost_per_message"`
	ModelComparison  []ModelEfficiency    `json:"model_comparison"`
	ThinkingImpact   ThinkingImpact      `json:"thinking_impact"`
	VampireSessions  []VampireSession     `json:"vampire_sessions"`
	CacheBySession   []SessionCacheInfo   `json:"cache_by_session"`
	CostDistribution []CostBucket         `json:"cost_distribution"`
	HealthScore      int                  `json:"health_score"`
	TotalSessions    int                  `json:"total_sessions"`
	TotalCost        float64              `json:"total_cost"`
}

type CostPerMessageStats struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P90    float64 `json:"p90"`
	P99    float64 `json:"p99"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

type ModelEfficiency struct {
	Model           string  `json:"model"`
	Sessions        int     `json:"sessions"`
	TotalCost       float64 `json:"total_cost"`
	TotalMessages   int     `json:"total_messages"`
	CostPerMessage  float64 `json:"cost_per_message"`
	AvgTokensPerMsg float64 `json:"avg_tokens_per_message"`
	CacheHitRate    float64 `json:"cache_hit_rate"`
	InputTokens     int64   `json:"input_tokens"`
	OutputTokens    int64   `json:"output_tokens"`
	CacheReadTokens int64   `json:"cache_read_tokens"`
}

type ThinkingImpact struct {
	WithThinking    EfficiencyGroup `json:"with_thinking"`
	WithoutThinking EfficiencyGroup `json:"without_thinking"`
	CostMultiplier  float64         `json:"cost_multiplier"`
}

type EfficiencyGroup struct {
	Sessions       int     `json:"sessions"`
	AvgCost        float64 `json:"avg_cost"`
	AvgCostPerMsg  float64 `json:"avg_cost_per_message"`
	TotalCost      float64 `json:"total_cost"`
	AvgDuration    float64 `json:"avg_duration"`
}

type VampireSession struct {
	SessionID     string  `json:"session_id"`
	Slug          string  `json:"slug"`
	FirstPrompt   string  `json:"first_prompt"`
	Cost          float64 `json:"cost"`
	Messages      int     `json:"messages"`
	CostPerMsg    float64 `json:"cost_per_message"`
	Duration      float64 `json:"duration_minutes"`
	PrimaryModel  string  `json:"primary_model"`
	HasThinking   bool    `json:"has_thinking"`
	HasCompaction bool    `json:"has_compaction"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
	StartTime     string  `json:"start_time"`
}

type SessionCacheInfo struct {
	SessionID    string  `json:"session_id"`
	Slug         string  `json:"slug"`
	FirstPrompt  string  `json:"first_prompt"`
	CacheHitRate float64 `json:"cache_hit_rate"`
	Cost         float64 `json:"cost"`
	Messages     int     `json:"messages"`
	WastedTokens int64   `json:"wasted_tokens"`
}

type CostBucket struct {
	Label string `json:"label"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int    `json:"count"`
}
