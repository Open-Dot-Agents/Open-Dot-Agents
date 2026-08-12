# Contributing to Open-Dot-Agents

Thank you for helping make repository-scoped agent configuration portable and
safe.

## Where changes belong

| Change | Repository |
| --- | --- |
| Normative behavior, schemas, examples, fixtures | `SPEC` |
| Reference command behavior and adapter implementation | `CLI` |
| Harness experiments and mapping research | `WORKBENCH` |
| Program-wide governance and public entry points | this repository |

An experimental mapping must not be described as supported until it has a
normative specification, conformance fixtures, and a passing implementation.

## Proposing a change

Open an issue before changes that alter canonical semantics, add a profile,
change a schema, or add a supported adapter. Explain the user problem,
compatibility impact, security considerations, migration path, and required
fixtures. Small documentation and defect fixes may be submitted directly.

Pull requests should be narrowly scoped, include tests or fixture changes for
observable behavior, and update documentation and changelogs when needed.
Generated harness files must not replace the portable `.agents/` source of
truth.

## Adapter requirements

An adapter proposal must identify the verified native harness version, all
supported profiles, native configuration paths and precedence, fields that
cannot round-trip, and the tests that back each claim. It must preserve
unsupported canonical data by refusing the operation or emitting an explicit
diagnostic; silent loss is not acceptable. Before calling an adapter
supported, record its exact evidence and limitations in the
[compatibility matrix](COMPATIBILITY.md); a successful CLI build or a
release-candidate artifact is not harness verification.

## Development expectations

Run the component's documented test command before submitting changes. Do not
commit credentials, real tokens, or private configuration. By participating,
you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
