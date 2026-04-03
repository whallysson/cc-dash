import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { CheckCircle2, Circle, Loader2 } from 'lucide-react'

export function Component() {
  const { data, loading } = useApi(() => api.todos())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  const allItems = data.flatMap(t => t.items.map(item => ({ ...item, file: t.name })))
  const statusIcon = (s: string) => {
    if (s === 'completed') return <CheckCircle2 size={14} className="text-green-500" />
    if (s === 'in_progress') return <Loader2 size={14} className="text-orange-400 animate-spin" />
    return <Circle size={14} className="text-zinc-600" />
  }

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">Todos ({allItems.length})</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg divide-y divide-zinc-800/50">
        {allItems.map((item, i) => (
          <div key={i} className="flex items-start gap-3 px-4 py-3">
            <div className="mt-0.5">{statusIcon(item.status)}</div>
            <div className="flex-1">
              <div className={`text-sm ${item.status === 'completed' ? 'text-zinc-600 line-through' : 'text-zinc-300'}`}>
                {item.content}
              </div>
              {item.priority && <span className="text-[10px] text-orange-500/70 mt-0.5">{item.priority}</span>}
            </div>
          </div>
        ))}
        {allItems.length === 0 && <div className="px-4 py-6 text-center text-zinc-600 text-sm">No todos found</div>}
      </div>
    </div>
  )
}
