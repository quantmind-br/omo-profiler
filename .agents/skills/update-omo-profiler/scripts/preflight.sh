#!/usr/bin/env bash
# Fases 0-2 do sync: pré-flight do clone upstream, fetch+pull --ff-only da
# branch dev, e diff binário do schema embarcado contra o upstream.
# Só muta o clone upstream (git pull); nunca escreve em omo-profiler.
set -euo pipefail

UPSTREAM_CLONE="${UPSTREAM_CLONE:-$HOME/dev/oh-my-openagent}"
OMO_PROFILER_DIR="${OMO_PROFILER_DIR:-$HOME/dev/omo-profiler}"
ANCHOR_FILE="$OMO_PROFILER_DIR/internal/schema/.upstream-sha"

abort() { echo "ABORT: $1" >&2; exit 1; }

# --- Fase 0: pré-flight -----------------------------------------------------
[ -d "$UPSTREAM_CLONE" ] || abort "clone não existe em $UPSTREAM_CLONE"
git -C "$UPSTREAM_CLONE" rev-parse --is-inside-work-tree >/dev/null 2>&1 \
  || abort "$UPSTREAM_CLONE não é um git repo"

ORIGIN_URL=$(git -C "$UPSTREAM_CLONE" remote get-url origin 2>/dev/null || echo "")
case "$ORIGIN_URL" in
  *code-yeongyu/oh-my-openagent.git*) ;;
  *) abort "origin inesperado: '$ORIGIN_URL' (esperado code-yeongyu/oh-my-openagent.git)" ;;
esac

BRANCH=$(git -C "$UPSTREAM_CLONE" symbolic-ref --short HEAD 2>/dev/null || echo "DETACHED")
[ "$BRANCH" = "dev" ] \
  || abort "branch atual é '$BRANCH', esperado 'dev'. Corrija: cd $UPSTREAM_CLONE && git checkout dev"

git -C "$UPSTREAM_CLONE" diff --quiet \
  || abort "árvore tracked suja (unstaged) em $UPSTREAM_CLONE. Stash/commit antes de continuar."
git -C "$UPSTREAM_CLONE" diff --cached --quiet \
  || abort "árvore tracked suja (staged) em $UPSTREAM_CLONE. Stash/commit antes de continuar."

# --- Fase 1: anchor + pull ---------------------------------------------------
ANCHOR_OLD=$(cat "$ANCHOR_FILE" 2>/dev/null || echo "")
BOOTSTRAP=0
if [ -z "$ANCHOR_OLD" ]; then
  ANCHOR_OLD=$(git -C "$UPSTREAM_CLONE" rev-parse HEAD)
  BOOTSTRAP=1
fi

git -C "$UPSTREAM_CLONE" fetch origin dev
git -C "$UPSTREAM_CLONE" pull --ff-only origin dev

ANCHOR_NEW=$(git -C "$UPSTREAM_CLONE" rev-parse HEAD)
UPSTREAM_VERSION=$(jq -r .version "$UPSTREAM_CLONE/package.json")

# --- Fase 2: diff de schema (contrato canônico) -----------------------------
UPSTREAM_SCHEMA="$UPSTREAM_CLONE/assets/oh-my-opencode.schema.json"
EMBEDDED_SCHEMA="$OMO_PROFILER_DIR/internal/schema/schema.json"
ROOT_SCHEMA="$OMO_PROFILER_DIR/oh-my-opencode.schema.json"

SCHEMA_CHANGED=0
diff -q "$UPSTREAM_SCHEMA" "$EMBEDDED_SCHEMA" >/dev/null 2>&1 || SCHEMA_CHANGED=1

echo "BOOTSTRAP=$BOOTSTRAP"
echo "ANCHOR_OLD=$ANCHOR_OLD"
echo "ANCHOR_NEW=$ANCHOR_NEW"
echo "UPSTREAM_VERSION=$UPSTREAM_VERSION"
echo "SCHEMA_CHANGED=$SCHEMA_CHANGED"
echo
sha256sum "$UPSTREAM_SCHEMA" "$EMBEDDED_SCHEMA" "$ROOT_SCHEMA"

if [ "$BOOTSTRAP" = "0" ] && [ "$ANCHOR_OLD" = "$ANCHOR_NEW" ] && [ "$SCHEMA_CHANGED" = "0" ]; then
  echo
  echo "NO_CHANGES: upstream sem novidades desde o último anchor. Nada a fazer."
  exit 0
fi

if [ "$BOOTSTRAP" = "1" ]; then
  echo
  echo "BOOTSTRAP_RUN: sem anchor anterior — Fase 4 (drift de código) será pulada nesta execução."
fi

if [ "$SCHEMA_CHANGED" = "1" ]; then
  echo
  echo "SCHEMA_CHANGED: Fase 3 (análise de impacto do schema) é obrigatória."
else
  echo
  echo "SCHEMA_UNCHANGED: contrato de schema inalterado — Fase 3 pode ser pulada."
fi
