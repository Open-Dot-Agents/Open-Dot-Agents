#!/usr/bin/env bash
set -euo pipefail

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd)
example_root="$repo_root/examples/copilot-project"
dota_bin="$repo_root/cli/bin/dota"
adapter_bin="$repo_root/cli/bin/dota-adapter-copilot"
adapter_id="org.open-dot-agents.copilot"
probe_root=$(mktemp -d "$repo_root/.dota-copilot-real.XXXXXX")
log_dir="$probe_root/.acceptance-logs"
acceptance_passed=0

cleanup() {
  if [ "$acceptance_passed" -eq 1 ]; then
    case "$probe_root" in
      "$repo_root"/.dota-copilot-real.*) rm -rf -- "$probe_root" ;;
      *) echo "Refusing to clean unexpected probe path: $probe_root" >&2 ;;
    esac
  else
    echo "Copilot acceptance failed; preserved probe and logs at: $probe_root" >&2
  fi
}
trap cleanup EXIT

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command not found: $1" >&2
    exit 1
  fi
}

require_command copilot
require_command git
require_command jq

if [ ! -x "$dota_bin" ] || [ ! -x "$adapter_bin" ]; then
  echo "dota binaries not found; run task cli:build" >&2
  exit 1
fi

cp -a "$example_root/." "$probe_root/"
git -C "$probe_root" init --quiet
mkdir -p "$log_dir"

"$dota_bin" adapter add --root "$probe_root" --id "$adapter_id" --version dev --path "$adapter_bin"
"$dota_bin" adapter doctor --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" validate --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" export --root "$probe_root" --adapter "$adapter_id" --dry-run --json >"$log_dir/export-plan.json"
jq -e '
  .command == "export" and .status == "ok"
  and .adapter.id == "org.open-dot-agents.copilot"
  and .changes.create == [
        ".github/agents/oda-reviewer.agent.md",
        ".github/copilot-instructions.md",
        ".github/copilot/settings.json",
        ".github/hooks/acceptance.json",
        ".github/instructions/acceptance.instructions.md",
        ".github/mcp.json",
        ".github/plugin/plugin.json",
        ".github/skills/oda-acceptance/SKILL.md",
        "services/api/.github/instructions/acceptance.instructions.md",
        "services/api/AGENTS.md"
      ]
  and (.changes.update == null) and (.changes.delete == null)
  and (.data.losses == []) and (.data.diagnostics == [])
' "$log_dir/export-plan.json" >/dev/null

"$dota_bin" export --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" check --root "$probe_root" --adapter "$adapter_id"

test -f "$probe_root/.github/copilot-instructions.md"
test -f "$probe_root/.github/copilot/settings.json"
test -f "$probe_root/.github/instructions/acceptance.instructions.md"
test -f "$probe_root/.github/agents/oda-reviewer.agent.md"
test -f "$probe_root/.github/hooks/acceptance.json"
test -f "$probe_root/.github/skills/oda-acceptance/SKILL.md"
test -f "$probe_root/.github/mcp.json"
test -f "$probe_root/.github/plugin/plugin.json"
test -f "$probe_root/.agents/.dota/export/$adapter_id.json"
test -f "$probe_root/services/api/AGENTS.md"
test -f "$probe_root/services/api/.github/instructions/acceptance.instructions.md"
jq -e '.version == 1 and (.hooks.userPromptSubmitted | length) == 1' "$probe_root/.github/hooks/acceptance.json" >/dev/null
jq -e '.mcpServers.openaiDeveloperDocs.type == "http" and .mcpServers.openaiDeveloperDocs.url == "https://developers.openai.com/mcp"' "$probe_root/.github/mcp.json" >/dev/null
jq -e '.enabledPlugins == {} and .extraKnownMarketplaces == {}' "$probe_root/.github/copilot/settings.json" >/dev/null

(
  cd "$probe_root"
  copilot --version >"$log_dir/copilot-version.txt"
  copilot skill list --json >"$log_dir/copilot-skills.json"
  copilot mcp list --json >"$log_dir/copilot-mcp.json"
  copilot --plugin-dir .github/plugin plugin list >"$log_dir/copilot-plugins.txt"
)

