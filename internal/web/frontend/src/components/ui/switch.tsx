import * as SwitchPrimitive from '@radix-ui/react-switch'
import { cn } from '../../lib/utils'

export function Switch({
  checked,
  onCheckedChange,
  className,
}: {
  checked: boolean
  onCheckedChange: (v: boolean) => void
  className?: string
}) {
  return (
    <SwitchPrimitive.Root
      checked={checked}
      onCheckedChange={onCheckedChange}
      className={cn(
        'relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border border-border transition-colors',
        checked ? 'bg-accent' : 'bg-surface-2',
        className,
      )}
    >
      <SwitchPrimitive.Thumb className="block h-3.5 w-3.5 translate-x-1 rounded-full bg-white transition-transform data-[state=checked]:translate-x-[18px]" />
    </SwitchPrimitive.Root>
  )
}
