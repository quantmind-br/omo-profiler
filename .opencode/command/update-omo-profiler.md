---
description: Sincroniza omo-profiler com upstream oh-my-openagent via clone local em /home/diogo/dev/oh-my-openagent
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, Agent
---

# Sincronização com upstream oh-my-openagent

omo-profiler gere profiles do arquivo `oh-my-openagent.json` (consumido pelo
plugin `oh-my-opencode`). Este comando alinha schema, tipos, defaults e
semântica com o upstream usando o **clone local canônico**.

| Recurso | Caminho |
|---|---|
| Clone upstream canônico | `/home/diogo/dev/oh-my-openagent` |
| Branch obrigatório | `dev` |
| Anchor de sincronização | `internal/schema/.upstream-sha` |
| Schema upstream (artefato) | `<clone>/assets/oh-my-opencode.schema.json` |
| Schema upstream (fonte) | `<clone>/src/config/schema/oh-my-opencode-config.ts` |

---

## Fase 0 — Pré-flight do clone

Em `/home/diogo/dev/oh-my-openagent`:

1. Diretório existe e é git repo (`git rev-parse --is-inside-work-tree`).
2. `origin` aponta para `code-yeongyu/oh-my-openagent.git`.
3. Branch atual = `dev` (`git symbolic-ref --short HEAD` deve retornar `dev`).
4. Árvore tracked limpa: `git diff --quiet` E `git diff --cached --quiet`.
   Untracked é OK.

Falha em qualquer item: **abortar** com instrução corretiva. Não tentar consertar
silenciosamente. Exemplos:

- HEAD detached → `cd /home/diogo/dev/oh-my-openagent && git checkout dev`
- Árvore suja → pedir ao usuário stash/commit antes de continuar.

## Fase 1 — Anchor + pull

```bash
cd /home/diogo/dev/oh-my-openagent

# Anchor anterior (de omo-profiler)
ANCHOR_OLD=$(cat /home/diogo/dev/omo-profiler/internal/schema/.upstream-sha 2>/dev/null)
BOOTSTRAP=0
if [ -z "$ANCHOR_OLD" ]; then
  ANCHOR_OLD=$(git rev-parse HEAD)
  BOOTSTRAP=1   # primeira execução, pula Fase 4
fi

git fetch origin dev
git pull --ff-only origin dev

ANCHOR_NEW=$(git rev-parse HEAD)
UPSTREAM_VERSION=$(jq -r .version package.json)
```

Se `ANCHOR_OLD == ANCHOR_NEW` E schema embedded já bate com `assets/`:
upstream sem novidades → reportar e **encerrar** (Fases 2-7 desnecessárias).

## Fase 2 — Diff de schema (contrato canônico)

Comparar bytes diretamente, sem rede:

```bash
diff -q /home/diogo/dev/oh-my-openagent/assets/oh-my-opencode.schema.json \
        /home/diogo/dev/omo-profiler/internal/schema/schema.json
sha256sum /home/diogo/dev/oh-my-openagent/assets/oh-my-opencode.schema.json \
          /home/diogo/dev/omo-profiler/internal/schema/schema.json \
          /home/diogo/dev/omo-profiler/oh-my-opencode.schema.json
```

- Idênticos → contrato inalterado, **pular Fase 3**.
- Divergem → Fase 3 obrigatória.

## Fase 3 — Análise de impacto schema (somente se schema mudou)

Mapear diffs em:

| Componente | Arquivo | O que verificar |
|------------|---------|-----------------|
| Tipos Go | `internal/config/types.go` | Structs, campos, JSON tags com `omitempty`, ponteiros `*bool` |
| Schema embarcado | `internal/schema/schema.json` | Substituir pelo `assets/oh-my-opencode.schema.json` upstream |
| Schema raiz | `oh-my-opencode.schema.json` | Cópia byte-exata do embarcado |
| Wizard steps | `internal/tui/views/wizard_*.go` | Novos campos editáveis precisam de UI |
| Testes | `internal/config/types_test.go` | Round-trip serialization dos novos campos |
| AGENTS.md | `internal/config/AGENTS.md` | Atualizar contagem de tipos/campos |

