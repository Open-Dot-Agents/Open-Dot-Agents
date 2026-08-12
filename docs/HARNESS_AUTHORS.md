# Harness Author Integration Guide

Harnesses should consume root and scoped `AGENTS.md`, `.agents/manifest.json`,
`.agents/tools/mcp.json`, and `.agents/skills` directly. Native consumption
avoids generated-file ownership and drift.

1. Reject unsupported major versions and unknown selected profiles.
2. Validate JSON Schema plus cross-file selection, path, and capability rules.
3. Resolve `urn:open-dot-agents:env:VARIABLE` only at execution time. Never log
   the value or replace a missing variable with literal text.
4. Preserve root-to-leaf instruction scope and nearest-file precedence.
5. Refuse activation before starting tools when required capabilities cannot
   be represented without loss.
6. Emit standard conformance result JSON and stable diagnostic codes.
7. Run the portable fixtures and a native black-box suite for every advertised
   harness version.

Propose an integration with the adapter issue template. A supported registry
entry needs immutable harness provenance, public evidence, limitations, and a
revalidation policy.
