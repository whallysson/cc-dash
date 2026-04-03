import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search } from 'lucide-react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatTokens, formatDuration, formatRelative, slugToName } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

const TOOL_COLORS: Record<string, string> = {
  Bash: 'bg-amber-500',
  Read: 'bg-blue-500',
  Edit: 'bg-blue-400',
  Write: 'bg-blue-300',
  Glob: 'bg-indigo-400',
  Grep: 'bg-indigo-500',
  Agent: 'bg-purple-500',
  WebFetch: 'bg-emerald-500',
  WebSearch: 'bg-emerald-400',
  LSP: 'bg-cyan-500',
}

function ProjectsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-1">
        <Skeleton className="h-7 w-32" />
        <Skeleton className="h-4 w-40" />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-64 rounded-xl" />
        ))}
      </div>
    </div>
  )
}

export function Component() {
  const { data, loading } = useApi(() => api.projects())
  const [query, setQuery] = useState('')

  if (loading || !data) return <ProjectsSkeleton />

  const filtered = query
    ? data.filter(p => {
        const q = query.toLowerCase()
        return slugToName(p.slug).toLowerCase().includes(q) ||
          p.project_path.toLowerCase().includes(q) ||
          (p.git_branches ?? []).some(b => b.toLowerCase().includes(q))
      })
    : data

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Projects</h2>
          <p className="text-sm text-muted-foreground">
            {filtered.length}{query ? ` of ${data.length}` : ''} projects tracked
          </p>
        </div>
        <div className="relative w-full sm:w-64">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <input
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search projects..."
            className="w-full rounded-lg border border-input bg-background pl-9 pr-4 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map(p => {
          const topTools = (p.top_tools ?? []).slice(0, 5)
          const maxCount = topTools[0]?.count || 1

          return (
            <Link key={p.slug || '_empty'} to={`/projects/${p.slug}`} className="group">
              <Card className="h-full transition-all duration-200 hover:border-primary/60 hover:shadow-md hover:shadow-primary/5 hover:-translate-y-0.5">
                <CardHeader className="pb-2">
                  <div className="flex items-start justify-between">
                    <div className="min-w-0 flex-1">
                      <CardTitle className="group-hover:text-primary transition-colors">
                        {slugToName(p.slug)}
                      </CardTitle>
                      <CardDescription className="truncate text-xs mt-0.5">
                        {p.project_path}
                      </CardDescription>
                    </div>
                    <span className="text-xs text-muted-foreground shrink-0 ml-2">
                      {formatRelative(p.last_active)}
                    </span>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-xs text-muted-foreground mb-2">
                    {p.session_count} sessions - {formatTokens(p.total_messages)} msgs - {formatDuration(p.total_duration_minutes)}
                  </p>

                  {topTools.length > 0 && (
                    <div className="flex flex-col gap-1.5 mb-3">
                      {topTools.map(t => {
                        const pct = (t.count / maxCount) * 100
                        const colorClass = TOOL_COLORS[t.name] || 'bg-muted-foreground'
                        return (
                          <div key={t.name} className="flex items-center gap-2">
                            <span className="w-16 text-xs font-mono text-muted-foreground truncate">{t.name}</span>
                            <div className="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
                              <div
                                className={`h-full rounded-full ${colorClass} transition-all duration-300`}
                                style={{ width: `${pct}%` }}
                              />
                            </div>
                            <span className="w-10 text-right text-[10px] text-muted-foreground tabular-nums">{t.count.toLocaleString()}</span>
                          </div>
                        )
                      })}
                    </div>
                  )}

                  {(p.git_branches ?? []).length > 0 && (
                    <div className="flex flex-wrap gap-1 mb-3">
                      {(p.git_branches ?? []).slice(0, 3).map(b => (
                        <Badge key={b} variant="secondary" className="text-[10px] py-0 font-mono">{b}</Badge>
                      ))}
                      {(p.git_branches ?? []).length > 3 && (
                        <Badge variant="outline" className="text-[10px] py-0">+{p.git_branches.length - 3}</Badge>
                      )}
                    </div>
                  )}

                  <div className="flex items-center justify-between pt-2 border-t border-border/50">
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {formatTokens(p.total_tokens)} tokens
                    </span>
                    <span className="text-sm font-bold tabular-nums">
                      {formatCost(p.total_cost)}
                    </span>
                  </div>
                </CardContent>
              </Card>
            </Link>
          )
        })}
      </div>

      {filtered.length === 0 && (
        <div className="flex items-center justify-center h-48 text-muted-foreground text-sm">
          No projects found.
        </div>
      )}
    </div>
  )
}
