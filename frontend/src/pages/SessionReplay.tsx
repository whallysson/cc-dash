import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { formatCost, formatTokens, formatDateTime, totalTokens } from '@/lib/format'
import { User, Bot, Wrench, Brain, ChevronDown, ArrowLeft } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import type { ReplayTurn } from '@/lib/types'

function ReplaySkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-7 w-96" />
        <div className="flex gap-2">
          <Skeleton className="h-5 w-16 rounded-full" />
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-5 w-14 rounded-full" />
        </div>
      </div>
      {Array.from({ length: 4 }).map((_, i) => (
        <Skeleton key={i} className="h-36 w-full rounded-xl" />
      ))}
    </div>
  )
}

function TurnCard({ turn }: { turn: ReplayTurn }) {
  const [showThinking, setShowThinking] = useState(false)
  const [expandedTools, setExpandedTools] = useState<Set<string>>(new Set())
  const isUser = turn.role === 'user'

  const toggleTool = (id: string) => {
    const next = new Set(expandedTools)
    next.has(id) ? next.delete(id) : next.add(id)
    setExpandedTools(next)
  }

  return (
    <Card className={isUser ? '' : 'bg-muted/30'}>
      <div className="flex items-center gap-2 px-4 py-2.5 border-b border-border/50">
        <div className={`flex items-center justify-center size-6 rounded-md ${isUser ? 'bg-chart-3/10 text-chart-3' : 'bg-primary/10 text-primary'}`}>
          {isUser ? <User className="size-3.5" /> : <Bot className="size-3.5" />}
        </div>
        <span className="text-xs font-semibold">{isUser ? 'User' : 'Assistant'}</span>
        {turn.model && (
          <span className="text-[10px] text-muted-foreground/60 font-mono">{turn.model}</span>
        )}
        <div className="flex-1" />
        {!isUser && (
          <span className="text-[10px] text-muted-foreground tabular-nums">
            {formatTokens(totalTokens(turn.tokens))} tokens | {formatCost(turn.cost)}
          </span>
        )}
        <span className="text-[10px] text-muted-foreground/50 tabular-nums">
          {formatDateTime(turn.timestamp)}
        </span>
      </div>

      {turn.has_thinking && turn.thinking_text && (
        <>
          <button
            onClick={() => setShowThinking(!showThinking)}
            className="w-full flex items-center gap-1.5 px-4 py-2 text-xs font-medium text-amber-600 dark:text-amber-400 hover:bg-amber-500/5 transition-colors border-b border-border/50"
          >
            <Brain className="size-3" />
            Thinking
            <ChevronDown className={`size-3 ml-auto transition-transform duration-200 ${showThinking ? 'rotate-180' : ''}`} />
          </button>
          {showThinking && (
            <div className="px-4 py-3 text-xs text-muted-foreground bg-amber-500/5 border-b border-border/50 whitespace-pre-wrap leading-relaxed max-h-72 overflow-y-auto">
              {turn.thinking_text}
            </div>
          )}
        </>
      )}

      {turn.text && (
        <CardContent className="py-3">
          <div className="text-sm whitespace-pre-wrap break-words leading-relaxed">
            {turn.text}
          </div>
        </CardContent>
      )}

      {(turn.tool_calls ?? []).length > 0 && (
        <div className="border-t border-border/50 px-4 py-2.5 flex flex-col gap-1.5">
          {(turn.tool_calls ?? []).map(tc => (
            <div key={tc.id} className="text-xs">
              <button
                onClick={() => toggleTool(tc.id)}
                className={`flex items-center gap-1.5 px-2 py-1 rounded-md transition-colors ${
                  tc.is_error
                    ? 'text-destructive bg-destructive/10 hover:bg-destructive/15'
                    : 'text-chart-3 bg-chart-3/10 hover:bg-chart-3/15'
                }`}
              >
                <Wrench className="size-2.5" />
                <span className="font-semibold">{tc.name}</span>
                <ChevronDown className={`size-2.5 ml-1 transition-transform duration-200 ${expandedTools.has(tc.id) ? 'rotate-180' : ''}`} />
              </button>
              {expandedTools.has(tc.id) && (
                <div className="mt-1.5 ml-4 flex flex-col gap-1.5">
                  {tc.input && (
                    <pre className="text-muted-foreground bg-muted/50 p-2.5 rounded-md text-[11px] overflow-x-auto max-h-40 overflow-y-auto leading-relaxed">
                      {tc.input}
                    </pre>
                  )}
                  {tc.result && (
                    <pre className="text-muted-foreground bg-muted/30 p-2.5 rounded-md text-[11px] overflow-x-auto max-h-40 overflow-y-auto leading-relaxed">
                      {tc.result}
                    </pre>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </Card>
  )
}

export function Component() {
  const { id } = useParams<{ id: string }>()
  const { data: session } = useApi(() => api.session(id!), [id])
  const { data: replay, loading } = useApi(() => api.replay(id!, 0, 200, true), [id])

  if (loading || !replay) return <ReplaySkeleton />

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Link
          to="/sessions"
          className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-primary mb-3 transition-colors"
        >
          <ArrowLeft className="size-3" />
          Back to sessions
        </Link>
        <h2 className="text-xl font-bold tracking-tight leading-tight">
          {session?.first_prompt?.slice(0, 80) || 'Session Replay'}
        </h2>
        <div className="flex gap-2 mt-3 flex-wrap">
          {replay.version && <Badge variant="outline">v{replay.version}</Badge>}
          {replay.git_branch && <Badge variant="secondary">{replay.git_branch}</Badge>}
          <Badge variant="secondary">
            {(replay.turns ?? []).filter(t => t.text || (t.tool_calls && t.tool_calls.length > 0) || (t.has_thinking && t.thinking_text)).length} turns
          </Badge>
          <Badge>{formatCost(replay.total_cost)}</Badge>
        </div>
      </div>

      <div className="flex flex-col gap-3">
        {replay.turns
          .filter(t => t.text || (t.tool_calls && t.tool_calls.length > 0) || (t.has_thinking && t.thinking_text))
          .map((turn) => {
          const compaction = replay.compactions?.find(c => c.turn_index === turn.index)
          return (
            <div key={turn.uuid || turn.index}>
              {compaction && (
                <div className="flex items-center gap-3 py-3 mb-3">
                  <div className="flex-1 h-px bg-amber-500/30" />
                  <span className="text-xs font-medium text-amber-600 dark:text-amber-400">
                    compaction
                  </span>
                  <div className="flex-1 h-px bg-amber-500/30" />
                </div>
              )}
              <TurnCard turn={turn} />
            </div>
          )
        })}
      </div>

      {replay.has_more && (
        <p className="text-center text-sm text-muted-foreground py-4">
          More turns available (offset: {replay.next_offset})
        </p>
      )}
    </div>
  )
}