jq -e --arg path "$probe_root/.github/skills/oda-acceptance" '
  [.[] | select(.name == "oda-acceptance" and .source == "project" and .path == $path and .enabled == true)] | length == 1
' "$log_dir/copilot-skills.json" >/dev/null
jq -e --arg path "$probe_root/.github/mcp.json" '
  .mcpServers.openaiDeveloperDocs
  | .enabled == true and .source == "workspace" and .sourcePath == $path
    and .type == "http" and .url == "https://developers.openai.com/mcp"
' "$log_dir/copilot-mcp.json" >/dev/null
grep -F 'oda-acceptance-plugin' "$log_dir/copilot-plugins.txt" >/dev/null

live_prompt="Do not modify files or run shell commands. Use the file-view tool to read $probe_root/acceptance.txt and $probe_root/.github/instructions/acceptance.instructions.md. Copy their exact acceptance phrases without renaming them. Invoke the oda-acceptance skill and capture its acceptance phrase. Call the openaiDeveloperDocs MCP server to search for GitHub Copilot CLI custom agents. Return one compact JSON object with keys project, rule, skill, and mcp."
if ! (
  cd "$probe_root"
  copilot --no-auto-update --no-remote --no-remote-export --disable-builtin-mcps --allow-all-tools --allow-url=developers.openai.com --output-format json -p "$live_prompt" >"$log_dir/copilot-live.jsonl" 2>"$log_dir/copilot-live.stderr"
); then
  echo "Live Copilot execution failed" >&2
  sed -n '1,160p' "$log_dir/copilot-live.stderr" >&2
  exit 1
fi

nested_prompt="Do not modify files or run shell commands. Use the file-view tool to read $probe_root/services/api/acceptance.txt, $probe_root/services/api/AGENTS.md, and $probe_root/services/api/.github/instructions/acceptance.instructions.md. Copy the exact nested acceptance phrases without renaming them. Return one compact JSON object with keys nested and nested_rule."
if ! (
  cd "$probe_root/services/api"
  copilot --no-auto-update --no-remote --no-remote-export --disable-builtin-mcps --allow-all-tools --output-format json -p "$nested_prompt" >"$log_dir/copilot-nested.jsonl" 2>"$log_dir/copilot-nested.stderr"
); then
  echo "Live Copilot nested-instruction execution failed" >&2
  sed -n '1,160p' "$log_dir/copilot-nested.stderr" >&2
  exit 1
fi

agent_prompt='Return the agent acceptance phrase from your custom-agent instructions and nothing else.'
if ! (
  cd "$probe_root"
  copilot --no-auto-update --no-remote --no-remote-export --disable-builtin-mcps --allow-all-tools --output-format json --agent oda-reviewer -p "$agent_prompt" >"$log_dir/copilot-agent.jsonl" 2>"$log_dir/copilot-agent.stderr"
); then
  echo "Live Copilot custom-agent execution failed" >&2
  sed -n '1,160p' "$log_dir/copilot-agent.stderr" >&2
  exit 1
fi

assert_live_marker() {
  local file="$1"
  local marker="$2"
  local description="$3"
  if ! grep -F "$marker" "$file" >/dev/null; then
    echo "Live Copilot acceptance missing $description ($marker)" >&2
    exit 1
  fi
}

assert_live_marker "$log_dir/copilot-live.jsonl" 'ODA_PROJECT_OK' 'project instruction marker'
assert_live_marker "$log_dir/copilot-live.jsonl" 'ODA_RULE_OK' 'scoped rule marker'
assert_live_marker "$log_dir/copilot-nested.jsonl" 'ODA_COPILOT_NESTED_OK' 'nested project instruction marker'
assert_live_marker "$log_dir/copilot-nested.jsonl" 'ODA_COPILOT_NESTED_RULE_OK' 'nested scoped-rule marker'

