---
name: agents-conformance
description: Verify Open-Dot-Agents repository changes against the canonical .agents tree, specification conformance baseline, Go CLI tests, Workbench projection tests, JSON validity, whitespace checks, and compatibility-claim boundaries. Use before final responses, commits, releases, adapter promotion, or when auditing whether the project is ready to claim standard status.
---

# Agents Conformance

## Core Checks

Prove only what the executed checks cover. Keep portable conformance, CLI unit
behavior, projection tests, and native harness evidence separate.

1. Check the worktree first with `git status --short`; preserve unrelated changes.
2. Validate edited JSON with `python3 -m json.tool <file>` for touched JSON files.
3. Validate the repository dogfood tree with `go run ./cmd/agents validate --source ../.agents` from `CLI`, or an equivalent freshly built `agents` binary.
4. Run the spec baseline: `python3 SPEC/conformance/run.py`.
5. Run CLI tests from `CLI`: `GOCACHE=/tmp/agents-gocache GOPATH=/tmp/agents-gopath go test ./...`.
6. Run Workbench projection tests from `WORKBENCH`: `python3 task/test/mcp_projections_test.py`.
7. Run `git diff --check`.

## Evidence Rules

Do not use green CLI or projection tests as native harness support evidence.
Only a version-pinned black-box harness run can support a compatibility matrix
upgrade.
