# Open-Dot-Agents

Open-Dot-Agents is the vendor-neutral standard for repository-scoped AI agent
configuration. A project owns one portable `.agents/` tree; explicitly locked,
out-of-process adapters translate that tree to and from individual harnesses.

The current contract is **1.0.0-rc.1**. It is a clean break: v0 trees and the
former CLI are intentionally unsupported.

## The standard

The specification preserves thirteen categories without pretending every
harness represents each one natively:

| Conformance tier | Categories |
| --- | --- |
| Core | `instructions`, `rules`, `agents`, `skills`, `tools` |
| Automation and packaging | `hooks`, `prompts`, `plugins` |
| Policy and configuration | `permissions`, `guardrails`, `profiles`, `settings` |
| Runtime state | `memories` |

Core implementations support the first tier. Full implementations support all
thirteen categories. Vendor-only data is preserved under
`.agents/extensions/<reverse-dns-id>/`; it never becomes a hidden fourteenth
category.

```text
.agents/
├── manifest.json
├── adapters.lock.json
├── agents/        ├── hooks/         ├── memories/      ├── profiles/
├── guardrails/    ├── instructions/  ├── permissions/   ├── prompts/
├── plugins/       ├── rules/         ├── settings/      ├── skills/
├── tools/
├── extensions/<reverse-dns-id>/
└── .dota/{import,export}/
```

The normative sources are:

- [Specification 1.0.0-rc.1](spec/v1/SPEC.md)
- [Adapter Protocol 1.0](spec/v1/ADAPTER_PROTOCOL.md)
- [Stable diagnostic registry](spec/v1/DIAGNOSTICS.md)
- [JSON Schemas](spec/v1/schema/)
- [Conformance fixtures](conformance/v1/)
- category contracts in [`.agents/`](.agents/)

## Quick start

Install the RC from the nested Go module:

```bash
go install github.com/Open-Dot-Agents/Open-Dot-Agents/cli/cmd/dota@v1.0.0-rc.1
dota --version
```

Or build the CLI and all three reference adapters from this checkout:

```bash
task cli:build
task cli:verify
```

Initialize and validate a tree:

```bash
dota init --root /path/to/project
dota validate --root /path/to/project
dota inspect --root /path/to/project
```

Adapters are never discovered from `PATH`. Add an executable explicitly; this
records its identity and SHA-256 checksum in `.agents/adapters.lock.json`:

```bash
dota adapter add \
  --root /path/to/project \
  --id org.open-dot-agents.codex \
  --version 1.0.0-rc.1 \
  --path /path/to/dota-adapter-codex

dota adapter doctor --root /path/to/project --adapter org.open-dot-agents.codex
dota export --root /path/to/project --adapter org.open-dot-agents.codex --dry-run --json
dota export --root /path/to/project --adapter org.open-dot-agents.codex
dota check --root /path/to/project --adapter org.open-dot-agents.codex
```

Local-path locks are for development and are rejected by `--ci`. Released
adapters use publisher manifests with per-platform artifact URLs and checksums:

```bash
dota adapter add --root /path/to/project --manifest https://example.org/adapter.json
dota adapter install --root /path/to/project
```

Reference adapter publisher manifests are attached to each v1 GitHub release as
`org.open-dot-agents.<target>.json`. They resolve directly installable binaries
for Linux, macOS, and Windows; release bundles remain available for manual
installation of `dota` and all three adapters together.

## Trust model

Adapters are JSON-RPC 2.0 executables using LSP `Content-Length` framing. They
receive a filtered snapshot and return diagnostics, explicit loss reports, and
a deterministic file plan. They do not write the workspace. The trusted `dota`
host validates paths, rejects symlinks and collisions, applies files atomically,
tracks ownership under `.agents/.dota/`, and protects modified output.

Checksums prove artifact integrity, not publisher trust. Review an adapter and
its publisher before locking it. See [SECURITY.md](SECURITY.md) for the threat
model and reporting process.

## Reference adapters

This repository ships external reference adapters for OpenAI Codex, GitHub
Copilot CLI, and Anthropic Claude Code. Their compatibility status describes
projection behavior, not features built into `dota`; see
[COMPATIBILITY.md](COMPATIBILITY.md) for the evidence boundary.

## Contributing

Specification changes use the RFC process in [CONTRIBUTING.md](CONTRIBUTING.md)
and [GOVERNANCE.md](GOVERNANCE.md). Before submitting a change, run:

```bash
task cli:verify
task cli:release:check  # when GoReleaser is installed
```

## License

Apache License 2.0 — see [LICENSE](../LICENSE).
