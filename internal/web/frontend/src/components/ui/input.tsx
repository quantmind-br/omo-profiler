import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        'h-9 w-full rounded-lg border border-border bg-surface-2 px-3 text-sm text-text placeholder:text-muted focus:border-accent focus:outline-none',
        className,
      )}
      {...props}
    />
  ),
)
Input.displayName = 'Input'
