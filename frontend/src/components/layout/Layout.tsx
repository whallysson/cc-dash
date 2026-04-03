import { Outlet, NavLink } from 'react-router-dom'
import {
  LayoutDashboard, MessageSquare, FolderKanban, DollarSign,
  Wrench, Activity, Clock, Brain, ListTodo, FileText,
  Settings, Download
} from 'lucide-react'

const nav = [
  { to: '/', icon: LayoutDashboard, label: 'Overview' },
  { to: '/sessions', icon: MessageSquare, label: 'Sessions' },
  { to: '/projects', icon: FolderKanban, label: 'Projects' },
  { to: '/costs', icon: DollarSign, label: 'Costs' },
  { to: '/tools', icon: Wrench, label: 'Tools' },
  { to: '/activity', icon: Activity, label: 'Activity' },
  { to: '/history', icon: Clock, label: 'History' },
  { to: '/memory', icon: Brain, label: 'Memory' },
  { to: '/todos', icon: ListTodo, label: 'Todos' },
  { to: '/plans', icon: FileText, label: 'Plans' },
  { to: '/settings', icon: Settings, label: 'Settings' },
  { to: '/export', icon: Download, label: 'Export' },
]

export function Layout() {
  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <aside className="w-56 bg-zinc-900 border-r border-zinc-800 flex flex-col shrink-0">
        <div className="p-4 border-b border-zinc-800">
          <h1 className="text-lg font-bold text-orange-500">cc-dash</h1>
          <p className="text-xs text-zinc-500 mt-0.5">Claude Code Analytics</p>
        </div>
        <nav className="flex-1 overflow-y-auto py-2">
          {nav.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-4 py-2 text-sm transition-colors ${
                  isActive
                    ? 'text-orange-400 bg-orange-500/10 border-r-2 border-orange-500'
                    : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/50'
                }`
              }
            >
              <Icon size={16} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="p-3 border-t border-zinc-800 text-xs text-zinc-600">
          v0.1.0
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-y-auto">
        <div className="p-6 max-w-7xl mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
