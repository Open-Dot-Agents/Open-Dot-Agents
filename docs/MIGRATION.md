# Migrating to Open-Dot-Agents

The portable `.agents/` tree is the source of truth. Root `AGENTS.md`, native
MCP files, and Claude bridge/skill files are compatibility or owned
projections; nested `AGENTS.md` files remain scoped portable instructions.

## Start a new repository

Build or install the reference CLI using the [installation guide](INSTALL.md),
then initialize and validate a starter tree:

```sh
agents init --root .
agents validate --root .
```

Edit `.agents/manifest.json` to select the optional `tools` and `skills`
profiles, place shared instructions in `.agents/AGENTS.md`, define MCP servers
in `.agents/tools/mcp.json`, and add skills below `.agents/skills/`. Keep the
generated root `AGENTS.md` compatibility link and add nested `AGENTS.md` files
where narrower instructions are needed.

## Import an existing configuration

The current reference CLI imports the supported native MCP, skill, and shared
`AGENTS.md` configuration for Copilot or Codex:

```sh
agents import --vendor copilot --root .
agents validate --root .
```

Import refuses to overwrite an existing portable target unless `--force` is
provided. Use `--force --backup` only when intentionally replacing it.

## Export a projection

Check the [compatibility matrix](COMPATIBILITY.md) and adapter capabilities
first:

```sh
agents capabilities --vendor copilot
agents plan --vendor copilot --root . --format json
agents apply --vendor copilot --root .
```

Existing native configuration is merged structurally. Equivalent entries can
be adopted; conflicts require an explicit forced backup:

```sh
agents apply --vendor copilot --adopt
agents apply --vendor copilot --force --backup
```

Commit the ownership state when committing generated projections. Do not
commit credentials, token values, or generated local overrides.
Consult [vendor mapping evidence](VENDOR_EVIDENCE.md) for documented native
differences, particularly Claude Code's `CLAUDE.md` instruction bridge and
the distinct MCP target files. An adapter should be used only when its exact
harness version is marked supported in the matrix.
