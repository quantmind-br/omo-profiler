import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { AlertTriangle, ArrowLeft, Plus, Save } from 'lucide-react'
import { api } from '../lib/api'
import type { ConfigObject, JSONSchemaNode, ValidationError } from '../lib/types'
import { cn, humanize } from '../lib/utils'
import { Button } from '../components/ui/button'
import { Card } from '../components/ui/card'
import { Input } from '../components/ui/input'
import { Badge } from '../components/ui/badge'
import { Spinner } from '../components/ui/spinner'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/tabs'
import { SchemaField, SchemaForm } from '../components/SchemaForm'
import { EntryList } from '../components/ui/entry-list'
import { JsonEditor } from '../components/JsonEditor'
import { useToast } from '../components/ui/toast'

const GENERAL_KEYS = [
  'hashline_edit',
  'telemetry',
  'model_fallback',
  'auto_update',
  'new_task_system_enabled',
  'default_run_agent',
  'disabled_mcps',
  'disabled_agents',
  'disabled_skills',
  'disabled_hooks',
  'disabled_commands',
  'disabled_tools',
  'disabled_providers',
]

const CURATED: Record<string, true> = Object.fromEntries(
  ['$schema', 'agents', 'categories', ...GENERAL_KEYS].map((k) => [k, true]),
)

