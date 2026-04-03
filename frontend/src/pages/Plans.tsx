import { useState } from 'react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatRelative } from '@/lib/format'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronDown, FileText } from 'lucide-react'

export function Component() {
  const { data, loading } = useApi(() => api.plans())
  const [expanded, setExpanded] = useState<string | null>(null)

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-14 rounded-xl" />)}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Plans</h2>
        <p className="text-sm text-muted-foreground">{data.length} plan files</p>
      </div>

      <div className="flex flex-col gap-2">
        {data.map(p => {
          const isExpanded = expanded === p.path
          return (
            <Card key={p.path}>
              <button
                onClick={() => setExpanded(isExpanded ? null : p.path)}
                className="w-full flex items-center justify-between px-4 py-2.5 text-left hover:bg-accent transition-colors rounded-xl"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <FileText className="size-4 text-muted-foreground shrink-0" />
                  <span className="text-sm font-medium truncate">{p.name}</span>
                </div>
                <div className="flex items-center gap-2 shrink-0 ml-2">
                  <span className="text-xs text-muted-foreground tabular-nums">{formatRelative(p.mod_time)}</span>
                  <ChevronDown className={`size-4 text-muted-foreground transition-transform ${isExpanded ? 'rotate-180' : ''}`} />
                </div>
              </button>
              {isExpanded && (
                <div className="border-t px-4 py-3">
                  <pre className="text-xs text-muted-foreground whitespace-pre-wrap max-h-[500px] overflow-y-auto leading-relaxed font-mono">
                    {p.content}
                  </pre>
                </div>
              )}
            </Card>
          )
        })}
        {data.length === 0 && (
          <Card className="px-4 py-8 text-center text-muted-foreground text-sm">No plans found</Card>
        )}
      </div>
    </div>
  )
}
