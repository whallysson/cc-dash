import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatDate } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

const CATEGORY_COLORS: Record<string, string> = {
  'file-io': '#60a5fa',
  shell: '#d97706',
  agent: '#a78bfa',
  web: '#22c55e',
  planning: '#fbbf24',
  mcp: '#34d399',
  skill: '#38bdf8',
  'code-intelligence': '#818cf8',
  other: '#64748b',
}

export function Component() {
  const { data, loading } = useApi(() => api.tools())

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-[400px] w-full rounded-xl" />
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <Skeleton className="h-[300px] rounded-xl" />
          <Skeleton className="h-[300px] rounded-xl" />
        </div>
      </div>
    )
  }

  const rankingData = data.tool_ranking.slice(0, 20)
  const maxCount = rankingData[0]?.count || 1

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Tools</h2>
        <p className="text-sm text-muted-foreground">Tool usage analytics and ranking</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tool Ranking (Top 20)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
            {rankingData.map((t, i) => {
              const pct = (t.count / maxCount) * 100
              return (
                <div key={t.name} className="flex items-center gap-3">
                  <span className="w-5 text-xs text-muted-foreground text-right tabular-nums">{i + 1}</span>
                  <span className="w-32 text-sm truncate font-medium">{t.name}</span>
                  <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full rounded-full transition-all duration-300"
                      style={{ width: `${pct}%`, backgroundColor: CATEGORY_COLORS[t.category] || '#64748b' }}
                    />
                  </div>
                  <span className="w-14 text-right text-xs text-muted-foreground tabular-nums">
                    {t.count.toLocaleString()}
                  </span>
                  <Badge variant="secondary" className="text-[10px]">{t.category}</Badge>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Feature Adoption</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-3">
              {data.feature_adoption.map(f => (
                <div key={f.feature}>
                  <div className="flex justify-between text-sm mb-1.5">
                    <span className="text-muted-foreground">{f.feature}</span>
                    <span className="tabular-nums">
                      {f.session_count} <span className="text-muted-foreground">({f.percentage.toFixed(0)}%)</span>
                    </span>
                  </div>
                  <div className="h-1.5 rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full rounded-full bg-primary/60 transition-all duration-300"
                      style={{ width: `${f.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
              {data.feature_adoption.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-4">No feature data</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Version History</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-1 max-h-80 overflow-y-auto">
              {data.version_history.map(v => (
                <div key={v.version} className="flex items-center justify-between py-1.5">
                  <span className="text-sm font-medium font-mono">{v.version}</span>
                  <span className="text-xs text-muted-foreground tabular-nums">
                    {formatDate(v.first_seen)} - {v.session_count} sessions
                  </span>
                </div>
              ))}
              {data.version_history.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-4">No version data</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
