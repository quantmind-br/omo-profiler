---
description: Sincroniza omo-profiler com upstream oh-my-openagent via clone local em /home/diogo/dev/oh-my-openagent
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, Agent
---

# Sincronização com upstream oh-my-openagent

Execute a skill `update-omo-profiler`. Ela é a **fonte única** do procedimento:
leia `.agents/skills/update-omo-profiler/SKILL.md` e siga as fases na ordem.

Este arquivo não repete o procedimento de propósito — a duplicação anterior
divergiu do upstream (apontava para o schema legado `oh-my-opencode.schema.json`
e para caminhos pré-monorepo) e escondeu uma migração inteira de formato.

Resumo operacional:

1. Fases 0-2 (mecânicas): `.agents/skills/update-omo-profiler/scripts/preflight.sh`
2. Fases 3-5 (julgamento): análise de impacto de schema e de drift de código,
   depois implementação — conforme as tabelas do SKILL.md.
3. Fases 6-7 (mecânicas): `.agents/skills/update-omo-profiler/scripts/finalize.sh`

Regra dura: **não avance o anchor** (`internal/schema/.upstream-sha`) antes de
implementar e validar tudo. `finalize.sh` é quem grava, e só após build/test/lint
e paridade de hash dos três schemas.