Regras (ver `internal/config/AGENTS.md`):
- JSON tags = chaves do schema upstream, exatamente
- `*bool` para distinguir `false` de ausente
- Todos tags com `omitempty`
- Structs pure data — sem métodos

Produzir relatório com: campos +/-/modificados e impacto por componente.

## Fase 4 — Drift de código upstream desde anchor

**Sempre executa** (exceto bootstrap), mesmo com schema idêntico. Captura
mudanças que o schema sozinho perde: defaults zod, TSDoc, enums novos,
renames, docs.

```bash
cd /home/diogo/dev/oh-my-openagent

git log ${ANCHOR_OLD}..${ANCHOR_NEW} --oneline -- \
  src/config/schema/ \
  src/plugin-config.ts \
  src/plugin-handlers/category-config-resolver.ts \
  src/plugin-handlers/agent-config-handler.ts \
  src/tools/delegate-task/constants.ts \
  docs/reference/configuration.md \
  docs/guide/ \
  package.json \
  CHANGELOG.md 2>/dev/null
```

Para cada commit relevante, examinar `git show --stat <sha>` e classificar:

| Sinal upstream | Implicação omo-profiler |
|---|---|
| Novo campo em `src/config/schema/oh-my-opencode-config.ts` | Adicionar a `types.go` + wizard se schema o expôs |
| Mudança em zod `.default(...)` | Atualizar wizard placeholder/hint |
| Novo TSDoc em campo | Atualizar tooltip/label do wizard |
| Novo enum value | Atualizar picker do wizard |
| Rename de campo no schema | BREAKING — avaliar migração de profiles existentes |
| Bump em `package.json` | Registrar para nota de commit |
| Mudança em `docs/reference/configuration.md` | Revisar texto do wizard |
| Novos defaults em `category-config-resolver.ts` | Sincronizar `wizard_categories_defaults.go` |

Se nenhum commit relevante → reportar "sem drift de código" e seguir.

**Bootstrap (`BOOTSTRAP=1`)**: pular Fase 4 — sem baseline para diff.

## Fase 5 — Implementação

Aplicar todas as mudanças identificadas nas Fases 3 e 4. Usar
`/oh-my-claudecode:autopilot` ou Edits diretos:

- Sincronizar schema embarcado e raiz (`cp` byte-exato do upstream)
- Atualizar tipos Go conforme Fase 3
- Atualizar wizard, defaults, tooltips conforme Fases 3 e 4
- Atualizar testes (round-trip dos novos campos)
- Atualizar `internal/config/AGENTS.md` (contagem de tipos)

## Fase 6 — Validação

```bash
cd /home/diogo/dev/omo-profiler
make build   # compila sem erros
make test    # todos passam
make lint    # sem novos warnings

# Confirmar paridade binária dos schemas
sha256sum /home/diogo/dev/oh-my-openagent/assets/oh-my-opencode.schema.json \
          internal/schema/schema.json \
          oh-my-opencode.schema.json
# Os três hashes DEVEM ser idênticos.
```

## Fase 7 — Persistir anchor

```bash
echo $ANCHOR_NEW > /home/diogo/dev/omo-profiler/internal/schema/.upstream-sha
```

Adicionar ao mesmo commit das mudanças. Mensagem sugerida:

```
chore(schema): sync to oh-my-openagent v${UPSTREAM_VERSION} (${ANCHOR_NEW:0:7})
```

---

## Comportamento de bootstrap

`internal/schema/.upstream-sha` ausente:
- Inicializar com SHA atual antes do pull.
- Pular Fase 4 nesta execução (sem baseline).
- Próxima execução já terá anchor para diff.

## Falhas e abort

Qualquer falha em pré-flight, pull (`--ff-only` rejeitou), build, test ou
hash mismatch da Fase 6: **abortar**, não persistir anchor, reportar causa
raiz com comando corretivo. Nunca commitar mudanças parciais.