export function EditorPage() {
  const { name = '' } = useParams()
  const navigate = useNavigate()
  const { toast } = useToast()

  const profileQ = useQuery({ queryKey: ['profile', name], queryFn: () => api.getProfile(name) })
  const schemaQ = useQuery({ queryKey: ['schema'], queryFn: api.getSchema, staleTime: Infinity })

  const [working, setWorking] = useState<ConfigObject>({})
  const [dirty, setDirty] = useState(false)
  const [errors, setErrors] = useState<ValidationError[]>([])
  const [saving, setSaving] = useState(false)
  const [section, setSection] = useState('General')

  useEffect(() => {
    if (profileQ.data) {
      setWorking(profileQ.data.config)
      setDirty(false)
    }
  }, [profileQ.data])

  useEffect(() => {
    if (!dirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      e.returnValue = ''
    }
    window.addEventListener('beforeunload', handler)
    return () => window.removeEventListener('beforeunload', handler)
  }, [dirty])

  const schema = schemaQ.data
  const otherSections = useMemo(() => {
    if (!schema?.properties) return []
    return Object.keys(schema.properties).filter((k) => !CURATED[k])
  }, [schema])

  function update(next: ConfigObject) {
    setWorking(next)
    setDirty(true)
    setErrors([])
  }

  function setField(key: string, value: unknown) {
    const next: ConfigObject = { ...working }
    if (value === undefined) delete next[key]
    else next[key] = value
    update(next)
  }

  async function save() {
    setSaving(true)
    try {
      const res = await api.validate(working, 'save')
      if (!res.valid) {
        setErrors(res.errors)
        toast({ title: 'Validation failed', description: `${res.errors.length} error(s)`, variant: 'error' })
        return
      }
      await api.saveProfile(name, working)
      setErrors([])
      setDirty(false)
      toast({ title: 'Saved', description: name, variant: 'success' })
    } catch (e) {
      toast({ title: 'Save failed', description: (e as Error).message, variant: 'error' })
    } finally {
      setSaving(false)
    }
  }

  if (profileQ.isLoading || schemaQ.isLoading) {
    return (
      <div className="flex justify-center p-10">
        <Spinner className="h-6 w-6" />
      </div>
    )
  }

  if (profileQ.isError || !schema) {
    return <div className="text-danger">Failed to load profile or schema.</div>
  }

  const sections = ['General', 'Agents', 'Categories', ...otherSections]

  return (
    <div className="mx-auto max-w-5xl space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => navigate('/profiles')} aria-label="Back">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-xl font-semibold text-text">{name}</h1>
            {dirty && <span className="text-xs text-warn">Unsaved changes</span>}
          </div>
        </div>
        <Button variant="primary" onClick={save} disabled={saving || !dirty}>
          {saving ? <Spinner /> : <Save className="h-4 w-4" />} Save
        </Button>
      </div>

      {profileQ.data?.hasLegacyFields && (
        <div className="flex items-start gap-2 rounded-lg border border-warn/40 bg-warn/10 p-3 text-sm text-warn">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
          <div>
            <div className="font-medium">Legacy / unknown fields present</div>
            <div className="mt-0.5 text-xs opacity-90">{profileQ.data.legacyFieldsWarning}</div>
          </div>
        </div>
      )}

      {errors.length > 0 && (
        <Card className="border-danger/40 bg-danger/10">
          <div className="text-sm font-medium text-danger">Validation errors</div>
          <ul className="mt-2 space-y-1 text-xs text-danger">
            {errors.map((e, i) => (
              <li key={i}>
                <span className="font-mono">{e.path}</span>: {e.message}
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Tabs defaultValue="form">
        <TabsList>
          <TabsTrigger value="form">Form</TabsTrigger>
          <TabsTrigger value="json">JSON</TabsTrigger>
        </TabsList>

        <TabsContent value="form" className="mt-4">
          <div className="flex gap-4">
            <nav className="w-48 shrink-0 space-y-0.5">
              {sections.map((s) => (
                <button
                  key={s}
                  onClick={() => setSection(s)}
                  className={cn(
                    'block w-full truncate rounded-md px-3 py-1.5 text-left text-sm transition-colors',
                    section === s ? 'bg-surface-2 text-text' : 'text-muted hover:text-text',
                  )}
                >
                  {s === 'General' || s === 'Agents' || s === 'Categories' ? s : humanize(s)}
                </button>
              ))}
            </nav>

            <div className="min-w-0 flex-1">
              <Card>
                {section === 'General' && (
                  <GeneralPanel schema={schema} working={working} setField={setField} />
                )}
                {section === 'Agents' && (
                  <EntityPanel
                    title="Agents"
                    parentSchema={schema.properties!.agents}
                    value={working.agents}
                    onChange={(v) => setField('agents', v)}
                    named
                  />
                )}
                {section === 'Categories' && (
                  <EntityPanel
                    title="Categories"
                    parentSchema={schema.properties!.categories}
                    value={working.categories}
                    onChange={(v) => setField('categories', v)}
                  />
                )}
                {!['General', 'Agents', 'Categories'].includes(section) && schema.properties![section] && (
                  <SchemaField
                    name={section}
                    schema={schema.properties![section]}
                    value={working[section]}
                    onChange={(v) => setField(section, v)}
                  />
                )}
              </Card>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="json" className="mt-4">
          <JsonTab working={working} onChange={update} />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function GeneralPanel({
  schema,
  working,
  setField,
}: {
  schema: JSONSchemaNode
  working: ConfigObject
  setField: (key: string, value: unknown) => void
}) {
  const props = schema.properties ?? {}
  return (
    <div className="space-y-5">
      {GENERAL_KEYS.filter((k) => props[k]).map((k) => (
        <SchemaField key={k} name={k} schema={props[k]} value={working[k]} onChange={(v) => setField(k, v)} />
      ))}
    </div>
  )
}

// EntityPanel renders a map of named entities (agents / categories). Each entry
// is edited via SchemaForm using the entry schema derived from the parent.
function EntityPanel({
  title,
  parentSchema,
  value,
  onChange,
  named = false,
}: {
  title: string
  parentSchema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
  named?: boolean
}) {
  const obj = (value && typeof value === 'object' && !Array.isArray(value) ? value : {}) as Record<string, unknown>
  const [newName, setNewName] = useState('')

  const fallback: JSONSchemaNode = {
    type: 'object',
    properties: { model: { type: 'string' }, variant: { type: 'string' }, category: { type: 'string' } },
  }

  function entrySchema(key: string): JSONSchemaNode {
    if (named && parentSchema.properties && parentSchema.properties[key]) return parentSchema.properties[key]
    if (parentSchema.additionalProperties && typeof parentSchema.additionalProperties === 'object') {
      return parentSchema.additionalProperties as JSONSchemaNode
    }
    if (named && parentSchema.properties) {
      // Use any known named-entity schema as a template.
      const first = Object.values(parentSchema.properties)[0]
      if (first) return first
    }
    return fallback
  }

  function setEntry(key: string, next: unknown) {
    const merged: Record<string, unknown> = { ...obj }
    if (next === undefined) delete merged[key]
    else merged[key] = next
    onChange(Object.keys(merged).length ? merged : undefined)
  }

  function add() {
    const k = newName.trim()
    if (!k || obj[k] !== undefined) return
    setEntry(k, {})
    setNewName('')
  }

  const knownNames =
    named && parentSchema.properties ? Object.keys(parentSchema.properties).filter((n) => obj[n] === undefined) : []
  const entries = Object.keys(obj)

  return (
    <div className="space-y-4">
      <EntryList
        keys={entries}
        countLabel={title.toLowerCase()}
        onRemove={(key) => setEntry(key, undefined)}
        renderBody={(key) => (
          <SchemaForm schema={entrySchema(key)} value={obj[key]} onChange={(v) => setEntry(key, v ?? {})} />
        )}
      />

      <div className="flex flex-wrap items-center gap-2">
        <Input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), add())}
          placeholder={`New ${title.toLowerCase().replace(/s$/, '')} name…`}
          className="w-56"
        />
        <Button variant="secondary" size="sm" onClick={add}>
          <Plus className="h-4 w-4" /> Add
        </Button>
        {knownNames.map((n) => (
          <Badge key={n} tone="muted" className="cursor-pointer hover:text-text" onClick={() => setEntry(n, {})}>
            + {n}
          </Badge>
        ))}
      </div>
    </div>
  )
}

function JsonTab({ working, onChange }: { working: ConfigObject; onChange: (v: ConfigObject) => void }) {
  const [text, setText] = useState(() => JSON.stringify(working, null, 2))
  const [parseError, setParseError] = useState<string | null>(null)
  const [validation, setValidation] = useState<ValidationError[]>([])

  // Re-sync when the working object changes from the form tab.
  useEffect(() => {
    setText(JSON.stringify(working, null, 2))
  }, [working])

  useEffect(() => {
    if (parseError) return
    const id = setTimeout(async () => {
      try {
        const parsed = JSON.parse(text)
        const res = await api.validate(parsed, 'save')
        setValidation(res.errors)
      } catch {
        /* ignore; parseError handles syntax */
      }
    }, 500)
    return () => clearTimeout(id)
  }, [text, parseError])

  function onEdit(next: string) {
    setText(next)
    try {
      const parsed = JSON.parse(next)
      setParseError(null)
      onChange(parsed as ConfigObject)
    } catch (e) {
      setParseError((e as Error).message)
    }
  }

  return (
    <div className="space-y-3">
      <JsonEditor value={text} onChange={onEdit} />
      {parseError && <p className="text-sm text-danger">JSON syntax error: {parseError}</p>}
      {!parseError && validation.length > 0 && (
        <div className="rounded-lg border border-danger/40 bg-danger/10 p-3">
          <div className="text-sm font-medium text-danger">Schema validation</div>
          <ul className="mt-2 space-y-1 text-xs text-danger">
            {validation.map((e, i) => (
              <li key={i}>
                <span className="font-mono">{e.path}</span>: {e.message}
              </li>
            ))}
          </ul>
        </div>
      )}
      {!parseError && validation.length === 0 && <p className="text-sm text-success">Valid.</p>}
    </div>
  )
}
