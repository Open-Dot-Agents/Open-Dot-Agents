#!/usr/bin/env bash
set -euo pipefail

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd)
example_root="$repo_root/examples/codex-project"
dota_bin="$repo_root/cli/bin/dota"
adapter_bin="$repo_root/cli/bin/dota-adapter-codex"
adapter_id="org.open-dot-agents.codex"
probe_root=$(mktemp -d "$repo_root/.dota-codex-real.XXXXXX")
log_dir="$probe_root/.acceptance-logs"
acceptance_passed=0

cleanup() {
  if [ "$acceptance_passed" -eq 1 ]; then
    case "$probe_root" in
      "$repo_root"/.dota-codex-real.*) rm -rf -- "$probe_root" ;;
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

if [ ! -x "$dota_bin" ] || [ ! -x "$adapter_bin" ]; then
  echo "dota binaries not found; run task cli:build" >&2
  exit 1
fi

cp -a "$example_root/." "$probe_root/"
mkdir -p "$log_dir"

"$dota_bin" adapter add --root "$probe_root" --id "$adapter_id" --version dev --path "$adapter_bin"
"$dota_bin" adapter doctor --root "$probe_root" --adapter "$adapter_id"

"$dota_bin" validate --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" export --root "$probe_root" --adapter "$adapter_id" --dry-run --json >"$log_dir/export-plan.json"
jq -e '
  .command == "export" and .status == "ok"
  and .adapter.id == "org.open-dot-agents.codex"
  and .changes.create == [".codex/agents/oda-reviewer.toml", ".codex/config.toml", ".codex/rules/acceptance.rules", "AGENTS.md", "services/payments/AGENTS.override.md", "services/search/TEAM_GUIDE.md"]
  and (.changes.update == null) and (.changes.delete == null)
  and (.data.losses == []) and (.data.diagnostics == [])
' "$log_dir/export-plan.json" >/dev/null

"$dota_bin" export --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" check --root "$probe_root" --adapter "$adapter_id"

test -f "$probe_root/AGENTS.md"
test -f "$probe_root/.codex/agents/oda-reviewer.toml"
test -f "$probe_root/.codex/config.toml"
test -f "$probe_root/.codex/rules/acceptance.rules"
test -f "$probe_root/services/payments/AGENTS.override.md"
test -f "$probe_root/services/search/TEAM_GUIDE.md"
test -f "$probe_root/.agents/.dota/export/$adapter_id.json"
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
  codex execpolicy check --pretty --rules .codex/rules/acceptance.rules -- git status --short >"$log_dir/codex-rule-check.json"
  codex mcp list --json >"$log_dir/codex-mcp.json"
  codex debug prompt-input 'Run the Open-Dot-Agents acceptance workflow.' >"$log_dir/codex-prompt.json"
  codex --cd services/payments debug prompt-input 'State the nested Codex acceptance phrase.' >"$log_dir/codex-nested-prompt.json"
  codex --cd services/search debug prompt-input 'State the fallback Codex acceptance phrase.' >"$log_dir/codex-fallback-prompt.json"
)

jq -e '.checks["auth.credentials"].status == "ok" and .checks["config.load"].status == "ok" and .checks["mcp.config"].status == "ok"' "$log_dir/codex-doctor.json" >/dev/null
jq -e '.decision == "allow"' "$log_dir/codex-rule-check.json" >/dev/null
jq -e '[.[] | select(.name == "openaiDeveloperDocs" and .enabled == true)] | length == 1' "$log_dir/codex-mcp.json" >/dev/null
jq -r '.[] | .content[]? | select(.type == "input_text") | .text' "$log_dir/codex-prompt.json" >"$log_dir/codex-prompt.txt"
grep -F 'ODA_PROJECT_OK' "$log_dir/codex-prompt.txt" >/dev/null
jq -r '.[] | .content[]? | select(.type == "input_text") | .text' "$log_dir/codex-nested-prompt.json" | grep -F 'ODA_CODEX_NESTED_OK' >/dev/null
jq -r '.[] | .content[]? | select(.type == "input_text") | .text' "$log_dir/codex-fallback-prompt.json" | grep -F 'ODA_CODEX_FALLBACK_OK' >/dev/null
test "$(grep -F -c -- '- oda-acceptance:' "$log_dir/codex-prompt.txt")" -eq 1
grep -F "$probe_root/.agents/skills/oda-acceptance/SKILL.md" "$log_dir/codex-prompt.txt" >/dev/null
if grep -F "$probe_root/.codex/skills/" "$log_dir/codex-prompt.txt" >/dev/null; then
  echo "Codex registered a duplicate generated skill" >&2
  exit 1
fi

