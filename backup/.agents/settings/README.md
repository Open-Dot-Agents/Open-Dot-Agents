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

## Reference adapter projection

- OpenAI Codex: `mapped` through `.agents/settings/codex.toml`. Import also
  records `codex.raw.toml` and `codex-agents/*.toml` as lossless provenance
  sidecars for comments and vendor fields that have no canonical equivalent.
- GitHub Copilot CLI: `mapped` byte-for-byte between
  `.agents/settings/copilot.json` and `.github/copilot/settings.json`, including
  repository `enabledPlugins` and `extraKnownMarketplaces` configuration.
- Claude Code: `unsupported` by the current adapter.

## Expected content

Base configuration values that represent the "current" session defaults.

## V1 contract

Portable settings use strict JSON and the fields in
`../../spec/v1/schema/settings.schema.json`. Vendor settings belong under
`../extensions/`.
