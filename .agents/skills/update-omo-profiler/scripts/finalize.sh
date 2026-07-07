#!/usr/bin/env bash
# Fases 6-7 do sync: valida (build/test/lint + paridade de hash dos 3
# schemas) e, só em caso de sucesso total, persiste o novo anchor upstream.
# Nunca escreve o anchor se qualquer verificação falhar.
set -euo pipefail

UPSTREAM_CLONE="${UPSTREAM_CLONE:-$HOME/dev/oh-my-openagent}"
OMO_PROFILER_DIR="${OMO_PROFILER_DIR:-$HOME/dev/omo-profiler}"
ANCHOR_FILE="$OMO_PROFILER_DIR/internal/schema/.upstream-sha"

abort() { echo "ABORT: $1 — anchor NÃO persistido, nada commitado." >&2; exit 1; }

cd "$OMO_PROFILER_DIR"
make build || abort "make build falhou"
make test  || abort "make test falhou"
make lint  || abort "make lint falhou"

UPSTREAM_SCHEMA="$UPSTREAM_CLONE/assets/oh-my-opencode.schema.json"
EMBEDDED_SCHEMA="$OMO_PROFILER_DIR/internal/schema/schema.json"
ROOT_SCHEMA="$OMO_PROFILER_DIR/oh-my-opencode.schema.json"

DISTINCT_HASHES=$(sha256sum "$UPSTREAM_SCHEMA" "$EMBEDDED_SCHEMA" "$ROOT_SCHEMA" | awk '{print $1}' | sort -u | wc -l)
if [ "$DISTINCT_HASHES" != "1" ]; then
  sha256sum "$UPSTREAM_SCHEMA" "$EMBEDDED_SCHEMA" "$ROOT_SCHEMA" >&2
  abort "hashes dos 3 schemas divergem (upstream/embedded/root devem ser idênticos)"
fi

ANCHOR_NEW=$(git -C "$UPSTREAM_CLONE" rev-parse HEAD)
UPSTREAM_VERSION=$(jq -r .version "$UPSTREAM_CLONE/package.json")

echo "$ANCHOR_NEW" > "$ANCHOR_FILE"

echo "OK: build/test/lint passaram, schemas idênticos, anchor persistido em $ANCHOR_FILE."
echo
echo "Sugestão de mensagem de commit (inclua o anchor no mesmo commit das mudanças):"
echo "  chore(schema): sync to oh-my-openagent v${UPSTREAM_VERSION} (${ANCHOR_NEW:0:7})"
