# Skills

Skills are on-demand task playbooks that load only when needed.

## Purpose

- package repeatable problem-solving procedures
- minimize prompt/context overhead
- keep specialist workflows discoverable and reusable

## Vendor mappings

- GitHub Copilot CLI: `.github/skills/<skill>/SKILL.md`
- Claude Code: skill-style reusable workflow folders
- OpenAI Codex CLI: reads `.agents/skills/<skill>/SKILL.md` directly
- `agentskills.io` compatible skill formats

## Expected content

Each skill uses a dedicated subfolder (for example `.github/skills/<name>/`) and
at minimum includes a `SKILL.md` with:

- trigger/usage description
- required context
- execution steps or heuristics

Optional scripts or resources can be added as needed.

All nested skill assets are preserved by conforming adapters.

## V1 contract

Skills conform directly to the Agent Skills specification, including directory
name, front-matter constraints, and progressive-disclosure layout. Dota does
not define a competing skill dialect.
