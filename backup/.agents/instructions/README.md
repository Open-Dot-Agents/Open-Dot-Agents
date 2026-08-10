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
- GitHub Copilot CLI: repository-path `AGENTS.md` files and
  `.github/copilot-instructions.md` files
- OpenAI Codex CLI: repository-path `AGENTS.md`, `AGENTS.override.md`, and
  configured fallback instruction filenames

For Copilot exports, canonical `instructions/AGENTS.md` remains the root
`AGENTS.md`; other instruction files are merged into
`.github/copilot-instructions.md`. Both files import back independently.

Lossless path-specific harness extensions are stored without changing their
bytes or relative paths:

- `instructions/copilot-project/**` maps to Copilot's nested repository
  instruction paths.
- `instructions/codex-project/**` maps to Codex's nested instructions,
  overrides, and filenames declared by `project_doc_fallback_filenames`.

Nested Copilot `.github/instructions/**/*.instructions.md` files use the
parallel `rules/copilot-project/**` passthrough because they remain scoped
rules. These reserved subtrees preserve harness precedence while keeping the
portable root forms concise.

## Expected content

One or more Markdown files containing high-level global guidance.

## V1 contract

Files are UTF-8 Markdown. Optional integer `priority` front matter controls
ascending order; equal priority is ordered by normalized relative path.
