# Copilot canonical instruction binding

Draft.2 preserves one canonical instruction source when a project also has a
separate Copilot instruction body. A verified root `AGENTS.md` link to
`.agents/AGENTS.md` establishes this native project binding:

```json
{"kind": "canonical-instructions", "source": "AGENTS.md"}
```

Only the `com.github.copilot` native project profile maps this artifact.
Its source must be exactly `AGENTS.md`, and `name` must be absent. This
registered source refers to the portable `.agents/AGENTS.md`. It is the fixed
exception to namespace-relative artifact sources. Metadata cannot select a
different core source or output path. See the
[example](../SPEC/examples/canonical-instructions/.agents/manifest.json).

The destination is root `AGENTS.md`. Apply preserves an existing verified
canonical link without file ownership. After relocation to a project without
that link, apply creates a managed regular root file. Later canonical edits
update this output. A separate native `instructions` artifact can own
`.github/copilot-instructions.md`. Its bytes and namespace source remain
independent, even if its source also has the basename `AGENTS.md`.

Import preserves both bodies and existing custom native source paths. Reimport
of the relocated regular root retains the binding. Conflicting runtime edits
cannot replace canonical policy through `--force`. Duplicate core bindings or
assignments to root `AGENTS.md` refuse before writes. An optional selected but
unmapped binding also blocks an unsafe fallback to the ordinary Copilot path.

Removing a binding refuses a known change to the reference base. A
reference-free transition can remove the owned root copy and restore the
ordinary Copilot destination. Same-repository file ownership can be released
when a verified canonical link replaces the managed root file. Foreign
ownership remains protected, including with force or adoption.

## Native evidence

The [source receipt](../WORKBENCH/evidence/native-draft2-debug/copilot-canonical-instructions.sources.json)
retains the live GitHub documentation index and custom instruction guide.
The guide describes relative references and session restart for changes. It
does not establish reference behavior through the canonical link.

The native tests use Copilot `1.0.83` on Linux, binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`,
isolated homes, preconfigured fixture trust, and a local model endpoint.
Two cases cover a canonical link alone and a canonical link beside a distinct
Copilot body. Each case runs three fresh native sessions:

| Phase | Observed result |
| --- | --- |
| Source link | Canonical body and root-relative policy load; a distinct native body and its own policy also load when present |
| Relocated regular root | The same bodies and references load after import and apply |
| Updated canonical core | The new core marker loads; the separate native body remains unchanged |

References through the source link resolve from the project root. Decoy
policies under `.agents` and the wrong `.github` reference base do not load.
Referenced project files remain external dependencies. Apply does not copy
them or grant trust. Source apply with `--adopt` preserves the link inode.

The [native runner](../WORKBENCH/conformance/run_native_copilot_canonical_instructions.py)
and [verifier](../WORKBENCH/conformance/verify_copilot_canonical_instructions.py)
check six correlated sessions, model requests, native file-read events and
results, completed turns, both import roundtrips, source and authority
preservation, and current Go hashes. The fixture model reads a known file.
These checks do not prove compliance with arbitrary instruction text,
arbitrary link discovery, other nested scopes, or live reload.

Current receipts are `copilot-canonical-instructions-link-only-user-instructions.json`
and `copilot-canonical-instructions-link-distinct-user-instructions.json`. The earlier
`copilot-canonical-instructions-before.json` retains successful source loading
followed by the old distinct-body import refusal. The failed Go baseline and
the first ownership test failure are also retained. The latter was a test
error: decoding into a reused map retained a removed key. The corrected test
starts from an empty registry.

## Verification and limits

The [Go tests](../CLI/internal/config/native_canonical_instructions_test.go)
cover import, relocation, edits, ownership, malformed metadata, scope,
duplicate targets, same-basename sources, removal, and rollback after the final
transaction write. Schema and example checks apply only to draft.2 behind
`--experimental`. Stable 1.0 and draft.1 semantics remain unchanged.

The combined record is
`verification-copilot-user-instructions-final.json`. Root, agent-file,
and recursive regression receipts use the `-user-instructions.json` suffix. Earlier
records retain their original source hashes. This result does not promote
adapter support or complete the Linux milestone.
