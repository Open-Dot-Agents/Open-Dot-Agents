# Agents

Definitions for custom sub-agents (a.k.a. "custom agents" or "subagents"):
named, specialized personas with their own instructions, allowed tools, and
optionally their own model — invoked to isolate a task in a separate context
window (e.g. a "security reviewer" or "test writer" agent).

Maps to: GitHub Copilot CLI custom agents (`.github/agents/*.md`), Claude Code
subagents (`.claude/agents/*.md`), OpenAI Codex CLI `AGENTS.md` agent roles.

Expected content: one file per agent, typically Markdown with front-matter
for name/description/allowed-tools, followed by the agent's system prompt.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`, including
the `agents/*.md` filename convention and required `description` front matter.
