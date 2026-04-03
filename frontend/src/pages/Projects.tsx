import { Link } from 'react-router-dom'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatTokens, formatDuration, formatRelative, slugToName } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

function ProjectsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-1">
        <Skeleton className="h-7 w-32" />
        <Skeleton className="h-4 w-40" />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-48 rounded-xl" />
        ))}
      </div>
    </div>
  )
}

export function Component() {
  const { data, loading } = useApi(() => api.projects())

  if (loading || !data) return <ProjectsSkeleton />

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Projects</h2>
        <p className="text-sm text-muted-foreground">
          {data.length} projects tracked
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {data.map(p => (
          <Link key={p.slug} to={`/projects/${p.slug}`} className="group">
            <Card className="h-full transition-all hover:border-primary/50 hover:shadow-sm">
              <CardHeader className="pb-2">
                <CardTitle className="group-hover:text-primary transition-colors">
                  {slugToName(p.slug)}
                </CardTitle>
                <CardDescription className="truncate text-xs">
                  {p.project_path}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <p className="text-xs text-muted-foreground">Sessions</p>
                    <p className="text-lg font-bold tabular-nums">{p.session_count}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Cost</p>
                    <p className="text-lg font-bold tabular-nums">{formatCost(p.total_cost)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Time</p>
                    <p className="text-lg font-bold tabular-nums">{formatDuration(p.total_duration_minutes)}</p>
                  </div>
                </div>
                <div className="flex items-center justify-between mt-4 pt-3 border-t border-border/50 text-xs text-muted-foreground">
                  <span className="tabular-nums">{formatTokens(p.total_tokens)} tokens</span>
                  <span>{formatRelative(p.last_active)}</span>
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>

      {data.length === 0 && (
        <div className="flex items-center justify-center h-48 text-muted-foreground text-sm">
          No projects found.
        </div>
      )}
    </div>
  )
}
