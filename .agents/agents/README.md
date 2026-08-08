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

## Contract

See `../schema/v0.0.1/agents.schema.json` for filename and front-matter requirements.
