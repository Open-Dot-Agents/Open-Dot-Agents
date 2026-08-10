# Plugins

Plugins are distributable bundles that combine multiple agent categories into a
single reusable package.

## Purpose

Use plugins when teams want to share:

- persona definitions
- tool integrations
- hooks and rules
- command/skill bundles

as one installable unit rather than individual file copies.

## Vendor mappings

- GitHub Copilot CLI: installable plugin packages through its marketplace model
- Other harnesses: usually emulate via direct repository sharing of equivalent
  category folders and files

## Reference adapter projection

- GitHub Copilot CLI: `mapped` between `.agents/plugins/copilot/**` and
  `.github/plugin/**`.
- OpenAI Codex: `mapped` through `.agents/plugins/codex.toml`.
- Claude Code: `unsupported` by the current adapter.

## Expected content

- plugin manifest (name, version, description, compatibility metadata)
- package contents for included categories
- optional installation metadata

## V1 contract

Configuration bundles use `plugin.json` and reverse-DNS identifiers. They are
distinct from executable dota adapter plugins. See
`../../spec/v1/schema/plugin.schema.json`.
