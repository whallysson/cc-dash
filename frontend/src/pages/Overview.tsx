import { Link } from 'react-router-dom'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, BarChart, Bar,
} from 'recharts'
import { MessageSquare, Zap, DollarSign, HardDrive } from 'lucide-react'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatTokens, formatCost, formatBytes, formatRelative, slugToName, totalTokens } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardAction } from '@/components/ui/card'
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

function StatCard({ title, value, sub, icon }: {
  title: string
  value: string
  sub?: string
  icon: React.ReactNode
}) {
  return (
    <Card>
      <CardHeader>
        <CardDescription>{title}</CardDescription>
        <CardAction>{icon}</CardAction>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold tabular-nums">{value}</div>
        {sub && <p className="text-xs text-muted-foreground mt-1">{sub}</p>}
      </CardContent>
    </Card>
  )
}

function OverviewSkeleton() {
  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-4 w-24" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-20" />
              <Skeleton className="mt-1 h-3 w-16" />
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-7">
        <Card className="col-span-1 lg:col-span-4">
          <CardContent><Skeleton className="h-[300px] w-full" /></CardContent>
        </Card>
        <Card className="col-span-1 lg:col-span-3">
          <CardContent><Skeleton className="h-[300px] w-full" /></CardContent>
        </Card>
      </div>
      <Card>
        <CardContent><Skeleton className="h-[200px] w-full" /></CardContent>
      </Card>
    </div>
  )
}

export function Component() {
  const { data, loading } = useApi(() => api.stats())

  if (loading || !data) return <OverviewSkeleton />

  const d = data
  const tokenTotal = d.total_tokens

  const dailyData = d.daily_activity.slice(-30).map(day => ({
    date: day.date.slice(5),
    messages: day.messages,
    sessions: day.sessions,
  }))

  const modelData = Object.entries(d.model_breakdown)
    .filter(([, v]) => v > 0)
    .sort(([, a], [, b]) => b - a)
    .map(([name, value]) => ({ name, value }))

  const hourData = Array.from({ length: 24 }, (_, h) => {
    const key = String(h).padStart(2, '0')
    const count = d.hour_counts[key] || d.hour_counts[String(h)] || 0
    return { hour: `${key}h`, count, h }
  })
  const topHours = new Set(
    [...hourData].sort((a, b) => b.count - a.count).slice(0, 3).map(x => x.h)
  )

  return (
    <div className="flex flex-col gap-4">
      {/* Stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Sessions"
          value={d.total_sessions.toLocaleString()}
          sub={`${d.total_projects} projects`}
          icon={<Zap className="size-4 text-muted-foreground" />}
        />
        <StatCard
          title="Total Messages"
          value={formatTokens(d.total_messages)}
          sub={`${formatTokens(tokenTotal)} tokens`}
          icon={<MessageSquare className="size-4 text-muted-foreground" />}
        />
        <StatCard
          title="Total Cost"
          value={formatCost(d.total_cost)}
          icon={<DollarSign className="size-4 text-muted-foreground" />}
        />
        <StatCard
          title="Storage"
          value={formatBytes(d.storage_bytes)}
          icon={<HardDrive className="size-4 text-muted-foreground" />}
        />
      </div>

      {/* Charts - 4+3 split */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-7">
        {/* Area chart - Usage Over Time */}
        <Card className="col-span-1 lg:col-span-4">
          <CardHeader>
            <CardTitle>Usage Over Time</CardTitle>
            <CardDescription>Messages and sessions over the last 30 days</CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={dailyData}>
                <defs>
                  <linearGradient id="gradMsg" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--chart-1)" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="var(--chart-1)" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="gradSess" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--chart-2)" stopOpacity={0.2} />
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
                />
                <Tooltip content={<ChartTip />} />
                <Area
                  type="monotone"
                  dataKey="messages"
                  stroke="var(--chart-1)"
                  fill="url(#gradMsg)"
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 3 }}
                />
                <Area
                  type="monotone"
                  dataKey="sessions"
                  stroke="var(--chart-2)"
                  fill="url(#gradSess)"
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 3 }}
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Pie chart - Model Distribution */}
        <Card className="col-span-1 lg:col-span-3">
          <CardHeader>
            <CardTitle>Model Distribution</CardTitle>
            <CardDescription>Token usage by model</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center">
            <ResponsiveContainer width="100%" height={200}>
              <PieChart>
                <Pie
                  data={modelData}
                  cx="50%"
                  cy="50%"
                  innerRadius={55}
                  outerRadius={85}
                  paddingAngle={2}
                  dataKey="value"
                  strokeWidth={0}
                >
                  {modelData.map((_, i) => (
                    <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip content={<ChartTip fmt={formatTokens} />} />
              </PieChart>
            </ResponsiveContainer>
            <div className="flex flex-wrap gap-3 justify-center mt-2">
              {modelData.map((m, i) => (
                <span key={m.name} className="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <span
                    className="size-2 rounded-full"
                    style={{ backgroundColor: CHART_COLORS[i % CHART_COLORS.length] }}
                  />
                  {m.name}
                </span>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Bar chart - Peak Hours */}
      <Card>
        <CardHeader>
          <CardTitle>Peak Hours</CardTitle>
          <CardDescription>Activity distribution across 24 hours</CardDescription>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={hourData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis
                dataKey="hour"
                stroke="var(--muted-foreground)"
                tick={{ fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                interval={1}
              />
              <YAxis
                stroke="var(--muted-foreground)"
                tick={{ fontSize: 10 }}
                tickLine={false}
                axisLine={false}
              />
              <Tooltip content={<ChartTip />} />
              <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                {hourData.map((entry) => (
                  <Cell
                    key={entry.h}
                    fill={topHours.has(entry.h) ? 'var(--primary)' : 'var(--muted)'}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Recent Sessions */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Sessions</CardTitle>
          <CardDescription>Your latest Claude Code sessions</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col">
            {d.recent_sessions.map((s) => (
              <Link
                key={s.session_id}
                to={`/sessions/${s.session_id}`}
                className="flex items-center justify-between py-3 px-2 -mx-2 rounded-lg hover:bg-muted/50 transition-colors group"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate group-hover:text-primary transition-colors">
                    {s.first_prompt || 'No prompt'}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {slugToName(s.slug)} - {formatRelative(s.end_time || s.start_time)}
                  </p>
                </div>
                <div className="text-right ml-4 shrink-0">
                  <p className="text-sm tabular-nums">{s.total_message_count} msgs</p>
                  <p className="text-xs text-muted-foreground tabular-nums">
                    {formatTokens(totalTokens(s.total_tokens))}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
