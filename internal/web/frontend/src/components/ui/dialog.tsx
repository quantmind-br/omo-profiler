import * as DialogPrimitive from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

export const Dialog = DialogPrimitive.Root
export const DialogTrigger = DialogPrimitive.Trigger

export function DialogContent({
  children,
  title,
  description,
  className,
}: {
  children: ReactNode
  title: ReactNode
  description?: ReactNode
  className?: string
}) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 z-[110] bg-black/60 backdrop-blur-sm" />
      <DialogPrimitive.Content
        className={cn(
          'fixed left-1/2 top-1/2 z-[111] w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-surface p-5 shadow-2xl focus:outline-none',
          className,
        )}
      >
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <DialogPrimitive.Title className="text-base font-semibold text-text">{title}</DialogPrimitive.Title>
            {description && (
              <DialogPrimitive.Description className="mt-1 text-sm text-muted">
                {description}
              </DialogPrimitive.Description>
            )}
          </div>
          <DialogPrimitive.Close className="text-muted hover:text-text" aria-label="Close">
            <X className="h-4 w-4" />
          </DialogPrimitive.Close>
        </div>
        {children}
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  )
}

export const DialogClose = DialogPrimitive.Close
