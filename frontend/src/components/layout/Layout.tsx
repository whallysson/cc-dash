import { Outlet, NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard, MessageSquare, FolderKanban, DollarSign,
  Wrench, Activity, Clock, Brain, ListTodo, FileText,
  Settings, Download, Terminal, HeartPulse, Star
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Sidebar, SidebarContent, SidebarFooter, SidebarGroup,
  SidebarGroupContent, SidebarGroupLabel, SidebarHeader,
  SidebarInset, SidebarMenu, SidebarMenuButton, SidebarMenuItem,
  SidebarProvider, SidebarRail,
} from '@/components/ui/sidebar'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ThemeToggle } from '@/components/theme-toggle'
import { cn } from '@/lib/utils'

const navAnalytics = [
  { to: '/', icon: LayoutDashboard, label: 'Overview' },
  { to: '/token-health', icon: HeartPulse, label: 'Token Health' },
  { to: '/sessions', icon: MessageSquare, label: 'Sessions' },
  { to: '/projects', icon: FolderKanban, label: 'Projects' },
  { to: '/costs', icon: DollarSign, label: 'Costs' },
  { to: '/tools', icon: Wrench, label: 'Tools' },
  { to: '/activity', icon: Activity, label: 'Activity' },
]

const navData = [
  { to: '/history', icon: Clock, label: 'History' },
  { to: '/memory', icon: Brain, label: 'Memory' },
  { to: '/todos', icon: ListTodo, label: 'Todos' },
  { to: '/plans', icon: FileText, label: 'Plans' },
]

const navOther = [
  { to: '/settings', icon: Settings, label: 'Settings' },
  { to: '/export', icon: Download, label: 'Export' },
]

function NavGroup({ label, items }: { label: string; items: typeof navAnalytics }) {
  const location = useLocation()

  const isActive = (to: string) => {
    if (to === '/') return location.pathname === '/'
    return location.pathname.startsWith(to)
  }

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{label}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map(({ to, icon: Icon, label }) => (
            <SidebarMenuItem key={to}>
              <SidebarMenuButton asChild isActive={isActive(to)} tooltip={label}>
                <NavLink to={to}>
                  <Icon />
                  <span>{label}</span>
                </NavLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}

function AppSidebar() {
  return (
    <Sidebar collapsible="icon" variant="inset">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <NavLink to="/">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                  <Terminal className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">cc-dash</span>
                  <span className="truncate text-xs text-muted-foreground">
                    Claude Code Analytics
                  </span>
                </div>
              </NavLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <NavGroup label="Analytics" items={navAnalytics} />
        <NavGroup label="Data" items={navData} />
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          {navOther.map(({ to, icon: Icon, label }) => (
            <SidebarMenuItem key={to}>
              <SidebarMenuButton asChild tooltip={label}>
                <NavLink to={to}>
                  <Icon />
                  <span>{label}</span>
                </NavLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
        <div className="px-2 py-1 text-[10px] text-muted-foreground/50 group-data-[collapsible=icon]:hidden">
          v0.1.0
        </div>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}

function PageTitle() {
  const location = useLocation()
  const titles: Record<string, string> = {
    '/': 'Overview',
    '/sessions': 'Sessions',
    '/projects': 'Projects',
    '/costs': 'Costs',
    '/tools': 'Tools',
    '/activity': 'Activity',
    '/token-health': 'Token Health',
    '/history': 'History',
    '/memory': 'Memory',
    '/todos': 'Todos',
    '/plans': 'Plans',
    '/settings': 'Settings',
    '/export': 'Export',
  }

  const path = location.pathname
  const title = titles[path] ||
    (path.startsWith('/sessions/') ? 'Session Replay' :
     path.startsWith('/projects/') ? 'Project Detail' : 'cc-dash')

  return (
    <h1 className="text-sm font-medium text-foreground">{title}</h1>
  )
}

export function Layout() {
  const defaultOpen = typeof document !== 'undefined'
    ? document.cookie.includes('sidebar_state=true') || !document.cookie.includes('sidebar_state=false')
    : true

  return (
    <SidebarProvider defaultOpen={defaultOpen}>
      <AppSidebar />
      <SidebarInset
        className={cn(
          '@container/content',
          'has-data-[layout=fixed]:h-svh',
          'peer-data-[variant=inset]:has-data-[layout=fixed]:h-[calc(100svh-(var(--spacing)*4))]'
        )}
      >
        <Header fixed>
          <PageTitle />
          <div className="ml-auto flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              asChild
              className="gap-1.5 text-amber-500 border-amber-500/40 hover:bg-amber-500/10 hover:text-amber-400 hover:border-amber-500/60"
            >
              <a href="https://github.com/whallysson/cc-dash" target="_blank" rel="noopener noreferrer">
                <Star className="size-3.5 fill-amber-500" />
                Star on GitHub
              </a>
            </Button>
            <ThemeToggle />
          </div>
        </Header>
        <Main>
          <Outlet />
        </Main>
      </SidebarInset>
    </SidebarProvider>
  )
}
