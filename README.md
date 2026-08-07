# Open-Dot-Agents

Open configuration standard for AI coding agents.

Every AI coding harness (GitHub Copilot CLI, OpenAI Codex CLI, Claude Code, ...)
invents its own directory layout and file format for the same handful of
customization concepts: persistent instructions, reusable prompts, custom
sub-agents, reusable skills, external tool connections, lifecycle hooks,
safety guardrails, and so on.

`Open-Dot-Agents` proposes a single, vendor-neutral `.agents/` folder that
groups these concepts into well-defined categories, so that tooling can adopt
one standard layout instead of a different one per vendor.

## Structure

```
.agents/
├── agents/       # Custom sub-agent / persona definitions
├── guardrails/   # Safety limits, sandboxing, read-only/restricted modes
├── hooks/        # Lifecycle automation (pre/post tool call, session start/end, ...)
├── instructions/ # Always-on, repo-wide behavioral instructions
├── memories/     # Persistent cross-session knowledge/context
├── permissions/  # Tool allow/deny/ask rules and approval policies
├── plugins/      # Installable bundles combining any of the above categories
├── profiles/     # Named, swappable presets of settings/permissions/tools
├── prompts/      # Reusable, parameterized prompt templates / slash commands
├── rules/        # Scoped behavioral rules (e.g. path- or language-specific)
├── settings/     # Base runtime configuration (model, env vars, sandbox mode, ...)
├── skills/       # Portable, on-demand task playbooks
├── tools/        # External tool/service connections (e.g. MCP servers)
├── mappings.yaml # Maps each category to the vendor docs it corresponds to
└── schema/      # Versioned machine-readable contracts
```

Each category folder has its own `README.md` describing its purpose and expected
content. `mappings.yaml` cross-references each category with the equivalent
feature documentation from existing harnesses, to help adopters and tooling authors
translate between the open standard and vendor-specific formats. The machine-readable
contracts are in:

- `.agents/schema/v1/agents.schema.json`
- `.agents/schema/v1/mappings.schema.json`

`manifest.json` records explicit versioning, canonical root semantics, and
compatibility notes for this repository version.

## Status

This is an early-stage proposal. Category boundaries and file formats are
still being refined — see individual category READMEs and `mappings.yaml`
for current scope and open questions.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
