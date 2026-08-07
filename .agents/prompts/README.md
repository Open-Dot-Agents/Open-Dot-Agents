# Prompts

Reusable, parameterized prompt templates invoked on demand (not automatically
loaded like `../instructions/`) — e.g. "review this PR", "write tests for
this file". Often exposed to users as slash commands.

Maps to: GitHub Copilot CLI prompt files (`.github/prompts/*.prompt.md`),
Claude Code custom slash commands, OpenAI Codex CLI custom slash
commands/workflows.

Expected content: one file per prompt, typically Markdown with placeholders
for variables/arguments.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`.
