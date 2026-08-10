# Agents

Defines reusable, named personas or sub-agents with dedicated instructions.

## Purpose

Use this category for role-specific behavior that should be reusable across tasks, such as:

- security review
- test design
- architecture and API design

## Mappings

- GitHub Copilot CLI: `.github/agents/*.md`
- OpenAI Codex CLI: `.codex/agents/*.toml`
- Claude Code: `.claude/agents/*.md`

## Expected content

- One file per agent, typically Markdown.
- Front matter includes:
  - `description` (required)
  - optional persona metadata (e.g. allowed tools, model preference).

Copilot agent Markdown is copied without rewriting its front matter. Codex TOML
is mapped to Markdown; imported TOML is retained under
`../settings/codex-agents/` so unknown fields and comments survive a round trip.

## V1 contract

Files use Markdown with YAML front matter. `name` and `description` are required;
portable optional fields are `tools`, `skills`, `permissionProfile`, `execution`,
and `maxTurns`. Vendor-only metadata belongs under `../extensions/`.
