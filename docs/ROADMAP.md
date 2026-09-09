# Roadmap

## Current development cycle

Command-hook completion is implemented locally: disabled-state refusal,
matcher checks, hook-only import, field-level Claude ownership, lifecycle
regressions, and deterministic CI coverage. Full native support remains unverified.

1. Fix tools-profile removal so it removes the owned MCP entries from native
   configuration and a later native run cannot load them.
2. Resolve native discovery of canonical skills when the skills profile is
   unselected. Refuse the mapping if the adapter cannot preserve selection.
3. Make Codex stdio environment references fail when their runtime source is
   missing. Apply-time validation alone cannot establish that runtime guarantee.
4. Rerun the failing extended cases after each fix, then rerun the full gate.
   Keep unsupported mappings and native hook-output differences explicit.
5. Run Claude separately before any three-adapter support claim.

## 1.0 ratification

- Freeze root/scoped instructions, manifest, MCP, hooks, skills, capability
  loss, and machine-readable conformance contracts.
- Publish independently tagged Spec and CLI components with immutable schemas,
  cross-platform binaries, checksums, SPDX SBOMs, and provenance.
- Pass public credentialed black-box evidence for exact Copilot CLI, Codex, and
  Claude Code versions. OpenCode remains a Workbench experiment.
- Publish the root program release only after all three registry rows satisfy
  the support-evidence gate.

## Native integration

- Submit native-consumption proposals to the supported harness projects.
- Add a second independent conforming implementation and keep fixture/result
  contracts language-neutral.
- Revalidate supported rows on each harness update and immediately downgrade a
  row when evidence expires or regresses.

## Adoption

- Record verified production adopters without inventing or inflating entries.
- Improve package-manager distribution only after final assets and provenance
  are stable.
- Use “de-facto standard” only when the governance adoption threshold is met.

New portable profiles for permissions, models, subagents, or prompts require
evidence of shared semantics, a public proposal, security analysis, fixtures,
and migration rules. They do not block the small 1.0 core. The 1.0 hooks
profile is intentionally limited to command-hook projection; native harness
execution evidence is still required before any adapter can be marked
conformance-supported.
