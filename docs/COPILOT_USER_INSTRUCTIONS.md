# Copilot user instructions

Draft.2 imports and projects `copilot-instructions.md` at the explicit
`--native-home`. The artifact kind remains `instructions`. The namespace source
can have a different name, such as `policy/local.md`; the native target remains
fixed. Reimport now reuses that declaration, including its metadata. It does
not add a second source based on the native filename. The same source-routing
fix has Go regression coverage for Codex user instructions.

Use `--experimental --scope user --native-home /absolute/native/home` for user
import, plan, and apply. The native CLI must use that home too. Copilot normally
uses `$HOME/.copilot`; `COPILOT_HOME` can select another native home. Setting the
adapter option does not change the running client's environment. See the
[example](../SPEC/examples/user-instructions/.agents/manifest.json).

User import does not import project instructions into the user body. Project
apply does not write the user instruction file. Duplicate user instruction
targets refuse import. Existing declared policy cannot be replaced through a
forced reimport. Another repository cannot acquire or remove an owned user
instruction file through force or adoption.

New user files use mode `0600`; existing file modes remain intact. Removing
the selected artifact removes only the owned instruction file. Referenced user
files remain external. Private backups and ownership records use the existing
transaction rules. Plan reports the external reference requirement and the
need for a new native session. Apply does not copy references or grant trust.

## Defect and evidence

The [failed native receipt](../WORKBENCH/evidence/native-draft2-debug/copilot-user-instructions-custom-source-before.json)
loads the same user body and references before and after relocation, then
reimport changes the declaration. The
[retained source inspection](../WORKBENCH/evidence/native-draft2-debug/copilot-user-instructions-source-duplication.json)
shows both `policy/local.md` and a new `copilot-instructions.md` assigned to the
same target. Source selection caused the loss; ownership did not change the
declaration.

The [source receipt](../WORKBENCH/evidence/native-draft2-debug/copilot-user-instructions.sources.json)
contains the live GitHub index and instruction guide. Native tests use Copilot
`1.0.83` on Linux, binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`,
isolated homes, preconfigured fixture trust, and a local model endpoint.

Three cases use the default user home, an explicit `COPILOT_HOME`, and a custom
canonical source path. Each case has four fresh native sessions:

| Phase | Observed result |
| --- | --- |
| Source | User and project bodies load together; user references and their child reference load from the native home |
| Relocated | The same bodies and references load after import and user apply; reimport preserves the source |
| Updated | The edited user marker loads in a new session |
| Removed | The removed user body and references are absent; project instructions still load |

Project reference decoys do not load. Model requests, session IDs, native read
events and results, and completed turns are correlated. The fixture model
reads a known project file. These observations do not prove that a model
follows arbitrary instructions, that arbitrary references work, or that active
sessions reload changed files.

The first Go test also counted the import lock outside `.agents`; the corrected
test compares canonical content. A native probe passed all four session checks
but incorrectly treated Copilot's own first-launch metadata update as an
adapter write. Both failures remain recorded. Current checks compare external
configuration bytes around each adapter call and separately verify unchanged
native trust entries.

## Verification

The [Go tests](../CLI/internal/config/native_user_instructions_test.go) check
source and metadata preservation, Codex source routing, changed policy refusal,
duplicate targets, scope, foreign ownership, file modes, removal, and rollback
after the final transaction write. The
[native verifier](../WORKBENCH/conformance/verify_copilot_user_instructions.py)
checks 12 native sessions and current Go hashes. Its tests reject false update
and removal claims, changed reference bases, and missing native effects.

Current user receipts use `copilot-user-instructions-*-final.json`. The combined
record is `verification-copilot-user-instructions-final.json`; other selected
skill and instruction regressions use `-user-instructions.json`. Older records
retain their original hashes. This work does not promote adapter support or
complete the milestone.
