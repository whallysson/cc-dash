import { Link } from 'react-router-dom'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts'
import {
  ShieldCheck, TrendingUp, Brain, Skull, Database,
  AlertTriangle, Zap, ArrowRight,
} from 'lucide-react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatTokens, formatDuration, formatRelative, slugToName } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

const CHART_COLORS = [
  'var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)',
  'var(--chart-4)', 'var(--chart-5)',
]

function ChartTip({ active, payload, label, fmt }: {
  active?: boolean
  payload?: Array<{ name?: string; value?: number; color?: string }>
  label?: string
  fmt?: (v: number) => string
}) {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-lg border bg-popover px-3 py-2 text-sm shadow-lg">
      <p className="text-muted-foreground mb-1">{label}</p>
      {payload.map((e, i) => (
        <p key={i} className="flex items-center gap-2">
          <span className="size-2 rounded-full" style={{ backgroundColor: e.color }} />
          <span className="font-medium tabular-nums">
            {fmt && e.value != null ? fmt(e.value) : e.value?.toLocaleString()}
          </span>
        </p>
      ))}
    </div>
  )
}

function HealthBadge({ score }: { score: number }) {
  const color = score >= 80 ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/20'
    : score >= 60 ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border-amber-500/20'
    : 'bg-red-500/15 text-red-600 dark:text-red-400 border-red-500/20'
  const label = score >= 80 ? 'Healthy' : score >= 60 ? 'Warning' : 'Critical'
  return (
    <Badge className={`${color} text-base px-3 py-1 gap-1.5`}>
      <ShieldCheck className="size-4" />
      {score}/100 - {label}
    </Badge>
  )
}

function TokenHealthSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-9 w-40" />
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i}><CardContent><Skeleton className="h-20 w-full" /></CardContent></Card>
        ))}
      </div>
      <div className="grid gap-4 lg:grid-cols-2">
        <Card><CardContent><Skeleton className="h-[300px] w-full" /></CardContent></Card>
        <Card><CardContent><Skeleton className="h-[300px] w-full" /></CardContent></Card>
      </div>
    </div>
  )
}

