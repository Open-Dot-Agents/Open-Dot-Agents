# Open-Dot-Agents

Open-Dot-Agents defines a practical, vendor-neutral standard for configuring AI
coding agents. Instead of maintaining separate formats per tool, teams author one
canonical `.agents/` tree and generate compatible target outputs.

## Why this project exists

Most harnesses (GitHub Copilot CLI, OpenAI Codex, Anthropic Claude Code, and
others) represent the same ideas with
different schemas. Open-Dot-Agents normalizes those ideas into one contract:

- always-on instructions and scoped rules
- agent and skill definitions
- hooks and lifecycle automation
- safety controls and permissions
- tooling integrations (MCP and beyond)

## Canonical `.agents/` layout

```text
.agents/
├── agents/        Persona and sub-agent definitions
├── guardrails/    Sandbox and hard runtime boundaries
├── hooks/         Lifecycle hook definitions
├── instructions/  Unconditional behavioral guidance
├── memories/      Persistent memory structures
├── permissions/   Allow/deny/ask action policies
├── plugins/       Installable configuration bundles
├── profiles/      Named runtime presets
├── prompts/       Reusable prompt templates
├── rules/         Scoped and conditional behavior
├── settings/      Base runtime defaults
├── skills/        On-demand task playbooks
├── tools/         External integrations
├── mappings.yaml  Target mapping and status metadata
└── schema/        Versioned machine-readable contracts
```

Category documentation:

- [`agents`](./.agents/agents/README.md)
- [`guardrails`](./.agents/guardrails/README.md)
- [`hooks`](./.agents/hooks/README.md)
- [`instructions`](./.agents/instructions/README.md)
- [`memories`](./.agents/memories/README.md)
- [`permissions`](./.agents/permissions/README.md)
- [`plugins`](./.agents/plugins/README.md)
- [`profiles`](./.agents/profiles/README.md)
- [`prompts`](./.agents/prompts/README.md)
- [`rules`](./.agents/rules/README.md)
- [`settings`](./.agents/settings/README.md)
- [`skills`](./.agents/skills/README.md)
- [`tools`](./.agents/tools/README.md)

Machine-readable contracts:

- `.agents/schema/v0.0.1/agents.schema.json`
- `.agents/schema/v0.0.1/mappings.schema.json`
- `.agents/manifest.json`

## Getting started in 10 minutes

Use this section as the canonical onboarding reference.

### 1) Install and verify the CLI

```bash
cd /path/to/Open-Dot-Agents
cd cli
go install ./cmd/oda
cd ..
task cli:build
export PATH="/path/to/Open-Dot-Agents/cli/bin:$PATH"
task cli:verify
```

### 2) Validate a target repository with `.agents`

```bash
oda validate --target all
oda generate --target all
oda check --target all
```

### 3) First-run mismatch flow

```bash
oda validate --target all --allow-unsupported
oda generate --target all --force
oda check --target all
```

### 4) Verify adapter behavior by target

- `task cli:test:copilot` — Copilot projection validation
- `task cli:test:codex` — Codex projection and MCP transport validation
- `task cli:test:claude` — Claude projection validation
- `task cli:test:surface` — CLI command smoke checks

### 5) Common failure patterns

- `schema not available`: ensure schema files exist under `.agents/schema/v0.0.1/`
- `unsupported populated categories`: either move/remove files from unsupported directories or use `--allow-unsupported`
- `generated output is stale` or `no generated compatibility manifest found`: run `oda generate` again

### 6) Contributor checks before PR

- Read `AGENTS.md`, root `README.md`, and relevant `.agents/*/README.md`
- Ensure `.agents/mappings.yaml` and behavior docs are aligned for any status change
- Keep category names exact (`instructions`, `rules`, `tools`, ...)
- Run `task cli:verify` before opening the PR

The CLI uses `v0.0.1` schema artifacts in `.agents/schema/v0.0.1/*` and mapping metadata in `.agents/mappings.yaml` to enforce:

- `supported` output emitted directly from `.agents`
- `mapped` output emitted through a target transform
- `partial` partial output projection for a category
- `unsupported` no output emitted by current adapters

These values describe adapter projection only, not full native capability.

## Supported harness tools

- [<img src="https://api.iconify.design/simple-icons/github.svg?color=%23181717" width="14" height="14" style="vertical-align:middle;" alt="GitHub Copilot CLI" /> GitHub Copilot CLI](https://docs.github.com/en/copilot)
- [<img src="https://api.iconify.design/simple-icons/openai.svg?color=%23412991" width="14" height="14" style="vertical-align:middle;" alt="OpenAI Codex" /> OpenAI Codex](https://learn.chatgpt.com/docs/)
- [<img src="https://api.iconify.design/simple-icons/anthropic.svg?color=%23121212" width="14" height="14" style="vertical-align:middle;" alt="Anthropic Claude Code" /> Anthropic Claude Code](https://code.claude.com/docs/)

Helpful flags and one-liners:

- `--target all` for registry-wide runs
- `--dry-run` to preview writes
- `--diff` for file-level change summaries
- `--format=json` for machine-readable output
- `--ci` for strict CI behavior

```bash
oda --root . --help
oda --root . validate --target all --format=json
oda --root . check --target all --ci
```

## Project status

The v0.0.1 canonical contract and adapters for GitHub Copilot CLI, OpenAI Codex, and
Anthropic Claude Code are implemented. Current work focuses on compatibility,
release stability, and adoption of the `.agents` mapping model. The initial
project release is planned as `v0.0.1`.

Mapping statuses describe what the current `oda` adapters emit, not every native
feature available in each harness. See [the v0.0.1 schema notes](.agents/schema/v0.0.1/README.md)
for the status definitions.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
