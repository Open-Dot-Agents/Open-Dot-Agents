# Changelog

All notable changes to Open-Dot-Agents are documented in this file.

## [v1.0.0-rc.1] - 2026-08-09

### Added

- Normative v1 specification, JSON Schemas, governance, security policy, and
  minimal, full, and invalid conformance fixtures.
- Conformance tiers preserving all thirteen categories and reverse-DNS vendor
  extension namespaces for lossless non-portable data.
- `dota` trusted-host CLI with stable diagnostics, deterministic inspection,
  adapter lock/install/doctor workflows, and import/export/check/clean.
- Language-neutral JSON-RPC 2.0 adapter protocol using LSP framing, filtered
  snapshots, explicit loss reports, and plan-only workspace access.
- External Codex, Copilot, and Claude reference adapters plus a Python
  conformance adapter.
- Checksum-pinned adapter locks, CI rejection of local adapters, atomic writes,
  ownership manifests, collision and traversal defenses, modified-file
  protection, backups, and deterministic cleanup.
- Cross-platform release archives containing the host and all reference
  adapters, directly installable adapter binaries, generated publisher
  manifests, SHA-256 checksums, SPDX SBOMs, provenance, and canonical schema
  publication through GitHub Pages.

### Breaking changes

- The command is `dota`; the previous command name is not retained as an alias.
- v1 rejects v0.0.1 trees. There is no compatibility or migration layer.
- Adapters must be explicitly locked executables; in-process adapters, implicit
  target discovery, and per-project mapping registries are removed.

### Known limitations

- Claude acceptance is deterministic and adapter-level; it does not yet include
  the authenticated end-to-end native matrix used for Codex and Copilot.

[v1.0.0-rc.1]: https://github.com/Open-Dot-Agents/Open-Dot-Agents/releases/tag/v1.0.0-rc.1
