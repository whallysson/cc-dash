import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Flame, Trophy, CalendarDays, Hash } from 'lucide-react'

function ChartTip({ active, payload, label }: { active?: boolean; payload?: Array<{ name?: string; value?: number; color?: string }>; label?: string }) {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-lg border bg-popover px-3 py-2 text-sm shadow-lg">
      <p className="text-muted-foreground mb-1">{label}</p>
      {payload.map((e, i) => (
        <p key={i} className="font-medium tabular-nums">{e.value?.toLocaleString()}</p>
      ))}
    </div>
  )
}

const STAT_CARDS = [
  { key: 'current_streak', label: 'Current Streak', suffix: 'd', icon: Flame },
  { key: 'longest_streak', label: 'Longest Streak', suffix: 'd', icon: Trophy },
  { key: 'active_days', label: 'Active Days', suffix: '', icon: CalendarDays },
  { key: 'total_days', label: 'Total Days', suffix: '', icon: Hash },
] as const

export function Component() {
  const { data, loading } = useApi(() => api.activity())

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-24 rounded-xl" />)}
        </div>
        <Skeleton className="h-[200px] w-full rounded-xl" />
      </div>
    )
  }

  const dateMap = new Map(data.heatmap.map(h => [h.date, h.count]))
  const maxCount = Math.max(...data.heatmap.map(h => h.count), 1)
  const today = new Date()

  const weeks: { date: string; count: number; dow: number }[][] = []
  let currentWeek: { date: string; count: number; dow: number }[] = []

  for (let i = 364; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const dateStr = d.toISOString().slice(0, 10)
    const dow = d.getDay()
    const count = dateMap.get(dateStr) || 0
    if (dow === 0 && currentWeek.length > 0) {
      weeks.push(currentWeek)
      currentWeek = []
    }
    currentWeek.push({ date: dateStr, count, dow })
  }
  if (currentWeek.length > 0) weeks.push(currentWeek)

  function getColor(count: number): string {
    if (count === 0) return 'bg-muted/30'
    const intensity = count / maxCount
    if (intensity > 0.75) return 'bg-primary'
    if (intensity > 0.5) return 'bg-primary/70'
    if (intensity > 0.25) return 'bg-primary/45'
    return 'bg-primary/25'
  }

  const dayOfWeekData = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'].map(day => ({
    day: day.slice(0, 3),
    count: data.day_of_week[day] || 0,
  }))

  const peakHoursData = Array.from({ length: 24 }, (_, h) => {
    const key = String(h).padStart(2, '0')
    return { hour: `${key}h`, count: data.peak_hours[key] || data.peak_hours[String(h)] || 0 }
  })

  const dayLabels = ['', 'Mon', '', 'Wed', '', 'Fri', '']

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Activity</h2>
        <p className="text-sm text-muted-foreground">Your coding activity over time</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {STAT_CARDS.map(({ key, label, suffix, icon: Icon }) => (
          <Card key={key}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">{label}</CardTitle>
              <Icon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tabular-nums">{data[key]}{suffix}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Activity Heatmap</CardTitle>
          <CardDescription>365-day contribution graph</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-[3px] overflow-x-auto pb-2">
            <div className="flex flex-col gap-[3px] mr-1 shrink-0">
              {dayLabels.map((l, i) => (
                <div key={i} className="h-[12px] text-[10px] text-muted-foreground leading-[12px]">{l}</div>
              ))}
            </div>
            {weeks.map((week, wi) => (
              <div key={wi} className="flex flex-col gap-[3px]">
                {Array.from({ length: 7 }, (_, dow) => {
                  const day = week.find(d => d.dow === dow)
                  if (!day) return <div key={dow} className="size-[12px]" />
                  return (
                    <div key={dow} className={`size-[12px] rounded-sm ${getColor(day.count)} group relative cursor-default`}>
                      <div className="absolute bottom-full mb-1 left-1/2 -translate-x-1/2 rounded-lg border bg-popover px-2 py-1 text-[10px] shadow-lg opacity-0 group-hover:opacity-100 whitespace-nowrap z-10 pointer-events-none">
                        {day.date}: {day.count} sessions
                      </div>
                    </div>
                  )
                })}
              </div>
            ))}
          </div>
          <div className="flex items-center gap-1.5 mt-3 justify-end">
            <span className="text-[10px] text-muted-foreground mr-1">Less</span>
            <div className="size-[10px] rounded-sm bg-muted/30" />
            <div className="size-[10px] rounded-sm bg-primary/25" />
            <div className="size-[10px] rounded-sm bg-primary/45" />
            <div className="size-[10px] rounded-sm bg-primary/70" />
            <div className="size-[10px] rounded-sm bg-primary" />
            <span className="text-[10px] text-muted-foreground ml-1">More</span>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Day of Week</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={dayOfWeekData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="day" stroke="var(--muted-foreground)" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} />
                <YAxis stroke="var(--muted-foreground)" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} />
                <Tooltip content={<ChartTip />} />
                <Bar dataKey="count" fill="currentColor" className="fill-primary" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Peak Hours</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={peakHoursData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="hour" stroke="var(--muted-foreground)" tick={{ fontSize: 10 }} tickLine={false} axisLine={false} interval={2} />
                <YAxis stroke="var(--muted-foreground)" tick={{ fontSize: 10 }} tickLine={false} axisLine={false} />
                <Tooltip content={<ChartTip />} />
                <Bar dataKey="count" fill="currentColor" className="fill-chart-3" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
