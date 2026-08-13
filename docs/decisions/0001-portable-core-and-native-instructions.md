# 0001: Portable Core and Native Instructions

- Status: accepted
- Decision date: 2026-08-12
- Decider: Maurizio Casciano

## Context

The portable tree needs one canonical location for instructions while target
harnesses commonly discover root or scoped instruction files. Independent
regular files at `.agents/AGENTS.md` and root `AGENTS.md` would create drift.

## Options

- Keep independent copies at both locations.
- Keep root instructions outside the otherwise canonical `.agents` tree.
- Canonicalize `.agents/AGENTS.md` and use a root compatibility link or native
  projection.

## Decision

Open-Dot-Agents 1.0 uses `.agents/AGENTS.md` as mandatory repository-wide
instructions. A root `AGENTS.md` should link to the canonical file for native
discovery, while nested `AGENTS.md` files retain scoped precedence. Optional
manifest profiles match the `.agents` directories: `tools` and `skills`.
Copilot CLI, Codex, and Claude Code are the first adapter targets; OpenCode
remains experimental.

## Effects

Repositories keep one canonical instruction file below `.agents` and avoid
copy drift through compatibility links or owned projections. Claude adapters
may create an owned import bridge. No profiles for hooks, permissions, models,
or subagents enter 1.0 without a separate proposal and evidence.
