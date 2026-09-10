# Release Process

## Candidate artifacts

The release-candidate workflow creates platform-specific archives, SHA-256
checksum files, SPDX JSON SBOMs, and GitHub provenance attestations, verifies
them, and attaches the files to a draft GitHub Release.
Candidates are suitable for evaluation only and must not be described as final
releases or as evidence of harness support until a maintainer publishes the
draft. Installation and checksum instructions are in [INSTALL.md](INSTALL.md).

The workflow requires a real pushed `v<version>` tag and a checked-in
`releases/v<version>.md` release-notes file. The tag must point at the
same root commit checked out by the workflow. Manual dispatch is allowed only
to rebuild a draft for an existing tag.

## Release inputs

Specification and CLI releases are independently versioned. Before publishing a
release, maintainers must review the changelog, compatibility matrix, security
issues, and migration impact. A release may not claim support for a harness or
profile without current conformance evidence.

Run `python3 CLI/scripts/check_compatibility.py` before any release. That check
keeps `CLI/compatibility.json`, the generated Markdown matrix, Reference CLI
capability summaries, and support-evidence rules aligned.

## Specification release checklist

1. Confirm normative text, schemas, examples, and fixtures agree.
2. Run schema and conformance validation.
3. Classify compatibility according to [VERSIONING.md](VERSIONING.md).
4. Run `python3 CLI/scripts/check_compatibility.py`.
5. Publish release notes describing additions, clarifications, migrations, and
   deprecations.
6. Update the compatibility matrix with the exact specification version.

## Reference CLI release checklist

1. Run unit, integration, and conformance tests on every supported platform.
2. Verify every adapter's declared capabilities against the pinned native
   harness version. If there is no native black-box evidence, leave the
   adapter marked not conformance-supported.
3. Run `python3 CLI/scripts/check_compatibility.py`.
4. Verify `agents version` reports the intended release version in an
   extracted artifact.
5. Create final release assets and publish their checksums with installation,
   upgrade, and rollback guidance.
6. Verify every SBOM and GitHub artifact attestation with the published
   installation commands.
7. Update the compatibility matrix and CLI changelog before release.

## Publishing v1.0.0

1. Confirm the credentialed native adapter workflow passed for all three pinned
   harnesses and the compatibility registry links its attested results.
2. Tag and push `v1.0.0` in `Agents-Spec` and `Agents-CLI`; verify immutable
   schema URLs and versioned `go install` from clean environments.
3. Confirm root, `SPEC`, `CLI`, and `WORKBENCH` commits are pushed and the
   root submodule pointers reference those pushed commits.
4. Confirm the `Verify` and `Security` workflows are green on the intended root commit.
5. Create and push the root release tag:

   ```sh
   git tag -a v1.0.0 -m "Open-Dot-Agents v1.0.0"
   git push origin v1.0.0
   ```

6. Confirm the `Release candidate` workflow for `v1.0.0` is green and created
   or updated the draft release.
7. Review the draft release against `releases/v1.0.0.md`, asset list,
   checksum files, compatibility matrix, and security policy.
8. Publish the draft from GitHub only after the review is complete.

## Incident releases

Security fixes and adapter regressions receive the smallest compatible release
practical. If a harness change invalidates a compatibility claim, mark that
entry unsupported or affected immediately, publish a mitigation, and restore
support only after a verified conformance run.

## Enforced core release checks

The release workflow now runs specification validation, canonical repository
validation, every deterministic Workbench test, and
`python3 CLI/scripts/check_compatibility.py --require-supported` before the
native jobs and artifact builds. A missing Workbench checkout is a failure.
All three registry rows must be `conformance-supported`; each needs a public
HTTPS `evidence_url` and an `evidence_sha256` for its versioned evidence bundle.
These identifiers supplement the existing capability and exact-version checks.
They do not replace verification of the linked artifact or its attestation.

The native workflow runs baseline and extended tests and checks their agreement
with `WORKBENCH/conformance/release_gate.py`. Release verification rejects failed
or incomplete cases, mixed CLI binaries, changed runner sources, mismatched
source commits, and evidence from uncommitted source changes. A preflight result
cannot replace a native result. Expected refusals are recorded separately;
they do not satisfy the registry's full-support requirement.

Each native workflow artifact contains a tar archive and its SHA-256 file.
Passing archives receive GitHub provenance attestations. The draft release
includes the archives so evidence can remain available beyond the workflow
artifact retention period. Publishing that draft remains a maintainer action.
Until reviewed archives are public, do not replace local evidence links with
invented release URLs or mark an adapter supported.

Verify downloaded evidence from a trusted project workflow:

```sh
sha256sum -c adapter-evidence-codex.tar.gz.sha256
gh attestation verify adapter-evidence-codex.tar.gz \
  --repo Open-Dot-Agents/Open-Dot-Agents
```

Extract the three verified archives in a clean checkout of their recorded
source commits, install the pinned dependencies, then run:

```sh
python3 WORKBENCH/conformance/release_gate.py --evidence-dir evidence --release
```

The current registry intentionally fails `--require-supported`. The refusal
fixes do not remove that release blocker. No specification semantics changed
in this implementation cycle; a future policy change still requires the
public proposal and decision process.

### Existing local tag

The core review found an existing local root `v1.0.0` tag at
`422c8596cfc684720e5c3214048de1568d4f896b`. Do not move or recreate it.
Before any future publication, verify remote release state and select an unused
root release identifier through the existing release process. This review did
not inspect or change remote tags. The current CLI/Spec version remains 1.0.0;
root release identifiers and component versions are independent.
