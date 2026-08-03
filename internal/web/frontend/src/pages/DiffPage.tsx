import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { GitCompareArrows } from 'lucide-react'
import { api } from '../lib/api'
import type { DiffLine } from '../lib/types'
import { cn } from '../lib/utils'
import { Card } from '../components/ui/card'
import { Select } from '../components/ui/select'
import { Button } from '../components/ui/button'
import { Spinner } from '../components/ui/spinner'

export function DiffPage() {
  const profilesQ = useQuery({ queryKey: ['profiles'], queryFn: api.listProfiles })
  const [left, setLeft] = useState('__active__')
  const [right, setRight] = useState('')
  const [pair, setPair] = useState<{ left: string; right: string } | null>(null)

  const diffQ = useQuery({
    queryKey: ['diff', pair?.left, pair?.right],
    queryFn: () => api.diff(pair!.left, pair!.right),
    enabled: !!pair,
  })

  const options = [
    { value: '__active__', label: 'Active config' },
    ...(profilesQ.data?.profiles.map((p) => ({ value: p.name, label: p.name })) ?? []),
  ]

  return (
    <div className="mx-auto max-w-6xl space-y-5">
      <h1 className="text-xl font-semibold text-text">Compare</h1>

      <Card>
        <div className="flex flex-wrap items-end gap-3">
          <div className="w-56">
            <label className="mb-1 block text-sm text-muted">Left</label>
            <Select value={left} onValueChange={setLeft} options={options} />
          </div>
          <div className="w-56">
            <label className="mb-1 block text-sm text-muted">Right</label>
            <Select value={right || undefined} onValueChange={setRight} options={options} placeholder="Select…" />
          </div>
          <Button variant="primary" disabled={!left || !right} onClick={() => setPair({ left, right })}>
            <GitCompareArrows className="h-4 w-4" /> Compare
          </Button>
        </div>
      </Card>

      {diffQ.isLoading && (
        <div className="flex justify-center p-6">
          <Spinner className="h-6 w-6" />
        </div>
      )}
      {diffQ.isError && <p className="text-sm text-danger">{(diffQ.error as Error).message}</p>}
      {diffQ.data && (
        <div className="grid grid-cols-2 gap-4">
          <DiffPane label={diffQ.data.leftLabel === '__active__' ? 'Active config' : diffQ.data.leftLabel} lines={diffQ.data.left} />
          <DiffPane label={diffQ.data.rightLabel === '__active__' ? 'Active config' : diffQ.data.rightLabel} lines={diffQ.data.right} />
        </div>
      )}
    </div>
  )
}

function DiffPane({ label, lines }: { label: string; lines: DiffLine[] }) {
  return (
    <Card className="p-0">
      <div className="border-b border-border px-4 py-2 text-sm font-medium text-text">{label}</div>
      <pre className="overflow-auto scrollbar-thin p-0 text-xs">
        {lines.map((l, i) => (
          <div
            key={i}
            className={cn(
              'flex px-3 py-0.5 font-mono',
              l.type === 1 && 'bg-success/15',
              l.type === 2 && 'bg-danger/15',
            )}
          >
            <span className="mr-3 w-8 shrink-0 select-none text-right text-muted">{l.lineNum || ''}</span>
            <span
              className={cn(
                'whitespace-pre-wrap break-all',
                l.type === 1 && 'text-success',
                l.type === 2 && 'text-danger',
                l.type === 0 && 'text-text',
              )}
            >
              {l.text || '\u00a0'}
            </span>
          </div>
        ))}
      </pre>
    </Card>
  )
}
