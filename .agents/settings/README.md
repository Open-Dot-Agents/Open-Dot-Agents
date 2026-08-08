# Settings

Settings define the base runtime configuration for the active agent session.

## Purpose

- set default model and runtime knobs
- define environment defaults
- set default sandbox/policy behavior
- establish defaults before overlays like `../permissions/`, `../profiles/`, and
  `../tools/` are applied

## Vendor mappings

- OpenAI Codex CLI: top-level `config.toml` settings
- Claude Code: non-permission/non-hook settings in `.claude/settings.json`
- GitHub Copilot CLI configuration primitives

## Expected content

Base configuration values that represent the "current" session defaults.

## Contract

Machine metadata is in `../schema/v1/agents.schema.json`.

See `../mappings.yaml` for canonical vendor documentation.
