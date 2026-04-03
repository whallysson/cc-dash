import { useParams } from 'react-router-dom'
import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'
import { formatCost, formatTokens, formatDateTime, totalTokens } from '../lib/format'
import { User, Bot, Wrench, Brain, ChevronDown } from 'lucide-react'
import { useState } from 'react'

export function Component() {
  const { id } = useParams<{ id: string }>()
  const { data: session } = useApi(() => api.session(id!), [id])
  const { data: replay, loading } = useApi(() => api.replay(id!, 0, 200), [id])

  if (loading || !replay) return <div className="text-zinc-500">Carregando replay...</div>

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold">{session?.first_prompt?.slice(0, 60) || 'Session Replay'}</h2>
          <div className="text-xs text-zinc-500 mt-1 flex gap-3">
            {replay.version && <span>v{replay.version}</span>}
            {replay.git_branch && <span>branch: {replay.git_branch}</span>}
            <span>{replay.turns.length} turns</span>
            <span>Cost: {formatCost(replay.total_cost)}</span>
          </div>
        </div>
      </div>

      {/* Turns */}
      <div className="space-y-3">
        {(replay.turns ?? []).map((turn) => (
          <TurnCard key={turn.uuid || turn.index} turn={turn} />
        ))}
        {(replay.compactions ?? []).map((_, i) => (
          <div key={`c-${i}`} className="flex items-center gap-2 py-2">
            <div className="flex-1 h-px bg-yellow-500/30" />
            <span className="text-xs text-yellow-500">compaction</span>
            <div className="flex-1 h-px bg-yellow-500/30" />
          </div>
        ))}
      </div>

      {replay.has_more && (
        <div className="text-center text-sm text-zinc-500">
          More turns available (offset: {replay.next_offset})
        </div>
      )}
    </div>
  )
}

function TurnCard({ turn }: { turn: import('../lib/types').ReplayTurn }) {
  const [showThinking, setShowThinking] = useState(false)
  const [expandedTools, setExpandedTools] = useState<Set<string>>(new Set())

  const isUser = turn.role === 'user'

  return (
    <div className={`rounded-lg border ${isUser ? 'bg-zinc-900 border-zinc-800' : 'bg-zinc-900/50 border-zinc-800/50'}`}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-zinc-800/50">
        {isUser ? (
          <User size={14} className="text-blue-400" />
        ) : (
          <Bot size={14} className="text-orange-400" />
        )}
        <span className="text-xs font-medium text-zinc-400">{isUser ? 'User' : 'Assistant'}</span>
        {turn.model && <span className="text-[10px] text-zinc-600">{turn.model}</span>}
        <div className="flex-1" />
        {!isUser && (
          <span className="text-[10px] text-zinc-600">
            {formatTokens(totalTokens(turn.tokens))} tokens | {formatCost(turn.cost)}
          </span>
        )}
        <span className="text-[10px] text-zinc-600">{formatDateTime(turn.timestamp)}</span>
      </div>

      {/* Thinking */}
      {turn.has_thinking && turn.thinking_text && (
        <button
          onClick={() => setShowThinking(!showThinking)}
          className="w-full flex items-center gap-1 px-4 py-1.5 text-xs text-purple-400 hover:bg-purple-500/5"
        >
          <Brain size={12} />
          <span>Thinking</span>
          <ChevronDown size={12} className={`transition-transform ${showThinking ? 'rotate-180' : ''}`} />
        </button>
      )}
      {showThinking && turn.thinking_text && (
        <div className="px-4 py-2 text-xs text-zinc-500 bg-purple-500/5 border-b border-zinc-800/30 whitespace-pre-wrap max-h-64 overflow-y-auto">
          {turn.thinking_text}
        </div>
      )}

      {/* Text */}
      {turn.text && (
        <div className="px-4 py-3 text-sm text-zinc-300 whitespace-pre-wrap break-words">
          {turn.text}
        </div>
      )}

      {/* Tool Calls */}
      {(turn.tool_calls ?? []).length > 0 && (
        <div className="border-t border-zinc-800/30 px-4 py-2 space-y-1.5">
          {(turn.tool_calls ?? []).map((tc) => (
            <div key={tc.id} className="text-xs">
              <button
                onClick={() => {
                  const s = new Set(expandedTools)
                  s.has(tc.id) ? s.delete(tc.id) : s.add(tc.id)
                  setExpandedTools(s)
                }}
                className={`flex items-center gap-1.5 px-2 py-1 rounded ${tc.is_error ? 'text-red-400 bg-red-500/10' : 'text-cyan-400 bg-cyan-500/10'} hover:brightness-110`}
              >
                <Wrench size={10} />
                <span className="font-medium">{tc.name}</span>
                <ChevronDown size={10} className={`transition-transform ${expandedTools.has(tc.id) ? 'rotate-180' : ''}`} />
              </button>
              {expandedTools.has(tc.id) && (
                <div className="mt-1 ml-4 space-y-1">
                  {tc.input && (
                    <pre className="text-zinc-500 bg-zinc-800/50 p-2 rounded text-[11px] overflow-x-auto max-h-32 overflow-y-auto">
                      {tc.input}
                    </pre>
                  )}
                  {tc.result && (
                    <pre className="text-zinc-400 bg-zinc-800/30 p-2 rounded text-[11px] overflow-x-auto max-h-32 overflow-y-auto">
                      {tc.result}
                    </pre>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
