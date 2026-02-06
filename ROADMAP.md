Ordem de Implementação
Fase 1 — Fundação (fazer primeiro)
| # | ID | Feature | Esforço |
|---|-----|---------|---------|
| 1 | UIUX-001 | Consolidar estilos duplicados (100+ inline → styles.go) | medium |
| 2 | UIUX-009 | Padronizar formato do help text | small |
> Por que primeiro: Toda mudança visual depois disso vira edit de 1 arquivo ao invés de 14. E padronizar help text antes de adicionar novos previne mais inconsistência.
---
Fase 2 — Quick Wins (trivial, impacto alto)
| # | ID | Feature | Esforço |
|---|-----|---------|---------|
| 3 | UIUX-006 | Select All / Deselect All nos hooks | trivial |
| 4 | UIUX-018 | Mostrar todos os erros de validação | trivial |
| 5 | UIUX-003 | Descrições nos wizard steps | trivial |
| 6 | UIUX-019 | Home/End em todas as listas | trivial |
| 7 | UIUX-014 | Ícones nos menu items do dashboard | trivial |
> Por que segundo: Cada um é <20 linhas de código. Entrega valor imediato e dá momentum.
---
Fase 3 — Core Usability (esforço médio, valor alto)
| # | ID | Feature | Esforço |
|---|-----|---------|---------|
| 8 | UIUX-004 | Confirmação ao cancelar wizard | medium |
| 9 | UIUX-005 | Agrupar hooks por categoria | medium |
| 10 | UIUX-002 | Dashboard com borda/card visual | small |
| 11 | UIUX-021 | Search / nos hooks | small |
> Por que terceiro: São as maiores melhorias visíveis pro usuário. UIUX-004 previne perda de dados, UIUX-005 transforma uma lista inutilizável em utilizável.
---
Fase 4 — Polish (refinamento)
| # | ID | Feature | Esforço |
|---|-----|---------|---------|
| 12 | UIUX-008 | Pre-fill path no export | small |
| 13 | UIUX-011 | Progress bar responsivo | small |
| 14 | UIUX-012 | Preview na seleção de template | medium |
| 15 | UIUX-015 | Atalhos numéricos 1-8 no dashboard | small |
| 16 | UIUX-016 | Loading contextual ("Switching to...") | small |
| 17 | UIUX-007 | Verificar/corrigir Esc no diff | trivial |
| 18 | UIUX-013 | Indicador ▾ nos dropdowns | trivial |
| 19 | UIUX-022 | Erros específicos no import | small |
