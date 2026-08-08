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

## Operator workflow

Canonical onboarding for command operators:

```bash
cd /path/to/Open-Dot-Agents
task cli:build
task cli:verify

cd /path/to/target-repo-with-agents
oda validate --root /path/to/repo-with-agents
oda generate --root /path/to/repo-with-agents --target all
oda check --root /path/to/repo-with-agents --target all
```

For first-run unsupported-category mismatches:

```bash
oda validate --root /path/to/repo-with-agents --target codex --allow-unsupported
oda generate --root /path/to/repo-with-agents --target codex --force
oda check --root /path/to/repo-with-agents --target codex
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

`import` requires one explicit source target. `--target all` is intentionally
rejected because multiple harnesses can define conflicting canonical files. With
`--force --backup`, import archives the existing `.agents/` tree before writing.

## Command examples

```bash
# run all targets
oda validate --root . --target all
oda generate --root . --target all --dry-run --diff
oda check --root . --target all --format=json --ci
```

Target-specific fixture checks:

- `task cli:test:copilot` validates fixture generation for Copilot-compatible projections and runs optional native discovery (`copilot` command present).
- `task cli:test:copilot:real` runs an isolated, authenticated Copilot CLI acceptance test against `examples/copilot-project`; it exercises native discovery and live instructions, rules, hooks, skills, MCP, and custom agents, and consumes model requests.
- `task cli:test:codex` validates Codex fixture behavior and MCP generation rules (including transport validation).
- `task cli:test:codex:real` runs an isolated, authenticated Codex acceptance test against `examples/codex-project`; it requires live services and consumes model tokens.
- `task cli:test:claude` validates Claude fixture behavior and MCP discovery (`claude` command present).
- `task cli:test:surface` runs command smoke checks with CLI help/completions.
- `task cli:verify` runs `test`, `test:copilot`, `test:codex`, and `test:claude` (plus `go vet`).

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
| `skills/<name>/...` | Copilot/Claude target directories; Codex reads `.agents/skills` directly |

Per-adapter status values map to: supported, mapped, partial, unsupported.

Troubleshooting:

- `schema not available`: ensure `.agents/schema/v0.0.1/agents.schema.json` and `.agents/schema/v0.0.1/mappings.schema.json` exist.
- `unsupported populated categories`: either remove non-README files from unsupported category directories or run with `--allow-unsupported`.
- `no generated compatibility manifest found` or `generated output is stale`: rerun `oda generate` (or `oda generate --force` if needed).
- `invalid adapter manifest` during `check`/`clean`: run generation from a clean working tree for the same target and review manual edits.
- `output ".codex/config.toml" exists but is not adapter-owned`: inspect `generate --dry-run --diff`; `--force` replaces the complete Codex config in v0.0.1.

## Native validation after generation

- Copilot: `copilot skill list --json`, `copilot mcp list --json`, and `task cli:test:copilot:real` (Copilot CLI has no standalone config schema or doctor command)
- Codex: `codex --strict-config doctor --json`, `codex mcp list --json`, `codex debug prompt-input`, and `task cli:test:codex:real`
- Claude: `claude mcp list`, `claude agents list`

Native inspection is optional; fixture checks remain the deterministic base.
Generated Codex TOML includes OpenAI's official schema directive. The real
acceptance task downloads that schema and validates the generated document
before asking the installed Codex CLI to load it strictly. The standalone
schema check used by that task is:

```bash
cd cli
go run ./cmd/codex-config-validator --config /path/to/.codex/config.toml
```
