import { useState, type ReactNode } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { Button } from './button'
import { Input } from './input'

// KeyValueEditor renders a map as key/value rows. The value cell is rendered
// via renderValue so callers (e.g. SchemaForm) can recurse into arbitrary value
// schemas; omit it for a plain string map.
export function KeyValueEditor({
  entries,
  onAdd,
  onRemove,
  renderValue,
}: {
  entries: Array<[string, unknown]>
  onAdd: (key: string) => void
  onRemove: (key: string) => void
  renderValue: (key: string, value: unknown) => ReactNode
}) {
  const [newKey, setNewKey] = useState('')

  function add() {
    const k = newKey.trim()
    if (!k || entries.some(([ek]) => ek === k)) return
    onAdd(k)
    setNewKey('')
  }

  return (
    <div className="space-y-2">
      {entries.map(([key, value]) => (
        <div key={key} className="flex items-start gap-2">
          <div className="w-40 shrink-0 pt-2 font-mono text-xs text-accent">{key}</div>
          <div className="min-w-0 flex-1">{renderValue(key, value)}</div>
          <Button variant="ghost" size="icon" onClick={() => onRemove(key)} aria-label={`Remove ${key}`}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
      <div className="flex gap-2">
        <Input
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), add())}
          placeholder="New key…"
          className="w-40"
        />
        <Button variant="secondary" size="sm" onClick={add}>
          <Plus className="h-4 w-4" /> Add
        </Button>
      </div>
    </div>
  )
}
