Com base na análise do arquivo `@schema.json` e na comparação com o código fonte da aplicação (especialmente `internal/config/types.go` e os arquivos em `internal/tui/views/`), aqui está o relatório dos parâmetros que **já existem na estrutura de dados (Go structs)**, mas **ainda não foram implementados na Interface de Usuário (TUI)** para edição.

### Resumo da Análise

O arquivo `internal/config/types.go` reflete corretamente o schema JSON. No entanto, os "Wizards" e "Editors" da TUI (`internal/tui/views/wizard_*.go`) ignoram várias seções complexas ou parâmetros detalhados.

### Parâmetros Faltantes na TUI (omo-profiler)

#### 1. Configuração de `categories` (Top-Level)
O schema define um objeto `categories` (ex: `general`, `planning`) que permite definir padrões de modelo, temperatura e configurações de pensamento ("thinking") para categorias de agentes.

*   **Status:** **Totalmente Ausente na TUI.**
*   **Onde deveria estar:** Provavelmente precisaria de um passo novo no Wizard (`WizardCategories`) ou uma seção expandida em `WizardOther`.
*   **Campos faltantes:**
    *   `model`
    *   `temperature`, `top_p`, `maxTokens`
    *   `thinking` (type, budgetTokens)
    *   `reasoningEffort`, `textVerbosity`
    *   `tools` (habilitação por categoria)

#### 2. Configuração Detalhada de `experimental`
Em `internal/tui/views/wizard_other.go`, a seção `experimental` implementa apenas os flags booleanos simples. Faltam os objetos complexos e valores numéricos.

*   **Status:** **Parcialmente Implementado.**
*   **Campos faltantes:**
    *   `preemptive_compaction_threshold` (float): A UI tem o booleano para ativar, mas não o campo para definir o limiar (ex: 0.8).
    *   `dynamic_context_pruning` (Object): A UI não permite configurar:
        *   `notification` (enum: off, minimal, detailed)
        *   `turn_protection` (enabled, turns)
        *   `protected_tools` (array de strings)
        *   `strategies` (deduplication, supersede_writes, purge_errors)

#### 3. Configuração Detalhada de `background_task`
Em `wizard_other.go`, apenas `defaultConcurrency` é configurável.

*   **Status:** **Parcialmente Implementado.**
*   **Campos faltantes:**
    *   `providerConcurrency` (Map string->int): Controle de concorrência por provedor (ex: openai: 5).
    *   `modelConcurrency` (Map string->int): Controle de concorrência por modelo.

#### 4. Configuração de `tools` dentro dos Agentes
Em `internal/tui/views/wizard_agents.go`, é possível configurar modelo, prompt, permissões, etc.

*   **Status:** **Parcialmente Implementado.**
*   **Campos faltantes:**
    *   `tools` (Map string->bool): O schema permite ativar/desativar ferramentas específicas para um agente (ex: `"bash": true`, `"webfetch": false`). A UI de agentes não possui interface para manipular esse mapa.

#### 5. Configuração Avançada de `skills` (Objeto)
O schema permite que `skills` seja um array de strings (implementado como lista de desativação) OU um objeto complexo definindo novas skills (`sources`, definições customizadas).

*   **Status:** **Apenas Visualização.**
*   **Onde:** `internal/tui/views/editor.go` (renderSkillsSection).
*   **Problema:** A aplicação trata `Skills` como `json.RawMessage`. O editor apenas exibe o JSON ("view only") ou permite desabilitar skills padrões na lista de exclusão. Não há interface visual para *criar* ou *editar* definições complexas de skills (sources, templates, etc.).

### Tabela de Ação Recomendada

| Seção | Parâmetro | Prioridade | Sugestão de Implementação |
| :--- | :--- | :--- | :--- |
| **Experimental** | `preemptive_compaction_threshold` | Alta | Adicionar `textinput` em `wizard_other.go` ao lado do checkbox. |
| **Experimental** | `dynamic_context_pruning` | Média | Criar sub-menu ou nova view devido à complexidade do objeto. |
| **Agentes** | `tools` | Alta | Adicionar lista de checkboxes (bash, edit, webfetch, etc.) no form expandido do agente. |
| **Categorias** | `categories` {*} | Média | Criar um novo passo no Wizard (`WizardCategories`) similar ao `WizardAgents`. |
| **Background** | `provider/modelConcurrency` | Baixa | Adicionar inputs de texto que aceitem formato "chave:valor". |

### Conclusão

Para que a aplicação `omo-profiler` cubra 100% do schema `@schema.json`, é necessário expandir significativamente o `WizardOther` para suportar configurações aninhadas (DCP, Concorrência) e adicionar suporte à edição da propriedade `tools` nos Agentes e a criação de `categories`.