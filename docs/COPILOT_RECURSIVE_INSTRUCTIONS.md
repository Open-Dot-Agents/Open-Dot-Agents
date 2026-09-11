# Copilot recursive instruction import

Draft.2 now imports nested Copilot instruction files and preserves their
relative paths. The previous importer read only the first directory level.
It omitted nested files even when Copilot loaded them at the source.

| Scope | Native directory | Canonical directory |
| --- | --- | --- |
| Project | `.github/instructions/` under the supplied project root | `.agents/native/com.github.copilot/scoped-instructions/` |
| User | `instructions/` under the explicit absolute `--native-home` | `.agents/native/com.github.copilot/scoped-instructions/` |

For example, `nested/deep/fixture.instructions.md` keeps this relative name
through import, apply, and reimport. A flat file with the same basename is a
separate asset. The existing draft.2 `scoped-instructions` artifact uses its
`name` field for this relative path. No schema version or portable profile
semantics change is required.

## Import and ownership

The [importer](../CLI/internal/config/native_recursive_instructions.go) walks
only the registered instruction directory. It does not recursively copy a
native home. It preserves instruction bytes and reports unrecognized files
as excluded. Symlinks, indirect parent paths, non-regular files, non-directory
instruction roots, and unsafe destination names refuse import before content
writes. The agent and hook registries retain their flat filename rules.

Each projected instruction is an owned file. User scope requires an absolute
native home and uses the private state registry. Another repository cannot
replace or remove the file through `--force`. New user files and backups use
`0600`. Existing transaction and ownership checks apply to nested files.

Plan reports a separate native action for user instructions: reading a
path-specific instruction outside trusted directories can require read
permission in Copilot. The native user can allow or deny that request. Apply
does not grant trust to the native home or approve future reads.

The [Go tests](../CLI/internal/config/native_recursive_instructions_test.go)
cover same-basename files, exact content, excluded files, repeat projection,
reimport, source preservation, cross-repository ownership refusal, removal,
private backups, and rollback after the last transaction write. A failure
after backup, configuration, removal, and ownership operations must restore
the original state. Project projection keeps the existing project file-mode
rules; the initial test expectation of private project files was corrected.

## Native evidence and limits

The [source record](../WORKBENCH/evidence/native-draft2-debug/copilot-recursive-instructions.sources.json)
contains the live GitHub documentation index, command reference, and custom
instruction guide. The guide lists recursive project and user directories and
describes `applyTo` matching. Documentation alone does not establish behavior.

The [runner](../WORKBENCH/conformance/run_native_copilot_recursive_instructions.py)
uses Copilot `1.0.83` on Linux, with binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.
It uses isolated homes and a local model endpoint. Fixture setup supplies
trust; the adapter does not write trust or account state.

Seven cases each run a source and relocated native process:

- Project and user instructions with `applyTo: "**"`: flat and nested bodies
  reach the model directly before and after relocation.
- Project and user instructions with `applyTo: "**/*.go"`, while the model
  reads a text file: the native catalog lists both files. The fixture model
  does not request their bodies. This does not prove native glob exclusion.
- Project scoped instructions: the fixture model reads both instruction paths
  from Copilot's native catalog, then reads the Go file. Both instruction bodies
  reach the model through native `view` results before and after relocation.
- User scoped instructions with approval: each of the two instruction reads
  receives one exact-path native approval, and both bodies reach the model.
- User scoped instructions with denial: the first instruction read is rejected,
  and the tested native turn ends without either instruction body.

The [verifier](../WORKBENCH/conformance/verify_copilot_recursive_instructions.py)
checks definition hashes, source preservation, source-code hashes, session IDs,
native file-read events, file contents in tool results, completed turns, and
instruction text in model requests. Nine verifier tests reject altered evidence
and wider fixture permission requests.
The native runs perform file reads; they do not claim that a model follows
arbitrary instruction text.

A separate [matching Go-file read](../WORKBENCH/evidence/native-draft2-debug/copilot-recursive-instructions-project-yes-before.json)
did not automatically load either `applyTo: "**/*.go"` body. The earlier test
expectation was incorrect. The native prompt lists the pattern, file path,
and description, and tells the model to use `view` to read the instructions.
That probe never requested the instruction files. A missing body did not mean
that discovery or pattern parsing had failed.

The [trigger verifier](../WORKBENCH/conformance/verify_copilot_instruction_triggers.py)
checks nine retained native experiments: file mentions, ACP resource links,
tracked files, a later user turn, denied and approved edits, and three pattern
controls. Non-global instructions appear in the native table. An approved
fixture edit changes the file even when the model skips instruction reads.
The `**/*` control loads directly. These observations distinguish native
discovery, model-guided selection, file-read permission, and model compliance.
They do not establish deterministic glob enforcement or mandatory instruction
compliance. Other metadata, imports, inherited discovery, and live reload
remain unverified.

The denied edit and the first denied user-instruction probe remain failed
attempts. Their initial checks incorrectly required an edit or later file read
after denial. Current tests check the denied instruction read and absent body;
they do not require the native client to continue after denial.

The retained [project](../WORKBENCH/evidence/native-draft2-debug/copilot-recursive-instructions-project-all-before.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-recursive-instructions-user-all-before.json)
failures show both unconditional bodies at the source, followed by only the
flat body after relocation. The [CLI import cases](../WORKBENCH/evidence/native-draft2-debug/copilot-recursive-instructions-import-before.json)
and [Go failure](../WORKBENCH/evidence/native-draft2-debug/copilot-recursive-instructions-go-before.json)
record the missing nested assets separately.

## Verification

Current native receipts use `copilot-recursive-instructions-*-user-instructions.json`.
Earlier receipts and frozen runners remain historical, including the wrong
automatic-injection expectations. Run repository checks and verify the current
native evidence with:

```sh
python3 WORKBENCH/conformance/verify_plugin_selections.py \
  --repository-only --with-copilot-recursive-instructions \
  --output /absolute/new-verification.json
```

The [combined record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-agent-instructions-final.json)
includes stable and draft conformance, Go race tests and vet, Workbench tests,
compatibility, repository validation, coverage, and the new native verifier.
Other native families retain their historical source hashes and are not
promoted by this record. Full adapter support and the wider milestone remain
incomplete.
