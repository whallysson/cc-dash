import { useParams, Link } from 'react-router-dom'
import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatCost, formatDuration, formatRelative, formatTokens, slugToName, totalTokens } from '../lib/format'

export function Component() {
  const { slug } = useParams<{ slug: string }>()
  const { data, loading } = useApi(() => api.project(slug!), [slug])

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  const { project: p, sessions } = data

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-xl font-bold">{slugToName(p.slug)}</h2>
        <div className="text-xs text-zinc-500 mt-1">{p.project_path}</div>
      </div>

      <div className="grid grid-cols-4 gap-3">
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
          <div className="text-xs text-zinc-500">Sessions</div>
          <div className="text-xl font-bold text-zinc-200">{p.session_count}</div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
          <div className="text-xs text-zinc-500">Cost</div>
          <div className="text-xl font-bold text-zinc-200">{formatCost(p.total_cost)}</div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
          <div className="text-xs text-zinc-500">Time</div>
          <div className="text-xl font-bold text-zinc-200">{formatDuration(p.total_duration_minutes)}</div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
          <div className="text-xs text-zinc-500">Tokens</div>
          <div className="text-xl font-bold text-zinc-200">{formatTokens(p.total_tokens)}</div>
        </div>
      </div>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-500 text-xs">
              <th className="text-left p-3">Prompt</th>
              <th className="text-right p-3">Messages</th>
              <th className="text-right p-3">Cost</th>
              <th className="text-right p-3">Duration</th>
              <th className="text-right p-3">When</th>
            </tr>
          </thead>
          <tbody>
            {sessions?.map(s => (
              <tr key={s.session_id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="p-3">
                  <Link to={`/sessions/${s.session_id}`} className="text-zinc-300 hover:text-orange-400 truncate block max-w-md">
                    {s.first_prompt?.slice(0, 80) || 'No prompt'}
                  </Link>
                </td>
                <td className="p-3 text-right text-zinc-400">{s.total_message_count}</td>
                <td className="p-3 text-right text-zinc-400">{formatCost(s.estimated_cost)}</td>
                <td className="p-3 text-right text-zinc-500">{formatDuration(s.duration_minutes)}</td>
                <td className="p-3 text-right text-zinc-600 text-xs">{formatRelative(s.start_time)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
