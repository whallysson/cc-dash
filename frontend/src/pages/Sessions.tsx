import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatDuration, formatRelative, formatTokens, slugToName, totalTokens } from '@/lib/format'
import { Search, ChevronLeft, ChevronRight } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const LIMIT = 50

const sortOptions = [
  { key: 'date', label: 'Date' },
  { key: 'cost', label: 'Cost' },
  { key: 'messages', label: 'Messages' },
  { key: 'duration', label: 'Duration' },
  { key: 'tokens', label: 'Tokens' },
] as const

function SessionsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <Skeleton className="h-7 w-32" />
          <Skeleton className="h-4 w-24" />
        </div>
        <div className="flex gap-1.5">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-7 w-16 rounded-lg" />
          ))}
        </div>
      </div>
      <Skeleton className="h-9 w-full rounded-lg" />
      <Skeleton className="h-[480px] w-full rounded-xl" />
    </div>
  )
}

export function Component() {
  const [page, setPage] = useState(1)
  const [sort, setSort] = useState<string>('date')
  const [query, setQuery] = useState('')
  const [search, setSearch] = useState('')

  const { data, loading } = useApi(
    () => api.sessions(page, LIMIT, sort, search),
    [page, sort, search]
  )

  const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setSearch(query)
    setPage(1)
  }

  if (loading || !data) return <SessionsSkeleton />

  const sessions = data.sessions ?? []
  const total = data.total ?? 0
  const totalPages = Math.ceil(total / LIMIT)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Sessions</h2>
          <p className="text-sm text-muted-foreground">
            {total.toLocaleString()} total sessions
          </p>
        </div>
        <div className="flex gap-1">
          {sortOptions.map(({ key, label }) => (
            <Button
              key={key}
              variant={sort === key ? 'default' : 'outline'}
              size="sm"
              onClick={() => { setSort(key); setPage(1) }}
            >
              {label}
            </Button>
          ))}
        </div>
      </div>

      <form onSubmit={handleSearch} className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <input
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search sessions by prompt, project..."
          className="w-full rounded-lg border border-input bg-background pl-9 pr-4 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors"
        />
      </form>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="min-w-[240px]">Prompt</TableHead>
              <TableHead>Project</TableHead>
              <TableHead className="text-right">Msgs</TableHead>
              <TableHead className="text-right">Tokens</TableHead>
              <TableHead className="text-right">Cost</TableHead>
              <TableHead className="text-right">Duration</TableHead>
              <TableHead className="text-center">Features</TableHead>
              <TableHead className="text-right">When</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sessions.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} className="h-24 text-center text-muted-foreground">
                  No sessions found.
                </TableCell>
              </TableRow>
            )}
            {sessions.map(s => {
              const features = [
                s.has_compaction && { label: 'compact', variant: 'outline' as const },
                s.uses_task_agent && { label: 'agent', variant: 'secondary' as const },
                s.uses_mcp && { label: 'mcp', variant: 'secondary' as const },
                s.has_thinking && { label: 'think', variant: 'secondary' as const },
                s.uses_web_search && { label: 'web', variant: 'secondary' as const },
              ].filter(Boolean)

              return (
                <TableRow key={s.session_id} className="group">
                  <TableCell className="max-w-xs">
                    <Link
                      to={`/sessions/${s.session_id}`}
                      className="truncate block font-medium text-foreground group-hover:text-primary transition-colors"
                    >
                      {s.first_prompt?.slice(0, 80) || 'No prompt'}
                    </Link>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {slugToName(s.slug)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {s.total_message_count}
                  </TableCell>
                  <TableCell className="text-right tabular-nums text-muted-foreground">
                    {formatTokens(totalTokens(s.total_tokens))}
                  </TableCell>
                  <TableCell className="text-right tabular-nums font-medium">
                    {formatCost(s.estimated_cost)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums text-muted-foreground">
                    {formatDuration(s.duration_minutes)}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1 justify-center flex-wrap">
                      {features.map(f => f && (
                        <Badge key={f.label} variant={f.variant}>{f.label}</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground text-xs tabular-nums">
                    {formatRelative(s.start_time)}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </Card>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              <ChevronLeft className="size-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              Next
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
