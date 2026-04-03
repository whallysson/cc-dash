import { useParams, Link } from 'react-router-dom'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatDuration, formatRelative, formatTokens, slugToName } from '@/lib/format'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ArrowLeft, MessageSquare, DollarSign, Clock, Cpu } from 'lucide-react'

function DetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-7 w-48" />
        <Skeleton className="h-4 w-64" />
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-28 rounded-xl" />
        ))}
      </div>
      <Skeleton className="h-[320px] w-full rounded-xl" />
    </div>
  )
}

const statCards = [
  { key: 'sessions', label: 'Sessions', icon: MessageSquare, format: (p: { session_count: number }) => String(p.session_count) },
  { key: 'cost', label: 'Total Cost', icon: DollarSign, format: (p: { total_cost: number }) => formatCost(p.total_cost) },
  { key: 'time', label: 'Total Time', icon: Clock, format: (p: { total_duration_minutes: number }) => formatDuration(p.total_duration_minutes) },
  { key: 'tokens', label: 'Total Tokens', icon: Cpu, format: (p: { total_tokens: number }) => formatTokens(p.total_tokens) },
] as const

export function Component() {
  const { slug } = useParams<{ slug: string }>()
  const { data, loading } = useApi(() => api.project(slug!), [slug])

  if (loading || !data) return <DetailSkeleton />

  const { project: p, sessions } = data

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Link
          to="/projects"
          className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-primary mb-3 transition-colors"
        >
          <ArrowLeft className="size-3" />
          Back to projects
        </Link>
        <h2 className="text-2xl font-bold tracking-tight">{slugToName(p.slug)}</h2>
        <p className="text-xs text-muted-foreground mt-1">{p.project_path}</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map(({ key, label, icon: Icon, format }) => (
          <Card key={key}>
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {label}
                </CardTitle>
                <Icon className="size-4 text-muted-foreground/50" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tabular-nums">{format(p)}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader className="border-b">
          <CardTitle>Sessions</CardTitle>
        </CardHeader>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="min-w-[240px]">Prompt</TableHead>
              <TableHead className="text-right">Msgs</TableHead>
              <TableHead className="text-right">Cost</TableHead>
              <TableHead className="text-right">Duration</TableHead>
              <TableHead className="text-right">When</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(!sessions || sessions.length === 0) && (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                  No sessions found.
                </TableCell>
              </TableRow>
            )}
            {sessions?.map(s => (
              <TableRow key={s.session_id} className="group">
                <TableCell className="max-w-md">
                  <Link
                    to={`/sessions/${s.session_id}`}
                    className="truncate block font-medium text-foreground group-hover:text-primary transition-colors"
                  >
                    {s.first_prompt?.slice(0, 80) || 'No prompt'}
                  </Link>
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {s.total_message_count}
                </TableCell>
                <TableCell className="text-right tabular-nums font-medium">
                  {formatCost(s.estimated_cost)}
                </TableCell>
                <TableCell className="text-right tabular-nums text-muted-foreground">
                  {formatDuration(s.duration_minutes)}
                </TableCell>
                <TableCell className="text-right text-muted-foreground text-xs tabular-nums">
                  {formatRelative(s.end_time || s.start_time)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </div>
  )
}
