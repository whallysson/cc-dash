import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, BarChart, Bar,
} from 'recharts'
import { DollarSign, TrendingDown, Database, Percent } from 'lucide-react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatTokens, slugToName } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

const CHART_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
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
          <span className="text-muted-foreground">{e.name}:</span>
          <span className="font-medium tabular-nums">
            {fmt && e.value != null ? fmt(e.value) : e.value?.toLocaleString()}
          </span>
        </p>
      ))}
    </div>
  )
}

function CostsSkeleton() {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <Skeleton className="h-7 w-32" />
          <Skeleton className="mt-1 h-4 w-48" />
        </div>
        <Skeleton className="h-6 w-20 rounded-full" />
      </div>
      <Card>
        <CardContent><Skeleton className="h-[260px] w-full" /></CardContent>
      </Card>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card><CardContent><Skeleton className="h-[320px] w-full" /></CardContent></Card>
        <Card><CardContent><Skeleton className="h-[320px] w-full" /></CardContent></Card>
      </div>
      <Card>
        <CardContent><Skeleton className="h-[250px] w-full" /></CardContent>
      </Card>
    </div>
  )
}

export function Component() {
  const { data, loading } = useApi(() => api.costs())

  if (loading || !data) return <CostsSkeleton />

  const dailyCostData = data.cost_by_date.slice(-30).map(d => ({
    date: d.date.slice(5),
    cost: Number(d.cost.toFixed(4)),
  }))

  const modelData = Object.values(data.cost_by_model)
    .filter(m => m.cost > 0)
    .sort((a, b) => b.cost - a.cost)
    .map(m => ({ name: m.model, value: m.cost }))

  const modelList = Object.values(data.cost_by_model).sort((a, b) => b.cost - a.cost)

  const projectData = data.cost_by_project.slice(0, 10).map(p => ({
    name: slugToName(p.slug),
    cost: p.cost,
  }))

  const cache = data.cache_efficiency

  return (
    <div className="flex flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Costs</h2>
          <p className="text-sm text-muted-foreground">Cost analytics and breakdown</p>
        </div>
        <Badge variant="secondary" className="h-7 px-3 text-base font-bold tabular-nums">
          <DollarSign className="mr-1 size-4" />
          {formatCost(data.total_cost)}
        </Badge>
      </div>

      {/* Daily Cost Area Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Daily Cost</CardTitle>
          <CardDescription>Cost trend over the last 30 days</CardDescription>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={260}>
            <AreaChart data={dailyCostData}>
              <defs>
                <linearGradient id="gradDailyCost" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--chart-2)" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="var(--chart-2)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis
                dataKey="date"
                stroke="var(--muted-foreground)"
                tick={{ fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                stroke="var(--muted-foreground)"
                tick={{ fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={v => `$${v}`}
              />
              <Tooltip content={<ChartTip fmt={formatCost} />} />
              <Area
                type="monotone"
                dataKey="cost"
                stroke="var(--chart-2)"
                fill="url(#gradDailyCost)"
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 3 }}
              />
            </AreaChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* 2-col grid: Pie + Cache */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Cost by Model - Pie */}
        <Card>
          <CardHeader>
            <CardTitle>Cost by Model</CardTitle>
            <CardDescription>Spending distribution across models</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center">
            <ResponsiveContainer width="100%" height={200}>
              <PieChart>
                <Pie
                  data={modelData}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={80}
                  paddingAngle={2}
                  dataKey="value"
                  strokeWidth={0}
                >
                  {modelData.map((_, i) => (
                    <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip content={<ChartTip fmt={formatCost} />} />
              </PieChart>
            </ResponsiveContainer>
            <div className="flex flex-col gap-2 mt-3 w-full">
              {modelList.map((m, i) => (
                <div key={m.model} className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <span
                      className="size-2 rounded-full shrink-0"
                      style={{ backgroundColor: CHART_COLORS[i % CHART_COLORS.length] }}
                    />
                    <span className="truncate">{m.model}</span>
                  </div>
                  <div className="flex items-center gap-3 shrink-0">
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {m.session_count} sessions
                    </span>
                    <span className="font-medium tabular-nums">{formatCost(m.cost)}</span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Cache Efficiency */}
        <Card>
          <CardHeader>
            <CardTitle>Cache Efficiency</CardTitle>
            <CardAction>
              <Percent className="size-4 text-muted-foreground" />
            </CardAction>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-5">
              {/* Hit rate bar */}
              <div>
                <div className="flex justify-between text-sm mb-2">
                  <span className="text-muted-foreground">Cache Hit Rate</span>
                  <span className="font-bold tabular-nums">{cache.cache_hit_rate.toFixed(1)}%</span>
                </div>
                <div className="h-2.5 rounded-full bg-muted overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${Math.min(cache.cache_hit_rate, 100)}%`,
                      backgroundColor: 'var(--chart-2)',
                    }}
                  />
                </div>
              </div>

              {/* Stats grid */}
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-1.5">
                    <Database className="size-3.5 text-muted-foreground" />
                    <p className="text-xs text-muted-foreground">Cache Reads</p>
                  </div>
                  <p className="text-lg font-bold tabular-nums">{formatTokens(cache.total_cache_reads)}</p>
                </div>
                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-1.5">
                    <Database className="size-3.5 text-muted-foreground" />
                    <p className="text-xs text-muted-foreground">Cache Writes</p>
                  </div>
                  <p className="text-lg font-bold tabular-nums">{formatTokens(cache.total_cache_writes)}</p>
                </div>
              </div>

              {/* Input tokens */}
              <div className="border-t pt-3">
                <p className="text-xs text-muted-foreground">Total Input Tokens</p>
                <p className="text-lg font-bold tabular-nums">{formatTokens(cache.total_input_tokens)}</p>
              </div>

              {/* Estimated savings */}
              {cache.estimated_savings > 0 && (
                <div className="border-t pt-3">
                  <div className="flex items-center gap-1.5">
                    <TrendingDown className="size-3.5 text-chart-2" />
                    <p className="text-xs text-muted-foreground">Estimated Savings</p>
                  </div>
                  <p className="text-2xl font-bold tabular-nums text-chart-2">
                    {formatCost(cache.estimated_savings)}
                  </p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Cost by Project - Horizontal Bar */}
      {projectData.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Cost by Project</CardTitle>
            <CardDescription>Top {projectData.length} projects by spending</CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={Math.max(projectData.length * 35, 120)}>
              <BarChart data={projectData} layout="vertical" margin={{ left: 80 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" horizontal={false} />
                <XAxis
                  type="number"
                  stroke="var(--muted-foreground)"
                  tick={{ fontSize: 11 }}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={v => `$${v}`}
                />
                <YAxis
                  type="category"
                  dataKey="name"
                  stroke="var(--muted-foreground)"
                  tick={{ fontSize: 12 }}
                  tickLine={false}
                  axisLine={false}
                  width={80}
                />
                <Tooltip content={<ChartTip fmt={formatCost} />} />
                <Bar
                  dataKey="cost"
                  fill="currentColor"
                  className="fill-primary"
                  radius={[0, 4, 4, 0]}
                  barSize={18}
                />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
