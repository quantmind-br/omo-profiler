import { useState } from 'react'
import { ChevronDown, ChevronRight, Plus, X } from 'lucide-react'
import type { JSONSchemaNode } from '../lib/types'
import { cn, humanize } from '../lib/utils'
import { Input } from './ui/input'
import { Textarea } from './ui/textarea'
import { Switch } from './ui/switch'
import { Select } from './ui/select'
import { TagInput } from './ui/tag-input'
import { MultiSelect } from './ui/multiselect'
import { ModelCombobox } from './ui/model-combobox'
import { KeyValueEditor } from './ui/key-value'
import { Button } from './ui/button'
import { EntryList } from './ui/entry-list'

// nodeType collapses a possibly-union JSON schema type into the primary
// non-null type we render.
function nodeType(schema: JSONSchemaNode): string | undefined {
  const t = schema.type
  if (Array.isArray(t)) return t.find((x) => x !== 'null')
  return t
}

const TEXTAREA_KEY = /prompt|custom_prompt|instruction|description/i

// SchemaForm renders an object's properties (schema.properties) as labeled
// fields. value is the current object; onChange receives the updated object,
// or undefined when it becomes empty (sparse).
export function SchemaForm({
  schema,
  value,
  onChange,
}: {
  schema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
}) {
  const obj = (value && typeof value === 'object' && !Array.isArray(value) ? value : {}) as Record<string, unknown>
  const props = schema.properties ?? {}
  const keys = Object.keys(props)

  function setKey(key: string, next: unknown) {
    const merged: Record<string, unknown> = { ...obj }
    if (next === undefined) delete merged[key]
    else merged[key] = next
    onChange(Object.keys(merged).length ? merged : undefined)
  }

  if (keys.length === 0) {
    // Object with no declared properties → treat as a free map.
    return <MapEditor schema={schema} value={obj} onChange={onChange} />
  }

  return (
    <div className="space-y-4">
      {keys.map((key) => (
        <SchemaField
          key={key}
          name={key}
          schema={props[key]}
          value={obj[key]}
          onChange={(v) => setKey(key, v)}
        />
      ))}
    </div>
  )
}

// SchemaField renders a single labeled field dispatched by its schema node.
export function SchemaField({
  name,
  schema,
  value,
  onChange,
}: {
  name: string
  schema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
}) {
  const label = schema.title || humanize(name)
  return (
    <div>
      <div className="mb-1 flex items-center gap-2">
        <label className="text-sm font-medium text-text">{label}</label>
        <span className="font-mono text-[10px] text-muted">{name}</span>
      </div>
      {schema.description && <p className="mb-1.5 text-xs text-muted">{schema.description}</p>}
      <FieldControl name={name} schema={schema} value={value} onChange={onChange} />
    </div>
  )
}

