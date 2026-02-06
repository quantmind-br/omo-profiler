# Plano de Cobertura de Testes - Meta: 90%

## Status Atual

### Cobertura Geral
- **Cobertura atual**: 23.7%
- **Meta**: 90%
- **Gap**: 66.3%

### Cobertura por Pacote/Módulo

| Pacote | Cobertura Atual | Meta | Gap | Prioridade |
|--------|----------------|------|-----|------------|
| `internal/diff` | 100.0% | 90% | +10% | Baixa |
| `internal/schema` | 87.9% | 90% | 2.1% | Média |
| `internal/config` | 83.3% | 90% | 6.7% | Média |
| `internal/backup` | 80.0% | 90% | 10.0% | Média |
| `internal/profile` | 80.4% | 90% | 9.6% | Média |
| `internal/models` | 79.4% | 90% | 10.6% | Média |
| `internal/tui/views` | 20.2% | 90% | 69.8% | **Alta** |
| `internal/cli/cmd` | 2.9% | 90% | 87.1% | **Alta** |
| `internal/cli` | 0.0% | 90% | 90.0% | **Alta** |
| `internal/tui` | 0.0% | 90% | 90.0% | **Alta** |
| `cmd/omo-profiler` | 0.0% | 50% | 50.0% | Média |

## Análise Detalhada

### Arquivos Sem Testes

#### Entry Points (Baixa Prioridade - Coverage Difícil)
- `cmd/omo-profiler/main.go` (14 linhas) - Entry point simples, testes de integração recomendados
- `internal/cli/root.go` (43 linhas) - Cobra setup, difícil de testar unidade
- `internal/tui/tui.go` (7 linhas) - Wrapper simples

#### Core TUI Application (Alta Prioridade)
- `internal/tui/app.go` (809 linhas) - **CRÍTICO**: Core da aplicação Bubble Tea
- `internal/tui/keys.go` (24 linhas) - Key bindings

#### Views com 0% Cobertura (Alta Prioridade)

**Views Simples (< 300 linhas):**
- `internal/tui/views/template_select.go` (141 linhas)
- `internal/tui/views/wizard_name.go` (158 linhas)
- `internal/tui/views/list.go` (289 linhas)
- `internal/tui/views/dashboard.go` (268 linhas)
- `internal/tui/views/diff.go` (401 linhas)
- `internal/tui/views/wizard.go` (410 linhas)
- `internal/tui/views/wizard_review.go` (222 linhas)
- `internal/tui/views/wizard_hooks.go` (286 linhas)

**Views Complexas (> 500 linhas):**
- `internal/tui/views/model_selector.go` (525 linhas)
- `internal/tui/views/model_import.go` (544 linhas)
- `internal/tui/views/model_registry.go` (619 linhas)
- `internal/tui/views/wizard_categories.go` (962 linhas)
- `internal/tui/views/wizard_agents.go` (1026 linhas)
- `internal/tui/views/wizard_other.go` (2132 linhas)

#### CLI Commands (Média Prioridade)
- `internal/cli/cmd/list.go` - Comando de listagem
- `internal/cli/cmd/current.go` - Comando de perfil atual
- `internal/cli/cmd/switch.go` - Comando de troca de perfil
- `internal/cli/cmd/import.go` - Comando de importação (tem init testado)
- `internal/cli/cmd/export.go` - Comando de exportação (tem init testado)
- `internal/cli/cmd/models.go` - Comando de modelos (tem init testado)

### Arquivos com Baixa Cobertura (<50%)

#### internal/tui/views
- `model_registry.go` (90.9% NewModelRegistry, mas Update 11.3%, View 0%)
- `model_import.go` (0% em Update, View)
- `wizard_categories.go` (SetConfig 6.8%, Update 21.8%, View 0%)
- `wizard_agents.go` (NewWizardAgents 12.5%, Init 33.3%, Update 22.7%, View 0%)
- `wizard_other.go` (SetConfig 14.4%, Update 8.0%, View 0%)
- `wizard_review.go` (Update 0%, View 0%, IsValid 0%)
- `model_selector.go` (tudo 0%)

#### internal/models
- `models.go` (List 0%, Save 66.7%)
- `modelsdev.go` (FetchModelsDevRegistry 0%)

#### internal/config
- `paths.go` (ModelsFile 0%)

#### internal/cli/cmd
- Cobertura geral 2.9% (apenas init functions)

### Código Crítico Não Testado

#### Bubble Tea Update Loops (CRÍTICO)
- Navegação entre estados em `app.go`
- Handler de mensagens (toastMsg, clearToastMsg, etc.)
- Operações assíncronas (switchProfileDoneMsg, deleteProfileDoneMsg, etc.)
- Import/Export profile logic

#### Form Wizard Logic (CRÍTICO)
- Validação de formulários em `wizard_categories.go`
- Validação de formulários em `wizard_agents.go`
- Validação de formulários em `wizard_other.go`
- Navegação entre steps do wizard

