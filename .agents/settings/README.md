# Settings

Base runtime configuration for an agent session — model selection, reasoning
effort, environment variables, default sandbox mode, and other options that
apply regardless of which agent/skill/tool is active. This is the
vendor-neutral equivalent of a harness's top-level config file, as opposed to
the more specific `../permissions/`, `../profiles/`, and `../tools/`
categories.

Maps to: OpenAI Codex CLI `config.toml` (top-level, non-profile keys),
Claude Code `.claude/settings.json` (non-permissions, non-hooks keys),
GitHub Copilot CLI configuration.

**Scope note:** keep tool wiring in `../tools/`, tool allow/deny/approval
rules in `../permissions/`, and named presets in `../profiles/` — `settings`
is for the single "current" base configuration.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`.
