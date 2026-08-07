# Schema contracts

This folder hosts the v1 machine-readable contracts for Open-Dot-Agents.

- `agents.schema.json` — canonical `.agents/` category contract
  (filenames, front-matter requirements, behavior flags, and category shapes).
- `mappings.schema.json` — schema for `.agents/mappings.yaml` metadata
  (targets, per-category status values, and optional compatibility notes).

Tooling that validates `.agents` trees should use `format_version: 1` and
`canonical_root: ".agents"` as the shared baseline across implementations.
