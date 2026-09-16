# 0002: Portable Models and Agents (proposal)

- Status: superseded by [0003](0003-model-selection-native-profile.md)
- Public comment opened: 2026-09-16
- Superseded: 2026-09-16 (comment period cut short; the model-selection
  subset of this proposal turned out to already be solved by the existing
  draft.2 native profile, see 0003)
- Decider: Maurizio Casciano

> **Superseded.** The model-selection design below was withdrawn before the
> comment period closed: `CLI/internal/config/native_registry.go` already
> recognizes `model` for both Codex and Copilot, and the existing draft.2
> native profile already projects it into each harness's real config file
> today. See [decision 0003](0003-model-selection-native-profile.md). The
> remaining 130 `models-agents` settings this record did not name explicitly
> (provider wiring, context window, auto-compaction, review model) are
> unaffected and remain open for a future proposal.

## Context

The [feature inventory](../FEATURE_INVENTORY.md) classifies 131 documented
Codex and Copilot settings under the `models-agents` area as
`vendor-extension`: real native config-file fields with no `.agents` portable
counterpart today. A user can only set them by editing `~/.codex/config.toml`
or Copilot's native config directly. Representative fields:

- Codex `config.toml`: `model`, `model_provider`, `model_context_window`,
  `model_reasoning_effort` (already partly covered by the accepted Copilot
  `reasoningEffort` audit, but with no Codex-side portable mapping),
  `model_auto_compact_token_limit`, `model_auto_compact_token_limit_scope`,
  `review_model`.
- Copilot: model selection and per-agent model/provider assignment in agent
  frontmatter and settings, which are native today with only the narrow
  `model` and `reasoningEffort` fields already audited as `validator-declared`.

`docs/ROADMAP.md` names this the next "Security and feature standardization"
phase: "Propose portable agent roles and model intent with named provider
extensions." This record opens that proposal for the primary model-selection
subset. It intentionally excludes reasoning-effort (already handled) and
provider credential fields (`requires_openai_auth` and similar stay native
per the existing no-credential-values rule).

## Options

- **Do nothing.** Leave all 131 fields native-only. Keeps the spec small but
  leaves model/provider selection, a highly used setting, entirely outside
  `.agents`, forcing per-harness duplication for any repository that pins a
  model.
- **Copy every vendor field verbatim into the schema.** Maximizes coverage but
  imports vendor-specific shapes (Codex's `model_provider` id lookup against
  `model_providers`, Copilot's own model catalog) as normative portable
  fields, coupling the spec to each vendor's current catalog and provider
  list. Rejected: this is the kind of exact-syntax coupling
  `FEATURE_INVENTORY.md` explicitly warns against ("a native syntax variant
  is not automatically equivalent to a portable field").
- **Add a narrow, named `models` extension with vendor-neutral fields plus
  explicit per-harness overrides for anything unresolvable portably.**
  Define a small portable surface — requested model identifier, optional
  per-harness model override, and a context-window hint — and leave provider
  wiring, auto-compaction thresholds, and review-model overrides as named
  extensions or native-only until a second implementation proves the mapping.

## Proposed decision (subject to the comment period)

Adopt the third option. Add an optional `models` block to the manifest
profile set (parallel to today's `tools`, `hooks`, and `skills` profiles):

- `models.default`: a vendor-neutral requested model identifier (free-form
  string; the adapter refuses activation if the target harness cannot
  resolve it rather than silently substituting a different model).
- `models.overrides.<harness>`: optional per-harness model identifier,
  applied only when `default` cannot be resolved identically across
  harnesses (for example Copilot and Codex naming the same model
  differently).
- Everything else observed in the 131-field inventory
  (`model_provider`, `model_context_window`,
  `model_auto_compact_token_limit`, `model_auto_compact_token_limit_scope`,
  `review_model`, and Copilot's model-catalog/provider settings) remains a
  named extension or native-only pending a second adapter's evidence, per
  the existing `vendor-extension` disposition. No credential or provider
  API-key values are ever stored portably.

## Compatibility and security effects

- Additive: existing manifests without a `models` block are unaffected;
  `SPEC/conformance/run.py` gains new valid/invalid fixtures for the new
  block but no existing fixture changes semantics.
- An adapter that cannot resolve `models.default` against its own model
  catalog must refuse activation and report the unresolved model rather than
  silently ignoring the field or substituting a default, consistent with the
  project's lossy-projection refusal rule.
- No secret or credential material is introduced; provider auth fields
  explicitly stay out of scope.

## Migration

None required for existing repositories. Adopting `models.default` is
opt-in. A follow-up CLI change would add Codex `model` /
`model_provider`-narrow projection and a corresponding Copilot projection,
each gated on its own native-evidence fixture before being marked
`validator-declared`, following the same audit process used for
`reasoningEffort` and MCP `timeout`.

## Dissent

None recorded yet; this record is open for comment through 2026-09-30.
