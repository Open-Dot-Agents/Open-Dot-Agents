---
name: agents-release
description: Prepare Open-Dot-Agents specification or reference CLI releases, release candidates, changelog entries, installation guidance, compatibility updates, tag checks, artifact checksums, and release notes. Use when publishing or reviewing a release without overclaiming adapter support.
---

# Agents Release

## Workflow

Keep release claims exactly bounded by verified artifacts and compatibility
evidence.

1. Read `docs/RELEASING.md`, `docs/VERSIONING.md`, `docs/CHANGELOG.md`, component changelogs, `docs/INSTALL.md`, `docs/COMPATIBILITY.md`, and `CLI/compatibility.json` before changing release wording.
2. Classify the release as specification, reference CLI, or release candidate. Do not describe candidate artifacts as final releases.
3. Confirm all normative changes have matching schemas, fixtures, docs, and migration notes.
4. Confirm all adapter support claims have current version-pinned native evidence. Otherwise keep statuses as implementation-only, experimental, planned, or not conformance-supported.
5. Build release notes from actual committed changes and test results. Do not infer signing, SBOM, provenance, package-manager support, or native harness support unless those assets and instructions exist.
6. Preserve independent versioning between the specification and reference CLI.

## Validation

Run the full relevant gates before release wording is finalized: spec
conformance, CLI tests, Workbench projections, JSON parsing for release data,
and `git diff --check`.
