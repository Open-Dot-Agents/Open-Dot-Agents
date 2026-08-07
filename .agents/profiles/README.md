# Profiles

Named, swappable presets that layer on top of `../settings/` for context
switching — e.g. a "dev" profile with a fast model and relaxed sandboxing, a
"ci" profile with a stricter model and `approval_policy = "never"`, or a
"deep-review" profile with a high-reasoning-effort model. Selected explicitly
at invocation time rather than being the always-active configuration.

Maps to: OpenAI Codex CLI named profiles (`[profiles.<name>]` in
`config.toml`, selected via `--profile <name>`). Other harnesses do not yet
have a first-class equivalent; until they do, treat any "config scope/mode
switch" mechanism (e.g. Claude Code's managed/user/project/local settings
scopes) as adjacent but not identical — those are layering/precedence rules,
not user-selectable named presets.

Expected content: one file (or section) per named profile, containing
overrides to apply on top of the base `../settings/`.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`.
