# 0001: Portable Core and Native Instructions

- Status: accepted
- Decision date: 2026-08-12
- Decider: Maurizio Casciano

## Context

The release candidate duplicated instructions under `.agents/AGENTS.md` and
root `AGENTS.md`. This created drift and destructive projection risk while the
target harnesses already converged on root or scoped instruction files.

## Options

- Keep the duplicate canonical copy.
- Accept both locations with precedence rules.
- Standardize root and nested `AGENTS.md` and keep `.agents` for structured
  portable configuration.

## Decision

Open-Dot-Agents 1.0 uses root and nested `AGENTS.md` directly. The portable
structured core remains manifest, MCP, and skills. Copilot CLI, Codex, and
Claude Code are the first adapter targets; OpenCode remains experimental.

## Effects

Repositories remove `.agents/AGENTS.md`. Claude adapters may create an owned
import bridge. Adapters report native discovery that cannot be suppressed as a
limitation. No profiles for hooks, permissions, models, or subagents enter 1.0
without a separate proposal and evidence.
