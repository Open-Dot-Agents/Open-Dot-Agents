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

## Reference adapter projection

- OpenAI Codex: `mapped` through `.agents/memories/codex.toml`.
- GitHub Copilot CLI and Claude Code: `unsupported` by the current adapters.

## V1 contract

Memories are runtime-created JSON records, never hand-authored policy. Records
must include provenance, creation time, and sensitivity and must not contain
secrets. See `../../spec/v1/schema/memory.schema.json`.
