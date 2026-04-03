import { useState } from 'react'
import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { Search } from 'lucide-react'

export function Component() {
  const [query, setQuery] = useState('')
  const [search, setSearch] = useState('')
  const { data, loading } = useApi(() => api.history(500, search), [search])

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">History ({data.length})</h2>
      <form onSubmit={e => { e.preventDefault(); setSearch(query) }} className="relative">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
        <input
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search commands..."
          className="w-full bg-zinc-900 border border-zinc-800 rounded-lg pl-9 pr-4 py-2 text-sm text-zinc-200 placeholder:text-zinc-600 focus:outline-none focus:border-orange-500/50"
        />
      </form>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg divide-y divide-zinc-800/50">
        {data.map((e, i) => (
          <div key={i} className="flex items-start gap-3 px-4 py-2.5">
            <span className="text-[10px] text-zinc-600 shrink-0 mt-0.5 font-mono">
              {new Date(e.timestamp).toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}
            </span>
            <span className="text-sm text-zinc-300 break-all">{e.display}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
