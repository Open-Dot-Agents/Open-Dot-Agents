# Guardrails

Safety limits on *what an agent is allowed to attempt at all* — sandboxing
mode (e.g. filesystem/network isolation), read-only vs. write-enabled
sessions, and other hard boundaries baked into an agent's operating
environment.

**Scope note:** this category is about the *environment/sandbox* an agent
runs in. It is distinct from `../permissions/`, which covers *which specific
tools/actions* an agent may invoke and whether each needs approval — the two
are related but not the same concept. If a harness only exposes a single
combined "sandbox + approval policy" setting (e.g. Codex `sandbox_mode` +
`approval_policy` in one config block), document it under `guardrails` here
and cross-link from `../permissions/README.md`.

Maps to: OpenAI Codex CLI sandbox modes, general agent sandboxing concepts.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`.