function FieldControl({
  name,
  schema,
  value,
  onChange,
}: {
  name: string
  schema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
}) {
  const type = nodeType(schema)

  // enum → Select (+ clear)
  if (Array.isArray(schema.enum) && schema.enum.length > 0) {
    const options = schema.enum.map((e) => ({ value: String(e), label: String(e) }))
    return (
      <div className="flex items-center gap-2">
        <Select
          value={value === undefined ? undefined : String(value)}
          onValueChange={(v) => onChange(v)}
          options={options}
        />
        {value !== undefined && (
          <button className="text-muted hover:text-text" onClick={() => onChange(undefined)} aria-label="Clear">
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    )
  }

  // Model fields → registry-backed combobox (models.json), preserving custom values.
  if (name === 'model') {
    return (
      <ModelCombobox
        value={typeof value === 'string' ? value : undefined}
        onChange={onChange}
      />
    )
  }

  if (name === 'fallback_models') {
    return <FallbackModelsEditor schema={schema} value={value} onChange={onChange} />
  }

  if (type === 'boolean') {
    return (
      <Switch checked={value === true} onCheckedChange={(v) => onChange(v ? true : undefined)} />
    )
  }

  if (type === 'integer' || type === 'number') {
    return (
      <Input
        type="number"
        value={value === undefined || value === null ? '' : String(value)}
        min={schema.minimum}
        max={schema.maximum}
        onChange={(e) => {
          const raw = e.target.value
          if (raw === '') return onChange(undefined)
          const n = type === 'integer' ? parseInt(raw, 10) : parseFloat(raw)
          onChange(Number.isNaN(n) ? undefined : n)
        }}
      />
    )
  }

  if (type === 'string') {
    const multiline = TEXTAREA_KEY.test(name) || (typeof schema.maxLength === 'number' && schema.maxLength >= 512)
    const strVal = typeof value === 'string' ? value : ''
    if (multiline) {
      return (
        <Textarea value={strVal} onChange={(e) => onChange(e.target.value || undefined)} />
      )
    }
    return <Input value={strVal} onChange={(e) => onChange(e.target.value || undefined)} />
  }

  if (type === 'array') {
    const items = schema.items ?? {}
    const arr = Array.isArray(value) ? (value as unknown[]).map(String) : []
    if (Array.isArray(items.enum) && items.enum.length > 0) {
      return (
        <MultiSelect
          value={arr}
          options={items.enum.map(String)}
          onChange={(v) => onChange(v.length ? v : undefined)}
        />
      )
    }
    if (nodeType(items) === 'string' || items.type === undefined) {
      return <TagInput value={arr} onChange={(v) => onChange(v.length ? v : undefined)} />
    }
    // Array of objects/other → JSON fallback.
    return <JsonField value={value} onChange={onChange} />
  }

  if (type === 'object' || schema.properties || schema.additionalProperties) {
    if (schema.properties && Object.keys(schema.properties).length > 0) {
      return <Collapsible name={name} schema={schema} value={value} onChange={onChange} />
    }
    return <MapEditor schema={schema} value={value} onChange={onChange} />
  }

  // Unknown shape → JSON fallback so any field remains editable.
  return <JsonField value={value} onChange={onChange} />
}

function Collapsible({
  name,
  schema,
  value,
  onChange,
}: {
  name: string
  schema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
}) {
  const hasValue = value !== undefined && value !== null
  const [open, setOpen] = useState(hasValue)
  return (
    <div className="rounded-lg border border-border">
      <div className="flex items-center justify-between px-3 py-2">
        <button
          className="flex items-center gap-1.5 text-sm text-text"
          onClick={() => setOpen((o) => !o)}
          type="button"
        >
          {open ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          {humanize(name)}
        </button>
        {hasValue && (
          <button className="text-muted hover:text-text" onClick={() => onChange(undefined)} aria-label="Remove">
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
      {open && (
        <div className="border-t border-border p-3">
          <SchemaForm schema={schema} value={value} onChange={onChange} />
        </div>
      )}
    </div>
  )
}

// MapEditor renders an object whose keys are user-defined (additionalProperties).
// When the value schema declares properties (a collection of named entries with
// settings, e.g. openclaw.gateways), each entry is an independently collapsible
// card via EntryMapEditor; otherwise it stays a flat key/value list.
function MapEditor({
  schema,
  value,
  onChange,
}: {
  schema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
}) {
  const obj = (value && typeof value === 'object' && !Array.isArray(value) ? value : {}) as Record<string, unknown>
  const valueSchema: JSONSchemaNode =
    schema.additionalProperties && typeof schema.additionalProperties === 'object'
      ? (schema.additionalProperties as JSONSchemaNode)
      : {}

  function setKey(key: string, next: unknown) {
    const merged: Record<string, unknown> = { ...obj }
    if (next === undefined) delete merged[key]
    else merged[key] = next
    onChange(Object.keys(merged).length ? merged : undefined)
  }

  const isEntryMap = !!valueSchema.properties && Object.keys(valueSchema.properties).length > 0
  if (isEntryMap) {
    return <EntryMapEditor obj={obj} valueSchema={valueSchema} setKey={setKey} />
  }

  const entries = Object.entries(obj)
  return (
    <KeyValueEditor
      entries={entries}
      onAdd={(key) => setKey(key, valueSchema.type === 'object' ? {} : '')}
      onRemove={(key) => setKey(key, undefined)}
      renderValue={(key, val) => (
        <FieldControl name={key} schema={valueSchema} value={val} onChange={(v) => setKey(key, v)} />
      )}
    />
  )
}

// EntryMapEditor renders an object-valued map (named entries with settings) as
// collapsible cards plus a key-add row.
function EntryMapEditor({
  obj,
  valueSchema,
  setKey,
}: {
  obj: Record<string, unknown>
  valueSchema: JSONSchemaNode
  setKey: (key: string, next: unknown) => void
}) {
  const [newKey, setNewKey] = useState('')
  const keys = Object.keys(obj)

  function add() {
    const k = newKey.trim()
    if (!k || obj[k] !== undefined) return
    setKey(k, {})
    setNewKey('')
  }

  return (
    <div className="space-y-3">
      <EntryList
        keys={keys}
        countLabel="entries"
        onRemove={(key) => setKey(key, undefined)}
        renderBody={(key) => (
          <SchemaForm schema={valueSchema} value={obj[key]} onChange={(v) => setKey(key, v ?? {})} />
        )}
      />
      <div className="flex gap-2">
        <Input
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), add())}
          placeholder="New key…"
          className="w-40"
        />
        <Button variant="secondary" size="sm" onClick={add}>
          <Plus className="h-4 w-4" /> Add
        </Button>
      </div>
    </div>
  )
}

// JsonField edits any value as raw JSON text with parse validation.
function JsonField({ value, onChange }: { value: unknown; onChange: (v: unknown) => void }) {
  const [text, setText] = useState(() => (value === undefined ? '' : JSON.stringify(value, null, 2)))
  const [error, setError] = useState<string | null>(null)

  function commit(next: string) {
    setText(next)
    if (next.trim() === '') {
      setError(null)
      onChange(undefined)
      return
    }
    try {
      onChange(JSON.parse(next))
      setError(null)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  return (
    <div>
      <Textarea value={text} onChange={(e) => commit(e.target.value)} className={cn('font-mono text-xs', error && 'border-danger')} />
      {error && <p className="mt-1 text-xs text-danger">{error}</p>}
    </div>
  )
}

const FALLBACK_OBJECT_SCHEMA: JSONSchemaNode = {
  type: 'object',
  properties: {
    model: { type: 'string' },
    reasoning: {
      anyOf: [
        { type: 'string', enum: ['off', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max', 'auto'] },
        { type: 'string' },
      ],
    },
    variant: { type: 'string' },
    reasoningEffort: {
      anyOf: [
        { type: 'string', enum: ['off', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max', 'auto'] },
        { type: 'string' },
      ],
    },
    temperature: { type: 'number', minimum: 0, maximum: 2 },
    top_p: { type: 'number', minimum: 0, maximum: 1 },
    maxTokens: { type: 'number' },
    max_tokens: { type: 'integer', exclusiveMinimum: 0 },
    thinking: {
      type: 'object',
      properties: {
        type: { type: 'string', enum: ['enabled', 'disabled'] },
        budgetTokens: { type: 'number' },
      },
    },
  },
}

// fallbackObjectSchema pulls the object-item schema out of the fallback_models
// anyOf so nested fields (reasoningEffort enum, thinking, numbers) render with
// the live schema; falls back to the constant if extraction fails.
function fallbackObjectSchema(schema: JSONSchemaNode): JSONSchemaNode {
  const variants = Array.isArray(schema.anyOf) ? schema.anyOf : []
  for (const v of variants) {
    if (v?.type === 'array' && v.items) {
      const it = v.items as JSONSchemaNode
      if (it.type === 'object' && it.properties) return it
      const inner = Array.isArray(it.anyOf) ? it.anyOf : []
      for (const iv of inner) if (iv?.type === 'object' && iv.properties) return iv
    }
  }
  return FALLBACK_OBJECT_SCHEMA
}

// FallbackModelsEditor edits agent/category fallback_models
// (string | string[] | (string|object)[]) as a registry-backed model list,
// preserving the original scalar-string shape and unknown object keys.
function FallbackModelsEditor({
  schema,
  value,
  onChange,
}: {
  schema: JSONSchemaNode
  value: unknown
  onChange: (v: unknown) => void
}) {
  const objSchema = fallbackObjectSchema(schema)
  const arr: unknown[] = Array.isArray(value) ? value : value === undefined || value === '' ? [] : [value]
  const wasScalar = typeof value === 'string'

  function commit(next: unknown[]) {
    if (next.length === 0) return onChange(undefined)
    if (wasScalar && next.length === 1 && typeof next[0] === 'string') return onChange(next[0])
    onChange(next)
  }
  const setItem = (i: number, v: unknown) => commit(arr.map((x, j) => (j === i ? v : x)))
  const removeItem = (i: number) => commit(arr.filter((_, j) => j !== i))

  return (
    <div className="space-y-2">
      {arr.map((item, i) => (
        <div key={i} className="flex items-start gap-2 rounded-lg border border-border p-2">
          <div className="min-w-0 flex-1">
            {typeof item === 'string' ? (
              <div className="flex items-center gap-2">
                <div className="min-w-0 flex-1">
                  <ModelCombobox value={item || undefined} onChange={(v) => setItem(i, v ?? '')} />
                </div>
                <button
                  type="button"
                  className="shrink-0 text-xs text-muted hover:text-text"
                  onClick={() => setItem(i, { model: item })}
                >
                  Options
                </button>
              </div>
            ) : (
              <SchemaForm
                schema={objSchema}
                value={item}
                onChange={(v) => (v === undefined ? removeItem(i) : setItem(i, v))}
              />
            )}
          </div>
          <button
            type="button"
            className="shrink-0 text-muted hover:text-text"
            onClick={() => removeItem(i)}
            aria-label="Remove model"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
      <button type="button" className="text-sm text-accent hover:underline" onClick={() => commit([...arr, ''])}>
        + Add fallback model
      </button>
    </div>
  )
}
