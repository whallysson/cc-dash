import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatRelative } from '../lib/format'
import { useState } from 'react'

export function Component() {
  const { data, loading } = useApi(() => api.plans())
  const [expanded, setExpanded] = useState<string | null>(null)

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">Plans ({data.length})</h2>
      <div className="space-y-2">
        {data.map(p => (
          <div key={p.path} className="bg-zinc-900 border border-zinc-800 rounded-lg">
            <button
              onClick={() => setExpanded(expanded === p.path ? null : p.path)}
              className="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-zinc-800/30"
            >
              <span className="text-sm text-zinc-300">{p.name}</span>
              <span className="text-xs text-zinc-600">{formatRelative(p.mod_time)}</span>
            </button>
            {expanded === p.path && (
              <div className="border-t border-zinc-800 px-4 py-3">
                <pre className="text-xs text-zinc-400 whitespace-pre-wrap max-h-96 overflow-y-auto">{p.content}</pre>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
