import { api } from '@/lib/api'
import { useApi } from '@/hooks/useApi'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Settings2, Puzzle } from 'lucide-react'

export function Component() {
  const { data, loading } = useApi(() => api.settings())

  if (loading || !data) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-[400px] w-full rounded-xl" />
      </div>
    )
  }

  const pluginList = Array.isArray(data.plugins) ? data.plugins : []

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Settings</h2>
        <p className="text-sm text-muted-foreground">Claude Code configuration</p>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center gap-2">
          <Settings2 className="size-4 text-muted-foreground" />
          <CardTitle>Configuration</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="text-xs text-muted-foreground whitespace-pre-wrap overflow-x-auto leading-relaxed font-mono bg-muted/30 rounded-lg p-4">
            {JSON.stringify(data.settings, null, 2)}
          </pre>
        </CardContent>
      </Card>

      {pluginList.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center gap-2">
            <Puzzle className="size-4 text-muted-foreground" />
            <div className="flex items-center gap-2">
              <CardTitle>Plugins</CardTitle>
              <Badge variant="secondary">{pluginList.length}</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <pre className="text-xs text-muted-foreground whitespace-pre-wrap leading-relaxed font-mono bg-muted/30 rounded-lg p-4">
              {JSON.stringify(pluginList, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
