---
name: update-omo-profiler
description: >
  Sincroniza o repositório omo-profiler com o upstream oh-my-openagent usando
  o clone local canônico em ~/dev/oh-my-openagent (branch dev). Compara o
  schema JSON embarcado (internal/schema/schema.json) contra o artefato
  upstream, propaga mudanças de schema para internal/config/types.go, os
  wizard steps, testes e AGENTS.md, detecta drift de código desde o último
  anchor (internal/schema/.upstream-sha) mesmo quando o schema não mudou,
  valida com build/test/lint e paridade de hash entre os 3 arquivos de
  schema, e só então persiste o novo anchor. Use sempre que o usuário pedir
  para atualizar/sincronizar o omo-profiler com o upstream, verificar se o
  schema ou os tipos Go estão desatualizados/com drift, checar novidades do
  oh-my-openagent, revisar campos de config novos/removidos, ou invocar
  /update-omo-profiler. Keywords: sync schema, upstream sync, schema drift,
  oh-my-openagent, oh-my-opencode.schema.json, .upstream-sha, wizard fields.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Task
---

# Sincronização omo-profiler ↔ upstream oh-my-openagent

omo-profiler gera profiles do arquivo `oh-my-openagent.json`, consumido pelo
plugin `oh-my-opencode`. Esta skill alinha schema, tipos Go, wizard e
semântica com o upstream usando o **clone local canônico** — nunca a API do
GitHub, sempre o git local.

| Recurso | Caminho |
|---|---|
| Clone upstream canônico | `~/dev/oh-my-openagent` |
| Branch obrigatório | `dev` |
| Anchor de sincronização | `internal/schema/.upstream-sha` |
| Schema upstream (artefato) | `<clone>/assets/oh-my-opencode.schema.json` |
| Schema upstream (fonte) | `<clone>/src/config/schema/oh-my-opencode-config.ts` |
| Schema embarcado | `internal/schema/schema.json` |
| Schema raiz (cópia) | `oh-my-opencode.schema.json` |

## Quando usar

- "atualiza/sincroniza o omo-profiler com o upstream"
- "verifica se o schema do omo-profiler está desatualizado"
- "tem novidade no oh-my-openagent pra puxar?"
- "checa drift de config/wizard contra o upstream"
- `/update-omo-profiler`

## Quando NÃO usar

- Mudar o schema/tipos por motivo que não seja alinhamento com upstream —
  edite `internal/config/types.go` diretamente.
- Trabalhar em uma branch do upstream diferente de `dev`, ou usar um fork —
  o anchor e o contrato desta skill assumem o clone canônico em `dev`.

## Fluxo: mecânico (script) vs. julgamento (você)

As fases 0, 1, 2, 6 e 7 são puramente mecânicas — determinísticas, sem
decisão de conteúdo — e estão nos scripts `scripts/preflight.sh` e
`scripts/finalize.sh`. As fases 3, 4 e 5 exigem ler diffs, classificar
impacto e editar código: são feitas por você, com as tabelas abaixo como
guia. Não pule etapas nem tente scriptar as fases de julgamento.

### Fase 0-2 — Pré-flight, pull, diff de schema

```bash
~/dev/skills/update-omo-profiler/scripts/preflight.sh
```

Variáveis de ambiente opcionais: `UPSTREAM_CLONE` (default
`~/dev/oh-my-openagent`), `OMO_PROFILER_DIR` (default `~/dev/omo-profiler`).

O script:
1. Valida que o clone existe, é git repo, `origin` aponta para
   `code-yeongyu/oh-my-openagent.git`, branch é `dev`, árvore tracked limpa.
   Qualquer falha **aborta** com a causa e o comando corretivo — nunca
   tenta consertar sozinho (ex.: HEAD detached → `git checkout dev`; árvore
   suja → peça ao usuário para stash/commit).
2. Lê o anchor anterior de `.upstream-sha`; se ausente, marca
   `BOOTSTRAP=1` (usa o HEAD atual como anchor "antigo" e sinaliza que a
   Fase 4 deve ser pulada nesta execução — não há baseline para diff).
3. `git fetch origin dev && git pull --ff-only origin dev`.
4. Compara `assets/oh-my-opencode.schema.json` (upstream) byte-a-byte contra
   `internal/schema/schema.json` (embarcado) e imprime os 3 sha256sum.
5. Se `ANCHOR_OLD == ANCHOR_NEW` e o schema já bate: imprime `NO_CHANGES` e
   sai — não há nada a fazer, encerre aqui e reporte ao usuário.
6. Caso contrário imprime `SCHEMA_CHANGED=0|1`, que decide se a Fase 3 é
   necessária, e segue para as fases manuais abaixo.

