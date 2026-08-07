# AGENTS.md

## What this repository is
- This repo is a **specification project**, not an executable app.
- The core artifact is the vendor-neutral `.agents/` standard documented in `README.md` and `.agents/**/README.md`.
- Most changes are documentation and schema/mapping updates; there is no discovered build/test pipeline.

## Big-picture architecture
- `README.md` defines the 13-category model and the intent behind each category.
- `.agents/mappings.yaml` is the cross-vendor translation layer (Copilot, Claude, Codex, Cursor, Windsurf, Devin).
- Each category directory under `.agents/` contains its local contract in `README.md` (scope, examples, file-shape expectations).
- Example integration config lives in `.agents/tools/mcp.example.json` (`mcpServers` map).

## Category boundaries you must preserve
- `instructions` = always-on, unconditional agent behavior context.
- `rules` = conditional/scoped behavior (triggered by path, task, or context).
- `permissions` = tool/action policy (allow/deny/ask).
- `guardrails` = runtime/sandbox boundaries (environment enforcement).
- `memories` = session-captured knowledge; not hand-authored policy text.
- `profiles` = named presets layered on settings at invocation time.

## Repo-specific authoring patterns
- Keep terminology aligned across `README.md`, `.agents/*/README.md`, and `.agents/mappings.yaml`.
- Treat `.agents/mappings.yaml` as the canonical place to add or update vendor links.
- Follow file-shape conventions documented per category:
  - skills: one subfolder per skill containing `SKILL.md` (+ optional scripts)
  - agents: one Markdown file per agent with front-matter + system prompt
  - prompts: one Markdown file per reusable prompt template
  - hooks: JSON/YAML definitions
- Keep examples vendor-neutral unless explicitly documenting a vendor mapping.

## Critical workflows (what to do when editing)
- Start by reading `README.md` and the target category `README.md` before editing.
- If changing a category definition, update all affected references in:
  - `README.md`
  - `.agents/mappings.yaml`
  - `.agents/<category>/README.md`
- If adding integrations, mirror existing style from `.agents/tools/mcp.example.json` and update mapping links.

## Known constraints and assumptions
- `.gitignore` currently tracks `.agents/` and ignores other dotfiles; do not assume hidden-tool configs are versioned unless intentionally added.
- No discoverable CI/test commands were found in tracked docs; verify expectations with maintainers before introducing automation.
- This repository appears to be an evolving proposal; favor minimal, precise diffs and preserve existing category semantics.

