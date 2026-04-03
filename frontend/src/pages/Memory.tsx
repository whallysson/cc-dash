import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatRelative, slugToName } from '../lib/format'
import { useState } from 'react'

export function Component() {
  const { data, loading, refetch } = useApi(() => api.memory())
  const [filter, setFilter] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  const types = [...new Set(data.map(m => m.type).filter(Boolean))]
  const filtered = filter ? data.filter(m => m.type === filter) : data

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">Memory ({data.length})</h2>
        <div className="flex gap-1.5">
          <button onClick={() => setFilter('')} className={`text-xs px-2 py-1 rounded ${!filter ? 'bg-orange-500/20 text-orange-400' : 'bg-zinc-800 text-zinc-500'}`}>All</button>
          {types.map(t => (
            <button key={t} onClick={() => setFilter(t)} className={`text-xs px-2 py-1 rounded ${filter === t ? 'bg-orange-500/20 text-orange-400' : 'bg-zinc-800 text-zinc-500'}`}>{t}</button>
          ))}
        </div>
      </div>

      <div className="space-y-2">
        {filtered.map(m => {
          const isExpanded = expanded.has(m.file_path)
          return (
            <div key={m.file_path} className="bg-zinc-900 border border-zinc-800 rounded-lg">
              <button
                onClick={() => {
                  const s = new Set(expanded)
                  isExpanded ? s.delete(m.file_path) : s.add(m.file_path)
                  setExpanded(s)
                }}
                className="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-zinc-800/30"
              >
                <div>
                  <div className="text-sm text-zinc-300">{m.name}</div>
                  <div className="text-xs text-zinc-600 mt-0.5">
                    {m.type && <span className="text-orange-500/70">[{m.type}]</span>}
                    {' '}{slugToName(m.slug)} - {formatRelative(m.mod_time)}
                  </div>
                </div>
                {m.description && <span className="text-xs text-zinc-600 max-w-xs truncate">{m.description}</span>}
              </button>
              {isExpanded && (
                <div className="border-t border-zinc-800 px-4 py-3">
                  <pre className="text-xs text-zinc-400 whitespace-pre-wrap max-h-64 overflow-y-auto">{m.content}</pre>
                  <div className="text-[10px] text-zinc-700 mt-2">{m.file_path}</div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
