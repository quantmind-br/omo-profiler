import { useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Copy, Download, Pencil, Play, Plus, Tag, Trash2, Upload } from 'lucide-react'
import { api, ApiError } from '../lib/api'
import type { ProfileListEntry } from '../lib/types'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Spinner } from '../components/ui/spinner'
import { Dialog, DialogContent } from '../components/ui/dialog'
import { useToast } from '../components/ui/toast'

export function ProfilesPage() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const navigate = useNavigate()
  const { data, isLoading } = useQuery({ queryKey: ['profiles'], queryFn: api.listProfiles })

  const [newOpen, setNewOpen] = useState(false)
  const [cloneFrom, setCloneFrom] = useState<string | null>(null)
  const [renameFrom, setRenameFrom] = useState<string | null>(null)
  const [deleteName, setDeleteName] = useState<string | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  function refresh() {
    qc.invalidateQueries({ queryKey: ['profiles'] })
    qc.invalidateQueries({ queryKey: ['active'] })
  }

  const activate = useMutation({
    mutationFn: (name: string) => api.activateProfile(name),
    onSuccess: (res, name) => {
      toast({ title: `Applied ${res.name}`, variant: 'success' })
      if (res.snapshot) {
        toast({ title: `Previous config saved as ${res.snapshot}`, variant: 'success' })
      }
      refresh()
    },
    onError: (e: Error) => toast({ title: 'Activation failed', description: e.message, variant: 'error' }),
  })

  async function onImportFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      const parsed = JSON.parse(await file.text())
      const name = file.name.replace(/\.json$/i, '')
      const res = await api.import(parsed, name)
      toast({
        title: `Imported as ${res.name}`,
        description: res.hadCollision ? 'Name collision resolved with suffix' : undefined,
        variant: 'success',
      })
      refresh()
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : (err as Error).message
      toast({ title: 'Import failed', description: msg, variant: 'error' })
    } finally {
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  return (
    <div className="mx-auto max-w-4xl space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Profiles</h1>
        <div className="flex gap-2">
          <input ref={fileRef} type="file" accept="application/json,.json" className="hidden" onChange={onImportFile} />
          <Button variant="secondary" onClick={() => fileRef.current?.click()}>
            <Upload className="h-4 w-4" /> Import
          </Button>
          <Button variant="primary" onClick={() => setNewOpen(true)}>
            <Plus className="h-4 w-4" /> New profile
          </Button>
        </div>
      </div>

      <Card className="p-0">
        {isLoading ? (
          <div className="p-6">
            <Spinner />
          </div>
        ) : data && data.profiles.length > 0 ? (
          <ul className="divide-y divide-border">
            {data.profiles.map((p) => (
              <ProfileRow
                key={p.name}
                profile={p}
                onActivate={() => activate.mutate(p.name)}
                activating={activate.isPending}
                onEdit={() => navigate(`/profiles/${encodeURIComponent(p.name)}/edit`)}
                onClone={() => setCloneFrom(p.name)}
                onRename={() => setRenameFrom(p.name)}
                onDelete={() => setDeleteName(p.name)}
              />
            ))}
          </ul>
        ) : (
          <div className="p-8 text-center text-sm text-muted">No profiles yet. Create one to get started.</div>
        )}
      </Card>

      {newOpen && (
        <NewProfileDialog
          existing={data?.profiles.map((p) => p.name) ?? []}
          onClose={() => setNewOpen(false)}
          onCreated={(name) => {
            toast({ title: `Created ${name}`, variant: 'success' })
            refresh()
            setNewOpen(false)
          }}
        />
      )}

      {cloneFrom && (
        <CloneDialog
          from={cloneFrom}
          onClose={() => setCloneFrom(null)}
          onCloned={(name) => {
            toast({ title: `Cloned to ${name}`, variant: 'success' })
            refresh()
            setCloneFrom(null)
          }}
        />
      )}

      {renameFrom && (
        <RenameDialog
          from={renameFrom}
          onClose={() => setRenameFrom(null)}
          onRenamed={(res) => {
            toast({ title: `Renamed to ${res.name}`, variant: 'success' })
            refresh()
            setRenameFrom(null)
          }}
        />
      )}

      {deleteName && (
        <ConfirmDeleteDialog
          name={deleteName}
          onClose={() => setDeleteName(null)}
          onDeleted={() => {
            toast({ title: `Deleted ${deleteName}`, variant: 'success' })
            refresh()
            setDeleteName(null)
          }}
        />
      )}
    </div>
  )
}

