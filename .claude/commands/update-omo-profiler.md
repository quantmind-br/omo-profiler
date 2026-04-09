---
description: Sincroniza a aplicação omo-profiler com o esquema upstream do oh-my-opencode.json
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, WebFetch, Agent
---

# Sincronização com esquema upstream do oh-my-opencode

A aplicação omo-profiler gerencia profiles do arquivo oh-my-opencode.json.
Este command garante que a aplicação esteja 100% alinhada com o esquema mais recente.

## Fase 1: Fetch e Comparação

1. Baixe o esquema upstream de:
   `https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/refs/heads/dev/assets/oh-my-opencode.schema.json`

2. Compare com o esquema embarcado em `internal/schema/schema.json`

3. Se idênticos, informe ao usuário e encerre.

## Fase 2: Relatório de Impacto

Analise as diferenças e mapeie o impacto nos seguintes arquivos:

| Componente | Arquivo | O que verificar |
|------------|---------|-----------------|
| Tipos Go | `internal/config/types.go` | Structs, campos, JSON tags com `omitempty`, ponteiros `*bool` |
| Schema embarcado | `internal/schema/schema.json` | Substituir pelo upstream |
| Schema raiz | `oh-my-opencode.schema.json` | Manter cópia sincronizada |
| Wizard steps | `internal/tui/views/wizard_*.go` | Novos campos editáveis precisam de UI |
| Testes | `internal/config/types_test.go` | Round-trip serialization dos novos campos |
| AGENTS.md | `internal/config/AGENTS.md` | Atualizar contagem de tipos/campos |

Regras (ver `internal/config/AGENTS.md`):
- JSON tags devem corresponder exatamente às chaves do schema upstream
- Usar `*bool` para distinguir `false` de ausente
- Todos os tags precisam de `omitempty`
- Structs devem ser pure data — sem métodos

Produza um relatório estruturado com: campos adicionados, removidos,
modificados, e impacto em cada componente acima.

## Fase 3: Implementação

Utilize a skill `/oh-my-claudecode:autopilot` para implementar todas as
modificações identificadas no relatório.

## Fase 4: Validação

Após a implementação:
1. Execute `make build` — deve compilar sem erros
2. Execute `make test` — todos os testes devem passar
3. Execute `make lint` — sem novos warnings
4. Verifique que o schema embarcado é idêntico ao upstream
