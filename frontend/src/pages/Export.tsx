import { useState } from 'react'
import { Download } from 'lucide-react'

export function Component() {
  const [exporting, setExporting] = useState(false)

  const handleExport = async () => {
    setExporting(true)
    try {
      const res = await fetch('/api/export', { method: 'POST' })
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
    <div className="space-y-4">
      <h2 className="text-xl font-bold">Export</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-center">
        <p className="text-sm text-zinc-400 mb-4">
          Export all session data and stats as a JSON file.
        </p>
        <button
          onClick={handleExport}
          disabled={exporting}
          className="inline-flex items-center gap-2 px-4 py-2 bg-orange-500 text-white rounded-lg hover:bg-orange-600 disabled:opacity-50 transition-colors"
        >
          <Download size={16} />
          {exporting ? 'Exporting...' : 'Download JSON'}
        </button>
      </div>
    </div>
  )
}
