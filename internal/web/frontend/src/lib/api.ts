import type {
  ActiveResponse,
  CatalogResponse,
  CreateProfileRequest,
  DiffResponse,
  ImportResult,
  JSONSchemaNode,
  ModelsResponse,
  ProfileDetail,
  ProfilesResponse,
  RegisteredModel,
  SchemaCheckResult,
  ValidateResult,
  ValidationError,
} from './types'

export class ApiError extends Error {
  status: number
  validationErrors?: ValidationError[]
  constructor(status: number, message: string, validationErrors?: ValidationError[]) {
    super(message)
    this.status = status
    this.validationErrors = validationErrors
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  const text = await res.text()
  let parsed: unknown = undefined
  if (text) {
    try {
      parsed = JSON.parse(text)
    } catch {
      parsed = text
    }
  }

  if (!res.ok) {
    const obj = (parsed ?? {}) as { error?: string; validationErrors?: ValidationError[] }
    throw new ApiError(res.status, obj.error || `HTTP ${res.status}`, obj.validationErrors)
  }

  return parsed as T
}

export const api = {
  // Profiles
  listProfiles: () => request<ProfilesResponse>('GET', '/api/profiles'),
  getProfile: (name: string) => request<ProfileDetail>('GET', `/api/profiles/${encodeURIComponent(name)}`),
  saveProfile: (name: string, config: unknown) =>
    request<{ ok: boolean }>('PUT', `/api/profiles/${encodeURIComponent(name)}`, config),
  createProfile: (req: CreateProfileRequest) => request<{ name: string }>('POST', '/api/profiles', req),
  deleteProfile: (name: string) =>
    request<{ ok: boolean }>('DELETE', `/api/profiles/${encodeURIComponent(name)}`),
  renameProfile: (name: string, newName: string) =>
    request<{ name: string }>('POST', `/api/profiles/${encodeURIComponent(name)}/rename`, { newName }),
  activateProfile: (name: string) =>
    request<{ ok: boolean; name: string; snapshot: string }>('POST', `/api/profiles/${encodeURIComponent(name)}/activate`),
  exportProfileUrl: (name: string) => `/api/profiles/${encodeURIComponent(name)}/export`,

  // Active / diff / import / validate / schema
  getActive: () => request<ActiveResponse>('GET', '/api/active'),
  diff: (left: string, right: string) =>
    request<DiffResponse>('GET', `/api/diff?left=${encodeURIComponent(left)}&right=${encodeURIComponent(right)}`),
  import: (config: unknown, name?: string) =>
    request<ImportResult>('POST', '/api/import', { name: name ?? '', config }),
  validate: (config: unknown, mode: 'strict' | 'save' = 'save') =>
    request<ValidateResult>('POST', `/api/validate?mode=${mode}`, config),
  getSchema: () => request<JSONSchemaNode>('GET', '/api/schema'),
  schemaCheck: () => request<SchemaCheckResult>('GET', '/api/schema-check'),

  // Models
  listModels: () => request<ModelsResponse>('GET', '/api/models'),
  createModel: (m: RegisteredModel) => request<RegisteredModel>('POST', '/api/models', m),
  updateModel: (provider: string, modelId: string, m: RegisteredModel) =>
    request<RegisteredModel>(
      'PUT',
      `/api/models/${encodeURIComponent(provider)}/${encodeURIComponent(modelId)}`,
      m,
    ),
  deleteModel: (provider: string, modelId: string) =>
    request<{ ok: boolean }>(
      'DELETE',
      `/api/models/${encodeURIComponent(provider)}/${encodeURIComponent(modelId)}`,
    ),
  modelsCatalog: () => request<CatalogResponse>('GET', '/api/models/catalog'),
}
