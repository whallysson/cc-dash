import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { CheckCircle2, Circle, Loader2 } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

function StatusIcon({ status }: { status: string }) {
  if (status === 'completed') return <CheckCircle2 className="size-4 text-chart-2" />
  if (status === 'in_progress') return <Loader2 className="size-4 text-primary animate-spin" />
  return <Circle className="size-4 text-muted-foreground/30" />
}

export function Component() {
  const { data, loading } = useApi(() => api.todos())

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-[300px] w-full rounded-xl" />
      </div>
    )
  }

  const allItems = data.flatMap(t => t.items.map(item => ({ ...item, file: t.name })))
  const completed = allItems.filter(i => i.status === 'completed').length
  const total = allItems.length
  const pct = total > 0 ? (completed / total) * 100 : 0

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Todos</h2>
          <p className="text-sm text-muted-foreground">{total} items</p>
        </div>
        {total > 0 && (
          <div className="flex items-center gap-2">
            <div className="w-24 h-1.5 rounded-full bg-muted overflow-hidden">
              <div
                className="h-full rounded-full bg-chart-2 transition-all duration-300"
                style={{ width: `${pct}%` }}
              />
            </div>
            <span className="text-xs text-muted-foreground tabular-nums">{completed}/{total}</span>
          </div>
        )}
      </div>

      <Card className="divide-y divide-border">
        {allItems.map((item, i) => (
          <div key={i} className="flex items-start gap-3 px-4 py-2.5 hover:bg-accent transition-colors">
            <div className="mt-0.5 shrink-0">
              <StatusIcon status={item.status} />
            </div>
            <div className="flex-1 min-w-0">
              <p className={`text-sm ${item.status === 'completed' ? 'text-muted-foreground line-through' : ''}`}>
                {item.content}
              </p>
              <div className="flex items-center gap-2 mt-1">
                {item.priority && (
                  <Badge variant={item.priority === 'high' ? 'destructive' : 'outline'}>
                    {item.priority}
                  </Badge>
                )}
                <span className="text-[10px] text-muted-foreground/40">{item.file}</span>
              </div>
            </div>
          </div>
        ))}
        {allItems.length === 0 && (
          <div className="px-4 py-8 text-center text-muted-foreground text-sm">No todos found</div>
        )}
      </Card>
    </div>
  )
}
