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

## Reference adapter projection

- OpenAI Codex: `mapped` through `.agents/permissions/codex.toml`, including
  approval policy and permission tables, plus recursive command-policy files:
  `.agents/permissions/codex-rules/**/*.rules` maps byte-for-byte to
  `.codex/rules/**/*.rules` in both directions.
- GitHub Copilot CLI: `mapped` between
  `.agents/permissions/copilot-allowed-models.txt` and the shared repository
  policy `.github/allowed_models.txt`. Personal approval caches and
  `.github/copilot/settings.local.json` are intentionally not imported.
- Claude Code: `unsupported` by the current adapter.

## Expected content

- allow/deny/ask action rules
- global approval policy
- per-tool gating (for example shell, network, and file write constraints)

Codex `.rules` files use Starlark and are loaded only when the project-local
`.codex` layer is trusted. Validate an exported file with `codex execpolicy
check --rules <file> -- <command>`.

## V1 contract

Portable policies use strict JSON. Decisions are `allow`, `ask`, or `deny`, with
deny taking precedence over ask and allow. See
`../../spec/v1/schema/policy.schema.json`.
