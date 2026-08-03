import * as TooltipPrimitive from '@radix-ui/react-tooltip'
import type { ReactNode } from 'react'

export function Tooltip({ content, children }: { content: ReactNode; children: ReactNode }) {
  return (
    <TooltipPrimitive.Provider delayDuration={300}>
      <TooltipPrimitive.Root>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            sideOffset={4}
            className="z-[130] max-w-xs rounded-md border border-border bg-surface-2 px-2 py-1 text-xs text-text shadow-lg"
          >
            {content}
            <TooltipPrimitive.Arrow className="fill-surface-2" />
          </TooltipPrimitive.Content>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  )
}
