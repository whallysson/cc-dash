export interface TokenUsage {
  input_tokens: number
  output_tokens: number
  cache_read_input_tokens: number
  cache_creation_input_tokens: number
}

export interface SessionMeta {
  session_id: string
  slug: string
  project_path: string
  start_time: string
  end_time: string
  duration_minutes: number
  user_message_count: number
  assistant_message_count: number
  total_message_count: number
  tool_counts: Record<string, number>
  model_tokens: Record<string, TokenUsage>
  total_tokens: TokenUsage
  estimated_cost: number
  first_prompt: string
  git_branch: string
  version: string
  entrypoint: string
  has_compaction: boolean
  has_thinking: boolean
  uses_mcp: boolean
  uses_task_agent: boolean
  uses_web_search: boolean
  uses_web_fetch: boolean
}

export interface ProjectSummary {
  slug: string
  project_path: string
  session_count: number
  total_messages: number
  total_tokens: number
  total_cost: number
  total_duration_minutes: number
  last_active: string
  git_branches: string[]
  top_tools: { name: string; count: number }[]
}

export interface OverviewStats {
  total_sessions: number
  total_messages: number
  total_tokens: number
  total_cost: number
  total_projects: number
  storage_bytes: number
  daily_activity: { date: string; sessions: number; messages: number; tokens: number }[]
  hour_counts: Record<string, number>
  model_breakdown: Record<string, number>
  recent_sessions: SessionMeta[]
}

export interface CostAnalytics {
  total_cost: number
  cost_by_date: { date: string; cost: number }[]
  cost_by_project: { slug: string; cost: number }[]
  cost_by_model: Record<string, {
    model: string
    tokens: TokenUsage
    cost: number
    session_count: number
  }>
  cache_efficiency: {
    total_cache_reads: number
    total_cache_writes: number
    total_input_tokens: number
    cache_hit_rate: number
    estimated_savings: number
  }
}

export interface ToolAnalytics {
  tool_ranking: { name: string; count: number; category: string; sessions: number }[]
  tools_by_category: Record<string, { name: string; count: number; category: string; sessions: number }[]>
  version_history: { version: string; first_seen: string; session_count: number }[]
  feature_adoption: { feature: string; session_count: number; percentage: number }[]
}

export interface ActivityData {
  heatmap: { date: string; count: number }[]
  current_streak: number
  longest_streak: number
  day_of_week: Record<string, number>
  peak_hours: Record<string, number>
  total_days: number
  active_days: number
}

export interface ReplayData {
  turns: ReplayTurn[]
  compactions: { turn_index: number; summary?: string }[]
  has_more: boolean
  next_offset: number
  total_cost: number
  slug: string
  version: string
  git_branch: string
}

export interface ReplayTurn {
  index: number
  role: 'user' | 'assistant'
  text: string
  model?: string
  tokens: TokenUsage
  tool_calls?: { id: string; name: string; input?: string; result?: string; is_error: boolean }[]
  has_thinking: boolean
  thinking_text?: string
  timestamp: string
  uuid: string
  cost: number
  duration_ms?: number
}

export interface HistoryEntry {
  display: string
  timestamp: number
  project: string
  sessionId: string
}

export interface MemoryEntry {
  file_path: string
  slug: string
  name: string
  description: string
  type: string
  content: string
  mod_time: string
  frontmatter: Record<string, string>
}

export interface PlanFile {
  name: string
  path: string
  content: string
  mod_time: string
}

export interface TodoFile {
  name: string
  path: string
  items: { id: string; content: string; status: string; priority?: string }[]
  mod_time: string
}

export interface EfficiencyData {
  cost_per_message: {
    mean: number
    median: number
    p90: number
    p99: number
    min: number
    max: number
  }
  model_comparison: {
    model: string
    sessions: number
    total_cost: number
    total_messages: number
    cost_per_message: number
    avg_tokens_per_message: number
    cache_hit_rate: number
    input_tokens: number
    output_tokens: number
    cache_read_tokens: number
  }[]
  thinking_impact: {
    with_thinking: { sessions: number; avg_cost: number; avg_cost_per_message: number; total_cost: number; avg_duration: number }
    without_thinking: { sessions: number; avg_cost: number; avg_cost_per_message: number; total_cost: number; avg_duration: number }
    cost_multiplier: number
  }
  vampire_sessions: {
    session_id: string
    slug: string
    first_prompt: string
    cost: number
    messages: number
    cost_per_message: number
    duration_minutes: number
    primary_model: string
    has_thinking: boolean
    has_compaction: boolean
    cache_hit_rate: number
    start_time: string
  }[]
  cache_by_session: {
    session_id: string
    slug: string
    first_prompt: string
    cache_hit_rate: number
    cost: number
    messages: number
    wasted_tokens: number
  }[]
  cost_distribution: { label: string; count: number }[]
  health_score: number
  total_sessions: number
  total_cost: number
}