#### Model Registry Logic (CRÍTICO)
- CRUD de modelos
- Validação de inputs
- Navegação no formulário

## Estratégia de Implementação

### Fase 1: Quick Wins (Semana 1-2)

**Objetivo: +15-20% cobertura**

#### 1.1 Helpers e Utilitários
- [ ] `internal/tui/app.go` - Funções auxiliares
  - `joinWithSeparator()` - Teste unitário simples
  - `placeholderView()` - Teste de renderização
  - `renderShortHelp()` - Teste para cada state
  - **Estimativa**: +2% cobertura, 1 hora

- [ ] `internal/tui/views/wizard_other.go` - Helpers
  - `parseMapStringInt()` - Parser de string para map
  - `serializeMapStringInt()` - Serializer de map para string
  - **Estimativa**: +1% cobertura, 1 hora

- [ ] `internal/tui/views/wizard_agents.go` - Helpers
  - `parseMapStringBool()` - Parser de string para map
  - `serializeMapStringBool()` - Serializer de map para string
  - **Estimativa**: +1% cobertura, 1 hora

#### 1.2 Views Simples - Construtores e Getters
- [ ] `internal/tui/views/template_select.go`
  - `NewTemplateSelect()` - Construtor
  - `SetSize()` - Setter de tamanho
  - **Estimativa**: +1% cobertura, 1 hora

- [ ] `internal/tui/views/wizard_name.go`
  - `NewWizardName()` - Construtor
  - `SetName()` - Setter
  - `GetName()` - Getter
  - `IsComplete()` - Validação
  - **Estimativa**: +1% cobertura, 1 hora

#### 1.3 Config e Models - Funções Faltantes
- [ ] `internal/config/paths.go`
  - `ModelsFile()` - Teste simples de retorno de path
  - **Estimativa**: +0.5% cobertura, 30 minutos

- [ ] `internal/models/models.go`
  - `List()` - Listagem de modelos
  - **Estimativa**: +1% cobertura, 1 hora

#### 1.4 CLI Commands - Testes de Integração Leves
- [ ] `internal/cli/cmd/current.go`
  - Teste de execução do comando
  - Mock de stdout
  - **Estimativa**: +1% cobertura, 1 hora

- [ ] `internal/cli/cmd/list.go`
  - Teste de listagem de perfis
  - Mock de profile.List()
  - **Estimativa**: +1% cobertura, 1 hora

**Total Fase 1: +10-12% cobertura, 8-10 horas**

### Fase 2: Views Moderadas (Semana 3-4)

**Objetivo: +15-20% cobertura**

#### 2.1 Dashboard View
- [ ] `internal/tui/views/dashboard.go`
  - `NewDashboard()` - Construtor
  - `Init()` - Inicialização
  - `Update()` - Handler de mensagens
    - tea.KeyMsg (navegação)
    - tea.WindowSizeMsg
  - `View()` - Renderização básica
  - `handleSelect()` - Navegação para itens
  - `Refresh()` - Refresh de dashboard
  - **Estratégia**: Usar padrão dos testes existentes
  - **Estimativa**: +3% cobertura, 3 horas

#### 2.2 List View
- [ ] `internal/tui/views/list.go`
  - `NewList()` - Construtor
  - `Init()` - Inicialização
  - `LoadProfiles()` - Carregamento de perfis
  - `Update()` - Handler de mensagens
    - Navegação (↑↓)
    - Enter (switch)
    - 'd' (delete)
    - 'e' (edit)
    - '/' (search)
    - Esc (back)
  - `View()` - Renderização básica
  - **Estratégia**: Mock de profile.List(), profile.GetActive()
  - **Estimativa**: +3% cobertura, 4 horas

#### 2.3 Diff View
- [ ] `internal/tui/views/diff.go`
  - `NewDiff()` - Construtor
  - `Init()` - Inicialização
  - `loadProfiles()` - Carregamento
  - `computeDiff()` - Cálculo de diff
  - `Update()` - Handler de mensagens
    - Navegação entre panes
    - Scroll
  - `View()` - Renderização
  - **Estratégia**: Mock de profile.List(), diff.ComputeDiff()
  - **Estimativa**: +3% cobertura, 4 horas

#### 2.4 Wizard Review
- [ ] `internal/tui/views/wizard_review.go`
  - `NewWizardReview()` - Construtor
  - `SetConfig()` - Setter de config
  - `validateAndPreview()` - Validação
  - `Update()` - Handler básico
  - `IsValid()` - Validação
  - `GetErrors()` - Getter de erros
  - **Estratégia**: Mock de schema.Validator
  - **Estimativa**: +2% cobertura, 2 horas

