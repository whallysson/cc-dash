import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'

export function Component() {
  const { data, loading } = useApi(() => api.activity())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  // Gerar grid de 52 semanas x 7 dias para heatmap
  const dateMap = new Map(data.heatmap.map(h => [h.date, h.count]))
  const maxCount = Math.max(...data.heatmap.map(h => h.count), 1)

  const today = new Date()
  const weeks: { date: string; count: number; dow: number }[][] = []
  let currentWeek: { date: string; count: number; dow: number }[] = []

  // Voltar 364 dias (52 semanas)
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
    if (count === 0) return 'bg-zinc-800/50'
    const intensity = count / maxCount
    if (intensity > 0.75) return 'bg-orange-500'
    if (intensity > 0.5) return 'bg-orange-600'
    if (intensity > 0.25) return 'bg-orange-700'
    return 'bg-orange-800'
  }

  const dayLabels = ['', 'Mon', '', 'Wed', '', 'Fri', '']

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold">Activity</h2>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-3">
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <div className="text-xs text-zinc-500">Current Streak</div>
          <div className="text-2xl font-bold text-orange-400">{data.current_streak}d</div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <div className="text-xs text-zinc-500">Longest Streak</div>
          <div className="text-2xl font-bold text-zinc-200">{data.longest_streak}d</div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <div className="text-xs text-zinc-500">Active Days</div>
          <div className="text-2xl font-bold text-zinc-200">{data.active_days}</div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <div className="text-xs text-zinc-500">Total Days</div>
          <div className="text-2xl font-bold text-zinc-200">{data.total_days}</div>
        </div>
      </div>

      {/* Heatmap */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-zinc-300 mb-3">Activity Heatmap</h3>
        <div className="flex gap-1">
          <div className="flex flex-col gap-1 mr-1">
            {dayLabels.map((l, i) => (
              <div key={i} className="h-3 text-[10px] text-zinc-600 leading-3">{l}</div>
            ))}
          </div>
          {weeks.map((week, wi) => (
            <div key={wi} className="flex flex-col gap-1">
              {Array.from({ length: 7 }, (_, dow) => {
                const day = week.find(d => d.dow === dow)
                if (!day) return <div key={dow} className="w-3 h-3" />
                return (
                  <div
                    key={dow}
                    className={`w-3 h-3 rounded-sm ${getColor(day.count)} group relative`}
                  >
                    <div className="absolute bottom-full mb-1 left-1/2 -translate-x-1/2 bg-zinc-700 text-[10px] px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap z-10">
                      {day.date}: {day.count}
                    </div>
                  </div>
                )
              })}
            </div>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Day of Week */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Day of Week</h3>
          <div className="space-y-2">
            {['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'].map(day => {
              const count = data.day_of_week[day] || 0
              const max = Math.max(...Object.values(data.day_of_week), 1)
              return (
                <div key={day} className="flex items-center gap-2">
                  <span className="w-12 text-xs text-zinc-500">{day.slice(0, 3)}</span>
                  <div className="flex-1 h-2 bg-zinc-800 rounded-full overflow-hidden">
                    <div className="h-full bg-orange-500/70 rounded-full" style={{ width: `${(count / max) * 100}%` }} />
                  </div>
                  <span className="w-8 text-right text-xs text-zinc-400">{count}</span>
                </div>
              )
            })}
          </div>
        </div>

        {/* Peak Hours */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-3">Peak Hours</h3>
          <div className="flex items-end gap-1 h-32">
            {Array.from({ length: 24 }, (_, h) => {
              const key = String(h).padStart(2, '0')
              const count = data.peak_hours[key] || data.peak_hours[String(h)] || 0
              const max = Math.max(...Object.values(data.peak_hours), 1)
              return (
                <div key={h} className="flex-1 group relative">
                  <div className="bg-orange-500/60 hover:bg-orange-400 rounded-t transition-colors" style={{ height: `${(count / max) * 100}%` }} />
                  <div className="absolute bottom-full mb-1 left-1/2 -translate-x-1/2 bg-zinc-800 text-[10px] px-1 py-0.5 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap z-10">
                    {key}:00 ({count})
                  </div>
                </div>
              )
            })}
          </div>
          <div className="flex justify-between text-[10px] text-zinc-600 mt-1">
            <span>00</span><span>06</span><span>12</span><span>18</span><span>23</span>
          </div>
        </div>
      </div>
    </div>
  )
}
