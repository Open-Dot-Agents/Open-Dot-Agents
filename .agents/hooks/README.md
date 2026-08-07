# Hooks

Automation triggered on lifecycle events of an agent session — e.g. before/
after a tool call, on session start/end, before context compaction. Used for
deterministic actions (linting, logging, notifications, policy enforcement)
that shouldn't consume the agent's own context or reasoning.

Maps to: GitHub Copilot CLI hooks (`.github/hooks/*.json`), Claude Code hooks
(registered in `.claude/settings.json`), OpenAI Codex CLI lifecycle hooks
(governed via `requirements.toml`).

Expected content: hook definitions specifying the triggering event and the
command/script to run, typically JSON or YAML.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`, including
`hooks/*.json` payload shape and Claude merge semantics.
