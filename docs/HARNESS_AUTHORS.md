# Harness Author Integration Guide

Harnesses should consume canonical `.agents/AGENTS.md`, optional scoped
`AGENTS.md`, `.agents/manifest.json`, `.agents/tools/mcp.json`, and
`.agents/skills` directly. A root `AGENTS.md` is a compatibility link or native
projection, not a second canonical source.

1. Reject unsupported major versions and unknown selected profiles.
2. Validate JSON Schema plus cross-file selection, path, and capability rules.
3. Resolve `urn:open-dot-agents:env:VARIABLE` only at execution time. Never log
   the value or replace a missing variable with literal text.
4. Apply `.agents/AGENTS.md` repository-wide and preserve nearest-file
   precedence for nested instructions.
5. Refuse activation before starting tools when required capabilities cannot
   be represented without loss.
6. Emit standard conformance result JSON and stable diagnostic codes.
7. Run the portable fixtures and a native black-box suite for every advertised
   harness version.

Propose an integration with the adapter issue template. A supported registry
entry needs immutable harness provenance, public evidence, limitations, and a
revalidation policy.
