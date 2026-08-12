---
name: agents-adapter
description: Promote or assess an Open-Dot-Agents adapter for a native harness, including upstream documentation review, version-pinned black-box evidence, capability/loss declarations, Workbench graduation, compatibility matrix updates, and conservative support claims. Use for Copilot, Codex, Claude Code, OpenCode, or any proposed new adapter.
---

# Agents Adapter

## Workflow

Treat adapter support as an evidence claim, not an implementation preference.
Search current official vendor documentation before relying on native behavior.

1. Identify the exact harness, version, OS, trust mode, command, and repository fixture under test.
2. Verify official current documentation for instructions, MCP, skills, path precedence, trust behavior, and reload behavior.
3. Build or select fixtures that cover nested instruction discovery, MCP discovery, native server startup, skill discovery, collisions, unsupported fields, and credential references.
4. Capture durable evidence using `WORKBENCH/evidence/ADAPTER_EVIDENCE_TEMPLATE.md` before editing support claims.
5. Update `docs/VENDOR_EVIDENCE.md` for upstream facts, `docs/COMPATIBILITY.md` and `CLI/compatibility.json` for support status, and CLI capabilities only when all claims are backed by evidence.
6. If a projection is lossy, make the adapter refuse activation or report every lost item before activation. Never silently omit selected canonical content.

## Promotion Bar

A supported adapter row must name exact Open-Dot-Agents and harness versions,
profile outcomes, evidence links, limitations, and required user actions. Unit
tests, docs-only research, or release-candidate artifacts are insufficient.