#### 2.5 Wizard Hooks
- [ ] `internal/tui/views/wizard_hooks.go`
  - `NewWizardHooks()` - Construtor
  - `SetConfig()` - Setter
  - `SetSize()` - Setter
  - `Apply()` - Aplicação de configuração
  - `Update()` - Handler básico
  - **Estratégia**: Testar update de campos
  - **Estimativa**: +2% cobertura, 2 horas

**Total Fase 2: +13-15% cobertura, 15-17 horas**

### Fase 3: Views Complexas - Parte 1 (Semana 5-6)

**Objetivo: +15-20% cobertura**

#### 3.1 Model Selector
- [ ] `internal/tui/views/model_selector.go`
  - `NewModelSelector()` - Construtor
  - `buildItems()` - Construção de itens
  - `Init()` - Inicialização
  - `Update()` - Handler principal
    - Navegação
    - Seleção
    - Filtragem
  - `GetSelectedModel()` - Getter
  - **Estratégia**: Mock de models.List(), testar estados
  - **Estimativa**: +4% cobertura, 5 horas

#### 3.2 Model Registry
- [ ] `internal/tui/views/model_registry.go`
  - `NewModelRegistry()` - Já testado
  - `Init()` - Inicialização
  - `getFilteredModels()` - Filtragem
  - `enterAddMode()` - Modo de adição
  - `enterEditMode()` - Modo de edição
  - `resetForm()` - Reset de formulário
  - `updateFormFocus()` - Update de foco
  - `validateAndSave()` - Validação e salvamento
  - `View()` - Renderização
  - **Estratégia**: Testar estados de edição, validação
  - **Estimativa**: +5% cobertura, 6 horas

#### 3.3 Model Import
- [ ] `internal/tui/views/model_import.go`
  - `NewModelImport()` - Já testado
  - `Init()` - Inicialização
  - `Update()` - Handler principal
    - Navegação de providers
    - Navegação de modelos
    - Seleção múltipla
    - Busca
  - `getFilteredModels()` - Filtragem
  - `importSelectedModels()` - Importação
  - **Estratégia**: Mock de FetchModelsDev(), models.Add()
  - **Estimativa**: +4% cobertura, 5 horas

**Total Fase 3: +13-15% cobertura, 16-18 horas**

### Fase 4: Views Complexas - Parte 2 (Semana 7-9)

**Objetivo: +15-20% cobertura**

#### 4.1 Wizard Categories
- [ ] `internal/tui/views/wizard_categories.go`
  - `NewWizardCategories()` - Já testado
  - `SetConfig()` - Setter de configuração
  - `Apply()` - Aplicação de configuração
  - `updateFieldFocus()` - Update de foco
  - `getLineForField()` - Obter linha do campo
  - `ensureFieldVisible()` - Garantir campo visível
  - `Update()` - Handler principal
  - `renderCategoryForm()` - Renderização do formulário
  - `View()` - View principal
  - **Estratégia**: Testar cada campo, estados de foco, validação
  - **Estimativa**: +6% cobertura, 8 horas

#### 4.2 Wizard Agents
- [ ] `internal/tui/views/wizard_agents.go`
  - `NewWizardAgents()` - Parcialmente testado
  - `SetConfig()` - Setter de configuração
  - `Apply()` - Aplicação de configuração
  - `updateFieldFocus()` - Update de foco
  - `getLineForField()` - Obter linha do campo
  - `ensureFieldVisible()` - Garantir campo visível
  - `Update()` - Handler principal
  - `renderAgentForm()` - Renderização do formulário
  - `View()` - View principal
  - **Estratégia**: Testar cada campo, estados de foco, validação
  - **Estimativa**: +6% cobertura, 8 horas

**Total Fase 4: +12-15% cobertura, 16-18 horas**

### Fase 5: Core TUI Application (Semana 10-12)

**Objetivo: +10-15% cobertura**

#### 5.1 App - Navigation
- [ ] `internal/tui/app.go` - Navegação
  - `navigateTo()` - Navegação entre estados
  - Testar cada transição de estado
  - **Estratégia**: Verificar state changes e Init calls
  - **Estimativa**: +3% cobertura, 4 horas

#### 5.2 App - Message Handlers
- [ ] `internal/tui/app.go` - Handlers de mensagens
  - Toast messages (toastMsg, clearToastMsg)
  - Navigation messages (NavToListMsg, NavToWizardMsg, etc.)
  - Operation messages (switchProfileDoneMsg, etc.)
  - **Estratégia**: Testar cada handler individualmente
  - **Estimativa**: +5% cobertura, 6 horas

#### 5.3 App - Key Messages
- [ ] `internal/tui/app.go` - Handler de teclas
  - Quit key
  - Help toggle
  - Back key
  - **Estratégia**: Testar comportamento por estado
  - **Estimativa**: +2% cobertura, 2 horas

