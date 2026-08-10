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

## Reference adapter projection

- GitHub Copilot CLI: `supported`
- OpenAI Codex: `mapped` through `.agents/hooks/codex.toml`
- Anthropic Claude Code: `supported`

## Expected content

Each hook entry should define:

- trigger/event name
- command or script to execute
- optional arguments and environment
- failure handling semantics where supported

V1 hook files are strict JSON and use the portable event, action, timeout,
priority, matcher, and failure semantics defined by the specification.

## V1 contract

See `../../spec/v1/schema/hooks.schema.json`. Adapter-specific hook fields must
be preserved under `../extensions/` and reported in loss diagnostics.
