# Changelog

All notable program-level changes are documented here. Component-specific
release histories live in their respective repositories.

## Unreleased

- Document and add regression coverage for portable model selection: the
  existing draft.2 native profile already projects a portable `model` field
  into Codex's `config.toml` and Copilot's `settings.json` at project or
  user scope; no new schema was needed. Adds
  `SPEC/examples/native-model-selection` (conformance-checked),
  `CLI/internal/config/native_model_selection_test.go`, and a new "Model
  selection" section in `docs/NATIVE_CONFIGURATION.md`. Decision 0002
  (proposed abstract `models` manifest field) is superseded by decision 0003.

- Fix CI regressions surfaced by PR #16: the "Workbench deterministic
  tests" job installed only `SPEC/conformance/requirements.txt`, leaving
  `pexpect` (required by several Workbench conformance tests) uninstalled;
  it now installs `WORKBENCH/conformance/requirements.txt`, which already
  pulls in the SPEC requirements transitively. Also update four leftover
  `1.0.83` literals in `WORKBENCH/conformance/synthetic_verifier_fixtures.py`
  to `1.0.84-9` (missed in the earlier Copilot pin bump), which were
  tripping `verify_copilot_skill_metadata`/`verify_copilot_parent_skills`/
  `verify_public_github_mcp` assertions. Also switch `security.yml`'s
  standalone `govulncheck` job to `go-version-file: CLI/go.mod` instead of
  a second hardcoded Go version, so `CLI/go.mod` is the single source of
  truth. Verified locally: all 160 Workbench conformance tests and 104
  task/test tests pass against an isolated venv that mirrors the CI
  install step exactly.

- Bump build and CI toolchain pins to the current latest stable releases:
  the reference CLI's `go.mod` directive to Go 1.27.1 (CI's `setup-go`
  steps already track `CLI/go.mod`), the standalone `govulncheck` job's
  Go pin to 1.27.1, the verification workflow's Python pin from 3.12 to
  3.14, the MCP server and adapter-conformance workflows' Node.js pin
  from 24 to 26, and the spec conformance suite's `jsonschema` pin from
  4.25.1 to 4.26.0. All Go, spec conformance, and Workbench test suites
  were re-run locally against the updated toolchains and pass unchanged.

- Update the pinned Copilot native harness from 1.0.83 to 1.0.84-9 in the
  reference CLI's version gate, compatibility summary, and Workbench test
  fixtures. Codex remains pinned at 0.154.0. The frozen 1,679-row feature
  inventory and its historical per-feature evidence claims are unchanged;
  a new native evidence campaign against 1.0.84-9 has not yet been run.

- Add transactional initial instruction links to stable file projections,
  with shared-link sync and protection for existing root instruction files.

- Add the missing `SPEC/.agents/plugins/.gitkeep` placeholder. Correct CLI
  validation of nested canonical instruction links and add the SPEC starter
  to the repository verification command matrix.

- Correct Codex TLS reference relocation. Record CA-only TLS delivery and
  separate HTTP identity failures for logs, traces, and metrics. Keep the
  complete failing exporter inactive to preserve authentication.

- Correct draft.2 Codex telemetry scope and exclude credentials in collector
  URLs and headers. Record local HTTP JSON log and trace evidence. Resolve
  exporter aliases in coverage without adding semantic features.

- Add draft.2 native configuration, scoped ownership, and a semantic coverage
  map. Correct numeric preservation, ownership and import defects, and coverage
  classification. Record the pinned Codex alias probe and debug review. The full
  configuration milestone and native adapter promotion remain incomplete.

- Record native isolation and model-tool approval tests. Preserve the Copilot
  local-network and Codex mandatory-ask refusals; do not extend coverage.

- Add the sourced Codex/Copilot feature inventory and experimental security
  draft. Validate and normalize permissions and sandbox requirements; refuse
  unsupported activation. Add a pinned Codex Linux direct-shell mapping with
  authority checks, owned security segments, and native enforcement evidence.
  Keep Copilot projection refused after a local-network mismatch.

- Fix repository validation without a manifest and projection of explicitly
  required unsupported capabilities. These checks implement existing rules.

- Add refusal checks for unselected canonical skills and Codex stdio
  environment references. Preserve canonical files and existing ownership.
- Require Workbench checks and all three native support rows before release.
- Add baseline/extended evidence agreement checks and versioned evidence
  bundles to the native and release workflows.
- Add eleven Claude extended cases; native verification needs credentials.

- Add extended native feature tests for latest Codex and Copilot, with HTTPS
  fixtures, terminal compaction, event filters, failure behavior, profile
  selection, and resumed-session checks. Preserve failures and source snapshots.
- Record conformance failures in tools-profile removal, unselected skill
  exposure, and Codex missing stdio environment references. Keep support status
  conservative.

- Refuse disabled hook projections that cannot preserve catalogue scope.
- Preserve unrelated Claude settings when hooks change or are removed; upgrade
  existing ownership records only when their recorded file is unchanged.
- Import hook-only repositories and preserve the native disabled state.
- Reject ignored matchers and align hook schema and CLI validation.
- Run all deterministic Workbench tests in CI, including hook lifecycle and
  evidence checks. Add session-hook and disabled-hook native evidence checks.

## [1.0.0] - 2026-08-12

- Established public governance, contribution, security, versioning, release,
  and proposal processes.
- Added the compatibility matrix and version-pinned vendor evidence record.
- Added CI gates for the specification conformance baseline, reference CLI,
  and experimental workbench projections.
- Added installable, platform-specific reference CLI release-candidate
  archives with SHA-256 checksums.
- Added compatibility drift checks that keep `CLI/compatibility.json`,
  `COMPATIBILITY.md`, and Reference CLI capability summaries aligned.
- Added official `llms.txt` source catalog and `$agent-docs` skill for vendor
  documentation research.
- Added a portable command-hook profile with schema, fixtures, reference CLI
  projections for Copilot CLI, Codex, and Claude Code, and conservative
  projection-only compatibility claims.
