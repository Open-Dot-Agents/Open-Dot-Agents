# Guardrails

Guardrails define hard runtime boundaries: what the agent environment can or
cannot do, regardless of its task intent.

## Purpose

- Filesystem scope (`read-only` vs. `read-write`)
- Network access policy
- Process execution limits
- Sandboxing mode and similar environmental restrictions

## Scope note

Guardrails are different from `../permissions/`:

- `guardrails`: what the environment allows at the platform level
- `permissions`: which tools/actions are permitted by policy and approval flow

If a harness exposes only a single combined control (for example, a broad
`sandbox_mode` setting), model it as `guardrails` and document the approval policy
separately in `../permissions/README.md`.

## Vendor mappings

- OpenAI Codex CLI sandbox modes
- Generic agent runtime sandbox models used across vendors

## Current oda projection

No v0.0.1 adapter emits this category; all three targets are `unsupported`.

## Expected content

This category typically maps to harness configuration blocks that define:

- read/write filesystem policy
- network restrictions
- command/session execution limits
- other non-negotiable execution boundaries

## Contract

Machine-readable fields are defined in `../schema/v0.0.1/agents.schema.json`.

See `../mappings.yaml` for canonical links.