#### 5.4 App - Async Operations
- [ ] `internal/tui/app.go` - Operações assíncronas
  - `doSwitchProfile()` - Teste com mock de profile.SetActive
  - `doDeleteProfile()` - Teste com mock de profile.Delete
  - `doImportProfile()` - Teste com mocks
  - `doExportProfile()` - Teste com mocks
  - **Estratégia**: Testar retorno de mensagens
  - **Estimativa**: +3% cobertura, 4 horas

**Total Fase 5: +13-15% cobertura, 16-18 horas**

### Fase 6: Wizard Other (Semana 13-14)

**Objetivo: +5-8% cobertura**

#### 6.1 Wizard Other - Core
- [ ] `internal/tui/views/wizard_other.go`
  - `NewWizardOther()` - Já testado
  - `SetConfig()` - Setter de configuração
  - `Apply()` - Aplicação de configuração
  - `toggleSection()` - Toggle de seções
  - `toggleSubItem()` - Toggle de sub-itens
  - **Estratégia**: Testar toggle de cada seção/sub-item
  - **Estimativa**: +2% cobertura, 4 horas

#### 6.2 Wizard Other - Update e View
- [ ] `internal/tui/views/wizard_other.go`
  - `Update()` - Handler principal
  - `renderSubSection()` - Renderização de sub-seção
  - `View()` - View principal
  - **Estratégia**: Testar estados, navegação, renderização
  - **Estimativa**: +3% cobertura, 6 horas

**Total Fase 6: +5% cobertura, 10 horas**

### Fase 7: Otimização e Edge Cases (Semana 15-16)

**Objetivo: Atingir 90% nos pacotes faltantes**

#### 7.1 internal/config - Completar
- [ ] `EnsureDirs()` - Testar branch não coberto
- [ ] Edge cases em paths

#### 7.2 internal/backup - Completar
- [ ] Error handling branches
- [ ] Edge cases em Create, List, Restore, Clean

#### 7.3 internal/profile - Completar
- [ ] GetActive() - branches não cobertos
- [ ] SetActive() - branches não cobertos
- [ ] Load() - branches não cobertos

#### 7.4 internal/models - Completar
- [ ] Save() - branches não cobertos
- [ ] Load() - branches não cobertos
- [ ] Exists() - branches não cobertos
- [ ] FetchModelsDevRegistry() - teste de integração

#### 7.5 internal/schema - Completar
- [ ] GetValidator() - branches não cobertos
- [ ] Validate() - branches não cobertos

**Total Fase 7: +3-5% cobertura, 8-10 horas**

## Checklist de Testes por Arquivo

### internal/tui/app.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `NewApp()` - Cenários: Verificar inicialização de todos os componentes
- [ ] `Init()` - Cenários: Verificar batch de comandos
- [ ] `Update()` - Cenários: Todos os tipos de mensagens (KeyMsg, WindowSizeMsg, custom msgs)
- [ ] `navigateTo()` - Cenários: Transição para cada estado
- [ ] `doSwitchProfile()` - Cenários: Sucesso, erro
- [ ] `doDeleteProfile()` - Cenários: Sucesso, erro
- [ ] `doImportProfile()` - Cenários: Sucesso, colisão de nome, erro de validação, erro de leitura
- [ ] `doExportProfile()` - Cenários: Sucesso, erro
- [ ] `showToast()` - Cenários: Cada tipo de toast
- [ ] `View()` - Cenários: Cada estado, loading, toast
- [ ] `renderShortHelp()` - Cenários: Cada estado
- [ ] `renderFullHelp()` - Cenários: Cada estado
- [ ] `placeholderView()` - Cenários: Renderização básica
- [ ] `joinWithSeparator()` - Cenários: Array vazio, um item, múltiplos itens

### internal/tui/views/list.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `Title()` - Cenários: Profile ativo, inativo
- [ ] `Description()` - Cenários: Profile ativo, inativo
- [ ] `FilterValue()` - Cenários: Nome do perfil
- [ ] `newListKeyMap()` - Cenários: Verificar bindings
- [ ] `NewList()` - Cenários: Inicialização
- [ ] `Init()` - Cenários: Carregamento de perfis
- [ ] `loadProfiles()` - Cenários: Com perfis, sem perfis, erro
- [ ] `LoadProfiles()` - Cenários: Recarregamento
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `Update()` - Cenários: Todas as teclas (↑↓, enter, e, d, n, /, esc)
- [ ] `View()` - Cenários: Com perfis, sem perfis, confirmando delete
- [ ] `SelectedProfile()` - Cenários: Com seleção, sem seleção
- [ ] `IsConfirmingDelete()` - Cenários: Estados de confirmação

### internal/tui/views/dashboard.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `NewDashboard()` - Cenários: Inicialização
- [ ] `Init()` - Cenários: Carregamento
- [ ] `loadActiveProfile()` - Cenários: Com perfil ativo, sem perfil, erro
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `Update()` - Cenários: Navegação, WindowSizeMsg
- [ ] `handleSelect()` - Cenários: Cada item do menu
- [ ] `View()` - Cenários: Renderização básica
- [ ] `renderMenu()` - Cenários: Menu completo
- [ ] `Refresh()` - Cenários: Recarregamento

