import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatCost, formatTokens, slugToName } from '../lib/format'

export function Component() {
  const { data, loading } = useApi(() => api.costs())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">Costs</h2>
        <div className="text-2xl font-bold text-orange-400">{formatCost(data.total_cost)}</div>
      </div>

      {/* Cost over time */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3">Daily Cost</h3>
        <div className="h-40 flex items-end gap-[2px]">
          {data.cost_by_date.slice(-30).map(d => {
            const max = Math.max(...data.cost_by_date.slice(-30).map(x => x.cost), 0.01)
            const h = (d.cost / max) * 100
            return (
              <div key={d.date} className="flex-1 group relative">
                <div className="bg-green-500/70 hover:bg-green-400 rounded-t transition-colors" style={{ height: `${Math.max(h, 2)}%` }} />
                <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 bg-zinc-800 text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap z-10">
                  {d.date}: {formatCost(d.cost)}
                </div>
              </div>
            )
          })}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Cost by Model */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Cost by Model</h3>
          <div className="space-y-3">
            {Object.values(data.cost_by_model)
              .sort((a, b) => b.cost - a.cost)
              .map(m => (
                <div key={m.model} className="flex items-center justify-between text-sm">
                  <div>
                    <div className="text-zinc-300">{m.model}</div>
                    <div className="text-[10px] text-zinc-600">
                      {m.session_count} sessions | {formatTokens(m.tokens.input_tokens + m.tokens.output_tokens)} tokens
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-zinc-200 font-medium">{formatCost(m.cost)}</div>
                    <div className="text-[10px] text-zinc-600">
                      {data.total_cost > 0 ? ((m.cost / data.total_cost) * 100).toFixed(1) : 0}%
                    </div>
                  </div>
                </div>
              ))}
          </div>
        </div>

        {/* Cache Efficiency */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Cache Efficiency</h3>
          <div className="space-y-4">
            <div>
              <div className="flex justify-between text-sm mb-1">
                <span className="text-zinc-400">Cache Hit Rate</span>
                <span className="text-zinc-200 font-medium">{data.cache_efficiency.cache_hit_rate.toFixed(1)}%</span>
              </div>
              <div className="h-3 bg-zinc-800 rounded-full overflow-hidden">
                <div className="h-full bg-emerald-500 rounded-full" style={{ width: `${data.cache_efficiency.cache_hit_rate}%` }} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3 text-xs">
              <div>
                <div className="text-zinc-500">Cache Reads</div>
                <div className="text-zinc-300 font-medium">{formatTokens(data.cache_efficiency.total_cache_reads)}</div>
              </div>
              <div>
                <div className="text-zinc-500">Cache Writes</div>
                <div className="text-zinc-300 font-medium">{formatTokens(data.cache_efficiency.total_cache_writes)}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Cost by Project */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3">Cost by Project</h3>
        <div className="space-y-2">
          {data.cost_by_project.slice(0, 15).map(p => {
            const pct = data.total_cost > 0 ? (p.cost / data.total_cost) * 100 : 0
            return (
              <div key={p.slug} className="flex items-center gap-3">
                <div className="w-40 text-sm text-zinc-400 truncate">{slugToName(p.slug)}</div>
                <div className="flex-1 h-2 bg-zinc-800 rounded-full overflow-hidden">
                  <div className="h-full bg-orange-500/70 rounded-full" style={{ width: `${pct}%` }} />
                </div>
                <div className="w-20 text-right text-sm text-zinc-300">{formatCost(p.cost)}</div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
