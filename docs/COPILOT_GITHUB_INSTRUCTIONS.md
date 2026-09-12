# Copilot GitHub instruction references

Draft.2 import preserves `.github/copilot-instructions.md` at its native
location when it contains potential `@` references. It stores the bytes in a
native `instructions` artifact and uses the existing `canonical-instructions`
binding for the portable root body. A separate root instruction body becomes
the core. With no root body, import retains existing portable policy or creates
the normal default core. Plain instruction text keeps its portable mapping.

This fixes a reproduced change in policy loading. Previously, import placed
the GitHub body in the portable core. Apply to a project with a root canonical
link then omitted the GitHub file. Copilot resolved `@policy.md` from the
project root and loaded a different policy. The
failed native receipt (`WORKBENCH/evidence/native-draft2-debug/copilot-github-reference-before.json`)
retains both native sessions. The
failed Go regression (`WORKBENCH/evidence/native-draft2-debug/copilot-github-reference-regression-before.json`)
retains six failing import and relocation cases.

Import preserves declared custom source paths and existing policy. It refuses
conflicting native root artifacts, malformed Markdown, unsafe paths, and an
ambiguous reference base in existing core content. Force cannot bypass these
checks. Apply also refuses a core containing potential references through a
root link without an explicit root binding. For older imports, review the
original instruction locations before declaring that binding. Do not label
GitHub-relative text as root-relative text without resolving its references.

## Native evidence

The source record (`WORKBENCH/evidence/native-draft2-debug/copilot-github-instructions.sources.json`)
contains the fetched GitHub documentation index and custom instruction guide.
The guide documents relative references and combined instruction discovery.

Tests use Copilot `1.0.83` on Linux with binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`,
isolated homes, fixture trust, and a local model provider. Three cases cover a
target with a canonical link, a regular root, and combined root references.
Each case runs source, relocated, updated, and removed sessions. The correct
GitHub policy loads; root decoys do not. Root references remain independent.
Removal deletes the owned instruction file and retains referenced files.

Current receipts are `copilot-github-reference-{link,regular,combined}-source-final.json`.
The [verifier](../WORKBENCH/conformance/verify_copilot_github_instructions.py)
checks twelve distinct sessions, model context, correlated native events,
observable file reads, completed turns, source and authority preservation,
roundtrips, updates, removals, and source hashes. Verifier tests reject wrong
reference bases, stale updates, retained removed bodies, unrelated sessions,
and missing read results. Root instruction regressions use `-github-reference-retry`
receipts, with `-github-reference-final` for the identical-body case. Three
concurrent runs timed out during initialization before adapter execution;
their receipts remain intact. Earlier receipts remain historical.

Referenced files remain external. Apply does not expand references, grant
trust, or change user configuration. These tests do not prove arbitrary model
compliance or live reload. Stable 1.0 and draft.1 semantics, adapter support
status, and release gates remain unchanged. The wider milestone is incomplete.