### internal/tui/views/diff.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `NewDiff()` - Cenários: Inicialização
- [ ] `Init()` - Cenários: Carregamento de perfis
- [ ] `loadProfiles()` - Cenários: Com perfis, sem perfis
- [ ] `computeDiff()` - Cenários: Com seleção, sem seleção
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `Update()` - Cenários: Navegação, seleção, scroll
- [ ] `handleNavigationKeys()` - Cenários: ↑↓, pgup, pgdn
- [ ] `handleSelectionKeys()` - Cenários: enter, tab
- [ ] `scrollBoth()` - Cenários: Scroll sincronizado
- [ ] `initViewports()` - Cenários: Inicialização
- [ ] `updateViewportContent()` - Cenários: Atualização de conteúdo
- [ ] `renderDiffPane()` - Cenários: Com diff, sem diff
- [ ] `View()` - Cenários: Renderização
- [ ] `borderColor()` - Cenários: Cada foco
- [ ] `renderSelector()` - Cenários: Renderização do seletor
- [ ] `ShouldReturn()` - Cenários: Condições de retorno

### internal/tui/views/wizard_other.go
**Cobertura atual**: 8.0%

**Funções não testadas**:
- [ ] `parseMapStringInt()` - Cenários: String vazia, um par, múltiplos pares, formato inválido
- [ ] `serializeMapStringInt()` - Cenários: Map vazio, um item, múltiplos itens
- [ ] `NewWizardOther()` - Já testado (100%)
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `SetConfig()` - Cenários: Config completa, config parcial, config vazia
- [ ] `Apply()` - Cenários: Aplicação de configuração
- [ ] `Update()` - Cenários: Navegação, toggle, edição de campos
- [ ] `toggleSection()` - Cenários: Toggle de cada seção
- [ ] `toggleSubItem()` - Cenários: Toggle de cada sub-item
- [ ] `renderContent()` - Já testado (96.6%)
- [ ] `renderSubSection()` - Cenários: Renderização de cada seção
- [ ] `View()` - Cenários: Renderização

### internal/tui/views/wizard_categories.go
**Cobertura atual**: 21.8%

**Funções não testadas**:
- [ ] `newCategoryConfig()` - Já testado (100%)
- [ ] `NewWizardCategories()` - Já testado (100%)
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `SetConfig()` - Cenários: Config completa, config parcial
- [ ] `Apply()` - Cenários: Aplicação de configuração
- [ ] `updateFieldFocus()` - Cenários: Navegação entre campos
- [ ] `getLineForField()` - Cenários: Cada campo
- [ ] `ensureFieldVisible()` - Cenários: Campo visível, campo fora de vista
- [ ] `Update()` - Cenários: Navegação, edição, model selector
- [ ] `renderContent()` - Já testado (95.5%)
- [ ] `renderCategoryForm()` - Já testado (91.5%)
- [ ] `View()` - Cenários: Renderização
- [ ] `handleSaveCustomModel()` - Cenários: Save, cancel
- [ ] `renderSaveCustomPrompt()` - Cenários: Renderização

### internal/tui/views/wizard_agents.go
**Cobertura atual**: 22.7%

**Funções não testadas**:
- [ ] `parseMapStringBool()` - Cenários: String vazia, um par, múltiplos pares, formato inválido
- [ ] `serializeMapStringBool()` - Cenários: Map vazio, um item, múltiplos itens
- [ ] `newAgentConfig()` - Já testado (100%)
- [ ] `NewWizardAgents()` - Cenários: Inicialização
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `SetConfig()` - Cenários: Config completa, config parcial
- [ ] `Apply()` - Cenários: Aplicação de configuração
- [ ] `updateFieldFocus()` - Cenários: Navegação entre campos
- [ ] `getLineForField()` - Cenários: Cada campo
- [ ] `ensureFieldVisible()` - Cenários: Campo visível, campo fora de vista
- [ ] `Update()` - Cenários: Navegação, edição
- [ ] `renderContent()` - Já testado (100%)
- [ ] `renderAgentForm()` - Já testado (73.7%)
- [ ] `View()` - Cenários: Renderização
- [ ] `handleSaveCustomModel()` - Cenários: Save, cancel
- [ ] `renderSaveCustomPrompt()` - Cenários: Renderização

### internal/tui/views/model_registry.go
**Cobertura atual**: 11.3%

