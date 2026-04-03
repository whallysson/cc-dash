import './index.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { Layout } from './components/layout/Layout'

const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, lazy: () => import('./pages/Overview') },
      { path: 'sessions', lazy: () => import('./pages/Sessions') },
      { path: 'sessions/:id', lazy: () => import('./pages/SessionReplay') },
      { path: 'projects', lazy: () => import('./pages/Projects') },
      { path: 'projects/:slug', lazy: () => import('./pages/ProjectDetail') },
      { path: 'costs', lazy: () => import('./pages/Costs') },
      { path: 'tools', lazy: () => import('./pages/Tools') },
      { path: 'activity', lazy: () => import('./pages/Activity') },
      { path: 'history', lazy: () => import('./pages/History') },
      { path: 'memory', lazy: () => import('./pages/Memory') },
      { path: 'todos', lazy: () => import('./pages/Todos') },
      { path: 'plans', lazy: () => import('./pages/Plans') },
      { path: 'settings', lazy: () => import('./pages/Settings') },
      { path: 'export', lazy: () => import('./pages/Export') },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>
)
