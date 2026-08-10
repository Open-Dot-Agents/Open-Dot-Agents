# Open-Dot-Agents v1 and dota RC.1

## Summary

Replace v0.0.1 with a conformance-first v1 standard and rename the CLI from oda to dota (“Dot Agents”). Preserve the thirteen-category taxonomy through conformance tiers, but redesign the schemas and runtime architecture
without backward compatibility.

Vendor adapters become separately distributed JSON-RPC executables. dota remains the trusted host responsible for validation, filesystem safety, ownership, atomic writes, and drift detection.

## V1 Standard

- Publish normative RFC 2119 specifications and JSON Schemas under spec/v1/, with canonical GitHub Pages schema URLs. Embed identical schemas in dota for deterministic offline validation; consumer repositories no longer copy
  schemas or .agents/mappings.yaml.

- Define these conformance tiers:
    - Core: instructions, rules, agents, skills, tools
    - Automation and packaging: hooks, prompts, plugins
    - Policy and configuration: permissions, guardrails, profiles, settings
    - Runtime state: memories

- Require processors to support Core; projects may omit empty categories. “Full” conformance requires all thirteen. Do not add a plans category.
- Standardize Markdown with YAML front matter for narrative assets and strict JSON for structured configuration. Adopt the Agent Skills specification (https://agentskills.io/specification) and MCP wire concepts directly
  instead of creating incompatible variants.

- Specify deterministic ordering, repository-relative POSIX paths, rule glob semantics, profile inheritance through JSON Merge Patch, conflict diagnostics, secret references, scope/precedence, and deny-over-ask-over-allow
  policy resolution.

- Treat memories as runtime-created state with provenance, timestamps, scope, expiry, and sensitivity metadata—not as hand-authored policy.
- Add .agents/extensions/<reverse-dns-id>/ for lossless vendor or implementation data. Extensions are not a fourteenth category; unknown extensions must be preserved, and adapters must report every lossy conversion.
- Replace the current meta-schema with schemas for actual repository artifacts, including:
    - .agents/manifest.json
    - every structured category document
    - Markdown front-matter contracts
    - .agents/adapters.lock.json
    - adapter publisher manifests, diagnostics, plans, and JSON output

- Use specVersion: "1.0.0-rc.1" in the manifest. Detect v0.0.1 trees and fail with a stable replacement/reinitialization diagnostic; provide no compatibility or migration implementation.

## dota and Adapter Architecture

- Hard-rename oda to dota: command path, help, docs, Task targets, release archives, environment variables, generated metadata, and examples. Do not retain an oda alias.
- Make dota vendor-neutral. Core commands become:
    - dota init
    - dota validate
    - dota inspect
    - dota adapter add|install|list|doctor
    - dota import|export|check|clean --adapter <id>
    - dota conformance tree|adapter

- Version machine output with outputVersion: 1 and stable diagnostics. Reserve exit codes: 0 success, 1 validation/runtime failure, 2 usage error, 3 drift/conformance mismatch, and 4 adapter trust or integrity failure.
- Define a JSON-RPC 2.0 adapter protocol over stdio using Content-Length framing. Required methods:
    - initialize for protocol negotiation
    - describe for ID, target, category support, input globs, capabilities, and limits
    - validate
    - exportPlan
    - importPlan
    - shutdown

- Normalize RPC paths as repository-relative POSIX paths. Represent non-text data as base64. Standardize diagnostics, source ranges, create/update/delete plans, provenance, and loss reports.
- Adapters never mutate the workspace. dota supplies approved input snapshots and applies returned plans after enforcing traversal, symlink, collision, ownership, digest, size, timeout, and atomic-write protections.
- Remove duplicated capability matrices from Go and per-project YAML. An adapter’s describe response is authoritative; the conformance suite verifies that its behavior matches those claims.
- Introduce an adapter publisher manifest with reverse-DNS ID, semantic version, protocol range, platform artifacts, SHA-256 digests, and optional Sigstore bundle.
- Commit .agents/adapters.lock.json with exact adapter versions, protocol versions, sources, platform digests, and granted capabilities. Nothing is discovered implicitly from PATH.
- dota adapter install resolves locked artifacts into a user cache and verifies integrity. Local-path adapters are allowed for development but rejected by --ci.
- Rename generated ownership state to .dota metadata and keep it adapter-scoped. Preserve existing protections for modified generated files, unowned collisions, backups, dry runs, and atomic cleanup.
- Externalize the existing Codex, Copilot, and Claude implementations as separately built dota-adapter-* reference executables without expanding their feature sets. The dota binary must not import their transformation
  packages; an architecture test enforces that boundary.

- Provide a Go adapter SDK plus a small non-Go conformance adapter to prove the protocol is genuinely language-neutral.
- Keep “configuration plugins” inside the canonical plugins category distinct from executable “adapter plugins.”

## Conformance, Documentation, and Governance

- Publish normative specification, category contracts, adapter protocol, security model, versioning policy, extension policy, and glossary. Generate the README compatibility matrix from adapter metadata.
- Add governance, contribution, security-reporting, RFC, and compatibility-change processes. Any normative change requires schemas, conformance vectors, changelog entry, and an accepted RFC.
- Create minimal, full, invalid, security, extension-preservation, and adapter-round-trip fixtures that can be consumed independently of the Go implementation.
- Base the portable capability model on the shared primitives documented by official OpenAI documentation (https://learn.chatgpt.com/docs/customization/overview), Claude Code, Agent Skills, and MCP, while keeping vendor
  behavior in adapters or namespaced extensions.

- Release dota, the three reference adapter executables, schemas, conformance fixtures, checksums, SBOMs, and provenance as v1.0.0-rc.1. Keep the published v0.0.1 history intact but label it superseded.
- Require feedback from at least one independently implemented adapter before promoting the contract to v1.0.0.

## Test and Acceptance Plan

- Validate every schema against JSON Schema 2020-12 and test valid/invalid examples for all thirteen categories.
- Test instruction ordering, rule activation, profile cycles and merges, permission precedence, hook failures, MCP transport validation, extension preservation, and memory provenance/sensitivity rules.
- Exercise protocol negotiation, malformed frames, crashes, timeouts, oversized payloads, duplicate IDs, unsupported methods, checksum failures, untrusted local adapters, and conflicting capability claims.
- Prove adapters cannot request or return escaping paths, symlinks, secret material, uncontrolled deletes, or direct filesystem writes through the protocol.
- Require deterministic repeated plans and byte-stable round trips, with explicit diagnostics for every loss.
- Run the current Codex, Copilot, and Claude golden fixtures through the external process boundary and retain their existing native acceptance gates without adding new harness functionality.
- Test Linux, macOS, and Windows adapter installation and framing; validate offline schema operation and committed-lock reproducibility.
- Make task cli:verify gate Go tests/vet, schemas, conformance fixtures, architecture boundaries, documentation links, release metadata, and git diff --check.

## Assumptions

- The redesign intentionally breaks v0.0.1 and provides no oda compatibility alias or migration command.
- The project remains named Open-Dot-Agents; only the CLI and its implementation metadata become dota.
- The thirteen categories remain, but conformance tiers prevent immature or runtime-specific categories from weakening the portable Core.
- No new harness target or current-adapter feature expansion belongs in this iteration.