# Open-Dot-Agents

Open-Dot-Agents is an Apache-2.0, vendor-neutral standard for
repository-scoped AI agent configuration. A project owns one portable
`.agents/` tree; explicitly versioned adapters project it to individual
harnesses without making the canonical model vendor dependent.

The program is composed of three independently versioned components:

| Component | Purpose |
| --- | --- |
| [Specification](SPEC) | The normative portable format, profiles, schemas, and conformance fixtures. |
| [Reference CLI](CLI) | The `agents` command for creating, validating, importing, planning, and applying configurations. |
| [Workbench](WORKBENCH) | Experimental mappings and harness research that must graduate through conformance before becoming standard. |

## Status

The project is preparing the Open-Dot-Agents 1.0 release. No native adapter is
currently marked supported in the [compatibility matrix](docs/COMPATIBILITY.md).
Release-candidate CLI archives are evaluation artifacts, not final release
assets; see [installation guidance](docs/INSTALL.md). The release workflow
publishes checksums, SPDX SBOMs, and GitHub provenance attestations only after
the three native adapter gates pass. Treat other mappings as experimental.

## Start here

1. Read the [specification](SPEC) to understand the canonical `.agents/`
   format and profile guarantees.
2. Install or build the [reference CLI](CLI) to manage a repository
   configuration.
3. Check the [compatibility matrix](docs/COMPATIBILITY.md) before relying on a
   native harness projection.
4. Consult [vendor mapping evidence](docs/VENDOR_EVIDENCE.md) before implementing
   or extending an adapter.
5. Use the [migration guide](docs/MIGRATION.md) to adopt the portable source of
   truth safely.
6. Follow the [installation guide](docs/INSTALL.md) for the reference CLI.
7. Use the [harness author guide](docs/HARNESS_AUTHORS.md) to implement native
   consumption and publish comparable conformance results.

All project-level documentation is indexed in [docs](docs/README.md).

## Project governance

The standard is developed in public. The rules for participating, reporting
security issues, releasing compatible versions, and making decisions are in:

- [Contributing](docs/CONTRIBUTING.md)
- [Code of Conduct](docs/CODE_OF_CONDUCT.md)
- [Security policy](docs/SECURITY.md)
- [Governance](docs/GOVERNANCE.md)
- [Versioning policy](docs/VERSIONING.md)
- [Release process](docs/RELEASING.md)
- [Roadmap](docs/ROADMAP.md)
- [Changelog](docs/CHANGELOG.md)
