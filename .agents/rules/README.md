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

- Markdown files with activation metadata (e.g. scope/tags/globs/trigger)
- optional model-invocation hinting where supported

## Contract

Machine metadata is in `../schema/v1/agents.schema.json`, including filename and
required front matter.

See `../mappings.yaml` for vendor documentation.
