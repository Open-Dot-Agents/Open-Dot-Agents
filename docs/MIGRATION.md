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

Edit `.agents/manifest.json` to select the optional `tools`, `hooks`, and
`skills` profiles, place shared instructions in `.agents/AGENTS.md`, define
MCP servers in `.agents/tools/mcp.json`, define command hooks in
`.agents/hooks/hooks.json`, and add skills below `.agents/skills/`. Keep the
generated root `AGENTS.md` compatibility link and add nested `AGENTS.md` files
where narrower instructions are needed.

## Import an existing configuration

The current reference CLI imports the supported native MCP, command-hook, skill,
and shared `AGENTS.md` configuration for Copilot CLI, Codex, or Claude Code:

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
differences, particularly Claude Code's `CLAUDE.md` instruction bridge and the
distinct MCP and hook target files. An adapter should be used only when its
exact harness version is marked supported in the matrix.

## Completing command hooks

Omitted or zero `timeoutSec` selects the target's native default. Use a positive
value when a specific duration is required. Omit `matcher` for events that do
not support filtering. Null fields, empty matchers, and blank commands fail
validation.

For Codex and Claude, a catalogue with `disableAllHooks: true` now fails before
writes. A failed apply does not disable an earlier projection. Remove `hooks`
from the manifest profiles and apply to remove owned hooks instead.

Claude ownership now records a hash of the `hooks` field. Existing whole-file
records are upgraded only when the file matches the recorded hash. If it has
changed, inspect the plan and use `--force --backup` only for an intended
replacement. Unrelated settings remain in place. Keep the updated CLI when
using the new ownership records; older binaries do not know the field hash.
