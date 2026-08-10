# AGENTS.md

## What this repository is

- This repo defines the vendor-neutral `.agents/` v1 standard and the Go-based
  `dota` trusted-host CLI.
- The normative contract is in `spec/v1/`; category READMEs explain authoring
  conventions and conformance fixtures make behavior executable.
- CLI build and verification tasks are `task cli:build`, `task cli:test`, and
  `task cli:verify`.

## Architecture

- `spec/v1/SPEC.md` defines the thirteen categories and conformance tiers.
- `spec/v1/ADAPTER_PROTOCOL.md` defines the language-neutral JSON-RPC adapter
  boundary.
- `cli/internal/dotacore` is the trusted host: validation, locks, snapshots,
  path policy, ownership, and atomic application.
- `cli/pkg/adapterprotocol` contains the public framing/client/server protocol.
- `cli/internal/adapterserver` externalizes the reference Codex, Copilot, and
  Claude transforms; the `dota` command must never import vendor adapters.
- `.agents/extensions/<reverse-dns-id>/` preserves non-portable source data.

## Category boundaries

- `instructions` = always-on, unconditional agent behavior context.
- `rules` = conditional/scoped behavior triggered by paths or context.
- `permissions` = tool/action policy with deny-over-ask-over-allow precedence.
- `guardrails` = hard runtime and sandbox boundaries that profiles cannot weaken.
- `memories` = runtime-captured knowledge, never hand-authored policy.
- `profiles` = named RFC 7396 settings overlays; inheritance cycles are invalid.

## Authoring patterns

- Keep terminology aligned across `README.md`, `spec/v1/`, category READMEs,
  schemas, fixtures, and `COMPATIBILITY.md`.
- Skills use one subfolder per skill containing `SKILL.md` plus optional assets.
- Agents and prompts use one Markdown file per definition with YAML front matter.
- Hooks and policy/configuration categories use the normative JSON Schemas.
- Keep examples portable; vendor-only data belongs in a reverse-DNS extension.
- Adapters must return deterministic plans and explicit loss reports.

## Critical workflows

- Read `README.md`, `spec/v1/SPEC.md`, and the target category README before
  changing a category definition.
- A contract change requires corresponding schema, fixture, documentation, and
  changelog updates and follows the RFC process in `CONTRIBUTING.md`.
- Adapter integrations must remain out of process, explicitly locked, checksum
  pinned, and unable to write the workspace directly.
- Use `task cli:verify` as the production gate for CLI, schema, conformance, or
  adapter changes. Use `task cli:release:check` for packaging changes.

## Compatibility semantics

Projection status describes adapter behavior, not native harness capability:

- `supported`: emitted directly from portable `.agents` data.
- `mapped`: emitted through a target transform.
- `partial`: only part is projected and losses are reported.
- `unsupported`: no output is emitted.

Version 1 is a clean break. Do not add v0 compatibility, legacy command aliases,
implicit target discovery, in-process adapter imports, or per-project mappings.
