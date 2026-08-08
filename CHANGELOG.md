# Changelog

All notable changes to Open-Dot-Agents are documented in this file.

## [v0.0.1] - 2026-08-09

### Added

- The versioned, machine-readable `.agents` contract for thirteen agent configuration categories.
- `oda validate`, `import`, `export`, `check`, `clean`, `guide`, and shell completion workflows.
- Repository-scoped adapters for GitHub Copilot CLI, OpenAI Codex, and Anthropic Claude Code.
- Safe import and export previews, ownership manifests, drift detection, backups, JSON output, and CI mode.
- Deterministic fixtures and native compatibility checks for the supported harness surfaces.
- Cross-platform `oda` release archives and SHA-256 checksums.

### Known limitations

- Codex registers and round-trips generated custom agents, but the authenticated live acceptance probe does not yet observe a spawned child thread. The strict diagnostic remains enabled and the exact boundary is documented in `COMPATIBILITY.md`.
- Claude acceptance is fixture-based with optional native MCP discovery; it does not yet have the authenticated end-to-end matrix used for Codex and Copilot.

[v0.0.1]: https://github.com/Open-Dot-Agents/Open-Dot-Agents/releases/tag/v0.0.1
