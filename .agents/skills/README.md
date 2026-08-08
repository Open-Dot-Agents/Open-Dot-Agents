# Skills

Skills are on-demand task playbooks that load only when needed.

## Purpose

- package repeatable problem-solving procedures
- minimize prompt/context overhead
- keep specialist workflows discoverable and reusable

## Vendor mappings

- GitHub Copilot CLI: `.github/skills/<skill>/SKILL.md`
- Claude Code: skill-style reusable workflow folders
- OpenAI Codex CLI: codex-specific skill/task folders
- `agentskills.io` compatible skill formats

## Expected content

Each skill uses a dedicated subfolder (for example `.github/skills/<name>/`) and
at minimum includes a `SKILL.md` with:

- trigger/usage description
- required context
- execution steps or heuristics

Optional scripts or resources can be added as needed.

## Contract

Machine metadata is in `../schema/v0.0.1/agents.schema.json`, including required file
layout and front matter.

See `../mappings.yaml` for vendor documentation.
