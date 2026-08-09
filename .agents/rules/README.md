# Rules

Rules are scoped or conditional instructions that apply only when a trigger, path,
or context is matched.

## Purpose

- encode path- or file-type specific behavior
- document conditional constraints
- apply policy only where it is relevant

## Scope note

Rules are distinct from `../instructions/`:

- `instructions`: unconditional, always active
- `rules`: conditional by activation criteria (path, language, trigger, etc.)

## Vendor mappings

- OpenAI: agent rule concept
- Any vendor with conditional instruction semantics

## Expected content

- Markdown files, optionally in nested directories, with activation metadata
  (e.g. scope/tags/globs/trigger)
- optional model-invocation hinting where supported

Copilot rule files discovered below the repository root retain their exact
repository-relative location under `rules/copilot-project/**`; portable rules
outside that reserved subtree continue to project into the root
`.github/instructions/` directory.

## V1 contract

Rule Markdown requires `id` and an `applyTo` array. Optional `exclude` patterns
take precedence over includes. Paths and globs are repository-relative.