### Fase 3 — Análise de impacto do schema (só se `SCHEMA_CHANGED=1`)

Diff `<clone>/assets/oh-my-opencode.schema.json` contra
`internal/schema/schema.json` e mapeie cada campo +/-/modificado nestes
componentes:

| Componente | Arquivo | O que verificar |
|---|---|---|
| Tipos Go | `internal/config/types.go` | Structs, campos, JSON tags com `omitempty`, ponteiros `*bool`/`*float64`/`*int` |
| Schema embarcado | `internal/schema/schema.json` | Substituir pelo `assets/oh-my-opencode.schema.json` upstream (byte-exato) |
| Schema raiz | `oh-my-opencode.schema.json` | Cópia byte-exata do embarcado |
| Wizard steps | `internal/tui/views/wizard_*.go` | Campos novos editáveis precisam de UI |
| Testes | `internal/config/types_test.go` | Round-trip de serialização dos campos novos |
| AGENTS.md | `internal/config/AGENTS.md` | Atualizar contagem de tipos/campos |

Regras (ver `internal/config/AGENTS.md` do repo omo-profiler):
- JSON tags = chaves do schema upstream, exatamente.
- `*bool`/`*float64`/`*int` para distinguir `false`/`0` de "ausente".
- Todo tag com `omitempty`.
- Structs são pure data — sem métodos.

Produza um relatório de campos +/-/modificados e impacto por componente
antes de editar — isso vira a base da Fase 5.

### Fase 4 — Drift de código desde o último anchor (sempre, exceto bootstrap)

Mesmo com schema idêntico, o upstream pode ter mudado defaults zod, TSDoc,
enums, renames ou docs — coisas que o diff de schema sozinho não pega:

```bash
git -C ~/dev/oh-my-openagent log ${ANCHOR_OLD}..${ANCHOR_NEW} --oneline -- \
  src/config/schema/ \
  src/plugin-config.ts \
  src/plugin-handlers/category-config-resolver.ts \
  src/plugin-handlers/agent-config-handler.ts \
  src/tools/delegate-task/constants.ts \
  docs/reference/configuration.md \
  docs/guide/ \
  package.json \
  CHANGELOG.md
```

Para cada commit relevante, `git show --stat <sha>` e classifique:

| Sinal upstream | Implicação omo-profiler |
|---|---|
| Novo campo em `src/config/schema/oh-my-opencode-config.ts` | Adicionar a `types.go` + wizard se o schema o expôs |
| Mudança em zod `.default(...)` | Atualizar wizard placeholder/hint |
| Novo TSDoc em campo | Atualizar tooltip/label do wizard |
| Novo enum value | Atualizar picker do wizard |
| Rename de campo no schema | BREAKING — avaliar migração de profiles existentes |
| Bump em `package.json` | Registrar para a mensagem de commit |
| Mudança em `docs/reference/configuration.md` | Revisar texto do wizard |
| Novos defaults em `category-config-resolver.ts` | Sincronizar `wizard_categories_defaults.go` |

Sem commits relevantes → reporte "sem drift de código" e siga. Em bootstrap
(`BOOTSTRAP=1`), pule esta fase — não há baseline para o `log` range.

### Fase 5 — Implementação

Aplique todas as mudanças identificadas nas Fases 3 e 4:
- Sincronize schema embarcado e raiz (cópia byte-exata do upstream).
- Atualize os tipos Go conforme a Fase 3.
- Atualize wizard, defaults e tooltips conforme Fases 3 e 4.
- Atualize testes (round-trip dos campos novos).
- Atualize `internal/config/AGENTS.md` (contagem de tipos).

### Fase 6-7 — Validação e persistência do anchor

```bash
~/dev/skills/update-omo-profiler/scripts/finalize.sh
```

Roda `make build`, `make test`, `make lint` e confere que os sha256 dos 3
arquivos de schema (upstream/embedded/root) são idênticos. **Só** se tudo
passar, grava o novo anchor em `internal/schema/.upstream-sha` e imprime a
mensagem de commit sugerida (`chore(schema): sync to oh-my-openagent
v<versão> (<sha curto>)`) para você incluir no mesmo commit das mudanças.
Qualquer falha aborta sem persistir o anchor e sem sugerir commit — nunca
commite mudanças parciais.

## Comportamento de bootstrap

`.upstream-sha` ausente (primeira execução):
- `preflight.sh` inicializa o anchor "antigo" com o SHA pré-pull e imprime
  `BOOTSTRAP_RUN`.
- Pule a Fase 4 nesta execução (sem baseline).
- A próxima execução já terá anchor para diff de drift de código.
