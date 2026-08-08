# Open Dot Agents CLI (`oda`)

`oda` is the command-line adapter for Open-Dot-Agents.
It validates canonical `.agents/` trees, projects to vendor formats, and tracks
generated files so regeneration stays safe and repeatable.

## Supported targets

- `copilot` (default)
- `codex`
- `claude`

## Quick start

```bash
cd cli
go install ./cmd/oda

# validate the local repo
oda validate --root /path/to/repo

# generate one target
oda generate --root /path/to/repo --target codex

# verify generated files are still current
oda check --root /path/to/repo --target codex

# remove only manifest-tracked outputs
oda clean --root /path/to/repo --target codex
```

## Core commands

- `validate` — verify `.agents/` contract, schema, mappings, and target compatibility
- `generate` — write vendor-specific files
- `check` — detect drift versus generated manifest
- `clean` — remove only adapter-generated files
- `import` — convert vendor files into canonical `.agents/`
- `guide` — generate a short vendor implementation guide
- `completion` — shell completion for `bash` and `zsh`

## Common options

- `--target <name|all>` — selected adapter target
- `--root <path>` — source repository path (default: current directory)
- `--format=json` — machine-readable status output
- `--dry-run` — plan-only execution
- `--diff` — show planned file additions/updates/deletes
- `--force` — overwrite adapter-owned generated files
- `--backup` — archive destination before overwrite
- `--allow-unsupported` — explicitly proceed with unsupported populated categories
- `--ci` — return non-zero on drift in scripted runs

## Command examples

```bash
# run all targets
oda validate --root . --target all
oda generate --root . --target all --dry-run --diff
oda check --root . --target all --format=json --ci
```

## Verification flow

```bash
task cli:build
task cli:test
task cli:verify
```

Fixture-driven tests validate `basic`, `complex`, and additional edge cases for each target in this repository.

## Compatibility matrix (quick reference)

| Source category | Generated output |
|---|---|
| `instructions/*.md` | Copilot, Codex, Claude vendor instruction files |
| `rules/*.md` | Copilot, Claude rule files |
| `agents/*.md` | Copilot, Codex, Claude agent files |
| `hooks/*.json` | Copilot, Claude hook projections |
| `tools/mcp.json` | Copilot, Codex, Claude MCP definitions |
| `skills/<name>/...` | target-native skill directories |

## Native validation after generation

- Copilot: `copilot instructions`, `copilot skill list`, `copilot mcp list`
- Codex: `codex mcp list`, `codex agent list`
- Claude: `claude mcp list`, `claude agents list`

Native inspection is optional; fixture checks remain the deterministic base.
