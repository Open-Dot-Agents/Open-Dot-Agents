# Dot Agents CLI (`dota`)

`dota` is the trusted host for the Open-Dot-Agents v1 contract. It validates
portable trees and invokes explicitly locked adapter executables; vendor
translation code never runs in the host process.

## Install

```bash
go install github.com/Open-Dot-Agents/Open-Dot-Agents/cli/cmd/dota@v1.0.0-rc.1
dota --version
```

Release archives contain `dota` and the Codex, Copilot, and Claude reference
adapter executables. Verify an archive against `checksums.txt` before use.

For local development:

```bash
task cli:build
export PATH="$PWD/cli/bin:$PATH"
```

## Commands

- `init` creates a `1.0.0-rc.1` manifest.
- `validate` validates the portable tree, optionally with one locked adapter.
- `inspect` reports the manifest, populated categories, and adapter locks.
- `adapter add|install|list|doctor` manages explicit adapter trust.
- `import|export` asks one adapter for a plan and safely applies it.
- `check` detects drift from the adapter's current deterministic plan.
- `clean` removes only unchanged files owned by one adapter operation.
- `conformance tree|adapter` runs the normative conformance entry points.

All operation commands require `--adapter <reverse-dns-id>`. There is no target
registry, implicit discovery, multi-target mode, or legacy command alias.

## Locked adapter workflow

```bash
root=/path/to/project
adapter=/path/to/dota-adapter-codex

dota validate --root "$root"
dota adapter add --root "$root" \
  --id org.open-dot-agents.codex \
  --version dev \
  --path "$adapter"
dota adapter doctor --root "$root" --adapter org.open-dot-agents.codex
dota export --root "$root" --adapter org.open-dot-agents.codex --dry-run --json
dota export --root "$root" --adapter org.open-dot-agents.codex
dota check --root "$root" --adapter org.open-dot-agents.codex
dota clean --root "$root" --adapter org.open-dot-agents.codex
```

Use `--operation import` with `check` and `clean` to select import ownership.
Use `--force` only after reviewing conflicts, and `--backup` when replacing
existing files. Machine-readable output uses `--json`.

`--ci` enforces release-adapter trust and rejects local executable paths. A
publisher manifest can be added and installed with:

```bash
dota adapter add --root "$root" --manifest https://publisher.example/adapter.json
dota adapter install --root "$root"
dota adapter doctor --root "$root" --adapter org.example.adapter --ci
```

The lock file is `.agents/adapters.lock.json` and should be committed. Installed
artifacts are checksum-verified. `dota` never searches `PATH` for an adapter.

## Ownership and safety

Operation state is host-owned:

```text
.agents/.dota/export/<adapter-id>.json
.agents/.dota/import/<adapter-id>.json
```

Adapters receive only regular files matching their declared input patterns.
The host rejects traversal, symlinks, duplicate paths, case-fold collisions,
illegal `.agents/` writes, malformed plans, modified owned files, and stale
output. Writes use same-directory temporary files and atomic rename.

## Conformance and development

```bash
task cli:test
task cli:test:conformance
task cli:verify
```

The language-neutral fixture adapter at
`conformance/adapters/python/dota_adapter_fixture.py` demonstrates the JSON-RPC
contract without importing Go packages. The protocol is specified in
[`spec/v1/ADAPTER_PROTOCOL.md`](../spec/v1/ADAPTER_PROTOCOL.md).

Authenticated native probes are intentionally separate because they need
installed harnesses, credentials, network access, and may consume paid usage:

```bash
task cli:test:codex:real
task cli:test:copilot:real
```
