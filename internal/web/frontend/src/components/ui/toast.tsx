import { createContext, useCallback, useContext, useState, type ReactNode } from 'react'
import { CheckCircle2, XCircle, Info, X } from 'lucide-react'
import { cn } from '../../lib/utils'

type ToastVariant = 'success' | 'error' | 'info'

interface Toast {
  id: number
  title: string
  description?: string
  variant: ToastVariant
}

interface ToastContextValue {
  toast: (t: { title: string; description?: string; variant?: ToastVariant }) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}

let nextId = 1

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const toast = useCallback(
    (t: { title: string; description?: string; variant?: ToastVariant }) => {
      const id = nextId++
      const item: Toast = { id, title: t.title, description: t.description, variant: t.variant ?? 'info' }
      setToasts((prev) => [...prev, item])
      setTimeout(() => remove(id), 5000)
    },
    [remove],
  )

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className="fixed top-4 right-4 z-[100] flex w-80 flex-col gap-2">
        {toasts.map((t) => (
          <ToastCard key={t.id} toast={t} onClose={() => remove(t.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

function ToastCard({ toast, onClose }: { toast: Toast; onClose: () => void }) {
  const Icon = toast.variant === 'success' ? CheckCircle2 : toast.variant === 'error' ? XCircle : Info
  const color =
    toast.variant === 'success' ? 'text-success' : toast.variant === 'error' ? 'text-danger' : 'text-accent'
  return (
    <div className="flex items-start gap-3 rounded-lg border border-border bg-surface-2 p-3 shadow-lg">
      <Icon className={cn('mt-0.5 h-5 w-5 shrink-0', color)} />
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium text-text">{toast.title}</div>
        {toast.description && <div className="mt-0.5 break-words text-xs text-muted">{toast.description}</div>}
      </div>
      <button onClick={onClose} className="text-muted hover:text-text" aria-label="Dismiss">
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}
