# Compatibility Matrix

This matrix is the authoritative public record of Open-Dot-Agents compatibility
claims. A blank, planned, or experimental entry is not a support guarantee.
Each supported entry must link to the exact conformance evidence, adapter
version, harness version, profiles, and known limitations.
Verified upstream mapping research is recorded separately in
[vendor mapping evidence](VENDOR_EVIDENCE.md).
The same current claims are also available as machine-readable data in
[`compatibility.json`](../compatibility.json).

## Current state

<!-- compatibility-table:start -->
| Adapter | Harness version | Instructions | MCP | Skills | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| Reference CLI: Copilot | Not pinned | CLI Projection Only | CLI Projection Only | CLI Projection Only | Not conformance supported | CLI unit tests only; no version-pinned native-harness black-box run |
| Reference CLI: Codex | Not pinned | CLI Projection Only | CLI Projection Only | CLI Projection Only | Not conformance supported | CLI unit tests only; no version-pinned native-harness black-box run |
| Reference CLI: Claude Code | Not pinned | Planned | Planned | Planned | Not conformance supported | No native adapter evidence or black-box run |
| Reference CLI: OpenCode | Not pinned | Planned | Workbench Projection Only | Planned | Not conformance supported | Workbench projection tests only; no version-pinned native-harness black-box run |
<!-- compatibility-table:end -->

## Ratification review — 2026-08-12

The specification baseline (14/14), reference CLI unit suite, and Workbench
MCP projection suite passed in this review. Root workflow YAML also parsed and
`git diff --check` passed. These are not native-harness evidence: the
Workbench suite reads checked-in JSON/TOML projections and the CLI suite uses
temporary filesystem fixtures.

No adapter has a completed evidence record naming an immutable native harness
version, platform, test date, and repeatable black-box results for instruction
discovery, MCP discovery and server startup, or skill discovery. Claude Code
has no native adapter evidence at all. Accordingly, no adapter can be
ratified as conformance-supported and the project remains a release candidate.
The Reference CLI's `capabilities` command reports the same conservative
compatibility status and evidence recorded in `compatibility.json`, alongside
the managed projection paths.

## Native probe — 2026-08-12

A read-only probe on Linux x86_64 observed GitHub Copilot CLI 1.0.79 and Codex
CLI 0.147.0 listing project-visible MCP configuration. Those observations do
not exercise instruction discovery, skill discovery, or MCP startup, so they
are insufficient for a support claim. OpenCode 1.18.16 was available, but its
MCP status command was intentionally not run because it could start enabled
stdio servers. Claude Code was not installed. A repeatable, isolated native
harness suite remains required before any row can move beyond release-candidate
status.

## Publication rules

An entry can be marked **Supported** only when it:

1. names the Open-Dot-Agents and native harness versions;
2. declares support for each profile as lossless, transformed, unsupported, or
   vendor extension;
3. links to a passing conformance run using the published fixtures;
4. lists all known limitations and required user actions; and
5. is regenerated or revalidated when either the adapter or harness changes.

The current-state table is generated from `compatibility.json` with
`python3 scripts/check_compatibility.py --write`. Release checks must run
`python3 scripts/check_compatibility.py` so the Markdown matrix, CLI capability
summaries, and support-evidence rules cannot drift.
