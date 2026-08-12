# Migrating to Open-Dot-Agents

The portable `.agents/` tree is the source of truth. Native harness files are
generated projections and should not be edited as an alternative canonical
configuration.

## Start a new repository

Build or install the reference CLI using the [installation guide](INSTALL.md),
then initialize and validate a starter tree:

```sh
agents init
agents validate
```

Edit `.agents/manifest.json` to select the required profiles, place shared
instructions in `.agents/AGENTS.md`, define MCP servers in
`.agents/tools/mcp.json`, and add skills below `.agents/skills/`.

## Import an existing configuration

The current reference CLI imports the supported native MCP, skill, and shared
`AGENTS.md` configuration for Copilot or Codex:

```sh
agents import --vendor copilot --source . --target .agents
agents validate --source .agents
```

Import refuses to overwrite an existing canonical target unless `--force` is
provided. Use `--force --backup --diff` when intentionally replacing a
configuration; inspect the backup and diff before committing the result.

## Export a projection

Check the [compatibility matrix](COMPATIBILITY.md) and adapter capabilities
first:

```sh
agents capabilities --vendor copilot
agents export --vendor copilot --source .agents --target . --diff
```

If native configuration already exists, require both an explicit overwrite and
a backup:

```sh
agents export --vendor copilot --force --backup --diff
```

Do not commit credentials, token values, or generated local overrides.
Consult [vendor mapping evidence](VENDOR_EVIDENCE.md) for documented native
differences, particularly Claude Code's `CLAUDE.md` instruction bridge and
the distinct MCP target files. An adapter should be used only when its exact
harness version is marked supported in the matrix.
