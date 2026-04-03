import { useState } from 'react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { Search, Terminal } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export function Component() {
  const [query, setQuery] = useState('')
  const [search, setSearch] = useState('')
  const { data, loading } = useApi(() => api.history(500, search), [search])

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-10 w-full rounded-lg" />
        <Skeleton className="h-[400px] w-full rounded-xl" />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">History</h2>
        <p className="text-sm text-muted-foreground">{data.length} commands</p>
      </div>

      <form onSubmit={e => { e.preventDefault(); setSearch(query) }} className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <input
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search commands..."
          className="w-full rounded-lg border border-border bg-background pl-9 pr-4 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors"
        />
      </form>

      <Card className="divide-y divide-border">
        {data.map((entry, i) => (
          <div key={i} className="flex items-start gap-3 px-4 py-2.5 hover:bg-accent transition-colors">
            <Terminal className="size-3.5 text-muted-foreground/50 shrink-0 mt-1" />
            <span className="text-[10px] text-muted-foreground shrink-0 mt-0.5 tabular-nums">
              {new Date(entry.timestamp).toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}
            </span>
            <span className="text-sm break-all font-mono">{entry.display}</span>
          </div>
        ))}
        {data.length === 0 && (
          <div className="px-4 py-8 text-center text-muted-foreground text-sm">No commands found</div>
        )}
      </Card>
    </div>
  )
}
