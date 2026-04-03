import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatCost, formatTokens, formatDuration, formatRelative, slugToName } from '../lib/format'

export function Component() {
  const { data, loading } = useApi(() => api.projects())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">Projects ({data.length})</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {data.map(p => (
          <Link
            key={p.slug}
            to={`/projects/${p.slug}`}
            className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 hover:border-orange-500/30 transition-colors"
          >
            <div className="text-sm font-medium text-zinc-200 truncate">{slugToName(p.slug)}</div>
            <div className="text-[10px] text-zinc-600 truncate mt-0.5">{p.project_path}</div>
            <div className="grid grid-cols-3 gap-2 mt-3 text-xs">
              <div>
                <div className="text-zinc-500">Sessions</div>
                <div className="text-zinc-300 font-medium">{p.session_count}</div>
              </div>
              <div>
                <div className="text-zinc-500">Cost</div>
                <div className="text-zinc-300 font-medium">{formatCost(p.total_cost)}</div>
              </div>
              <div>
                <div className="text-zinc-500">Time</div>
                <div className="text-zinc-300 font-medium">{formatDuration(p.total_duration_minutes)}</div>
              </div>
            </div>
            <div className="flex items-center justify-between mt-3 text-[10px] text-zinc-600">
              <span>{formatTokens(p.total_tokens)} tokens</span>
              <span>{formatRelative(p.last_active)}</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
