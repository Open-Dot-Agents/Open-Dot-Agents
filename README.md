# Open-Dot-Agents

Open-Dot-Agents defines a practical, vendor-neutral standard for configuring AI
coding agents. Instead of maintaining separate formats per tool, teams author one
canonical `.agents/` tree and generate compatible target outputs.

## Why this project exists

Most harnesses (Copilot, Codex, Claude, and others) represent the same ideas with
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

- `.agents/schema/v1/agents.schema.json`
- `.agents/schema/v1/mappings.schema.json`
- `.agents/manifest.json`

## Quick start

```bash
cd cli
go install ./cmd/oda
oda validate --root /path/to/repo
oda generate --root /path/to/repo --target all
oda check --root /path/to/repo --target all
```

Helpful flags:

- `--target all` for registry-wide runs
- `--dry-run` to preview writes
- `--diff` for file-level change summaries
- `--format=json` for machine-readable output
- `--ci` for strict CI behavior

## Roadmap and status

See [ROADMAP.md](./ROADMAP.md) for the milestone plan and current status.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
