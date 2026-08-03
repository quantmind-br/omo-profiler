import { forwardRef, type TextareaHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(
  ({ className, ...props }, ref) => (
    <textarea
      ref={ref}
      className={cn(
        'min-h-[80px] w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm text-text placeholder:text-muted focus:border-accent focus:outline-none',
        className,
      )}
      {...props}
    />
  ),
)
Textarea.displayName = 'Textarea'