assert_completed_view() {
  local file="$1"
  local path="$2"
  if ! jq -s -e --arg path "$path" '
    . as $events
    | any($events[];
        .type == "tool.execution_start"
        and .data.toolName == "view"
        and .data.arguments.path == $path
        and (.data.toolCallId as $id
          | any($events[];
              .type == "tool.execution_complete"
              and .data.toolCallId == $id
              and .data.success == true)))
  ' "$file" >/dev/null; then
    echo "Live Copilot acceptance missing a completed view of $path" >&2
    exit 1
  fi
}

assert_completed_view "$log_dir/copilot-live.jsonl" "$probe_root/acceptance.txt"
assert_completed_view "$log_dir/copilot-live.jsonl" "$probe_root/.github/instructions/acceptance.instructions.md"
assert_completed_view "$log_dir/copilot-nested.jsonl" "$probe_root/services/api/acceptance.txt"
assert_completed_view "$log_dir/copilot-nested.jsonl" "$probe_root/services/api/AGENTS.md"
assert_completed_view "$log_dir/copilot-nested.jsonl" "$probe_root/services/api/.github/instructions/acceptance.instructions.md"
assert_live_marker "$log_dir/copilot-live.jsonl" 'ODA_SKILL_OK' 'skill marker'
assert_live_marker "$log_dir/copilot-agent.jsonl" 'ODA_AGENT_OK' 'custom-agent marker'
assert_live_marker "$log_dir/copilot-hook.txt" 'ODA_HOOK_OK' 'hook marker'
if ! jq -s -e '
  . as $events
  | any($events[];
      .type == "tool.execution_start"
      and .data.mcpServerName == "openaiDeveloperDocs"
      and (.data.toolCallId as $id
        | any($events[];
            .type == "tool.execution_complete"
            and .data.toolCallId == $id
            and .data.success == true)))
' "$log_dir/copilot-live.jsonl" >/dev/null; then
  echo "Live Copilot acceptance missing a completed openaiDeveloperDocs MCP call" >&2
  exit 1
fi

roundtrip_root="$probe_root/.copilot-roundtrip"
mkdir -p "$roundtrip_root/.agents"
cp "$probe_root/.agents/manifest.json" "$roundtrip_root/.agents/manifest.json"
cp -a "$probe_root/.github" "$roundtrip_root/.github"
cp -a "$probe_root/services" "$roundtrip_root/services"
"$dota_bin" adapter add --root "$roundtrip_root" --id "$adapter_id" --version dev --path "$adapter_bin"
"$dota_bin" import --root "$roundtrip_root" --adapter "$adapter_id"
"$dota_bin" validate --root "$roundtrip_root"
"$dota_bin" export --root "$roundtrip_root" --adapter "$adapter_id" --force
diff -ru "$probe_root/.github" "$roundtrip_root/.github" >"$log_dir/copilot-roundtrip.diff"
diff -ru "$probe_root/services" "$roundtrip_root/services" >"$log_dir/copilot-instructions-roundtrip.diff"
test -f "$roundtrip_root/.agents/.dota/import/$adapter_id.json"
rm -rf -- "$roundtrip_root"

printf '\nDrift marker.\n' >>"$probe_root/.agents/instructions/acceptance.md"
if "$dota_bin" check --root "$probe_root" --adapter "$adapter_id" >"$log_dir/drift-check.txt" 2>&1; then
  echo "Expected canonical instruction drift to fail check --ci" >&2
  exit 1
fi
"$dota_bin" export --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" check --root "$probe_root" --adapter "$adapter_id"
"$dota_bin" clean --root "$probe_root" --adapter "$adapter_id"

if [ -d "$probe_root/.github" ]; then
  test -z "$(find "$probe_root/.github" -type f -print -quit)"
fi
test -f "$probe_root/services/api/acceptance.txt"
test ! -e "$probe_root/services/api/AGENTS.md"
test -f "$probe_root/.agents/skills/oda-acceptance/SKILL.md"

acceptance_passed=1
echo "Real .agents <-> Copilot CLI acceptance passed."
