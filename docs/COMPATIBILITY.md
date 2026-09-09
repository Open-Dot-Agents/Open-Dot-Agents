# Compatibility Matrix

This matrix is the authoritative public record of Open-Dot-Agents compatibility
claims. A blank, planned, or experimental entry is not a support guarantee.
Each supported entry must link to the exact conformance evidence, adapter
version, harness version, profiles, and known limitations.
Verified upstream mapping research is recorded separately in
[vendor mapping evidence](VENDOR_EVIDENCE.md).
The same current claims are also available as machine-readable data in
[`CLI/compatibility.json`](../CLI/compatibility.json).
That registry declares every 1.0 capability separately. A profile summary does
not override an `unsupported` or unverified capability.

## Current state

<!-- compatibility-table:start -->
| Adapter | Harness version | Instructions | Tools | Hooks | Skills | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Reference CLI: Copilot | 1.0.83 | Projection Only | CLI Projection Only | CLI Projection Only | CLI Projection Only | Not conformance supported | Local native feature tests recorded with latest pinned copilot; tools-profile removal and unselected skill exposure fail; environment-reference mappings remain refused |
| Reference CLI: Codex | 0.153.4 | Projection Only | CLI Projection Only | CLI Projection Only | CLI Projection Only | Not conformance supported | Local native feature tests recorded with latest pinned codex; tools-profile removal and unselected skill exposure fail; missing stdio environment references also fail |
| Reference CLI: Claude Code | 2.1.229 | Transformed | CLI Projection Only | CLI Projection Only | CLI Projection Only | Not conformance supported | CLI unit tests and pinned preflight only; no passing native-harness black-box run |
<!-- compatibility-table:end -->

## Ratification review — 2026-08-12

The specification baseline (17/17), reference CLI unit suite, and Workbench
MCP projection suite passed in this review. Root workflow YAML also parsed and
`git diff --check` passed. These are not native-harness evidence: the
Workbench suite reads checked-in JSON/TOML projections and the CLI suite uses
temporary filesystem fixtures.

Workbench now pins the native harness versions used for preflight and future
evidence runs. No adapter has a completed passing evidence record with
platform, test date, and repeatable black-box results for instruction discovery,
MCP discovery and server startup, hook execution, or skill discovery. Claude
Code is implemented at the projection layer but has no native evidence.
Accordingly, no adapter can be ratified as
conformance-supported and the project remains a release candidate.
The Reference CLI's `capabilities` command reports the same conservative
compatibility status and evidence recorded in `CLI/compatibility.json`, alongside
the managed projection paths.

## Historical native preflight — 2026-08-13

That preflight used GitHub Copilot CLI 1.0.79, Codex CLI 0.147.0, and
Claude Code 2.1.229. The local preflight found a reference `agents` binary and
observed Copilot, Codex, and Claude Code installed at the pinned versions. The
accepted non-interactive credential variables were absent for all three
harnesses. These preflight records do not exercise instruction discovery, hook
execution, skill discovery, or MCP startup, so they are insufficient for a
support claim. The credentialed, isolated native harness suite must pass before
any stable row can move beyond release-candidate status.

## Publication rules

An entry can be marked **Supported** only when it:

1. names the Open-Dot-Agents and native harness versions;
2. declares support for each profile as lossless, transformed, unsupported, or
   vendor extension;
3. links to a passing conformance run using the published fixtures;
4. lists all known limitations and required user actions; and
5. is regenerated or revalidated when either the adapter or harness changes.

The current-state table is generated from `CLI/compatibility.json` with
`python3 CLI/scripts/check_compatibility.py --write`. Release checks must run
`python3 CLI/scripts/check_compatibility.py` so the Markdown matrix, CLI capability
summaries, and support-evidence rules cannot drift.

## Hook completion review — 2026-09-09

Disabled catalogues are projected only for Copilot. Codex does not document
`disableAllHooks`; Claude's switch affects settings beyond the owned catalogue.
The CLI refuses these two projections with `ODA-HOOK-0001`. Refusal leaves the
previous native configuration unchanged. Remove the `hooks` profile and apply
to remove owned hooks. Claude removal preserves unrelated settings.

Copilot matchers are limited to `PreToolUse`, `PostToolUse`, `PermissionRequest`,
`PreCompact`, and `SubagentStart`. Codex and Claude matchers on `UserPromptSubmit`
and `Stop` are refused. Native timeout defaults and matcher syntax are not
normalized. These limits appear in the CLI capability output and registry.

The latest stable npm releases checked on 2026-09-09 are Copilot 1.0.83 and
Codex 0.153.4. The test pins and package integrity values now match them.
The baseline suite passes for both targets with existing CLI logins. Extended
native testing now covers remote HTTPS, environment references, all canonical
hook events, filtering, timeouts, instruction precedence, skill resources,
profile selection, trust, permission decisions, and restart/resume refresh.
See the [extended Workbench report](../WORKBENCH/evidence/EXTENDED_NATIVE_TESTS.md).

The extended tests found conformance failures: removing the tools profile
leaves its MCP server active; both harnesses discover canonical skills even
when the skills profile is unselected; Codex also activates a stdio server
when a referenced environment variable is missing. Support remains
`not-conformance-supported`. Claude is outside this extended test cycle.

Use only the latest stable Codex and Copilot releases for new test cycles.
Resolve `latest` before the cycle, then pin the exact versions and integrity
values for reproducible execution and evidence.
