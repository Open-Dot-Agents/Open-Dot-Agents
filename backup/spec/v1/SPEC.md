# Open-Dot-Agents Specification 1.0.0-rc.1

This document is normative. The key words **MUST**, **MUST NOT**, **REQUIRED**,
**SHOULD**, **SHOULD NOT**, and **MAY** are to be interpreted as described by
RFC 2119 and RFC 8174.

## 1. Repository root

A conforming tree is rooted at `.agents/` and MUST contain `manifest.json`.
Empty category directories MAY be omitted. Paths in manifests, diagnostics,
and adapter messages MUST be UTF-8, slash-separated, relative to the repository
root, and MUST NOT contain `.` or `..` segments.

The thirteen categories and their conformance tiers are:

| Tier | Categories |
| --- | --- |
| Core | `instructions`, `rules`, `agents`, `skills`, `tools` |
| Automation and packaging | `hooks`, `prompts`, `plugins` |
| Policy and configuration | `permissions`, `guardrails`, `profiles`, `settings` |
| Runtime state | `memories` |

A processor claiming Core conformance MUST implement every Core category. A
processor claiming Full conformance MUST implement all tiers. Category support
does not require a project to populate the corresponding directory.

## 2. Authoring formats

Narrative assets use UTF-8 Markdown with YAML 1.2 front matter. Structured
assets use strict RFC 8259 JSON; duplicate object keys and non-finite numbers
MUST be rejected. Parsers MUST reject symlinked category files.

- `instructions/**/*.md` contains unconditional guidance. Processors order
  files by ascending `priority` front matter (default `0`), then normalized path.
- `rules/**/*.md` requires `id` and an `applyTo` array of repository-relative
  glob patterns. A rule activates when at least one include matches and no
  `exclude` pattern matches.
- `agents/*.md` requires `name` and `description`; optional portable fields are
  `tools`, `skills`, `permissionProfile`, `execution`, and `maxTurns`.
- `prompts/*.md` requires `name` and `description`; optional `inputs` defines
  named string parameters. Prompt placeholders use `{{inputs.<name>}}`.
- `skills/*/SKILL.md` MUST conform to the Agent Skills specification. Skills MAY
  contain arbitrary supporting files; symlinks are forbidden.
- `tools/mcp.json` uses an `mcpServers` map and supports `stdio` and
  `streamable-http` transports. Secrets MUST be environment references, never
  literal credentials.
- Structured categories use the schemas in `schema/`. Profiles apply RFC 7396
  JSON Merge Patch and MUST reject inheritance cycles.

## 3. Policy semantics

Permission decisions are `allow`, `ask`, or `deny`. When multiple matching
rules apply, `deny` wins over `ask`, which wins over `allow`; equal decisions
use the most specific matcher and then source order. Guardrails describe hard
runtime limits and MUST NOT be weakened by profiles or permissions.

Hooks are deterministic lifecycle automation. A hook declares an event,
optional matcher, action, timeout, and failure behavior. Processors MUST execute
matching hooks in ascending priority and declaration order.

Memories are runtime-created JSON records, not authored policy. Every record
MUST include provenance and creation time and MAY include expiry, scope, and
sensitivity. Secret material MUST NOT be persisted as memory.

## 4. Extensions and loss

Non-portable data belongs under `.agents/extensions/<reverse-dns-id>/`. The ID
MUST contain at least one dot and use lowercase alphanumerics, dots, and hyphens.
Extensions do not form a fourteenth category. Processors MUST preserve unknown
extension files byte-for-byte.

An adapter MUST return a loss report for each source field or artifact that it
cannot represent. Silent loss is non-conforming. A loss may be resolved by
preserving the source under the adapter's extension namespace.

## 5. Versioning and compatibility

`specVersion` follows Semantic Versioning. Processors MUST reject an unsupported
major version. Pre-release versions require an exact match. Version 1 has no
compatibility or migration requirement for v0.0.1 trees.

## 6. Conformance

Conformance is established by the published fixtures and `dota conformance`.
Implementations MUST produce stable diagnostic codes, deterministic plans, and
identical output for identical input. The conformance suite is normative when
prose and a fixture disagree until the discrepancy is resolved through the RFC
process. Host diagnostic identifiers and exit-status families are registered in
[`DIAGNOSTICS.md`](DIAGNOSTICS.md).
