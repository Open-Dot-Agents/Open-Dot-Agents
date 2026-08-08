# Permissions

Permissions define policy for what the agent is allowed to do at runtime and what
requires explicit consent.

## Scope note

Permissions sit above guardrails:

- `guardrails`: environment capability limits (what is technically possible)
- `permissions`: governance limits (what is allowed, blocked, or requires approval)

If a harness combines these concepts, document the environment controls under
`../guardrails/` and place approval policy here.

## Vendor mappings

- OpenAI Codex CLI: `approval_policy` and policy-style command controls
- Claude Code: `permissions` section in `.claude/settings.json`

## Current oda projection

No v0.0.1 adapter emits this category; all three targets are `unsupported`.

## Expected content

- allow/deny/ask action rules
- global approval policy
- per-tool gating (for example shell, network, and file write constraints)

## Contract

Machine metadata is in `../schema/v0.0.1/agents.schema.json`.

See `../mappings.yaml` for canonical vendor documentation.
