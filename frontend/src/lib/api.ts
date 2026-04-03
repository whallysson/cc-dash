import type { OverviewStats, SessionMeta, ProjectSummary, CostAnalytics, ToolAnalytics, ActivityData, ReplayData, HistoryEntry, MemoryEntry, PlanFile, TodoFile } from './types'

const BASE = ''

async function get<T>(url: string): Promise<T> {
  const res = await fetch(`${BASE}${url}`)
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json() as Promise<T>
}

export const api = {
  stats: () => get<OverviewStats>('/api/stats'),

  sessions: (page = 1, limit = 50, sort = 'date', q = '') =>
    get<{ sessions: SessionMeta[]; total: number; page: number }>
      (`/api/sessions?page=${page}&limit=${limit}&sort=${sort}&q=${encodeURIComponent(q)}`),

  session: (id: string) => get<SessionMeta>(`/api/sessions/${id}`),

  replay: (id: string, offset = 0, limit = 0) =>
    get<ReplayData>(`/api/sessions/${id}/replay?offset=${offset}&limit=${limit}`),

  projects: () => get<ProjectSummary[]>('/api/projects'),

  project: (slug: string) =>
    get<{ project: ProjectSummary; sessions: SessionMeta[] }>(`/api/projects/${slug}`),

  costs: () => get<CostAnalytics>('/api/costs'),
  tools: () => get<ToolAnalytics>('/api/tools'),
  activity: () => get<ActivityData>('/api/activity'),

  history: (limit = 200, q = '') =>
    get<HistoryEntry[]>(`/api/history?limit=${limit}&q=${encodeURIComponent(q)}`),

  memory: () => get<MemoryEntry[]>('/api/memory'),
  plans: () => get<PlanFile[]>('/api/plans'),
  todos: () => get<TodoFile[]>('/api/todos'),

  settings: () => get<{ settings: Record<string, unknown>; plugins: unknown[] }>('/api/settings'),
}