function ProfileRow({
  profile,
  onActivate,
  activating,
  onEdit,
  onClone,
  onRename,
  onDelete,
}: {
  profile: ProfileListEntry
  onActivate: () => void
  activating: boolean
  onEdit: () => void
  onClone: () => void
  onRename: () => void
  onDelete: () => void
}) {
  return (
    <li className="flex items-center justify-between gap-3 px-4 py-3">
      <div className="flex items-center gap-2">
        <span className="font-medium text-text">{profile.name}</span>
        {profile.active && <Badge tone="success">active</Badge>}
      </div>
      <div className="flex items-center gap-1">
        <Button size="sm" variant="secondary" onClick={onActivate} disabled={profile.active || activating}>
          <Play className="h-3.5 w-3.5" /> Activate…
        </Button>
        <Button size="sm" variant="ghost" onClick={onEdit}>
          <Pencil className="h-3.5 w-3.5" /> Edit
        </Button>
        <Button size="sm" variant="ghost" onClick={onClone}>
          <Copy className="h-3.5 w-3.5" /> Clone
        </Button>
        <Button size="sm" variant="ghost" onClick={onRename}>
          <Tag className="h-3.5 w-3.5" /> Rename
        </Button>
        <a href={api.exportProfileUrl(profile.name)} download>
          <Button size="sm" variant="ghost">
            <Download className="h-3.5 w-3.5" /> Export
          </Button>
        </a>
        <Button size="sm" variant="ghost" onClick={onDelete} className="text-danger">
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>
    </li>
  )
}

function NewProfileDialog({
  existing,
  onClose,
  onCreated,
}: {
  existing: string[]
  onClose: () => void
  onCreated: (name: string) => void
}) {
  const [name, setName] = useState('')
  const [mode, setMode] = useState<'blank' | 'default' | 'clone'>('blank')
  const [from, setFrom] = useState(existing[0] ?? '')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit() {
    setBusy(true)
    setError(null)
    try {
      const fromValue = mode === 'blank' ? '' : mode === 'default' ? '__default__' : from
      await api.createProfile({ name: name.trim(), from: fromValue })
      onCreated(name.trim())
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent title="New profile" description="Create a blank profile, seed from the default template, or clone an existing one.">
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm text-muted">Name</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="my-profile" autoFocus />
          </div>
          <div>
            <label className="mb-1 block text-sm text-muted">Seed</label>
            <Select
              value={mode}
              onValueChange={(v) => setMode(v as typeof mode)}
              options={[
                { value: 'blank', label: 'Blank' },
                { value: 'default', label: 'Default template' },
                { value: 'clone', label: 'Clone existing' },
              ]}
            />
          </div>
          {mode === 'clone' && (
            <div>
              <label className="mb-1 block text-sm text-muted">Source profile</label>
              <Select value={from} onValueChange={setFrom} options={existing.map((n) => ({ value: n, label: n }))} />
            </div>
          )}
          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" onClick={submit} disabled={busy || !name.trim() || (mode === 'clone' && !from)}>
              {busy ? <Spinner /> : 'Create'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function CloneDialog({ from, onClose, onCloned }: { from: string; onClose: () => void; onCloned: (name: string) => void }) {
  const [name, setName] = useState(`${from}-copy`)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit() {
    setBusy(true)
    setError(null)
    try {
      await api.createProfile({ name: name.trim(), from })
      onCloned(name.trim())
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent title={`Clone ${from}`}>
        <div className="space-y-4">
          <Input value={name} onChange={(e) => setName(e.target.value)} autoFocus />
          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" onClick={submit} disabled={busy || !name.trim()}>
              {busy ? <Spinner /> : 'Clone'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

type RenameResult = { name: string }

function RenameDialog({
  from,
  onClose,
  onRenamed,
}: {
  from: string
  onClose: () => void
  onRenamed: (res: RenameResult) => void
}) {
  const [name, setName] = useState(from)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit() {
    setBusy(true)
    setError(null)
    try {
      const res = await api.renameProfile(from, name.trim())
      onRenamed(res)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent title={`Rename ${from}`}>
        <div className="space-y-4">
          <Input value={name} onChange={(e) => setName(e.target.value)} autoFocus />
          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" onClick={submit} disabled={busy || !name.trim() || name.trim() === from}>
              {busy ? <Spinner /> : 'Rename'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function ConfirmDeleteDialog({
  name,
  onClose,
  onDeleted,
}: {
  name: string
  onClose: () => void
  onDeleted: () => void
}) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function submit() {
    setBusy(true)
    setError(null)
    try {
      await api.deleteProfile(name)
      onDeleted()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent title={`Delete ${name}?`} description="This permanently removes the profile from the omo document. This cannot be undone.">
        <div className="space-y-4">
          {error && <p className="text-sm text-danger">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="danger" onClick={submit} disabled={busy}>
              {busy ? <Spinner /> : 'Delete'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
