import { cn } from '../../lib/utils'

export function MultiSelect({
  value,
  options,
  onChange,
}: {
  value: string[]
  options: string[]
  onChange: (v: string[]) => void
}) {
  function toggle(opt: string) {
    if (value.includes(opt)) {
      onChange(value.filter((v) => v !== opt))
    } else {
      onChange([...value, opt])
    }
  }

  if (options.length === 0) {
    return <p className="text-xs text-muted">No options available.</p>
  }

  return (
    <div className="flex flex-wrap gap-1.5">
      {options.map((opt) => {
        const active = value.includes(opt)
        return (
          <button
            key={opt}
            type="button"
            onClick={() => toggle(opt)}
            className={cn(
              'rounded-md border px-2 py-1 text-xs transition-colors',
              active
                ? 'border-accent bg-accent/15 text-accent'
                : 'border-border bg-surface-2 text-muted hover:text-text',
            )}
          >
            {opt}
          </button>
        )
      })}
    </div>
  )
}
