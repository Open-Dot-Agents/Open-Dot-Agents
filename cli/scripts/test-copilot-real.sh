#!/usr/bin/env bash
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
example_root="$repo_root/examples/copilot-project"
oda_bin="$repo_root/cli/bin/oda"
probe_root=$(mktemp -d "$repo_root/.oda-copilot-real.XXXXXX")
log_dir="$probe_root/.acceptance-logs"
acceptance_passed=0

cleanup() {
  if [ "$acceptance_passed" -eq 1 ]; then
    case "$probe_root" in
      "$repo_root"/.oda-copilot-real.*) rm -rf -- "$probe_root" ;;
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

if [ ! -x "$oda_bin" ]; then
  echo "oda binary not found at $oda_bin; run task cli:build" >&2
  exit 1
fi

cp -a "$example_root/." "$probe_root/"
git -C "$probe_root" init --quiet
mkdir -p "$probe_root/.agents/schema/v0.0.1" "$log_dir"
cp "$repo_root/.agents/manifest.json" "$probe_root/.agents/manifest.json"
cp "$repo_root/.agents/mappings.yaml" "$probe_root/.agents/mappings.yaml"
cp "$repo_root/.agents/schema/v0.0.1/agents.schema.json" "$probe_root/.agents/schema/v0.0.1/agents.schema.json"
cp "$repo_root/.agents/schema/v0.0.1/mappings.schema.json" "$probe_root/.agents/schema/v0.0.1/mappings.schema.json"

"$oda_bin" --root "$probe_root" --target copilot validate
"$oda_bin" --root "$probe_root" --target copilot generate --dry-run --diff --format=json >"$log_dir/generate-plan.json"
jq -e '
  .targets == [{
    command: "generate",
    target: "copilot",
    status: "ok",
    plan: {
      create: [
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
      ],
      update: [],
      delete: []
    },
    diff: [
      "A .github/agents/oda-reviewer.agent.md",
      "A .github/copilot-instructions.md",
      "A .github/copilot/settings.json",
      "A .github/hooks/acceptance.json",
      "A .github/instructions/acceptance.instructions.md",
      "A .github/mcp.json",
      "A .github/plugin/plugin.json",
      "A .github/skills/oda-acceptance/SKILL.md",
      "A services/api/.github/instructions/acceptance.instructions.md",
      "A services/api/AGENTS.md"
    ]
  }]
' "$log_dir/generate-plan.json" >/dev/null

"$oda_bin" --root "$probe_root" --target copilot generate
"$oda_bin" --root "$probe_root" --target copilot check --ci

test -f "$probe_root/.github/copilot-instructions.md"
test -f "$probe_root/.github/copilot/settings.json"
test -f "$probe_root/.github/instructions/acceptance.instructions.md"
test -f "$probe_root/.github/agents/oda-reviewer.agent.md"
test -f "$probe_root/.github/hooks/acceptance.json"
test -f "$probe_root/.github/skills/oda-acceptance/SKILL.md"
test -f "$probe_root/.github/mcp.json"
test -f "$probe_root/.github/plugin/plugin.json"
test -f "$probe_root/.github/.open-dot-agents.json"
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

live_prompt='Do not modify files or run shell commands. State the project acceptance phrase. For acceptance.txt, state the scoped rule phrase. Invoke the oda-acceptance skill and capture its acceptance phrase. Call the openaiDeveloperDocs MCP server to search for GitHub Copilot CLI custom agents. Return one compact JSON object with keys project, rule, skill, and mcp.'
if ! (
  cd "$probe_root"
  copilot --no-auto-update --no-remote --no-remote-export --disable-builtin-mcps --allow-all-tools --allow-url=developers.openai.com --output-format json -p "$live_prompt" >"$log_dir/copilot-live.jsonl" 2>"$log_dir/copilot-live.stderr"
); then
  echo "Live Copilot execution failed" >&2
  sed -n '1,160p' "$log_dir/copilot-live.stderr" >&2
  exit 1
fi

nested_prompt="You must use the file-view tool to read $probe_root/services/api/acceptance.txt. Do not modify files or run shell commands. Then copy the exact nested Copilot acceptance phrase and the exact nested scoped-rule phrase. Return one compact JSON object with keys nested and nested_rule."
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
if ! jq -s -e '
  . as $events
  | any($events[];
      .type == "tool.execution_start"
      and .data.toolName == "view"
      and (.data.toolCallId as $id
        | any($events[];
            .type == "tool.execution_complete"
            and .data.toolCallId == $id
            and .data.success == true)))
' "$log_dir/copilot-nested.jsonl" >/dev/null; then
  echo "Live Copilot nested acceptance missing a completed file-view call" >&2
  exit 1
fi
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
mkdir -p "$roundtrip_root"
cp -a "$probe_root/.github" "$roundtrip_root/.github"
cp -a "$probe_root/services" "$roundtrip_root/services"
"$oda_bin" --root "$roundtrip_root" --target copilot import
"$oda_bin" --root "$roundtrip_root" --target copilot export --force
diff -ru "$probe_root/.github" "$roundtrip_root/.github" >"$log_dir/copilot-roundtrip.diff"
diff -ru "$probe_root/services" "$roundtrip_root/services" >"$log_dir/copilot-instructions-roundtrip.diff"
test -f "$roundtrip_root/.agents/.open-dot-agents-import-copilot.json"
rm -rf -- "$roundtrip_root"

printf '\nDrift marker.\n' >>"$probe_root/.agents/instructions/acceptance.md"
if "$oda_bin" --root "$probe_root" --target copilot check --ci >"$log_dir/drift-check.txt" 2>&1; then
  echo "Expected canonical instruction drift to fail check --ci" >&2
  exit 1
fi
"$oda_bin" --root "$probe_root" --target copilot generate
"$oda_bin" --root "$probe_root" --target copilot check --ci
"$oda_bin" --root "$probe_root" --target copilot clean

test ! -e "$probe_root/.github"
test -f "$probe_root/services/api/acceptance.txt"
test ! -e "$probe_root/services/api/AGENTS.md"
test -f "$probe_root/.agents/skills/oda-acceptance/SKILL.md"

acceptance_passed=1
echo "Real .agents <-> Copilot CLI acceptance passed."
