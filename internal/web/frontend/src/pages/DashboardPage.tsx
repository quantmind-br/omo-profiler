import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowRight, Cpu, GitCompareArrows, ListChecks, ShieldCheck } from 'lucide-react'
import { api } from '../lib/api'
import { Card, CardHeader } from '../components/ui/card'
import { Badge } from '../components/ui/badge'
import { Spinner } from '../components/ui/spinner'

const ACTIONS = [
  { to: '/profiles', title: 'Profiles', desc: 'Switch, create, clone, rename, edit, import & export', icon: ListChecks },
  { to: '/models', title: 'Model registry', desc: 'Manage models and import from models.dev', icon: Cpu },
  { to: '/diff', title: 'Compare', desc: 'Side-by-side diff of any two profiles', icon: GitCompareArrows },
  { to: '/schema-check', title: 'Schema check', desc: 'Compare embedded schema against upstream', icon: ShieldCheck },
]

export function DashboardPage() {
  const active = useQuery({ queryKey: ['active'], queryFn: api.getActive })
  const profiles = useQuery({ queryKey: ['profiles'], queryFn: api.listProfiles })

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-text">Dashboard</h1>
        <p className="mt-1 text-sm text-muted">Manage profiles in your ~/.omo/omo.json document.</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader title="Active profile" />
          {active.isLoading ? (
            <Spinner />
          ) : active.data?.modified ? (
            <div className="flex items-center gap-2">
              <span className="text-lg font-medium text-text">custom</span>
              <Badge tone="warn">matches no profile</Badge>
            </div>
          ) : active.data?.profileName ? (
            <div className="flex items-center gap-2">
              <span className="text-lg font-medium text-text">{active.data.profileName}</span>
              <Badge tone="success">applied</Badge>
            </div>
          ) : (
            <p className="text-sm text-muted">No profile applied.</p>
          )}
        </Card>
        <Card>
          <CardHeader title="Profiles" />
          <div className="text-lg font-medium text-text">
            {profiles.isLoading ? <Spinner /> : `${profiles.data?.profiles.length ?? 0} saved`}
          </div>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {ACTIONS.map((a) => (
          <Link key={a.to} to={a.to}>
            <Card className="group h-full transition-colors hover:border-accent">
              <div className="flex items-start justify-between">
                <a.icon className="h-6 w-6 text-accent" />
                <ArrowRight className="h-4 w-4 text-muted transition-transform group-hover:translate-x-0.5" />
              </div>
              <div className="mt-3 text-sm font-semibold text-text">{a.title}</div>
              <div className="mt-1 text-xs text-muted">{a.desc}</div>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}
