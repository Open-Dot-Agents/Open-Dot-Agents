# Changelog

All notable program-level changes are documented here. Component-specific
release histories live in their respective repositories.

## Unreleased

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
