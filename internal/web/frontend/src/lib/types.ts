// Types mirroring the Go API responses (see internal/web/handlers*.go).

// The active profile is detected by comparing the root against stored profiles,
// so the root *is* the effective configuration — no env vars, no hint file.
export interface ActiveInfo {
  documentExists: boolean
  profileName: string
  modified: boolean
}

export interface ProfileListEntry {
  name: string
  active: boolean
}

export interface ProfilesResponse {
  active: ActiveInfo
  profiles: ProfileListEntry[]
}

export type ConfigObject = Record<string, unknown>

export interface ProfileDetail {
  name: string
  config: ConfigObject
  fieldPresence: Record<string, boolean> | null
  hasLegacyFields: boolean
  legacyFieldsWarning: string
}

export interface ActiveResponse extends ActiveInfo {
  config: ConfigObject
}

export interface ValidationError {
  path: string
  message: string
}

export interface ValidateResult {
  valid: boolean
  errors: ValidationError[]
}

export interface DiffLine {
  text: string
  type: number // 0 equal, 1 added, 2 removed
  lineNum: number
}

export interface DiffResponse {
  leftLabel: string
  rightLabel: string
  left: DiffLine[]
  right: DiffLine[]
}

export interface ImportResult {
  name: string
  hadCollision: boolean
}

export interface SchemaCheckResult {
  identical: boolean
  diff: string
}

export interface RegisteredModel {
  displayName: string
  modelId: string
  provider: string
}

export interface ModelGroup {
  provider: string
  models: RegisteredModel[]
}

export interface ModelsResponse {
  groups: ModelGroup[]
}

export interface CatalogModel {
  id: string
  name: string
  family: string
  reasoning: boolean
  toolCall: boolean
  attachment: boolean
  context: number
  output: number
  capabilities: string
}

export interface CatalogProvider {
  id: string
  name: string
  models: CatalogModel[]
}

export interface CatalogResponse {
  providers: CatalogProvider[]
}

export interface CreateProfileRequest {
  name: string
  from: string
}

// ---- JSON Schema (draft-07 subset we render) ----

export interface JSONSchemaNode {
  type?: string | string[]
  properties?: Record<string, JSONSchemaNode>
  additionalProperties?: boolean | JSONSchemaNode
  items?: JSONSchemaNode
  enum?: unknown[]
  description?: string
  minimum?: number
  maximum?: number
  minLength?: number
  maxLength?: number
  default?: unknown
  required?: string[]
  title?: string
  anyOf?: JSONSchemaNode[]
  [key: string]: unknown
}
