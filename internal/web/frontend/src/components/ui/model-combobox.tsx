import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, X } from 'lucide-react'
import { api } from '../../lib/api'
import type { RegisteredModel } from '../../lib/types'
import { cn } from '../../lib/utils'
import { Input } from './input'

function modelValue(m: RegisteredModel): string {
  return m.provider ? `${m.provider}/${m.modelId}` : m.modelId
}

// ModelCombobox is a searchable dropdown for a single `model` string field,
// sourced from the registry served by GET /api/models. It preserves any
// current value not present in the registry and always allows custom entry so
// editing is never blocked when the registry is empty or the request fails.
export function ModelCombobox({
  value,
  onChange,
  placeholder = 'Select a model…',
}: {
  value: string | undefined
  onChange: (v: string | undefined) => void
  placeholder?: string
}): JSX.Element {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  // Same queryKey as ModelsPage so react-query dedupes to one request even
  // with many combobox instances mounted at once.
  const { data, isLoading, isError } = useQuery({ queryKey: ['models'], queryFn: api.listModels })
  const models: RegisteredModel[] = data?.groups.flatMap((g) => g.models) ?? []

  useEffect(() => {
    if (!open) return
    function onMouseDown(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
        setQuery('')
      }
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        setOpen(false)
        setQuery('')
      }
    }
    document.addEventListener('mousedown', onMouseDown)
    document.addEventListener('keydown', onKeyDown)
    return () => {
      document.removeEventListener('mousedown', onMouseDown)
      document.removeEventListener('keydown', onKeyDown)
    }
  }, [open])

  const q = query.trim().toLowerCase()
  const filtered =
    q === ''
      ? models
      : models.filter(
          (m) =>
            m.displayName.toLowerCase().includes(q) ||
            m.modelId.toLowerCase().includes(q) ||
            m.provider.toLowerCase().includes(q),
        )

  const selectedModel = value ? models.find((m) => modelValue(m) === value || m.modelId === value) : undefined
  const isCustom = value !== undefined && value !== '' && !selectedModel
  const showPin = isCustom && (q === '' || (value ?? '').toLowerCase().includes(q))

  const trimmed = query.trim()
  const showCustomEntry =
    trimmed !== '' && !models.some((m) => modelValue(m) === trimmed || m.modelId === trimmed) && trimmed !== value

  return (
    <div ref={ref} className="relative">
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => setOpen((o) => !o)}
          className="inline-flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-border bg-surface-2 px-3 text-sm text-text focus:border-accent focus:outline-none"
        >
          <span className={cn('truncate', !value && 'text-muted')}>
            {selectedModel ? selectedModel.displayName : value ? value : placeholder}
          </span>
          <ChevronDown className="h-4 w-4 shrink-0 text-muted" />
        </button>
        {value !== undefined && (
          <button
            type="button"
            className="text-muted hover:text-text"
            onClick={() => onChange(undefined)}
            aria-label="Clear"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {open && (
        <div className="absolute z-[120] mt-1 max-h-72 w-full overflow-auto rounded-lg border border-border bg-surface-2 p-1 shadow-xl">
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
            placeholder="Search or type a model…"
            className="mb-1"
          />
          {isLoading && <div className="px-2 py-1.5 text-xs text-muted">Loading models…</div>}
          {showPin && (
            <button
              type="button"
              onClick={() => {
                setOpen(false)
                setQuery('')
              }}
              className="flex w-full items-center gap-2 rounded-md border border-accent bg-accent/15 px-2 py-1.5 text-left text-sm text-accent"
            >
              <span className="truncate">{value}</span>
              <span className="shrink-0 text-xs text-muted">· current (not in registry)</span>
            </button>
          )}
          {filtered.map((m) => (
            <button
              key={`${m.provider}/${m.modelId}`}
              type="button"
              onClick={() => {
                onChange(modelValue(m))
                setOpen(false)
                setQuery('')
              }}
              className={cn(
                'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-accent/20',
                modelValue(m) === value || m.modelId === value ? 'bg-accent/15 text-accent' : 'text-text',
              )}
            >
              <span className="truncate">{m.displayName}</span>
              <span className="shrink-0 font-mono text-xs text-muted">{m.modelId}</span>
              <span className="shrink-0 text-xs text-muted">{m.provider}</span>
            </button>
          ))}
          {showCustomEntry && (
            <button
              type="button"
              onClick={() => {
                onChange(trimmed)
                setOpen(false)
                setQuery('')
              }}
              className="flex w-full items-center rounded-md px-2 py-1.5 text-left text-sm text-text hover:bg-accent/20"
            >
              Use &quot;{trimmed}&quot;
            </button>
          )}
          {!isLoading && filtered.length === 0 && query === '' && (
            <div className="px-2 py-1.5 text-xs text-muted">
              {isError
                ? 'Model registry unavailable — type to set a value.'
                : 'No registered models — type to set a value.'}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
