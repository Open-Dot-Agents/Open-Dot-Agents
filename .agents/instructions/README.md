# Instructions

Always-on, repo-wide (or user-wide) behavioral instructions that are loaded
into every agent session automatically — coding conventions, build/test
commands, architectural context, communication style.

**Scope note vs. `../rules/`:** across vendors this same concept goes by
different names — Claude Code's always-loaded `CLAUDE.md`, GitHub Copilot's
`copilot-instructions.md`, and generic `AGENTS.md` are all "instructions" in
this standard's sense: unconditional, always-applied context. Use `../rules/`
instead for *scoped/conditional* behavior (e.g. rules that only apply to
certain paths, languages, or triggers). If a harness doesn't distinguish the
two, treat its equivalent file as `instructions` by default.

Maps to: Claude Code `CLAUDE.md`, GitHub Copilot CLI custom instructions
(`.github/copilot-instructions.md`, `AGENTS.md`), OpenAI Codex CLI `AGENTS.md`.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`, including
the `.md` file convention and parser expectations for this category.
