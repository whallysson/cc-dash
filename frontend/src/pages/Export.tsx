import { useState } from 'react'
import { Download, FileJson } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export function Component() {
  const [exporting, setExporting] = useState(false)

  const handleExport = async () => {
    setExporting(true)
    try {
      const res = await fetch('/api/export', { method: 'POST' })
      if (!res.ok) throw new Error(`Export failed: ${res.status}`)
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'cc-dash-export.json'
      a.click()
      URL.revokeObjectURL(url)
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Export</h2>
        <p className="text-sm text-muted-foreground">Download your data</p>
      </div>

      <div className="flex items-center justify-center py-12">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle>Export Data</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4 pb-8">
            <div className="flex size-14 items-center justify-center rounded-full bg-primary/10">
              <FileJson className="size-6 text-primary" />
            </div>
            <p className="text-sm text-muted-foreground text-center max-w-xs">
              Export all session data, analytics and configuration as a JSON file.
            </p>
            <Button onClick={handleExport} disabled={exporting} size="lg">
              <Download className="size-4 mr-2" />
              {exporting ? 'Exporting...' : 'Download JSON'}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
