# Repository Agent Instructions

Always use ASD-STE100 Simplified Technical English.
Use the project skills for recurring work:

- `$agents-spec` for normative specification, schema, fixture, and migration changes.
- `$agents-cli` for Go reference CLI and adapter projection changes.
- `$agents-conformance` for validation, compatibility-claim audits, and final verification.
- `$agents-adapter` for native harness evidence and adapter promotion work.
- `$agents-release` for release notes, installation guidance, and artifact claim checks.
- `$agent-docs` for official llms.txt documentation sources before vendor,
  MCP, skills, tool, API, or adapter research.

Keep root and nested `AGENTS.md` files as portable instructions. Keep the
manifest, MCP catalogue, skills, and implementation state under `.agents/`.
Do not mark an adapter supported without version-pinned native harness evidence.
