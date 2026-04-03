import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatTokens, formatCost, formatBytes, formatRelative, slugToName, totalTokens } from '../lib/format'
import { Link } from 'react-router-dom'

function StatCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
      <div className="text-xs text-zinc-500 uppercase tracking-wide">{label}</div>
      <div className="text-2xl font-bold text-zinc-100 mt-1">{value}</div>
      {sub && <div className="text-xs text-zinc-500 mt-1">{sub}</div>}
    </div>
  )
}

export function Component() {
  const { data, loading } = useApi(() => api.stats())

  if (loading || !data) {
    return <div className="text-zinc-500">Carregando...</div>
  }

  const d = data
  const tokenTotal = d.total_tokens

  // Modelo dominante
  const topModel = Object.entries(d.model_breakdown)
    .sort(([, a], [, b]) => b - a)[0]

  // Horas de pico
  const peakHour = Object.entries(d.hour_counts)
    .sort(([, a], [, b]) => b - a)[0]

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold">Overview</h2>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        <StatCard label="Sessions" value={String(d.total_sessions)} />
        <StatCard label="Messages" value={formatTokens(d.total_messages)} />
        <StatCard label="Tokens" value={formatTokens(tokenTotal)} />
        <StatCard label="Cost" value={formatCost(d.total_cost)} />
        <StatCard label="Projects" value={String(d.total_projects)} />
        <StatCard label="Storage" value={formatBytes(d.storage_bytes)} />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Daily Activity */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Daily Activity</h3>
          <div className="h-48 flex items-end gap-[2px]">
            {d.daily_activity.slice(-30).map((day) => {
              const maxSessions = Math.max(...d.daily_activity.slice(-30).map(d => d.sessions))
              const height = maxSessions > 0 ? (day.sessions / maxSessions) * 100 : 0
              return (
                <div key={day.date} className="flex-1 group relative">
                  <div
                    className="bg-orange-500/80 hover:bg-orange-400 rounded-t transition-colors"
                    style={{ height: `${Math.max(height, 2)}%` }}
                  />
                  <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 bg-zinc-800 text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap z-10">
                    {day.date}: {day.sessions} sessions
                  </div>
                </div>
              )
            })}
          </div>
          <div className="flex justify-between text-[10px] text-zinc-600 mt-1">
            <span>{d.daily_activity.slice(-30)[0]?.date}</span>
            <span>{d.daily_activity.slice(-1)[0]?.date}</span>
          </div>
        </div>

        {/* Model Breakdown */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Model Usage</h3>
          <div className="space-y-3">
            {Object.entries(d.model_breakdown)
              .filter(([, v]) => v > 0)
              .sort(([, a], [, b]) => b - a)
              .map(([model, tokens]) => {
                const pct = tokenTotal > 0 ? (tokens / tokenTotal) * 100 : 0
                return (
                  <div key={model}>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-zinc-300">{model}</span>
                      <span className="text-zinc-500">{formatTokens(tokens)} ({pct.toFixed(1)}%)</span>
                    </div>
                    <div className="h-2 bg-zinc-800 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-orange-500 rounded-full"
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                  </div>
                )
              })}
          </div>
          {topModel && (
            <div className="mt-3 text-xs text-zinc-500">
              Primary: {topModel[0]} | Peak hour: {peakHour?.[0]}:00
            </div>
          )}
        </div>
      </div>

      {/* Peak Hours */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3">Peak Hours</h3>
        <div className="flex items-end gap-1 h-24">
          {Array.from({ length: 24 }, (_, h) => {
            const key = String(h).padStart(2, '0')
            const count = d.hour_counts[key] || d.hour_counts[String(h)] || 0
            const max = Math.max(...Object.values(d.hour_counts), 1)
            const height = (count / max) * 100
            return (
              <div key={h} className="flex-1 group relative">
                <div
                  className="bg-orange-500/60 hover:bg-orange-400 rounded-t transition-colors"
                  style={{ height: `${Math.max(height, 2)}%` }}
                />
                <div className="absolute bottom-full mb-1 left-1/2 -translate-x-1/2 bg-zinc-800 text-[10px] px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap z-10">
                  {key}:00 - {count}
                </div>
              </div>
            )
          })}
        </div>
        <div className="flex justify-between text-[10px] text-zinc-600 mt-1">
          <span>00</span><span>06</span><span>12</span><span>18</span><span>23</span>
        </div>
      </div>

      {/* Recent Sessions */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3">Recent Sessions</h3>
        <div className="space-y-2">
          {d.recent_sessions.map((s) => (
            <Link
              key={s.session_id}
              to={`/sessions/${s.session_id}`}
              className="flex items-center justify-between p-2 rounded hover:bg-zinc-800/50 transition-colors group"
            >
              <div className="flex-1 min-w-0">
                <div className="text-sm text-zinc-300 truncate group-hover:text-orange-400">
                  {s.first_prompt || 'No prompt'}
                </div>
                <div className="text-xs text-zinc-600 mt-0.5">
                  {slugToName(s.slug)} - {formatRelative(s.start_time)}
                </div>
              </div>
              <div className="text-right ml-4 shrink-0">
                <div className="text-xs text-zinc-400">{s.total_message_count} msgs</div>
                <div className="text-xs text-zinc-600">{formatTokens(totalTokens(s.total_tokens))}</div>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}