**Funções não testadas**:
- [ ] `newModelRegistryKeyMap()` - Já testado (100%)
- [ ] `NewModelRegistry()` - Já testado (90.9%)
- [ ] `rebuildFlatModels()` - Já testado (100%)
- [ ] `getFilteredModels()` - Cenários: Com filtro, sem filtro
- [ ] `Init()` - Cenários: Inicialização
- [ ] `Update()` - Cenários: Navegação, edição, delete
- [ ] `enterAddMode()` - Já testado (100%)
- [ ] `enterEditMode()` - Cenários: Entrar em modo edição
- [ ] `resetForm()` - Já testado (100%)
- [ ] `updateFormFocus()` - Cenários: Navegação entre campos
- [ ] `updateFocusedInput()` - Cenários: Update de cada input
- [ ] `getFocusedInputValue()` - Cenários: Cada campo focado
- [ ] `validateAndSave()` - Cenários: Validação sucesso, erro, save
- [ ] `View()` - Cenários: Renderização em cada modo
- [ ] `renderList()` - Cenários: Com modelos, sem modelos
- [ ] `renderModelsList()` - Cenários: Lista de modelos
- [ ] `renderForm()` - Cenários: Formulário em cada modo
- [ ] `renderDeleteConfirm()` - Cenários: Confirmação de delete
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `IsEditing()` - Já testado (100%)

### internal/tui/views/model_import.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `newModelImportKeyMap()` - Já testado (100%)
- [ ] `NewModelImport()` - Já testado (100%)
- [ ] `Init()` - Cenários: Inicialização
- [ ] `fetchModelsDevCmd()` - Cenários: Fetch sucesso, erro
- [ ] `Update()` - Cenários: Provider list, model list, error state
- [ ] `handleProviderListKeys()` - Cenários: Navegação, seleção
- [ ] `handleModelListKeys()` - Cenários: Navegação, seleção múltipla, busca
- [ ] `handleErrorKeys()` - Cenários: Esc em erro
- [ ] `getFilteredModels()` - Cenários: Com filtro, sem filtro
- [ ] `importSelectedModels()` - Cenários: Import sucesso, parcial, erro
- [ ] `View()` - Cenários: Loading, provider list, model list, error
- [ ] `renderLoading()` - Cenários: Renderização
- [ ] `renderProviderList()` - Cenários: Com providers, sem providers
- [ ] `renderModelList()` - Cenários: Com modelos, sem modelos
- [ ] `renderError()` - Cenários: Renderização de erro
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `IsEditing()` - Já testado (100%)

### internal/tui/views/model_selector.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `newModelSelectorKeyMap()` - Cenários: Key bindings
- [ ] `NewModelSelector()` - Cenários: Inicialização
- [ ] `buildItems()` - Cenários: Com modelos, sem modelos
- [ ] `rebuildItems()` - Cenários: Rebuild
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `isSelectable()` - Cenários: Item selecionável, não selecionável
- [ ] `findNextSelectable()` - Cenários: Próximo selecionável
- [ ] `listHeight()` - Cenários: Altura da lista
- [ ] `headerHeight()` - Cenários: Altura do header
- [ ] `ensureCursorVisible()` - Cenários: Cursor visível, fora de vista
- [ ] `Update()` - Cenários: Navegação, seleção, filtro
- [ ] `View()` - Cenários: Renderização
- [ ] `renderList()` - Cenários: Lista de itens
- [ ] `renderCustomMode()` - Cenários: Modo custom
- [ ] `GetSelectedModel()` - Cenários: Com seleção, sem seleção

### internal/tui/views/wizard_review.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `newWizardReviewKeyMap()` - Já testado (100%)
- [ ] `NewWizardReview()` - Já testado (100%)
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `SetConfig()` - Já testado (100%)
- [ ] `validateAndPreview()` - Cenários: Validação sucesso, erro
- [ ] `Update()` - Cenários: Navegação
- [ ] `View()` - Cenários: Renderização com/sem erros
- [ ] `IsValid()` - Cenários: Válido, inválido
- [ ] `GetErrors()` - Cenários: Com erros, sem erros

### internal/tui/views/wizard_hooks.go
**Cobertura atual**: 33.3%

**Funções não testadas**:
- [ ] `newWizardHooksKeyMap()` - Já testado (100%)
- [ ] `NewWizardHooks()` - Já testado (100%)
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `SetConfig()` - Cenários: Config completa, config parcial
- [ ] `Apply()` - Cenários: Aplicação de configuração
- [ ] `Update()` - Cenários: Navegação, edição
- [ ] `renderContent()` - Já testado (100%)
- [ ] `View()` - Cenários: Renderização

### internal/tui/views/template_select.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `newTemplateSelectKeyMap()` - Cenários: Key bindings
- [ ] `NewTemplateSelect()` - Cenários: Inicialização
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `Update()` - Cenários: Navegação, seleção, cancel
- [ ] `View()` - Cenários: Renderização

