# AGENTS.md

## What this repository is
- This repo defines the vendor-neutral `.agents/` standard and includes the Go-based `oda` compatibility CLI.
- The standard is documented in `README.md` and `.agents/**/README.md`; schemas and mappings are machine-readable.
- CLI build and verification tasks are exposed through `task cli:build`, `task cli:test`, and `task cli:verify`.

## Big-picture architecture
- `README.md` defines the 13-category model and the intent behind each category.
- `.agents/mappings.yaml` is the cross-vendor translation layer (GitHub Copilot CLI, OpenAI Codex, Anthropic Claude Code).
- Each category directory under `.agents/` contains its local contract in `README.md` (scope, examples, file-shape expectations).
- Portable MCP integration config lives in `.agents/tools/mcp.json` (`mcpServers` map).

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

## Mapping and compatibility guide

- Status values describe projection behavior, not native harness capability:
  - `supported` emitted directly from `.agents`
  - `mapped` emitted through a target transform
  - `partial` only part of the category is currently projected
  - `unsupported` not emitted by v0.0.1 adapters
- `v0.0.1` mapping matrix:

| Category | Copilot | Codex | Claude |
|---|---|---|---|
| `agents` | mapped | mapped | mapped |
| `instructions` | supported | supported | supported |
| `rules` | supported | unsupported | supported |
| `hooks` | supported | mapped | supported |
| `tools` | mapped | mapped | mapped |
| `skills` | mapped | supported | mapped |
| `guardrails` | unsupported | mapped | unsupported |
| `memories` | unsupported | mapped | unsupported |
| `permissions` | mapped | mapped | unsupported |
| `plugins` | mapped | mapped | unsupported |
| `profiles` | unsupported | mapped | unsupported |
| `prompts` | mapped | unsupported | unsupported |
| `settings` | mapped | mapped | unsupported |

## Critical workflows (what to do when editing)
- Start by reading `README.md` and the target category `README.md` before editing.
- If changing a category definition, update all affected references in:
  - `README.md`
  - `.agents/mappings.yaml`
  - `.agents/<category>/README.md`
- If adding integrations, mirror existing style from `.agents/tools/mcp.example.json` and update mapping links.

## Known constraints and assumptions
- `.gitignore` currently tracks `.agents/` and ignores other dotfiles; do not assume hidden-tool configs are versioned unless intentionally added.
- Use `task cli:verify` as the production verification gate for CLI or mapping changes.
- This repository appears to be an evolving proposal; favor minimal, precise diffs and preserve existing category semantics.
