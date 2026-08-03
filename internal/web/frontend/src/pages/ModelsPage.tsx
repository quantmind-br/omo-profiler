import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Download, Pencil, Plus, Trash2 } from 'lucide-react'
import { api, ApiError } from '../lib/api'
import type { CatalogModel, RegisteredModel } from '../lib/types'
import { cn } from '../lib/utils'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Badge } from '../components/ui/badge'
import { Spinner } from '../components/ui/spinner'
import { Dialog, DialogContent } from '../components/ui/dialog'
import { useToast } from '../components/ui/toast'

export function ModelsPage() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { data, isLoading } = useQuery({ queryKey: ['models'], queryFn: api.listModels })

  const [editModel, setEditModel] = useState<RegisteredModel | null>(null)
  const [addOpen, setAddOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)

  function refresh() {
    qc.invalidateQueries({ queryKey: ['models'] })
  }

  const del = useMutation({
    mutationFn: (m: RegisteredModel) => api.deleteModel(m.provider, m.modelId),
    onSuccess: (_res, m) => {
      toast({ title: `Removed ${m.displayName}`, variant: 'success' })
      refresh()
    },
    onError: (e: Error) => toast({ title: 'Delete failed', description: e.message, variant: 'error' }),
  })

  return (
    <div className="mx-auto max-w-4xl space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Model registry</h1>
        <div className="flex gap-2">
          <Button variant="secondary" onClick={() => setImportOpen(true)}>
            <Download className="h-4 w-4" /> Import from models.dev
          </Button>
          <Button variant="primary" onClick={() => setAddOpen(true)}>
            <Plus className="h-4 w-4" /> Add model
          </Button>
        </div>
      </div>

      {isLoading ? (
        <Spinner />
      ) : data && data.groups.length > 0 ? (
        <div className="space-y-4">
          {data.groups.map((g) => (
            <Card key={g.provider || '_'} className="p-0">
              <div className="border-b border-border px-4 py-2 text-sm font-medium text-text">
                {g.provider || 'No provider'}
              </div>
              <ul className="divide-y divide-border">
                {g.models.map((m) => (
                  <li key={`${m.provider}/${m.modelId}`} className="flex items-center justify-between px-4 py-2.5">
                    <div className="min-w-0">
                      <div className="truncate text-sm text-text">{m.displayName}</div>
                      <div className="truncate font-mono text-xs text-muted">{m.modelId}</div>
                    </div>
                    <div className="flex gap-1">
                      <Button size="sm" variant="ghost" onClick={() => setEditModel(m)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button size="sm" variant="ghost" className="text-danger" onClick={() => del.mutate(m)}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </li>
                ))}
              </ul>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <p className="text-sm text-muted">No models registered yet.</p>
        </Card>
      )}

      {addOpen && (
        <ModelDialog
          title="Add model"
          onClose={() => setAddOpen(false)}
          onSubmit={async (m) => {
            await api.createModel(m)
            toast({ title: `Added ${m.displayName}`, variant: 'success' })
            refresh()
            setAddOpen(false)
          }}
        />
      )}

      {editModel && (
        <ModelDialog
          title="Edit model"
          initial={editModel}
          onClose={() => setEditModel(null)}
          onSubmit={async (m) => {
            await api.updateModel(editModel.provider, editModel.modelId, m)
            toast({ title: `Updated ${m.displayName}`, variant: 'success' })
            refresh()
            setEditModel(null)
          }}
        />
      )}

      {importOpen && <ImportDialog onClose={() => setImportOpen(false)} onDone={refresh} />}
    </div>
  )
}

function ModelDialog({
  title,
  initial,
  onClose,
  onSubmit,
}: {
  title: string
  initial?: RegisteredModel
  onClose: () => void
  onSubmit: (m: RegisteredModel) => Promise<void>
}) {
  const [displayName, setDisplayName] = useState(initial?.displayName ?? '')
  const [modelId, setModelId] = useState(initial?.modelId ?? '')
  const [provider, setProvider] = useState(initial?.provider ?? '')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit() {
    setBusy(true)
    setError(null)
    try {
      await onSubmit({ displayName: displayName.trim(), modelId: modelId.trim(), provider: provider.trim() })
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent title={title}>
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-sm text-muted">Display name</label>
            <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} autoFocus />
          </div>
          <div>
            <label className="mb-1 block text-sm text-muted">Model ID</label>
            <Input value={modelId} onChange={(e) => setModelId(e.target.value)} placeholder="provider/model-name" />
          </div>
          <div>
            <label className="mb-1 block text-sm text-muted">Provider</label>
            <Input value={provider} onChange={(e) => setProvider(e.target.value)} />
          </div>
          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" onClick={submit} disabled={busy || !displayName.trim() || !modelId.trim()}>
              {busy ? <Spinner /> : 'Save'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function ImportDialog({ onClose, onDone }: { onClose: () => void; onDone: () => void }) {
  const { toast } = useToast()
  const catalogQ = useQuery({ queryKey: ['catalog'], queryFn: api.modelsCatalog })
  const [providerId, setProviderId] = useState<string>('')
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Map<string, CatalogModel & { providerId: string }>>(new Map())
  const [importing, setImporting] = useState(false)

  const provider = catalogQ.data?.providers.find((p) => p.id === providerId)
  const filtered = useMemo(() => {
    const models = provider?.models ?? []
    const q = search.trim().toLowerCase()
    if (!q) return models
    return models.filter((m) => m.name.toLowerCase().includes(q) || m.id.toLowerCase().includes(q))
  }, [provider, search])

  function toggle(m: CatalogModel) {
    const key = `${providerId}/${m.id}`
    setSelected((prev) => {
      const next = new Map(prev)
      if (next.has(key)) next.delete(key)
      else next.set(key, { ...m, providerId })
      return next
    })
  }

  async function doImport() {
    const items = [...selected.values()]
    if (items.length === 0) return
    setImporting(true)
    let added = 0
    let skipped = 0
    for (const m of items) {
      try {
        await api.createModel({ displayName: m.name, modelId: m.id, provider: m.providerId })
        added++
      } catch (e) {
        if (e instanceof ApiError && e.status === 409) skipped++
        else throw e
      }
    }
    setImporting(false)
    toast({ title: `Imported ${added} model(s)`, description: skipped ? `${skipped} duplicate(s) skipped` : undefined, variant: 'success' })
    onDone()
    onClose()
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent title="Import from models.dev" className="max-w-2xl" description="Pick a provider, then select models to add to your registry.">
        {catalogQ.isLoading ? (
          <div className="flex justify-center p-6">
            <Spinner className="h-6 w-6" />
          </div>
        ) : catalogQ.isError ? (
          <p className="text-sm text-danger">{(catalogQ.error as Error).message}</p>
        ) : (
          <div className="space-y-3">
            <div className="flex gap-2">
              <div className="w-56">
                <Select
                  value={providerId || undefined}
                  onValueChange={(v) => {
                    setProviderId(v)
                    setSearch('')
                  }}
                  placeholder="Provider…"
                  options={(catalogQ.data?.providers ?? []).map((p) => ({ value: p.id, label: `${p.name} (${p.models.length})` }))}
                />
              </div>
              <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search models…" disabled={!providerId} />
            </div>

            <div className="max-h-80 overflow-auto scrollbar-thin rounded-lg border border-border">
              {!providerId ? (
                <p className="p-4 text-sm text-muted">Select a provider to list its models.</p>
              ) : filtered.length === 0 ? (
                <p className="p-4 text-sm text-muted">No models match.</p>
              ) : (
                <ul className="divide-y divide-border">
                  {filtered.map((m) => {
                    const key = `${providerId}/${m.id}`
                    const isSel = selected.has(key)
                    return (
                      <li
                        key={key}
                        onClick={() => toggle(m)}
                        className={cn('flex cursor-pointer items-center justify-between px-3 py-2', isSel && 'bg-accent/10')}
                      >
                        <div className="min-w-0">
                          <div className="truncate text-sm text-text">{m.name}</div>
                          <div className="truncate font-mono text-[11px] text-muted">
                            {m.id} {m.capabilities}
                          </div>
                        </div>
                        {isSel && <Badge tone="accent">selected</Badge>}
                      </li>
                    )
                  })}
                </ul>
              )}
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-muted">{selected.size} selected</span>
              <div className="flex gap-2">
                <Button variant="ghost" onClick={onClose}>
                  Cancel
                </Button>
                <Button variant="primary" onClick={doImport} disabled={importing || selected.size === 0}>
                  {importing ? <Spinner /> : 'Import selected'}
                </Button>
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
