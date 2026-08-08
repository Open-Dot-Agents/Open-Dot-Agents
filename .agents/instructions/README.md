# Instructions

Instructions are unconditional, always-on context that applies to every agent
session. They provide the baseline operating model for the workspace.

## Purpose

- repository architecture and conventions
- coding and review standards
- common build/test expectations
- persistent communication and behavior expectations

## Scope note

Differentiate clearly from `../rules/`:

- `instructions`: always active, no trigger required
- `rules`: conditionally active based on path, trigger, or context

When a harness conflates the two, treat its equivalent file as
`instructions` by default.

## Vendor mappings

- Claude Code: `CLAUDE.md`
- GitHub Copilot CLI: `AGENTS.md`, `.github/copilot-instructions.md`
- OpenAI Codex CLI: `AGENTS.md`

## Expected content

One or more Markdown files containing high-level global guidance.

## Contract

Machine metadata is in `../schema/v1/agents.schema.json`, including filename and
parser expectations.

See `../mappings.yaml` for canonical vendor documentation.