export function Component() {
  const { data, loading } = useApi(() => api.efficiency())

  if (loading || !data) return <TokenHealthSkeleton />

  const cpm = data.cost_per_message
  const thinking = data.thinking_impact

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Token Health</h2>
          <p className="text-sm text-muted-foreground">
            Efficiency analysis across {data.total_sessions} sessions
          </p>
        </div>
        <HealthBadge score={data.health_score} />
      </div>

      {/* 1. Cost Per Message Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader>
            <CardDescription>Median Cost/Msg</CardDescription>
            <CardAction><TrendingUp className="size-4 text-muted-foreground" /></CardAction>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tabular-nums">{formatCost(cpm.median)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Mean: {formatCost(cpm.mean)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardDescription>P90 Cost/Msg</CardDescription>
            <CardAction><AlertTriangle className="size-4 text-muted-foreground" /></CardAction>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tabular-nums">{formatCost(cpm.p90)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              P99: {formatCost(cpm.p99)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardDescription>Thinking Multiplier</CardDescription>
            <CardAction><Brain className="size-4 text-muted-foreground" /></CardAction>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tabular-nums">
              {thinking.cost_multiplier > 0 ? `${thinking.cost_multiplier.toFixed(1)}x` : 'N/A'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              vs non-thinking sessions
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardDescription>Total Spend</CardDescription>
            <CardAction><Zap className="size-4 text-muted-foreground" /></CardAction>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold tabular-nums">{formatCost(data.total_cost)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              {data.total_sessions} sessions
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 2. Model Comparison + Cost Distribution */}
      <div className="grid gap-4 lg:grid-cols-7">
        {/* Model Comparison */}
        <Card className="col-span-1 lg:col-span-4">
          <CardHeader>
            <CardTitle>Opus vs Sonnet</CardTitle>
            <CardDescription>Side-by-side cost efficiency by model</CardDescription>
          </CardHeader>
          <CardContent>
            {data.model_comparison.length > 0 ? (
              <div className="flex flex-col gap-4">
                {data.model_comparison.map((m, i) => (
                  <div key={m.model} className="flex flex-col gap-2 p-3 rounded-lg bg-muted/30">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="size-3 rounded-full" style={{ backgroundColor: CHART_COLORS[i % CHART_COLORS.length] }} />
                        <span className="font-medium text-sm">{m.model}</span>
                        <Badge variant="outline">{m.sessions} sessions</Badge>
                      </div>
                      <span className="font-bold tabular-nums">{formatCost(m.total_cost)}</span>
                    </div>
                    <div className="grid grid-cols-4 gap-4 text-xs">
                      <div>
                        <p className="text-muted-foreground">Cost/Msg</p>
                        <p className="font-semibold tabular-nums text-sm">{formatCost(m.cost_per_message)}</p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Tokens/Msg</p>
                        <p className="font-semibold tabular-nums text-sm">{formatTokens(Math.round(m.avg_tokens_per_message))}</p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Cache Hit</p>
                        <p className="font-semibold tabular-nums text-sm">{m.cache_hit_rate.toFixed(1)}%</p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Messages</p>
                        <p className="font-semibold tabular-nums text-sm">{m.total_messages.toLocaleString()}</p>
                      </div>
                    </div>
                    <div className="h-1.5 rounded-full bg-muted overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all duration-500"
                        style={{
                          width: `${Math.min(m.cache_hit_rate, 100)}%`,
                          backgroundColor: CHART_COLORS[i % CHART_COLORS.length],
                        }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-8">No model data available</p>
            )}
          </CardContent>
        </Card>

        {/* Cost Distribution */}
        <Card className="col-span-1 lg:col-span-3">
          <CardHeader>
            <CardTitle>Cost/Msg Distribution</CardTitle>
            <CardDescription>How your sessions distribute by cost per message</CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={data.cost_distribution} margin={{ left: -10 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis
                  dataKey="label"
                  stroke="var(--muted-foreground)"
                  tick={{ fontSize: 10 }}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  stroke="var(--muted-foreground)"
                  tick={{ fontSize: 10 }}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip content={<ChartTip />} />
                <Bar dataKey="count" fill="var(--chart-1)" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* 3. Thinking Impact */}
      <Card>
        <CardHeader>
          <CardTitle>Thinking Impact</CardTitle>
          <CardDescription>
            How extended thinking affects your costs
            {thinking.cost_multiplier > 1.5 && (
              <span className="ml-2 text-amber-500 dark:text-amber-400">
                - thinking costs {thinking.cost_multiplier.toFixed(1)}x more per message
              </span>
            )}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-lg border p-4">
              <div className="flex items-center gap-2 mb-3">
                <Brain className="size-4 text-chart-4" />
                <span className="font-medium text-sm">With Thinking</span>
                <Badge variant="outline">{thinking.with_thinking.sessions} sessions</Badge>
              </div>
              <div className="grid grid-cols-3 gap-3 text-xs">
                <div>
                  <p className="text-muted-foreground">Avg Cost/Msg</p>
                  <p className="font-bold tabular-nums text-lg">{formatCost(thinking.with_thinking.avg_cost_per_message)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Avg Session</p>
                  <p className="font-bold tabular-nums text-lg">{formatCost(thinking.with_thinking.avg_cost)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Avg Duration</p>
                  <p className="font-bold tabular-nums text-lg">{formatDuration(thinking.with_thinking.avg_duration)}</p>
                </div>
              </div>
            </div>
            <div className="rounded-lg border p-4">
              <div className="flex items-center gap-2 mb-3">
                <Zap className="size-4 text-chart-2" />
                <span className="font-medium text-sm">Without Thinking</span>
                <Badge variant="outline">{thinking.without_thinking.sessions} sessions</Badge>
              </div>
              <div className="grid grid-cols-3 gap-3 text-xs">
                <div>
                  <p className="text-muted-foreground">Avg Cost/Msg</p>
                  <p className="font-bold tabular-nums text-lg">{formatCost(thinking.without_thinking.avg_cost_per_message)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Avg Session</p>
                  <p className="font-bold tabular-nums text-lg">{formatCost(thinking.without_thinking.avg_cost)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Avg Duration</p>
                  <p className="font-bold tabular-nums text-lg">{formatDuration(thinking.without_thinking.avg_duration)}</p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 4. Vampire Sessions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Skull className="size-5 text-destructive" />
            Vampire Sessions
          </CardTitle>
          <CardDescription>Top 10 most expensive sessions - where your tokens went</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col">
            {data.vampire_sessions.map((v, i) => (
              <Link
                key={v.session_id}
                to={`/sessions/${v.session_id}`}
                className="flex items-center gap-4 py-3 px-2 -mx-2 rounded-lg hover:bg-muted/50 transition-colors group"
              >
                <span className="text-sm font-bold text-muted-foreground w-6 text-right tabular-nums">
                  #{i + 1}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate group-hover:text-primary transition-colors">
                    {v.first_prompt || 'No prompt'}
                  </p>
                  <div className="flex items-center gap-2 mt-0.5 flex-wrap">
                    <span className="text-xs text-muted-foreground">{slugToName(v.slug)}</span>
                    <Badge variant="outline" className="text-[10px] py-0">{v.primary_model}</Badge>
                    {v.has_thinking && <Badge variant="secondary" className="text-[10px] py-0">thinking</Badge>}
                    {v.has_compaction && <Badge variant="secondary" className="text-[10px] py-0">compaction</Badge>}
                    <span className="text-[10px] text-muted-foreground/50">{formatRelative(v.start_time)}</span>
                  </div>
                </div>
                <div className="text-right shrink-0">
                  <p className="text-sm font-bold tabular-nums text-destructive">{formatCost(v.cost)}</p>
                  <p className="text-xs text-muted-foreground tabular-nums">
                    {formatCost(v.cost_per_message)}/msg - {v.messages} msgs - {formatDuration(v.duration_minutes)}
                  </p>
                </div>
                <ArrowRight className="size-4 text-muted-foreground/30 group-hover:text-primary transition-colors shrink-0" />
              </Link>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* 5. Cache Efficiency by Session */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="size-5 text-chart-3" />
            Worst Cache Performance
          </CardTitle>
          <CardDescription>Sessions with lowest cache hit rates (min 5 messages)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
            {data.cache_by_session.map((c) => (
              <Link
                key={c.session_id}
                to={`/sessions/${c.session_id}`}
                className="flex items-center gap-3 py-2 px-2 -mx-2 rounded-lg hover:bg-muted/50 transition-colors group"
              >
                <div className="w-16 shrink-0">
                  <div className="h-1.5 rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full rounded-full bg-destructive"
                      style={{ width: `${Math.max(c.cache_hit_rate, 2)}%` }}
                    />
                  </div>
                  <p className="text-[10px] text-muted-foreground text-center mt-0.5 tabular-nums">
                    {c.cache_hit_rate}%
                  </p>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm truncate group-hover:text-primary transition-colors">
                    {c.first_prompt || 'No prompt'}
                  </p>
                  <span className="text-xs text-muted-foreground">{slugToName(c.slug)}</span>
                </div>
                <div className="text-right shrink-0 text-xs text-muted-foreground tabular-nums">
                  <p>{formatCost(c.cost)}</p>
                  <p>{c.messages} msgs</p>
                </div>
              </Link>
            ))}
            {data.cache_by_session.length === 0 && (
              <p className="text-sm text-muted-foreground text-center py-4">No sessions with poor cache performance</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