# shellcheck disable=SC2016 # $oda-acceptance is literal Codex skill syntax.
live_prompt='Do not read project files directly. State the project acceptance phrase. Invoke $oda-acceptance and capture its acceptance phrase. Call the openaiDeveloperDocs MCP server to search for the Codex custom agent file schema. Return one compact JSON object with keys project, skill, and mcp.'
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
if ! jq -s -e 'any(.[]; .type == "item.completed" and .item.type == "mcp_tool_call" and .item.server == "openaiDeveloperDocs" and .item.status == "completed")' "$log_dir/codex-live.jsonl" >/dev/null; then
  echo "Live Codex acceptance missing a completed openaiDeveloperDocs MCP call" >&2
  exit 1
fi

agent_marker="ODA_AGENT_OK_$(date +%s%N)"
sed -i "s/ODA_AGENT_OK/$agent_marker/" "$probe_root/.codex/agents/oda-reviewer.toml"
agent_live_prompt='Use the collaboration spawn tool exactly once to delegate to the custom agent named oda-reviewer with fork_turns set to none. Ask it for the agent acceptance phrase, wait for the actual child result, and return that result. Do not read project files and do not infer or copy the agent marker yourself. If the spawn does not complete, return an error instead of an acceptance phrase.'
if ! (
  cd "$probe_root"
  codex --strict-config exec --ephemeral --sandbox read-only --json "$agent_live_prompt" >"$log_dir/codex-agent-live.jsonl" 2>"$log_dir/codex-agent-live.stderr"
); then
  echo "Live Codex custom-agent execution failed" >&2
  sed -n '1,160p' "$log_dir/codex-agent-live.stderr" >&2
  exit 1
fi
if ! grep -F "$agent_marker" "$log_dir/codex-agent-live.jsonl" >/dev/null; then
  echo "Live Codex custom-agent acceptance missing agent marker" >&2
  exit 1
fi
if jq -s -e 'any(.[]; .item.type == "command_execution")' "$log_dir/codex-agent-live.jsonl" >/dev/null; then
  echo "Live Codex custom-agent acceptance read project files instead of delegating" >&2
  exit 1
fi
sed -i "s/$agent_marker/ODA_AGENT_OK/" "$probe_root/.codex/agents/oda-reviewer.toml"

roundtrip_root="$probe_root/.codex-roundtrip"
mkdir -p "$roundtrip_root/.agents"
cp "$probe_root/.agents/manifest.json" "$roundtrip_root/.agents/manifest.json"
cp -a "$probe_root/.codex" "$roundtrip_root/.codex"
cp "$probe_root/AGENTS.md" "$roundtrip_root/AGENTS.md"
cp -a "$probe_root/.agents/skills" "$roundtrip_root/.agents/skills"
cp -a "$probe_root/services" "$roundtrip_root/services"
"$dota_bin" adapter add --root "$roundtrip_root" --id "$adapter_id" --version dev --path "$adapter_bin"
"$dota_bin" import --root "$roundtrip_root" --adapter "$adapter_id"
"$dota_bin" validate --root "$roundtrip_root"
"$dota_bin" export --root "$roundtrip_root" --adapter "$adapter_id" --force
diff -ru "$probe_root/.codex" "$roundtrip_root/.codex" >"$log_dir/codex-roundtrip.diff"
cmp "$probe_root/AGENTS.md" "$roundtrip_root/AGENTS.md"
diff -ru "$probe_root/services" "$roundtrip_root/services" >"$log_dir/codex-instructions-roundtrip.diff"
test -f "$roundtrip_root/.agents/.dota/import/$adapter_id.json"
rm -rf -- "$roundtrip_root"

if ! jq -s -e 'any(.[]; .type == "item.completed" and .item.type == "collab_tool_call" and .item.tool == "wait" and .item.status == "completed")' "$log_dir/codex-agent-live.jsonl" >/dev/null; then
  echo "Live Codex acceptance missing a completed custom-agent wait" >&2
  sed -n '1,160p' "$log_dir/codex-agent-live.stderr" >&2
  exit 1
fi

printf '\nDrift marker.\n' >>"$probe_root/.agents/instructions/acceptance.md"
if "$dota_bin" check --root "$probe_root" --adapter "$adapter_id" >"$log_dir/drift-check.txt" 2>&1; then
  echo "Expected canonical instruction drift to fail check --ci" >&2
  exit 1
fi
"$dota_bin" export --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" check --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" clean --root "$probe_root" --adapter "$adapter_id"

test ! -e "$probe_root/AGENTS.md"
if [ -d "$probe_root/.codex" ]; then
  test -z "$(find "$probe_root/.codex" -type f -print -quit)"
fi
test -f "$probe_root/.agents/skills/oda-acceptance/SKILL.md"

acceptance_passed=1
echo "Real .agents <-> Codex acceptance passed."
