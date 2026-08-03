import type { HTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type Tone = 'accent' | 'success' | 'danger' | 'warn' | 'muted'

const tones: Record<Tone, string> = {
  accent: 'bg-accent/15 text-accent',
  success: 'bg-success/15 text-success',
  danger: 'bg-danger/15 text-danger',
  warn: 'bg-warn/15 text-warn',
  muted: 'bg-surface-2 text-muted',
}

export function Badge({ tone = 'muted', className, ...props }: HTMLAttributes<HTMLSpanElement> & { tone?: Tone }) {
  return (
    <span
      className={cn('inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium', tones[tone], className)}
      {...props}
    />
  )
}
