import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatCost, formatDuration, formatRelative, formatTokens, slugToName, totalTokens } from '../lib/format'
import { Search, Zap, Bot, Globe, Brain } from 'lucide-react'

export function Component() {
  const [page, setPage] = useState(1)
  const [sort, setSort] = useState('date')
  const [query, setQuery] = useState('')
  const [search, setSearch] = useState('')

  const { data, loading } = useApi(
    () => api.sessions(page, 50, sort, search),
    [page, sort, search]
  )

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setSearch(query)
    setPage(1)
  }

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  const { sessions, total } = data
  const totalPages = Math.ceil(total / 50)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">Sessions ({total})</h2>
        <div className="flex gap-2">
          {['date', 'cost', 'messages', 'duration', 'tokens'].map(s => (
            <button
              key={s}
              onClick={() => { setSort(s); setPage(1) }}
              className={`text-xs px-2.5 py-1 rounded ${sort === s ? 'bg-orange-500/20 text-orange-400' : 'bg-zinc-800 text-zinc-500 hover:text-zinc-300'}`}
            >
              {s}
            </button>
          ))}
        </div>
      </div>

      {/* Search */}
      <form onSubmit={handleSearch} className="relative">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
        <input
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search sessions..."
          className="w-full bg-zinc-900 border border-zinc-800 rounded-lg pl-9 pr-4 py-2 text-sm text-zinc-200 placeholder:text-zinc-600 focus:outline-none focus:border-orange-500/50"
        />
      </form>

      {/* Table */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-500 text-xs">
              <th className="text-left p-3 font-medium">Prompt</th>
              <th className="text-left p-3 font-medium">Project</th>
              <th className="text-right p-3 font-medium">Msgs</th>
              <th className="text-right p-3 font-medium">Tokens</th>
              <th className="text-right p-3 font-medium">Cost</th>
              <th className="text-right p-3 font-medium">Duration</th>
              <th className="text-center p-3 font-medium">Features</th>
              <th className="text-right p-3 font-medium">When</th>
            </tr>
          </thead>
          <tbody>
            {sessions?.map(s => (
              <tr key={s.session_id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="p-3 max-w-xs">
                  <Link
                    to={`/sessions/${s.session_id}`}
                    className="text-zinc-300 hover:text-orange-400 truncate block"
                  >
                    {s.first_prompt?.slice(0, 80) || 'No prompt'}
                  </Link>
                </td>
                <td className="p-3 text-zinc-500 text-xs">{slugToName(s.slug)}</td>
                <td className="p-3 text-right text-zinc-400">{s.total_message_count}</td>
                <td className="p-3 text-right text-zinc-400">{formatTokens(totalTokens(s.total_tokens))}</td>
                <td className="p-3 text-right text-zinc-400">{formatCost(s.estimated_cost)}</td>
                <td className="p-3 text-right text-zinc-500">{formatDuration(s.duration_minutes)}</td>
                <td className="p-3 text-center">
                  <div className="flex gap-1 justify-center">
                    {s.has_compaction && <Zap size={12} className="text-yellow-500" title="Compaction" />}
                    {s.uses_task_agent && <Bot size={12} className="text-blue-400" title="Agent" />}
                    {s.uses_mcp && <Globe size={12} className="text-green-400" title="MCP" />}
                    {s.has_thinking && <Brain size={12} className="text-purple-400" title="Thinking" />}
                  </div>
                </td>
                <td className="p-3 text-right text-zinc-600 text-xs">{formatRelative(s.start_time)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-3 py-1 text-sm bg-zinc-800 text-zinc-400 rounded disabled:opacity-30"
          >
            Prev
          </button>
          <span className="text-sm text-zinc-500">{page} / {totalPages}</span>
          <button
            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="px-3 py-1 text-sm bg-zinc-800 text-zinc-400 rounded disabled:opacity-30"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
