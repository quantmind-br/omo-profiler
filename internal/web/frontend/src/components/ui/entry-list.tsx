import { useEffect, useRef, useState, type ReactNode } from 'react'
import { ChevronDown, ChevronRight, Trash2 } from 'lucide-react'
import { cn } from '../../lib/utils'
import { Button } from './button'

// EntryList renders a collection of named entries as independently collapsible
// cards. All entries start collapsed (only the name/header shown); newly added
// keys auto-expand. A top toolbar toggles expand-all / collapse-all. Edit / add /
// remove logic stays with the caller via renderBody / onRemove.
export function EntryList({
  keys,
  countLabel,
  renderBody,
  onRemove,
}: {
  keys: string[]
  countLabel: string
  renderBody: (key: string) => ReactNode
  onRemove: (key: string) => void
}) {
  const [open, setOpen] = useState<Set<string>>(() => new Set())
  const prev = useRef<string[]>(keys)
  const keysSig = keys.join('\u0000')

  useEffect(() => {
    const before = prev.current
    setOpen((cur) => {
      const next = new Set<string>()
      for (const k of keys) if (cur.has(k) || !before.includes(k)) next.add(k)
      return next
    })
    prev.current = keys
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [keysSig])

  function toggle(key: string) {
    setOpen((cur) => {
      const next = new Set(cur)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted">
          {keys.length} {countLabel} configured.
        </div>
        {keys.length > 0 && (
          <div className="flex gap-1">
            <Button variant="ghost" size="sm" onClick={() => setOpen(new Set(keys))}>
              Expand all
            </Button>
            <Button variant="ghost" size="sm" onClick={() => setOpen(new Set())}>
              Collapse all
            </Button>
          </div>
        )}
      </div>

      {keys.map((key) => {
        const isOpen = open.has(key)
        return (
          <div key={key} className="rounded-lg border border-border">
            <div className={cn('flex items-center justify-between px-3 py-2', isOpen && 'border-b border-border')}>
              <button
                type="button"
                className="flex min-w-0 items-center gap-1.5 text-left"
                onClick={() => toggle(key)}
                aria-expanded={isOpen}
              >
                {isOpen ? (
                  <ChevronDown className="h-4 w-4 shrink-0 text-muted" />
                ) : (
                  <ChevronRight className="h-4 w-4 shrink-0 text-muted" />
                )}
                <span className="truncate font-mono text-sm text-accent">{key}</span>
              </button>
              <Button variant="ghost" size="icon" onClick={() => onRemove(key)} aria-label={`Remove ${key}`}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
            {isOpen && <div className="p-3">{renderBody(key)}</div>}
          </div>
        )
      })}
    </div>
  )
}
