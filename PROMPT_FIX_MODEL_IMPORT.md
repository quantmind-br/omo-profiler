# Fix: Model Import Duplicate Check Should Consider Provider + ModelID

## Problema

Durante a importação de modelos do models.dev, a verificação de duplicidade considera **apenas o `ModelID`**, ignorando o `Provider`. Isso significa que se um modelo como `gpt-4` já existe na base de dados local com o provider `openai`, ao tentar importar o mesmo `gpt-4` de um provider diferente (ex: `azure`), a importação é **silenciosamente pulada** como duplicata — mesmo sendo uma combinação válida de provider + model.

## Contexto da Arquitetura

A estrutura `RegisteredModel` possui 3 campos:

```go
type RegisteredModel struct {
    DisplayName string `json:"displayName"`
    ModelID     string `json:"modelId"` // atualmente tratado como chave única
    Provider    string `json:"provider"`
}
```

O problema está em **dois pontos**:

### 1. `ModelsRegistry.Add()` — `internal/models/models.go:93-100`

```go
func (r *ModelsRegistry) Add(m RegisteredModel) error {
    for _, existing := range r.Models {
        if existing.ModelID == m.ModelID {  // <-- VERIFICA SÓ ModelID
            return &ModelExistsError{ModelID: m.ModelID}
        }
    }
    r.Models = append(r.Models, m)
    return r.Save()
}
```

### 2. `ModelsRegistry.Get()` — `internal/models/models.go`

O método `Get()` também busca apenas por `ModelID`, o que afeta operações de lookup.

## Comportamento Esperado

- Um modelo é considerado **único** pela combinação `(Provider, ModelID)`.
- O mesmo `ModelID` com `Provider` diferente deve ser tratado como um **novo registro válido**.
- Exemplo: `{Provider: "openai", ModelID: "gpt-4"}` e `{Provider: "azure", ModelID: "gpt-4"}` devem coexistir.

## Tarefas de Implementação

### 1. Atualizar `ModelsRegistry.Add()`
- Modificar a verificação de duplicidade para comparar **`Provider` + `ModelID`** em conjunto.
- Atualizar `ModelExistsError` para incluir ambos os campos na mensagem de erro.

### 2. Atualizar `ModelsRegistry.Get()`
- Criar nova assinatura: `Get(provider, modelID string) *RegisteredModel`
- Ou criar método separado: `GetByProviderAndModelID(provider, modelID string) *RegisteredModel`
- Manter retrocompatibilidade se `Get()` for usado em outros lugares (avaliar impacto).

### 3. Atualizar `Exists()` (func standalone)
- Atualizar assinatura para aceitar `provider` + `modelID` ou criar `ExistsForProvider(provider, modelID string) bool`.

### 4. Atualizar pontos de chamada
- `importSelectedModels()` em `internal/tui/views/model_import.go` — usa `registry.Add()`.
- `internal/tui/views/model_registry.go` — usa `Get()`, `Update()`, `Delete()`.
- Qualquer outro lugar que use `Get()` ou `Exists()` (buscar com grep).

### 5. Atualizar testes
- Adicionar testes de cobertura para:
  - Adicionar mesmo `ModelID` com providers diferentes → **deve sucesso**.
  - Adicionar mesmo `ModelID` com mesmo provider → **deve falhar** com `ModelExistsError`.
  - Verificar que `Get()` retorna o modelo correto para o par `(provider, modelID)`.

## Arquivos Envolvidos

| Arquivo | Ação |
|---------|------|
| `internal/models/models.go` | Core: `Add()`, `Get()`, `Exists()`, `ModelExistsError` |
| `internal/tui/views/model_import.go` | Import loop usa `registry.Add()` |
| `internal/tui/views/model_registry.go` | CRUD manual pode usar `Get()` |
| `internal/models/models_test.go` | Adicionar/atualizar testes |

## Constraints

- **Não alterar** a estrutura `RegisteredModel` (manter os 3 campos atuais).
- **Não alterar** a lógica de fetch do models.dev (`modelsdev.go`).
- Seguir convenções do projeto: sem `//nolint`, sem type assertion shortcuts, usar `testify` para assertions.
- Tests obrigatórios: co-localizados com `_test.go`, usar `config.SetBaseDir(t.TempDir())`.

## Critérios de Aceitação

1. `make test` passa sem erros.
2. `make lint` passa sem warnings.
3. É possível importar o mesmo `ModelID` de providers diferentes sem skip.
4. Importar o mesmo `(Provider, ModelID)` ainda resulta em skip com `ModelExistsError`.
