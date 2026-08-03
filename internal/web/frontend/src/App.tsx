import { NavLink, Navigate, Route, Routes } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Boxes, GitCompareArrows, LayoutDashboard, ListChecks, ShieldCheck, Cpu } from 'lucide-react'
import { api } from './lib/api'
import { cn } from './lib/utils'
import { Badge } from './components/ui/badge'
import { DashboardPage } from './pages/DashboardPage'
import { ProfilesPage } from './pages/ProfilesPage'
import { EditorPage } from './pages/EditorPage'
import { ModelsPage } from './pages/ModelsPage'
import { DiffPage } from './pages/DiffPage'
import { SchemaCheckPage } from './pages/SchemaCheckPage'

const NAV = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/profiles', label: 'Profiles', icon: ListChecks, end: false },
  { to: '/models', label: 'Models', icon: Cpu, end: false },
  { to: '/diff', label: 'Compare', icon: GitCompareArrows, end: false },
  { to: '/schema-check', label: 'Schema', icon: ShieldCheck, end: false },
]

export default function App() {
  const { data: active } = useQuery({ queryKey: ['active'], queryFn: api.getActive })

  return (
    <div className="flex h-full">
      <aside className="flex w-60 shrink-0 flex-col border-r border-border bg-surface">
        <div className="flex items-center gap-2 px-5 py-4">
          <Boxes className="h-6 w-6 text-accent" />
          <div className="leading-tight">
            <div className="text-sm font-semibold text-text">omo-profiler</div>
            <div className="text-[11px] text-muted">omo.json profiles</div>
          </div>
        </div>
        <nav className="flex-1 space-y-1 px-3 py-2">
          {NAV.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors',
                  isActive ? 'bg-surface-2 text-text' : 'text-muted hover:text-text hover:bg-surface-2',
                )
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center justify-between border-b border-border px-6">
          <div className="text-sm text-muted">Profile manager</div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted">Active:</span>
            {active?.modified ? (
              <Badge tone="warn">custom (matches no profile)</Badge>
            ) : active?.profileName ? (
              <Badge tone="success">{active.profileName}</Badge>
            ) : (
              <Badge tone="muted">none</Badge>
            )}
          </div>
        </header>

        <main className="min-h-0 flex-1 overflow-auto scrollbar-thin p-6">
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/profiles" element={<ProfilesPage />} />
            <Route path="/profiles/:name/edit" element={<EditorPage />} />
            <Route path="/models" element={<ModelsPage />} />
            <Route path="/diff" element={<DiffPage />} />
            <Route path="/schema-check" element={<SchemaCheckPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  )
}