### internal/tui/views/wizard_name.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `newWizardNameKeyMap()` - Já testado (100%)
- [ ] `NewWizardName()` - Já testado (100%)
- [ ] `Init()` - Cenários: Inicialização
- [ ] `SetSize()` - Cenários: Dimensões variadas
- [ ] `SetName()` - Já testado (100%)
- [ ] `validate()` - Cenários: Nome válido, inválido, vazio
- [ ] `Update()` - Cenários: Input, Enter, Esc
- [ ] `View()` - Cenários: Renderização
- [ ] `IsComplete()` - Cenários: Completo, incompleto
- [ ] `GetName()` - Cenários: Obter nome

### internal/models/models.go
**Cobertura atual**: 79.4%

**Funções parcialmente testadas**:
- [ ] `Load()` - Cenários: Arquivo não existe, JSON inválido
- [ ] `Save()` - Cenários: Erro de escrita
- [ ] `List()` - Cenários: Sem modelos, com modelos
- [ ] `Exists()` - Cenários: Branch não coberto

### internal/models/modelsdev.go
**Cobertura atual**: 0%

**Funções não testadas**:
- [ ] `FetchModelsDevRegistry()` - Cenários: HTTP sucesso, erro, parse JSON

### internal/config/paths.go
**Cobertura atual**: 83.3%

**Funções não testadas**:
- [ ] `ModelsFile()` - Cenários: Retorno de path
- [ ] `EnsureDirs()` - Cenários: Diretórios existem, não existem, erro

### internal/profile/active.go
**Cobertura atual**: 85.7%

**Branches não cobertas**:
- [ ] `loadActiveState()` - Cenários: Arquivo não existe
- [ ] `GetActive()` - Cenários: Orphan true, IsOrphan branch
- [ ] `SetActive()` - Cenários: Perfil não existe

### internal/profile/profile.go
**Cobertura atual**: 87.5%

**Branches não cobertas**:
- [ ] `Load()` - Cenários: JSON inválido
- [ ] `Save()` - Cenários: Erro de escrita

### internal/schema/validator.go
**Cobertura atual**: 87.9%

**Branches não cobertas**:
- [ ] `GetValidator()` - Cenários: Schema não carregado
- [ ] `Validate()` - Cenários: Erro de validação

### internal/cli/cmd/
**Cobertura atual**: 2.9%

**Funções não testadas** (apenas init está testado):
- [ ] `list.go` - Cenários: Listagem com/som perfis, erro
- [ ] `current.go` - Cenários: Com perfil ativo, sem perfil
- [ ] `switch.go` - Cenários: Switch sucesso, erro

## Recomendações Técnicas

### Setup de Infraestrutura de Testes

#### 1. Test Helpers
Criar `internal/testutil/helpers.go`:
```go
package testutil

// Helpers comuns para testes
func SetupTestDir(t *testing.T) string
func CreateTestConfig(t *testing.T, cfg config.Config) string
func CreateTestProfile(t *testing.T, name string) *profile.Profile
// Mocks de Bubbletea
```

#### 2. Mocks para Dependências Externas
- **Bubble Tea**: Criar wrappers testáveis
- **HTTP**: Mock de responses para `FetchModelsDevRegistry()`
- **Filesystem**: Usar `t.TempDir()` consistentemente

#### 3. Test Patterns

**Pattern para Views Bubble Tea:**
```go
func TestViewUpdate(t *testing.T) {
    // Arrange
    view := NewView()
    view.SetSize(80, 24)

    // Act
    updated, cmd := view.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // Assert
    assertViewNotNil(t, updated)
    assertCommandNotNil(t, cmd)
}
```

**Pattern para Async Operations:**
```go
func TestAsyncOperation(t *testing.T) {
    // Arrange
    app := NewApp()
    mockProfile := &profile.Profile{...}

    // Act
    cmd := app.doSwitchProfile("test")
    msg := cmd()

    // Assert
    switchMsg, ok := msg.(switchProfileDoneMsg)
    assert.True(t, ok)
    assert.Nil(t, switchMsg.err)
}
```

#### 4. Coverage no CI/CD
Adicionar ao `.github/workflows/test.yml`:
```yaml
- name: Test with coverage
  run: go test -coverprofile=coverage.out ./...

- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Coverage: $COVERAGE%"
    if (( $(echo "$COVERAGE < 90" | bc -l) )); then
      echo "Coverage below 90%"
      exit 1
    fi
```

### Boas Práticas

1. **Table-Driven Tests**: Usar para múltiplos cenários
2. **Subtests**: Organizar testes relacionados com `t.Run()`
3. **Test Helpers**: Criar funções helper para setups comuns
4. **Mocks**: Usar interfaces para dependências externas
5. **t.Cleanup**: Limpar recursos após testes
6. **t.TempDir**: Usar para arquivos temporários
7. **Assertions**: Considerar usar testify/assert para asserts mais expressivos

### Ferramentas Sugeridas

#### 1. Go Testing Tools
```bash
go install github.com/golang/mock/mockgen@latest  # Mock generation
go install github.com/stretchr/testify@latest      # Assertions
```

