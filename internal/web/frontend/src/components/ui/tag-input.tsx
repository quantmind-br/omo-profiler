import { useState, type KeyboardEvent } from 'react'
import { X } from 'lucide-react'
import { Input } from './input'

export function TagInput({
  value,
  onChange,
  placeholder = 'Add item and press Enter',
}: {
  value: string[]
  onChange: (v: string[]) => void
  placeholder?: string
}) {
  const [draft, setDraft] = useState('')

  function add() {
    const v = draft.trim()
    if (!v || value.includes(v)) {
      setDraft('')
      return
    }
    onChange([...value, v])
    setDraft('')
  }

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      e.preventDefault()
      add()
    } else if (e.key === 'Backspace' && draft === '' && value.length > 0) {
      onChange(value.slice(0, -1))
    }
  }

  return (
    <div className="rounded-lg border border-border bg-surface-2 p-2">
      {value.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {value.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1 rounded-md bg-accent/15 px-2 py-0.5 text-xs text-accent"
            >
              {tag}
              <button onClick={() => onChange(value.filter((t) => t !== tag))} aria-label={`Remove ${tag}`}>
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      )}
      <Input
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={onKeyDown}
        onBlur={add}
        placeholder={placeholder}
        className="h-7 border-0 bg-transparent px-1 focus:border-0"
      />
    </div>
  )
}
