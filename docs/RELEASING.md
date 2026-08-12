# Release Process

## Candidate artifacts

The release-candidate workflow creates unsigned, platform-specific archives
and SHA-256 checksum files, verifies them, and attaches them to a draft GitHub
Release. It does not publish packages, signatures, SBOMs, or provenance.
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

Run `python3 scripts/check_compatibility.py` before any release. That check
keeps `compatibility.json`, the generated Markdown matrix, Reference CLI
capability summaries, and support-evidence rules aligned.

## Specification release checklist

1. Confirm normative text, schemas, examples, and fixtures agree.
2. Run schema and conformance validation.
3. Classify compatibility according to [VERSIONING.md](VERSIONING.md).
4. Run `python3 scripts/check_compatibility.py`.
5. Publish release notes describing additions, clarifications, migrations, and
   deprecations.
6. Update the compatibility matrix with the exact specification version.

## Reference CLI release checklist

1. Run unit, integration, and conformance tests on every supported platform.
2. Verify every adapter's declared capabilities against the pinned native
   harness version. If there is no native black-box evidence, leave the
   adapter marked not conformance-supported.
3. Run `python3 scripts/check_compatibility.py`.
4. Verify `agents version` reports the intended release version in an
   extracted artifact.
5. Create final release assets and publish their checksums with installation,
   upgrade, and rollback guidance.
6. If artifact signing, SBOMs, or provenance are introduced, publish the
   corresponding verification instructions with those assets; do not claim
   them before then.
7. Update the compatibility matrix and CLI changelog before release.

## Publishing v1.0.0

1. Confirm root, `SPEC`, `CLI`, and `WORKBENCH` commits are pushed and the
   root submodule pointers reference those pushed commits.
2. Confirm the `Verify` workflow is green on the intended root commit.
3. Create and push the root release tag:

   ```sh
   git tag -a v1.0.0 -m "Open-Dot-Agents v1.0.0"
   git push origin v1.0.0
   ```

4. Confirm the `Release candidate` workflow for `v1.0.0` is green and created
   or updated the draft release.
5. Review the draft release against `releases/v1.0.0.md`, asset list,
   checksum files, compatibility matrix, and security policy.
6. Publish the draft from GitHub only after the review is complete.

## Incident releases

Security fixes and adapter regressions receive the smallest compatible release
practical. If a harness change invalidates a compatibility claim, mark that
entry unsupported or affected immediately, publish a mitigation, and restore
support only after a verified conformance run.
