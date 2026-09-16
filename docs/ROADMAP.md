# Roadmap

## Development preset (active)

Deliver one editable development policy: automatic project edits, tests,
builds, formatting, and local commits; approval for push, publish, deployment,
and other external changes. See [setup and current status](DEVELOPMENT_PRESET.md).

1. Practical mode separates agent guidance from native settings. Starter,
   migration, validation, projection, update, and removal are implemented.
2. Codex uses a workspace profile and user-reviewed on-request approval.
   Copilot receives guidance while its native permissions remain unchanged.
3. The 2026-09-16 workflow baseline passed 28 required native cases: 20 Codex
   cases and eight Copilot cases. Five Codex observations remain guidance-only.
   This includes setup, adoption, updates, conflicts, removal, rollback,
   project/user scope, relocation, and Codex local operations and boundaries.
   See the [recorded versions and hashes](../WORKBENCH/evidence/DEVELOPMENT_WORKFLOWS.md).
4. Add read-only `agents doctor --experimental --vendor codex|copilot` to
   explain requested policy, disk configuration, and unknown session authority.
   The command, lifecycle checks, clean-source gate, and refreshed
   installed-binary campaign passed. See the recorded doctor acceptance.
   Full adapter support remains false.
5. Keep strict mode refused where native enforcement is not equivalent.
   Keep automatic repairs, native smoke tests in doctor, user-home migration,
   new profiles, Claude verification, and the broad settings/security audit
   deferred. Publish component commits before root references only after
   separate approval.

Native execution evidence remains separate from deterministic lifecycle tests.
The prior passing receipt remains historical evidence after source changes;
it cannot qualify the new doctor implementation. Full adapter support is not
claimed.

## Current draft.2 cycle

Complete Codex/Copilot global and scoped configuration evidence before adding
more native feature families. Keep Claude native tests deferred and the
three-adapter release gate unchanged.

1. Verify remote component commits and a fresh recursive checkout.
2. Keep deterministic verifier tests independent of ignored receipts. All
   verifier test families now use synthetic inputs. The clean-source CI check
   copies source files without ignored artifacts and runs the deterministic
   suites. Real native receipts remain in the separate evidence verification
   path.
3. Run global defaults, project overrides, fallback, inherited skills,
   conflicts, relocation, removal, and rollback against the pinned binaries.
4. Review the 231 validator declarations by activation and scope. Review the
   81 security requirements separately. Retain unsafe-mapping refusals.
5. Publish tested component commits before the root integration reference.

See the [milestone audit](NATIVE_MILESTONE_AUDIT.md) for current reproduction
failures and the difference between evidence links and behavior verification.

## Security and feature standardization

The [inventory](FEATURE_INVENTORY.md) and [security draft](SECURITY_PROFILES.md)
are implemented for review. The first Codex Linux direct-shell mapping now
has native enforcement, refusal, and lifecycle evidence. Copilot prerequisites
were prepared locally, but host Unix-socket access does not preserve local
network denial. See the [scenario record](../WORKBENCH/evidence/SECURITY_SCENARIOS.json).
The next phases are:

1. Extend Codex authority and scope checks only with native evidence. Native
   model-tool allow/deny tests now exist, but `on-request` is not portable
   mandatory `ask`. Implement a lossless mapping before extending coverage.
   Credential isolation and other platforms remain unverified.
2. Resolve the Copilot loopback mismatch before implementing that mapping.
   Test portable policies after apply; native settings alone do not establish
   portable enforcement. Keep pending scenarios explicit.
3. Propose portable agent roles and model intent with named provider extensions.
   [Decision 0002](decisions/0002-portable-models-and-agents.md) opens the
   public comment period (through 2026-09-30) for the model-selection subset
   of the 131 `models-agents` vendor-extension settings.
4. Propose planning artifacts, context controls, and session transfer rules.
5. Assess plugin metadata, LSP scope, automation, and remote protocol extensions.
6. Submit the security proposal for public comment and record the maintainer
   decision before ratification. No public proposal has been submitted here.

## Previous development cycle

Command-hook completion is implemented locally: disabled-state refusal,
matcher checks, hook-only import, field-level Claude ownership, lifecycle
regressions, and deterministic CI coverage. Full native support remains unverified.

Tools-profile cleanup is implemented for all three CLI adapters. Native
removal and re-enabling pass on Codex 0.154.0 and Copilot 1.0.83. Claude has
deterministic regression coverage; its native test remains deferred. See the
[extended report](../WORKBENCH/evidence/EXTENDED_NATIVE_TESTS.md).

1. CLI refusal guards now cover unselected non-empty canonical skills on
   Codex and Copilot and Codex stdio environment references. Native discovery
   itself is unchanged. Refusal is not full profile support.
2. Assess the reviewed native results and their retained retries. Copilot
   recovered after network errors; earlier selected-skill failures remain
   recorded. Passing cases do not remove explicit capability refusals.
3. Claude native verification is skipped for this development cycle at the
   maintainer's request. Its baseline and eleven implemented extended cases
   remain unverified; the existing three-adapter release gate is unchanged.
4. Publish reviewed, versioned native evidence only after successful gates.
   The workflow now packages baseline results, extended results, source
   snapshots, checksums, and reproduction inputs. Local results remain local.
5. Keep release blocked until all three registry rows meet the existing support
   policy. Unsupported capabilities and expected refusals do not meet that bar.

The core audit also fixed public CLI validation without a manifest and
explicit unsupported capability requirements. The portable specification and
schemas are unchanged. See the [completion review](../WORKBENCH/evidence/CORE_COMPLETION.md)
for validation results and release blockers.

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
