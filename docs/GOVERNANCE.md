# Governance

## Principles

Open-Dot-Agents is developed in public and optimized for portability, explicit
compatibility, user safety, and implementability. The specification is
vendor-neutral; a vendor does not receive special control over the canonical
format.

## Maintainer and decisions

The founding maintainer is Maurizio Casciano (`@MaurizioCasciano`). The
maintainer reviews pull requests, triages issues, cuts releases, publishes
compatibility evidence, and has final decision authority. This project uses a
single-maintainer model; participation does not imply decision authority.

Changes that alter a normative profile, schema, compatibility guarantee, or
governance rule require a public proposal, at least 14 calendar days for
comment, and a documented decision under `docs/decisions/`. The maintainer may
shorten the period only for a privately reported security issue and must record
the reason after coordinated disclosure. Each record lists the problem,
options, compatibility and security effects, dissent, and migration.

The maintainer discloses material conflicts of interest in the proposal. If the
maintainer is unavailable for 90 days, they may name a successor in a signed
repository change. If no successor exists, contributors may fork under the
Apache-2.0 license; no contributor inherits credentials or authority by
activity alone.

Protected branches require passing CI, resolved review conversations, signed
off commits, and no force pushes. Administrative bypass is reserved for
security response and recovery, and must be documented afterward.

## Adding a standard capability

New profiles and adapters begin as proposals. To become supported, they need
normative documentation, schema changes where applicable, security analysis,
valid and invalid fixtures, a reference or independent implementation, and
passing conformance evidence. Breaking changes follow the major-version
process in [VERSIONING.md](VERSIONING.md).

## Adoption claims

The project reports measurable facts and does not call itself a de-facto
standard until all of these are public:

1. two independently maintained conforming implementations;
2. three non-demo public production adopters; and
3. one acknowledged or merged integration with a supported harness.

Entries are recorded in [implementations](IMPLEMENTATIONS.md) and
[adopters](ADOPTERS.md). The maintainer verifies links but does not certify
security or fitness for a third-party product.
