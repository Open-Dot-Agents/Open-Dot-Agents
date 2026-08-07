# Skills

Portable, on-demand task playbooks — self-contained folders of instructions
(and optionally scripts/resources) for a specific repeatable task, loaded
just-in-time when relevant rather than kept in context at all times.

Maps to: GitHub Copilot CLI agent skills (`.github/skills/<skill>/SKILL.md`),
Claude Code skills (`.claude/skills/`), OpenAI Codex CLI skills, the
vendor-neutral `agentskills.io` format.

Expected content: one subfolder per skill, containing at minimum a
`SKILL.md` with name/description/trigger metadata plus instructions, and
optionally supporting scripts.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`, including
the `skills/*/SKILL.md` file convention and required front matter.
