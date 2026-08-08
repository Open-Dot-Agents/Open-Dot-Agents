# Profiles

Profiles are named runtime presets that layer on top of base
`../settings/` for mode-specific behavior.

## Purpose

- switch between development, CI, review, and release modes
- preserve reusable policy bundles without editing base settings
- avoid one-off temporary manual configuration

## Vendor mappings

- OpenAI Codex CLI: named profiles under `[profiles.<name>]` in `config.toml`
- Harnesses without native profile support: map nearest equivalent as needed (for
  example scope/precedence settings), but keep it as a distinct layer only when
  user-selected switching is explicit

## Expected content

- one profile entry per mode with override values only
- minimal diff versus `../settings/` defaults
- optional profile-specific policy/tool adjustments

## Contract

Machine metadata is in `../schema/v1/agents.schema.json`.

See `../mappings.yaml` for vendor documentation.
