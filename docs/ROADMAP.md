# Roadmap to Open-Dot-Agents 1.0

## Foundation

Publish the normative portable model, JSON Schemas, versioning rules, security
guidance, governance, and valid/invalid examples.

## Conformance

Create implementation-neutral fixtures for instructions, MCP, and skills.
Publish a compatibility matrix with evidence for every adapter and harness
version.

## Reference implementation

Evolve the Go CLI to validate canonical trees, expose adapter capabilities,
project instructions/MCP/skills safely, and retain explicit diff, backup, and
loss diagnostics.

## Adapter releases

Deliver verified repository-scoped adapters for GitHub Copilot CLI, OpenAI
Codex, Claude Code, and OpenCode. Native formats, paths, and precedence will
be pinned per harness release and tested rather than assumed.

## Ratification and adoption

Publish final release artifacts, migration guides, examples, and adapter author
documentation. Add signing, SBOMs, or provenance only with the corresponding
published verification material. Ratify 1.0 after all guaranteed adapters pass
the declared profile conformance suite and their limitations are visible in the
compatibility matrix.
