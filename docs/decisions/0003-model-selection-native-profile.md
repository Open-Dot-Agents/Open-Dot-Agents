# 0003: Model Selection Uses the Existing Native Profile

- Status: accepted
- Decision date: 2026-09-16
- Decider: Maurizio Casciano
- Supersedes: [0002](0002-portable-models-and-agents.md) (proposed, comment
  period waived by explicit maintainer instruction to act on the best
  available solution immediately)

## Context

[Decision 0002](0002-portable-models-and-agents.md) proposed a new top-level
`models.default` / `models.overrides.<harness>` manifest block to close the
131-setting `models-agents` `vendor-extension` gap in the
[feature inventory](../FEATURE_INVENTORY.md), reasoning from the documented
setting list alone.

Direct inspection of `CLI/internal/config/native_registry.go` and the
existing `1.1.0-draft.2` native-profile implementation shows this reasoning
was incomplete: **the CLI already ships a working, tested mechanism for this
exact gap.** The draft.2 native profile (`--experimental`, manifest
`"version": "1.1.0-draft.2"`, `"profiles": ["native"]`) lets a repository
author a `config` artifact under `.agents/native/<namespace>/` containing any
field the target harness's native-registry allow-list recognizes. `model` is
already in that allow-list for both harnesses, at both scopes:

- Codex (`com.openai.codex`): `model` (project and user scope),
  `model_provider` (user scope only — Codex 0.154.0 ignores a project-scope
  override).
- Copilot (`com.github.copilot`): `model` (project and user scope).

This was verified directly in this session:

1. `TestNativeCopilotJSONCRoundTrip` (existing, already-passing test in
   `CLI/internal/config/native_test.go`) already round-trips Copilot's
   `model` field through `settings.json`.
2. A live smoke test with the built CLI binary confirmed `agents apply
   --experimental` writes a requested `model` into both `.codex/config.toml`
   (`model = "gpt-5.5"`) and `.github/copilot/settings.json`
   (`{"model": "gpt-5.5"}`) from a portable `.agents/native/*/config.*`
   source, at project scope, for both harnesses.
3. A live smoke test confirmed the project-scope `model_provider` refusal is
   explicit and diagnosed (not a silent drop): `agents: required native
   field /model_provider cannot activate: Codex 0.154.0 ignores project
   model_provider; use native user scope where the setting has a supported
   mapping`.

## Decision

Withdraw the abstract `models` manifest field proposed in decision 0002.
Model selection for Codex and Copilot is **already portable today** through
the existing draft.2 native profile; no new schema, manifest field, or CLI
projection code is needed. This is a stronger, evidence-based answer than a
new abstract field: it is already implemented, already tested, and already
covers more of the `models-agents` inventory (any native-registry-recognized
scalar, not just `model`) than the proposed block would have.

Concretely:

1. Added `SPEC/examples/native-model-selection/{codex,copilot}` as runnable,
   conformance-checked examples (`SPEC/conformance/native_draft.py`) showing
   the exact `.agents/native/<namespace>/` fixture for each harness.
2. Added `CLI/internal/config/native_model_selection_test.go` as a permanent
   regression test asserting the projected `model` value lands in each
   harness's real native configuration file.
3. Documented the mechanism explicitly in `docs/NATIVE_CONFIGURATION.md`
   under "Model selection" so it is discoverable without reverse-engineering
   the native registry.
4. Left the `models-agents` `vendor-extension` classification in
   `.agents/features/codex-copilot.json` unchanged for now: draft.2 is an
   experimental, `--experimental`-gated escape hatch, not the stable 1.0
   portable contract, and the write has not yet been proven to change
   runtime harness behavior (no native-evidence audit yet, per the standing
   `validator-declared` → native-evidence promotion process). A future
   native-evidence campaign (same process as the `reasoningEffort` and MCP
   `timeout` audits) can promote `model` once a fixture proves the harness
   actually honors the requested value at runtime, and can separately assess
   whether `model` (and only `model`, not the wider `models-agents` set)
   deserves promotion into the stable 1.0 contract given how narrow and
   already-solved this specific gap turned out to be.

## Compatibility and security effects

No schema, manifest, or CLI behavior changed for the stable 1.0 or draft.1
contract. The two new example directories and one new Go test are additive.
Draft.2's existing credential-exclusion and native-registry validation rules
already govern this mechanism unchanged; the diagnostics-only project-scope
`model_provider` refusal is existing behavior, not new in this record.

## Migration

None. Repositories that already use draft.2 can add `model` to an existing
`config` artifact today; repositories not yet using draft.2 can copy
`SPEC/examples/native-model-selection`.

## Dissent

None recorded.
