# Prompts

Prompts are reusable, parameterized templates invoked on demand. Unlike
`../instructions/`, they are not loaded automatically and are selected
explicitly by user intent.

## Purpose

- provide repeatable command-and-response formats
- standardize reviews, triage, refactoring, and other recurring tasks
- keep context focused through bounded templates

## Vendor mappings

- GitHub Copilot CLI: `.github/prompts/*.prompt.md`
- Claude Code: slash-command based custom prompts
- OpenAI Codex CLI: custom slash commands / workflows

## Current oda projection

- GitHub Copilot CLI: `mapped` bidirectionally to
  `.github/prompts/*.prompt.md`.
- OpenAI Codex and Claude Code: `unsupported` by the current adapters.

## Expected content

One file per prompt template, typically Markdown with variable placeholders.

## Contract

Machine metadata is in `../schema/v0.0.1/agents.schema.json`.

See `../mappings.yaml` for canonical links.
