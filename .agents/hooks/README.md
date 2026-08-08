# Hooks

Hooks are deterministic automations bound to session lifecycle events. They run
outside core reasoning and are useful for policy checks, logging, linting, and
notification workflows.

## Purpose

Typical hook points include:

- pre/post tool invocation
- session start/end
- before compaction or handoff
- custom task-specific triggers (where supported)

## Vendor mappings

- GitHub Copilot CLI hooks in `.github/hooks/*.json`
- Claude Code hooks in `.claude/settings.json`
- OpenAI Codex CLI hooks in `.codex/config.toml`

## Current oda projection

- GitHub Copilot CLI: `supported`
- OpenAI Codex: `mapped` through `.agents/hooks/codex.toml`
- Anthropic Claude Code: `supported`

## Expected content

Each hook entry should define:

- trigger/event name
- command or script to execute
- optional arguments and environment
- failure handling semantics where supported

Format is usually JSON or YAML depending on the target harness.

## Contract

Machine metadata is in `../schema/v0.0.1/agents.schema.json`, including
`hooks/*.json` payload expectations and merge behavior.

See `../mappings.yaml` for links to vendor docs.
