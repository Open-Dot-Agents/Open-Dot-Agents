# Memories

Memories preserve knowledge that should persist across sessions and reduce repeated
onboarding overhead.

## Purpose

- capture recurring user preferences
- preserve conventions discovered over time
- retain continuity notes for future sessions

## Scope note

`memories/` is intentionally different from `../rules/` and `../instructions/`:

- `rules` and `instructions` are authored, static configuration
- `memories` represent accumulated context and recall state

Not all tools expose first-class memory files yet. This category provides a vendor
neutral home where memory-capable harnesses can consume it consistently.

## Vendor mappings

- GitHub Copilot CLI: experimental memory-oriented behavior
- Claude Code / OpenAI Codex CLI: fold long-horizon context into
  instruction-like config when no native memory file exists

## Current oda projection

No v0.0.1 adapter emits this category; all three targets are `unsupported`.

## Expected content

Structured or semi-structured notes that represent persisted context,
including:

- task preferences
- workspace conventions
- persistent user instructions

## Contract

Machine metadata is in `../schema/v0.0.1/agents.schema.json`.

See `../mappings.yaml` for vendor references.
