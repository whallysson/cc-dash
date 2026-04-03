import { api } from '../lib/api'
import { useApi } from '../hooks/useApi'

export function Component() {
  const { data, loading } = useApi(() => api.settings())

  if (loading || !data) return <div className="text-zinc-500">Carregando...</div>

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">Settings</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <pre className="text-xs text-zinc-400 whitespace-pre-wrap overflow-x-auto">
          {JSON.stringify(data.settings, null, 2)}
        </pre>
      </div>
      {data.plugins && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-semibold text-zinc-300 mb-2">Plugins</h3>
          <pre className="text-xs text-zinc-400 whitespace-pre-wrap">
            {JSON.stringify(data.plugins, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
