# Open-Dot-Agents Milestone Plan

This plan turns the current implementation into an explicit open standard with a stable
adapter framework and stronger interoperability surface.

## Track 1 — Formalize the standard contract (highest priority)

### Milestone 1.1: Machine-readable schema and manifest

- [x] Add `.agents/schema/v1/agents.schema.json`
  - Canonical keys and required shape for `.agents` categories:
    - `instructions`, `rules`, `agents`, `hooks`, `permissions`, `guardrails`,
      `tools`, `skills`, `profiles`, `plugins`, `memories`, `prompts`, `settings`.
  - Validate:
    - file names and extensions per category
    - category behavior (`required`, `optional`, `forbidden`)
    - front-matter requirements by category (for example `agents` needs `description`,
      `hooks` needs expected object shape, `skills/SKILL.md` required metadata).
- [x] Add `.agents/schema/v1/mappings.schema.json`
    - Define mapping metadata shape and required keys:
      - target identity (`id`, `name`, `docs`)
    - per-category category status (`supported|unsupported|partial|mapped`)
    - notes and migration constraints.
- [x] Add `.agents/manifest.json`
  - Include at least:
    - `format_version` (starts at `1`)
    - `canonical_root`: `.agents`
    - `supports`: list of enabled categories
    - `unsupported` + rationale
    - `compatibility_notes`
  - Keep parser-friendly and optional for legacy trees until phase 2.
- [x] Add docs hooks
  - Update `README.md` with explicit schema references and contract versioning.
  - Add concise schema usage notes to each affected category README.

### Milestone 1.2: Explicit unsupported/partial/mapped status per target

- [x] Extend `.agents/mappings.yaml` entries to carry status metadata for each target-category pair
  (not only links).
- [x] Make supported/unsupported semantics machine-readable for automation.

### Milestone 1.3: Validate schema in CI/test surface

 - [x] Add a validator utility in `cli/internal/schema/*` for local validation mode.
- [x] Add tests in `cli/internal/adapter/adapter_test.go` (or a new `schema_test.go`) for:
   - malformed schema violations
   - required keys
   - status resolution per target.

## Track 2 — Make `oda` a stable adapter framework

### Milestone 2.1: Target registry + capability descriptors

- [x] Add a renderer registry in `cli/internal/adapter/renderer.go`.
- [x] Add exported adapter metadata structure for capabilities and target introspection.
- [x] Add `--target all` behavior in `cli/cmd/oda/main.go`:
  - loop through all registered targets
  - emit per-target status/error summaries.

### Milestone 2.2: CLI ergonomics and output modes

- [x] Add `--format=json` to all command results (machine-readable status payloads).
- [x] Add `--dry-run` to preview write operations without mutation.
- [x] Add `--diff` to show file-level summary before generate.
- [x] Add `--ci` mode:
  - non-zero on drift for generate dry-run and check.

### Milestone 2.3: Reverse compatibility import path

- [x] Add `oda import <vendor>` command path.
- [x] Add importer interfaces and tests in CLI internals:
  - `cli/internal/adapter/import.go` (or target package)
  - round-trips from Copilot/Codex/Claude generated files into `.agents`.
- [x] Add conflict/merge diagnostics and manual override flow.

## Track 3 — Expand conversion and diagnostics

### Milestone 3.1: Additional targets

- [x] Add source mapping README updates and fixture metadata templates for each adapter.
- [x] Add one adapter at a time using registry pattern:
  - Cursor
  - Windsurf
  - Copilot Chat (if applicable)
- [x] For each new target, include mapping contract in `.agents/mappings.yaml`.
- [x] Add fixture coverage scaffolding under `cli/testdata/shared/` for cursor/windsurf/copilot-chat.

### Milestone 3.2: MCP transport hardening

- [x] Codex:
  - explicitly enumerate supported transports (`stdio`, `local`, `http`, `streamable-http`).
  - reject unsupported/malformed transports in a single validation pass.
- [x] Add equivalent transport tests in `cli/internal/adapter/adapter_test.go`.

### Milestone 3.3: Rule/hook conflict and diagnostics

- [x] Detect duplicate rule events/paths across adapters.
- [x] Add cross-format transform tests:
  - malformed hook duplicates
  - duplicate events
  - cross-target conversion edge cases.

## Track 4 — Adoption, QA, and docs

### Milestone 4.1: Fixture hardening

- [x] Extend fixture matrix:
  - malformed category files
  - conflicting filenames across targets
  - nested skill payloads and nested assets
  - empty optional categories
  - migration from legacy vendor files.
- [x] Keep fixtures in `cli/testdata/shared/{basic,complex,...}` plus expected snapshots.

### Milestone 4.2: Tooling and publishing

- [x] Add a generator for short “Vendor implementation guide” from schema + mappings
  (markdown output under `README.md` or a new `docs/vendors/<vendor>.md` as needed).
- [x] Add a command or docs path so vendors can ingest `.agents` directly.
- [x] Publish explicit deprecation/compat notes in `.agents/manifest.json`.

## Proposed implementation order

1. Track 1.1 → Track 1.3 (stabilize contract and validator surface).
2. Track 2.1 → Track 2.2 (CLI framework hardening).
3. Track 4.1 (tests/fixtures now that command behavior is stable).
4. Track 3.x (new targets and conversion breadth) in small increments.
5. Track 2.3 and Track 4.2 (reverse import and vendor guide).

## Suggested tracking format

- Use one checkbox per file-level ticket above and commit in small logical groups.
- Treat each milestone as "complete" only when:
  - unit tests pass for affected behavior
- docs and mapping files are updated consistently:
  - `README.md`
  - `.agents/mappings.yaml`
  - `cli/README.md`
  - affected `.agents/<category>/README.md`.
