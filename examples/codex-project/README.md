# Codex acceptance project

This project is a complete canonical `.agents` example for exercising the
Open-Dot-Agents Codex adapter. Codex config fields are split across their
canonical settings, guardrails, hooks, memories, permissions, plugins, and
profiles categories on import and merged on export. Project command policies
round-trip between `.agents/permissions/codex-rules/**/*.rules` and
`.codex/rules/**/*.rules`; this fixture validates one with Codex's native
`execpolicy check` command.
The native discovery gate also checks a nested `AGENTS.override.md` and a
configured `TEAM_GUIDE.md` fallback without making additional model calls.

Run the end-to-end acceptance flow from the repository root:

```bash
task cli:test:codex:real
```

The task copies this project into an ignored temporary directory inside the
trusted checkout, generates Codex output there, and invokes Codex against the
generated configuration. It requires an authenticated Codex CLI, network
access, `jq`, and consumes model tokens. Before live discovery, it validates
the generated TOML against OpenAI's official Codex config JSON Schema and
loads it through the installed CLI with `--strict-config`.

Adopt an existing hand-authored `.codex/config.toml` with
`oda import --target codex` before exporting. The import retains its raw bytes
and unknown fields; review `oda export --target codex --dry-run --diff` before
using `--force --backup` on an unowned target.
