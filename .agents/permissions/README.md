# Permissions

Rules governing *which specific tools/actions* an agent may invoke, and
whether each requires human approval before running — tool allow/deny/ask
lists, approval policies (e.g. "never", "on-request", "untrusted"), and
per-action gating (shell, network, filesystem writes, MCP calls).

**Scope note vs. `../guardrails/`:** `guardrails` covers the sandbox/
environment an agent runs inside (what it's *physically capable of*, e.g.
network access at the OS level); `permissions` covers the *policy layer* on
top (what it's *allowed to attempt* and whether approval is required first).
Some harnesses expose both as a single setting (e.g. Codex's
`sandbox_mode` + `approval_policy` pair) — in that case document the sandbox
half under `guardrails` and the approval/allow-list half here, cross-linking
both READMEs.

Maps to: OpenAI Codex CLI `approval_policy` (config.toml), Claude Code
`permissions` (allow/deny/ask) in `.claude/settings.json`.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`.