#### 2. Coverage Tools
```bash
go tool cover -html=coverage.out    # HTML report
go tool cover -func=coverage.out    # Function report
gocov test ./... | gocov report    # Alternative coverage
```

#### 3. Test Execution
```bash
go test -v ./...                    # Verbose output
go test -race ./...                 # Race detection
go test -coverprofile=coverage.out ./...
go test -run TestSpecificFunction ./...
```

## Métricas de Acompanhamento

### Objetivos Semanais

| Semana | Cobertura Meta | Pacotes/Módulos Focus | Status |
|--------|----------------|----------------------|--------|
| 1 | 35% | Helpers, Config, Models simples | Pendente |
| 2 | 40% | CLI commands básicos | Pendente |
| 3 | 50% | Dashboard, List views | Pendente |
| 4 | 55% | Diff, Wizard Review | Pendente |
| 5 | 62% | Model Selector | Pendente |
| 6 | 68% | Model Registry, Model Import | Pendente |
| 7 | 74% | Wizard Categories | Pendente |
| 8 | 78% | Wizard Agents | Pendente |
| 9 | 82% | Core TUI - Navigation | Pendente |
| 10 | 85% | Core TUI - Message Handlers | Pendente |
| 11 | 87% | Core TUI - Async Operations | Pendente |
| 12 | 88% | Wizard Other | Pendente |
| 13 | 89% | Otimizações gerais | Pendente |
| 14 | 90% | Edge cases e validação final | Pendente |

## Riscos e Mitigações

### Riscos Identificados

1. **Risco**: Complexidade do Bubble Tea
   - **Impacto**: Alto
   - **Mitigação**: Criar test helpers específicos para Bubble Tea, focar em testes de mensagens e estados

2. **Risco**: Views muito grandes (wizard_other.go com 2132 linhas)
   - **Impacto**: Médio
   - **Mitigação**: Quebrar testes em funções menores, testar por seção

3. **Risco**: Dependências externas (HTTP, filesystem)
   - **Impacto**: Médio
   - **Mitigação**: Usar t.TempDir(), criar wrappers testáveis

4. **Risco**: Testes de UI são frágeis
   - **Impacto**: Médio
   - **Mitigação**: Focar em lógica de negócio, não em renderização visual

5. **Risco**: 90% pode não ser atingível em alguns pacotes
   - **Impacto**: Baixo
   - **Mitigação**: Documentar exceções, focar em código crítico

6. **Risco**: Tempo estimado pode ser insuficiente
   - **Impacto**: Alto
   - **Mitigação**: Priorizar quick wins primeiro, ajustar estimativas

## Conclusão

### Resumo Executivo

O projeto **omo-profiler** atualmente possui **23.7% de cobertura de testes**, com uma meta de atingir **90%**. Os pacotes `internal/diff` (100%), `internal/schema` (87.9%), `internal/config` (83.3%), `internal/backup` (80%), `internal/profile` (80.4%) e `internal/models` (79.4%) já possuem boa cobertura.

O maior gap está nos pacotes TUI:
- `internal/tui/views` (20.2%)
- `internal/tui` (0%)
- `internal/cli/cmd` (2.9%)
- `internal/cli` (0%)

### Estratégia

O plano está organizado em **7 fases** ao longo de **14-16 semanas**:

1. **Quick Wins** (Semana 1-2): Helpers, utilitários, funções simples
2. **Views Moderadas** (Semana 3-4): Dashboard, List, Diff, Wizard Review
3. **Views Complexas - Parte 1** (Semana 5-6): Model Selector, Registry, Import
4. **Views Complexas - Parte 2** (Semana 7-9): Wizard Categories, Agents
5. **Core TUI Application** (Semana 10-12): App navigation, message handlers
6. **Wizard Other** (Semana 13-14): Maior view do projeto
7. **Otimização Final** (Semana 15-16): Edge cases, validações

### Próximos Passos Imediatos

**Começar pela Fase 1 - Quick Wins:**

1. Testar helpers em `app.go`: `joinWithSeparator()`, `placeholderView()`
2. Testar parsers em `wizard_other.go`: `parseMapStringInt()`, `serializeMapStringInt()`
3. Testar parsers em `wizard_agents.go`: `parseMapStringBool()`, `serializeMapStringBool()`
4. Completar `internal/config/paths.go`: `ModelsFile()`
5. Completar `internal/models/models.go`: `List()`

Esses quick wins devem adicionar **+10-12% de cobertura** em **8-10 horas** de trabalho.

### Comando para Verificar Progresso

```bash
# Executar todos os testes com coverage
go test -coverprofile=coverage.out ./...

# Ver detalhes
go tool cover -func=coverage.out

# Ver HTML report
go tool cover -html=coverage.out

# Ver coverage por pacote
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out | grep total
```
