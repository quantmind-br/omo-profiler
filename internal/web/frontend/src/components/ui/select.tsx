import * as SelectPrimitive from '@radix-ui/react-select'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from '../../lib/utils'

export interface SelectOption {
  value: string
  label: string
}

export function Select({
  value,
  onValueChange,
  options,
  placeholder = 'Select…',
  className,
}: {
  value: string | undefined
  onValueChange: (v: string) => void
  options: SelectOption[]
  placeholder?: string
  className?: string
}) {
  return (
    <SelectPrimitive.Root value={value} onValueChange={onValueChange}>
      <SelectPrimitive.Trigger
        className={cn(
          'inline-flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-border bg-surface-2 px-3 text-sm text-text focus:border-accent focus:outline-none',
          className,
        )}
      >
        <SelectPrimitive.Value placeholder={placeholder} />
        <SelectPrimitive.Icon>
          <ChevronDown className="h-4 w-4 text-muted" />
        </SelectPrimitive.Icon>
      </SelectPrimitive.Trigger>
      <SelectPrimitive.Portal>
        <SelectPrimitive.Content
          position="popper"
          sideOffset={4}
          className="z-[120] max-h-72 overflow-hidden rounded-lg border border-border bg-surface-2 shadow-xl"
        >
          <SelectPrimitive.Viewport className="p-1">
            {options.map((opt) => (
              <SelectPrimitive.Item
                key={opt.value}
                value={opt.value}
                className="relative flex cursor-pointer select-none items-center rounded-md py-1.5 pl-8 pr-3 text-sm text-text outline-none data-[highlighted]:bg-accent/20"
              >
                <SelectPrimitive.ItemIndicator className="absolute left-2">
                  <Check className="h-4 w-4 text-accent" />
                </SelectPrimitive.ItemIndicator>
                <SelectPrimitive.ItemText>{opt.label}</SelectPrimitive.ItemText>
              </SelectPrimitive.Item>
            ))}
          </SelectPrimitive.Viewport>
        </SelectPrimitive.Content>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  )
}
