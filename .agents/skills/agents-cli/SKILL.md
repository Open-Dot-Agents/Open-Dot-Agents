---
name: agents-cli
description: Modify the Open-Dot-Agents Go reference CLI, import/export/convert behavior, adapter projections, validation logic, capabilities output, overwrite/diff/backup handling, or CLI documentation. Use for tasks under CLI or for code changes that project canonical .agents content into native harness files.
---

# Agents CLI

## Workflow

Use the CLI as a reference implementation of the portable contract, not as an
independent standard.

1. If `.codegraph/` exists, use CodeGraph before reading or changing Go code. Name the symbols or files involved, such as `ExportWithOptions`, `ImportWithOptions`, `Validate`, `VendorCapabilities`, `CLI/cmd/agents/main.go`, and `CLI/internal/config/config.go`.
2. Preserve overwrite safety, symlink rejection, backup behavior, and explicit loss diagnostics. Do not silently drop canonical data to satisfy a native format.
3. Keep supported target names and paths aligned across implementation, CLI help, tests, `CLI/README.md`, `docs/COMPATIBILITY.md`, and `compatibility.json`.
4. For adapter changes, update both renderer/importer behavior and capability status. Do not mark an adapter conformance-supported from unit tests alone.
5. Use temporary Go caches in restricted environments: `GOCACHE=/tmp/agents-gocache GOPATH=/tmp/agents-gopath go test ./...`.

## Validation

Run `go test ./...` from `CLI`. For behavior that affects canonical fixtures or
Workbench projections, also run `python3 SPEC/conformance/run.py` and
`python3 task/test/mcp_projections_test.py` from `WORKBENCH`.
