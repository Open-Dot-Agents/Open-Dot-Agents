#!/usr/bin/env bash
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
example_root="$repo_root/examples/codex-project"
oda_bin="$repo_root/cli/bin/oda"
probe_root=$(mktemp -d "$repo_root/.oda-codex-real.XXXXXX")
log_dir="$probe_root/.acceptance-logs"
acceptance_passed=0

cleanup() {
  if [ "$acceptance_passed" -eq 1 ]; then
    case "$probe_root" in
      "$repo_root"/.oda-codex-real.*) rm -rf -- "$probe_root" ;;
      *) echo "Refusing to clean unexpected probe path: $probe_root" >&2 ;;
    esac
  else
    echo "Codex acceptance failed; preserved probe and logs at: $probe_root" >&2
  fi
}
trap cleanup EXIT

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command not found: $1" >&2
    exit 1
  fi
}

require_command codex
require_command jq

if [ ! -x "$oda_bin" ]; then
  echo "oda binary not found at $oda_bin; run task cli:build" >&2
  exit 1
fi

# Keep the checked-in example self-contained and synchronized with the
# canonical v0.0.1 contract used by the CLI.
cmp "$repo_root/.agents/manifest.json" "$example_root/.agents/manifest.json"
cmp "$repo_root/.agents/mappings.yaml" "$example_root/.agents/mappings.yaml"
cmp "$repo_root/.agents/schema/v0.0.1/agents.schema.json" "$example_root/.agents/schema/v0.0.1/agents.schema.json"
cmp "$repo_root/.agents/schema/v0.0.1/mappings.schema.json" "$example_root/.agents/schema/v0.0.1/mappings.schema.json"

cp -a "$example_root/." "$probe_root/"
mkdir -p "$log_dir"

"$oda_bin" --root "$probe_root" --target codex validate
"$oda_bin" --root "$probe_root" --target codex generate --dry-run --diff --format=json >"$log_dir/generate-plan.json"
jq -e '
  .targets == [{
    command: "generate",
    target: "codex",
    status: "ok",
    plan: {
      create: [".codex/agents/oda-reviewer.toml", ".codex/config.toml", "AGENTS.md"],
      update: [],
      delete: []
    },
    diff: ["A .codex/agents/oda-reviewer.toml", "A .codex/config.toml", "A AGENTS.md"]
  }]
' "$log_dir/generate-plan.json" >/dev/null

"$oda_bin" --root "$probe_root" --target codex generate
"$oda_bin" --root "$probe_root" --target codex check --ci

test -f "$probe_root/AGENTS.md"
test -f "$probe_root/.codex/agents/oda-reviewer.toml"
test -f "$probe_root/.codex/config.toml"
test -f "$probe_root/.codex/.open-dot-agents.json"
test ! -e "$probe_root/.codex/skills"
test -f "$probe_root/.agents/skills/oda-acceptance/SKILL.md"
grep -Fx '#:schema https://developers.openai.com/codex/config-schema.json' "$probe_root/.codex/config.toml" >/dev/null

(
  cd "$repo_root/cli"
  go run ./cmd/codex-config-validator --config "$probe_root/.codex/config.toml"
) >"$log_dir/codex-schema-validation.txt"

(
  cd "$probe_root"
  codex --version >"$log_dir/codex-version.txt"
  codex --strict-config doctor --json >"$log_dir/codex-doctor.json" || true
  codex mcp list --json >"$log_dir/codex-mcp.json"
  codex debug prompt-input 'Run the Open-Dot-Agents acceptance workflow.' >"$log_dir/codex-prompt.json"
)

jq -e '.checks["auth.credentials"].status == "ok" and .checks["config.load"].status == "ok" and .checks["mcp.config"].status == "ok"' "$log_dir/codex-doctor.json" >/dev/null
jq -e '[.[] | select(.name == "openaiDeveloperDocs" and .enabled == true)] | length == 1' "$log_dir/codex-mcp.json" >/dev/null
jq -r '.[] | .content[]? | select(.type == "input_text") | .text' "$log_dir/codex-prompt.json" >"$log_dir/codex-prompt.txt"
grep -F 'ODA_PROJECT_OK' "$log_dir/codex-prompt.txt" >/dev/null
test "$(grep -F -c -- '- oda-acceptance:' "$log_dir/codex-prompt.txt")" -eq 1
grep -F "$probe_root/.agents/skills/oda-acceptance/SKILL.md" "$log_dir/codex-prompt.txt" >/dev/null
if grep -F "$probe_root/.codex/skills/" "$log_dir/codex-prompt.txt" >/dev/null; then
  echo "Codex registered a duplicate generated skill" >&2
  exit 1
fi

live_prompt='Do not read project files directly. State the project acceptance phrase. Invoke $oda-acceptance and capture its acceptance phrase. Call the openaiDeveloperDocs MCP server to search for the Codex custom agent file schema. Then you must call the collaboration spawn tool with agent_type oda-reviewer and fork_turns none, ask it for the agent acceptance phrase, and wait for its actual result. A passing response requires the real collaboration tool call; do not infer or copy the agent marker yourself. Return one compact JSON object with keys project, skill, mcp, and agent.'
if ! (
  cd "$probe_root"
  codex --strict-config exec --ephemeral --sandbox read-only --json "$live_prompt" >"$log_dir/codex-live.jsonl" 2>"$log_dir/codex-live.stderr"
); then
  echo "Live Codex execution failed" >&2
  sed -n '1,160p' "$log_dir/codex-live.stderr" >&2
  exit 1
fi

assert_live_marker() {
  local marker="$1"
  local description="$2"
  if ! grep -F "$marker" "$log_dir/codex-live.jsonl" >/dev/null; then
    echo "Live Codex acceptance missing $description ($marker)" >&2
    sed -n '1,160p' "$log_dir/codex-live.stderr" >&2
    exit 1
  fi
}

assert_live_marker 'ODA_PROJECT_OK' 'project instruction marker'
assert_live_marker 'ODA_SKILL_OK' 'skill marker'
assert_live_marker 'ODA_AGENT_OK' 'custom-agent marker'
if ! jq -s -e 'any(.[]; .type == "item.completed" and .item.type == "mcp_tool_call" and .item.server == "openaiDeveloperDocs" and .item.status == "completed")' "$log_dir/codex-live.jsonl" >/dev/null; then
  echo "Live Codex acceptance missing a completed openaiDeveloperDocs MCP call" >&2
  exit 1
fi
if ! jq -s -e 'any(.[]; .type == "item.completed" and .item.type == "collab_tool_call" and .item.status == "completed" and (.item.receiver_thread_ids | length) > 0)' "$log_dir/codex-live.jsonl" >/dev/null; then
  echo "Live Codex acceptance missing a completed custom-agent child thread" >&2
  sed -n '1,160p' "$log_dir/codex-live.stderr" >&2
  exit 1
fi

printf '\nDrift marker.\n' >>"$probe_root/.agents/instructions/acceptance.md"
if "$oda_bin" --root "$probe_root" --target codex check --ci >"$log_dir/drift-check.txt" 2>&1; then
  echo "Expected canonical instruction drift to fail check --ci" >&2
  exit 1
fi
"$oda_bin" --root "$probe_root" --target codex generate
"$oda_bin" --root "$probe_root" --target codex check --ci
"$oda_bin" --root "$probe_root" --target codex clean

test ! -e "$probe_root/AGENTS.md"
test ! -e "$probe_root/.codex"
test -f "$probe_root/.agents/skills/oda-acceptance/SKILL.md"

acceptance_passed=1
echo "Real .agents -> Codex acceptance passed."
