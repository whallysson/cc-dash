import { useState } from 'react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatRelative, slugToName } from '@/lib/format'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronDown, Brain } from 'lucide-react'

export function Component() {
  const { data, loading } = useApi(() => api.memory())
  const [filter, setFilter] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        <div className="flex gap-1.5">
          {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-8 w-16 rounded-md" />)}
        </div>
        <Skeleton className="h-[400px] w-full rounded-xl" />
      </div>
    )
  }

  const types = [...new Set(data.map(m => m.type).filter(Boolean))]
  const filtered = filter ? data.filter(m => m.type === filter) : data

  function toggleExpand(path: string) {
    const next = new Set(expanded)
    next.has(path) ? next.delete(path) : next.add(path)
    setExpanded(next)
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Memory</h2>
          <p className="text-sm text-muted-foreground">{data.length} memory entries</p>
        </div>
      </div>

      <div className="flex flex-wrap gap-1.5">
        <Button variant={!filter ? 'default' : 'outline'} size="sm" onClick={() => setFilter('')}>All</Button>
        {types.map(t => (
          <Button key={t} variant={filter === t ? 'default' : 'outline'} size="sm" onClick={() => setFilter(t)}>{t}</Button>
        ))}
      </div>

      <div className="flex flex-col gap-2">
        {filtered.map(m => {
          const isExpanded = expanded.has(m.file_path)
          return (
            <Card key={m.file_path}>
              <button
                onClick={() => toggleExpand(m.file_path)}
                className="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-muted/30 transition-colors rounded-xl"
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <Brain className="size-4 text-muted-foreground shrink-0" />
                    <span className="text-sm font-medium truncate">{m.name}</span>
                    {m.type && <Badge variant="secondary">{m.type}</Badge>}
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5 truncate pl-6">
                    {slugToName(m.slug)} - {formatRelative(m.mod_time)}
                  </p>
                </div>
                <ChevronDown className={`size-4 text-muted-foreground transition-transform shrink-0 ml-2 ${isExpanded ? 'rotate-180' : ''}`} />
              </button>
              {isExpanded && (
                <div className="border-t px-4 py-3">
                  {m.description && (
                    <p className="text-xs text-muted-foreground mb-2">{m.description}</p>
                  )}
                  <pre className="text-xs text-muted-foreground whitespace-pre-wrap max-h-64 overflow-y-auto leading-relaxed font-mono">{m.content}</pre>
                  <p className="text-[10px] text-muted-foreground/40 mt-3 font-mono">{m.file_path}</p>
                </div>
              )}
            </Card>
          )
        })}
        {filtered.length === 0 && (
          <Card className="px-4 py-8 text-center text-muted-foreground text-sm">No memory entries found</Card>
        )}
      </div>
    </div>
  )
}
