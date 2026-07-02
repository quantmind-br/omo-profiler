#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'Usage: %s <workspace> <codebase-root>\n' "${0##*/}" >&2
}

section_file() {
  local title=$1
  local path=$2
  printf '\n## %s\n\n' "$title"
  if [[ -s "$path" ]]; then
    sed -n '1,$p' "$path"
    printf '\n'
  else
    printf 'n/a — %s was not generated.\n' "$path"
  fi
}

[[ $# -eq 2 ]] || { usage; exit 2; }
workspace=$1
codebase_root=$2

if [[ ! -d "$workspace" ]]; then
  printf 'Workspace does not exist or is not a directory: %s\n' "$workspace" >&2
  exit 1
fi
if [[ ! -d "$codebase_root" ]]; then
  printf 'Codebase root does not exist or is not a directory: %s\n' "$codebase_root" >&2
  exit 1
fi

workspace=$(cd "$workspace" && pwd -P)
codebase_root=$(cd "$codebase_root" && pwd -P)
workspace_report="$workspace/REFACTOR_PLAN.md"

{
  printf '# TUI refactor plan\n\n'
  printf "> Generated from \`%s\`.\n\n" "$workspace"

  printf '## 10-second summary\n\n'
  if [[ -s "$workspace/meta.json" ]]; then
    printf -- "- Metadata: \`%s\`\n" "$workspace/meta.json"
  else
    printf -- "- Metadata: n/a — \`meta.json\` was not generated.\n"
  fi
  if [[ -s "$workspace/findings.json" ]]; then
    printf -- "- Findings: \`%s\`\n" "$workspace/findings.json"
  fi
  if [[ -s "$workspace/plan-items.json" ]]; then
    printf -- "- Structured plan items: \`%s\`\n" "$workspace/plan-items.json"
  fi

  section_file 'Current design' "$workspace/01-current-design.md"
  section_file 'Gap analysis' "$workspace/02-gap-analysis.md"

  printf '\n## Target design\n\n'
  if [[ -s "$workspace/03-target-design/00-overview.md" ]]; then
    sed -n '1,$p' "$workspace/03-target-design/00-overview.md"
    printf '\n'
  else
    printf 'n/a — target overview was not generated.\n'
  fi

  printf '\n### Target design files\n\n'
  if [[ -d "$workspace/03-target-design" ]]; then
    while IFS= read -r path; do
      printf -- "- \`%s\`\n" "$path"
    done < <(find "$workspace/03-target-design" -type f | sort)
  else
    printf 'n/a — target design directory was not generated.\n'
  fi

  section_file 'Refactor plan' "$workspace/04-refactor-plan.md"

  printf '\n## Before gallery\n\n'
  if [[ -d "$workspace/before" ]] && find "$workspace/before" -type f -print -quit | grep -q .; then
    while IFS= read -r path; do
      printf -- "- \`%s\`\n" "$path"
    done < <(find "$workspace/before" -type f | sort)
  else
    printf 'n/a — no before captures were generated.\n'
  fi

  printf '\n## Next step\n\n'
  printf "To implement, execute this plan milestone by milestone. To verify afterwards, run \`tui-validator\` on the result and diff against \`before/\`.\n"
} > "$workspace_report"

canonical="$codebase_root/TUI_REFACTOR.md"
if [[ -w "$codebase_root" ]]; then
  cp "$workspace_report" "$canonical"
  printf '%s\n' "$canonical"
else
  printf 'Codebase root is not writable; kept workspace report only: %s\n' "$workspace_report" >&2
  printf '%s\n' "$workspace_report"
fi
