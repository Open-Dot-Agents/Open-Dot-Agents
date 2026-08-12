---
name: agents-spec
description: Update the Open-Dot-Agents portable specification, manifest or MCP schema, profile semantics, examples, fixtures, governance-facing standard text, or migration rules. Use when a task changes normative behavior under SPEC, public compatibility semantics, schema versioning, or the canonical .agents configuration contract.
---

# Agents Spec

## Workflow

Use the specification as the source of truth. Keep implementation, examples,
schemas, docs, and compatibility claims aligned without silently widening the
standard.

1. Inspect the current normative surface before editing: `SPEC/spec/1.0/SPECIFICATION.md`, `SPEC/spec/1.0/schemas/`, `SPEC/examples/`, `SPEC/conformance/`, `VERSIONING.md`, and `GOVERNANCE.md`.
2. Classify the change as clarification, compatible addition, or breaking change. Follow `VERSIONING.md`; do not change `version: "1.0.0"` semantics casually.
3. Update normative prose and schemas together. JSON Schema alone is not enough when cross-file profile behavior or capability-loss rules change.
4. Add or update valid and invalid fixtures for every observable semantic change. Prefer small fixtures that isolate one rule.
5. Update migration, compatibility, and changelog docs when a user or adapter author must act differently.
6. Keep adapter claims conservative. A spec change does not make a native harness supported without version-pinned evidence.

## Validation

Run `python3 SPEC/conformance/run.py`. If CLI behavior changes or generated
projections depend on the spec change, also run the focused CLI and Workbench
tests that cover that behavior.
