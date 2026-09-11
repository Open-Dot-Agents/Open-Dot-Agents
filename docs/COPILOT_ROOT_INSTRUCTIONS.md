# Copilot root instruction import

Draft.2 project import now reads a regular root `AGENTS.md`. It imports the
bytes into `.agents/AGENTS.md` and leaves the source file unchanged. The
Copilot projection destination remains `.github/copilot-instructions.md`,
unless the project has a verified canonical compatibility link or an explicit
`canonical-instructions` binding. The
[canonical instruction report](COPILOT_CANONICAL_INSTRUCTIONS.md) covers that
fixed root destination and separate native instruction bodies.

Import accepts identical root and native instruction bodies. A distinct root
body or a root file with potential `@` references now enters the native
`agent-instructions` mapping. This retains its original file location and
reference base. See the [native agent instruction report](COPILOT_AGENT_INSTRUCTIONS.md)
for the mapping and native scope controls. Earlier conflict and reference
refusals remain in the historical evidence.

Existing canonical portable policy stays protected during additive import.
Projection still refuses an unmanaged regular root file that differs from the
canonical body unless an explicit native artifact preserves that separate body.
It must not leave stale root instructions active beside new portable instructions.

User-scope import does not read project root instructions. Stable 1.0 and
draft.1 behavior are unchanged. Nested instruction discovery, other agent
instruction filenames, and live reload are outside this result.

## Evidence

The [original CLI reproduction](../WORKBENCH/evidence/native-draft2-debug/copilot-root-instructions-import-before.json)
records a root-only import failure and a successful import that omits a
distinct root body. The [failed regression](../WORKBENCH/evidence/native-draft2-debug/copilot-root-instructions-regression-before.json)
also records acceptance of a directory in place of root `AGENTS.md`.

The [source record](../WORKBENCH/evidence/native-draft2-debug/copilot-root-instructions.sources.json)
contains the live GitHub documentation index and custom instruction guide.
The guide documents combined instruction sources, identical-body removal,
relative file references, and session restart for instruction changes.

The [native runner](../WORKBENCH/conformance/run_native_copilot_root_instructions.py)
uses Copilot `1.0.83`, binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`,
isolated homes, and a local model endpoint. Fixture setup supplies trust;
the adapter does not change trust or credentials.

Four cases cover eight native sessions:

- Root-only and identical-body sources load before and after relocation.
- Distinct root and Copilot bodies both load before and after relocation.
- A root `@policy.md` loads the same root policy before and after relocation.

All cases preserve import, apply, and reimport data. A retained earlier manual
copy into `.github` loads a different policy; the new native mapping keeps the
root file at `AGENTS.md` and does not perform that conversion.

The [verifier](../WORKBENCH/conformance/verify_copilot_root_instructions.py)
checks model requests, session IDs, native read events and results, completed
turns, roundtrip results, and implementation hashes. The fixture model reads
one known file; this does not prove compliance with arbitrary instructions.
The [Go tests](../CLI/internal/config/native_root_instructions_test.go) check
byte preservation, repeated import, conflicting canonical policy, force,
malformed roots, scope isolation, and stale-root projection refusal.

Current receipts end in `-user-instructions.json`. Earlier receipts are historical.
This result does not promote adapter support or complete the Linux milestone.
