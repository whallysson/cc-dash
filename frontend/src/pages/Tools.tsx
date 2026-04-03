import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatDate } from '../lib/format'

export function Component() {
  const { data, loading } = useApi(() => api.tools())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  const categoryColors: Record<string, string> = {
    'file-io': 'bg-blue-500', shell: 'bg-amber-500', agent: 'bg-purple-500',
    web: 'bg-green-500', planning: 'bg-cyan-500', skill: 'bg-pink-500',
    mcp: 'bg-emerald-500', other: 'bg-zinc-500', 'code-intelligence': 'bg-indigo-500',
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold">Tools</h2>

      {/* Tool Ranking */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3">Tool Ranking (top 20)</h3>
        <div className="space-y-2">
          {data.tool_ranking.slice(0, 20).map((t, i) => {
            const maxCount = data.tool_ranking[0]?.count || 1
            const pct = (t.count / maxCount) * 100
            return (
              <div key={t.name} className="flex items-center gap-2">
                <span className="w-5 text-xs text-zinc-600 text-right">{i + 1}</span>
                <span className="w-28 text-sm text-zinc-300 truncate">{t.name}</span>
                <div className="flex-1 h-2 bg-zinc-800 rounded-full overflow-hidden">
                  <div className={`h-full rounded-full ${categoryColors[t.category] || 'bg-zinc-500'}`} style={{ width: `${pct}%` }} />
                </div>
                <span className="w-16 text-right text-xs text-zinc-400">{t.count.toLocaleString()}</span>
                <span className={`text-[10px] px-1.5 py-0.5 rounded ${categoryColors[t.category] || 'bg-zinc-700'} bg-opacity-20 text-zinc-400`}>
                  {t.category}
                </span>
              </div>
            )
          })}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Feature Adoption */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Feature Adoption</h3>
          <div className="space-y-3">
            {data.feature_adoption.map(f => (
              <div key={f.feature}>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-zinc-400">{f.feature}</span>
                  <span className="text-zinc-300">{f.session_count} ({f.percentage.toFixed(0)}%)</span>
                </div>
                <div className="h-2 bg-zinc-800 rounded-full overflow-hidden">
                  <div className="h-full bg-orange-500/70 rounded-full" style={{ width: `${f.percentage}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Version History */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Version History</h3>
          <div className="space-y-1.5 max-h-64 overflow-y-auto">
            {data.version_history.map(v => (
              <div key={v.version} className="flex items-center justify-between text-sm py-1">
                <span className="text-zinc-300 font-mono">{v.version}</span>
                <span className="text-zinc-600 text-xs">{formatDate(v.first_seen)} - {v.session_count} sessions</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
