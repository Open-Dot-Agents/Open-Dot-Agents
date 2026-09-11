# Copilot native agent instruction files

Draft.2 can preserve native agent instruction files without a change to their
bytes or locations. The `agent-instructions` artifact kind registers four
Copilot project paths: `AGENTS.md`, `CLAUDE.md`, `.claude/CLAUDE.md`, and
`GEMINI.md`. It cannot select an arbitrary output path or a user-scope target.

Import keeps plain root instructions in the portable core when conversion is
lossless. A root file with potential `@` references, or a body that differs
from `.github/copilot-instructions.md`, stays in the native namespace as an
`agent-instructions` artifact. The Copilot instruction body enters the portable
core. Other registered agent instruction files stay in their native format.
Existing canonical policy and custom native source paths remain protected.

This replaces the earlier root-reference and distinct-body import refusals
with a tested mapping. It does not combine instruction bodies or translate
native reference syntax. Each native file returns to its original location.
Referenced project files remain external dependencies. Plan reports this
requirement; apply does not copy them or change native trust.

An unselected, unmanaged regular root file that differs from the canonical
body still blocks a new portable projection. An explicit native artifact can
preserve that separate body. Duplicate assignments to one native target remain
an error, including with force. Each selected native file uses the existing
ownership, conflict, backup, removal, and transaction rules.

An existing canonical root symlink uses the fixed `canonical-instructions`
binding. It retains a distinct `.github/copilot-instructions.md` body and the
root reference base after relocation. See the
[canonical instruction report](COPILOT_CANONICAL_INSTRUCTIONS.md) for that
separate native check and its limits.

## Native behavior

The [source record](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-instructions.sources.json)
contains the live GitHub index and instruction guide. The guide lists all four
agent instruction paths. The tests use Copilot `1.0.83` on Linux, binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`,
isolated homes, and a local model endpoint.

Four cases each run before import and after relocation:

| Source and session location | Observed result |
| --- | --- |
| All registered files; repository-root session | Root, Claude, Gemini, and Copilot bodies load; `.claude/CLAUDE.md` does not load |
| All registered files; `.claude` session | All bodies load, including both Claude instruction files |
| Only `.claude/CLAUDE.md`; repository-root session | Its body does not load |
| Only `.claude/CLAUDE.md`; `.claude` session | Its body loads |

Native `@` references in root, Claude, and Copilot instructions keep the same
file base after relocation. A reference within the root policy also loads its
child policy. `GEMINI.md` keeps the literal reference; the native client does
not load that referenced file. Referenced files remain unchanged.

The first tests incorrectly expected `.claude/CLAUDE.md` to load from the
repository root. Three failed attempts remain in the evidence directory.
Root `CLAUDE.md` precedence did not explain the result: the isolated file also
did not load in a root session. The `.claude` working-directory controls
establish the observed scope. Later file-triggered discovery, other nested
scopes, and live reload remain unverified.

## Verification

The [Go tests](../CLI/internal/config/native_agent_instructions_test.go) cover
exact bytes and paths, repeated import, custom source routing, portable policy,
duplicate assignments, unsafe sources, scope restrictions, removal, backups,
and rollback after the final transaction write.
The [specification example](../SPEC/examples/native-agent-instructions/.agents/manifest.json)
and schema cases cover the metadata and path contract.

The [native runner](../WORKBENCH/conformance/run_native_copilot_agent_instructions.py)
and [verifier](../WORKBENCH/conformance/verify_copilot_agent_instructions.py)
check eight correlated native sessions, instruction and reference text in model
requests, file-read events and results, source preservation, and Go source
hashes. Separate tests reject missing reference text, missing native effects,
and a false scope claim. The fixture model reads one known file; this is not
proof that a model follows arbitrary instruction text.

Current native receipts use `copilot-agent-instructions-*-user-instructions.json`.
The plain-root and reference regressions use `copilot-root-instructions-*-user-instructions.json`.
The recursive instruction regressions use `copilot-recursive-instructions-*-user-instructions.json`.
Older receipts remain historical. The combined record is
`verification-copilot-user-instructions-final.json`. This work does not promote
adapter support, clear portable security requirements, or complete the milestone.
