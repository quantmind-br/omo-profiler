# update-omo-profiler — notas de manutenção

Registros e decisões da skill. O `SKILL.md` contém apenas instruções de runtime.

## 2026-07-28 — retarget para `omo.schema.json` (upstream v4.19.2)

O upstream unificou a configuração: o arquivo canônico passou a ser
`~/.omo/omo.json[c]`, descrito por `assets/omo.schema.json`. O antigo
`oh-my-openagent.json[c]` / `oh-my-opencode.json[c]` só é lido pelo motor de
migração.

Consequência para esta skill: comparar `assets/oh-my-opencode.schema.json`
deixou de ser um sinal válido de sincronização. Aquele artefato continua sendo
gerado pelo upstream e é embutido **verbatim** como o sub-schema
`properties["[opencode]"]` dentro de `omo.schema.json` — ou seja, ele pode
permanecer idêntico enquanto o modelo de configuração inteiro muda em volta
dele. Foi exatamente o que aconteceu: o schema legado batia byte-a-byte
enquanto o formato, os caminhos, as camadas e a semântica de ativação já
tinham migrado.

Mudanças aplicadas:

- Alvo do diff: `assets/omo.schema.json` (upstream) ↔ `internal/schema/schema.json`
  (embarcado) ↔ `omo.schema.json` (raiz). O antigo `oh-my-opencode.schema.json`
  da raiz foi removido do repo.
- Caminhos da Fase 4 atualizados para o monorepo
  (`packages/omo-config-core/`, `packages/omo-opencode/`), incluindo
  `loader/paths.ts`, `loader/resolution.ts`, `loader/merge.ts` e
  `config-migration/`, que governam caminho de arquivo, precedência de ativação
  e merge de camadas — nada disso aparece no diff do schema.
- Aviso explícito no SKILL.md: `SCHEMA_UNCHANGED` não autoriza pular a Fase 4.
- Aviso explícito: nunca avançar o anchor antes de implementar as Fases 3-5.

## Ruído de submódulo no clone upstream

`preflight.sh` usa `git diff --quiet --ignore-submodules`. O clone canônico
mantém ponteiros de submódulo dessincronizados em
`packages/shared-skills/upstreams/*` como estado normal; sem a flag, o
pré-flight abortava sempre e o procedimento inteiro ficava inexecutável.
Arquivo tracked modificado continua abortando — só o ruído de ponteiro é
tolerado.

## Fonte única

`.opencode/command/update-omo-profiler.md` era uma cópia integral do
procedimento e divergiu (apontava para o schema legado e para caminhos
pré-monorepo). Foi reduzido a um ponteiro para o `SKILL.md`. Não reintroduza a
duplicação.
