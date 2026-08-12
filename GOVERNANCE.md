# Governance

## Principles

Open-Dot-Agents is developed in public and optimized for portability, explicit
compatibility, user safety, and implementability. The specification is
vendor-neutral; a vendor does not receive special control over the canonical
format.

## Maintainers and decisions

Repository maintainers review pull requests, triage issues, cut releases, and
publish compatibility evidence. Material decisions are recorded in issues or
architecture decision records with the problem, considered options, decision,
and migration impact.

Changes that alter a normative profile, schema, compatibility guarantee, or
governance rule require a public proposal, at least two maintainer approvals,
and a documented decision. Maintainers aim for consensus; if consensus cannot
be reached, a simple majority of active maintainers decides. An active
maintainer is one who has made a substantive contribution or review within the
previous six months.

## Adding a standard capability

New profiles and adapters begin as proposals. To become supported, they need
normative documentation, schema changes where applicable, security analysis,
valid and invalid fixtures, a reference or independent implementation, and
passing conformance evidence. Breaking changes follow the major-version
process in [VERSIONING.md](VERSIONING.md).
