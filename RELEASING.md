# Release Process

## Candidate artifacts

The release-candidate workflow creates unsigned, platform-specific archives
and SHA-256 checksum files as GitHub Actions workflow artifacts. It does not
create GitHub Release assets or publish packages, signatures, SBOMs, or
provenance. Candidates are suitable for evaluation only and must not be
described as final releases or as evidence of harness support. Installation
and checksum instructions are in [INSTALL.md](INSTALL.md).

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
   harness version.
3. Run `python3 scripts/check_compatibility.py`.
4. Verify `agents version` reports the intended release version in an
   extracted artifact.
5. Create final release assets and publish their checksums with installation,
   upgrade, and rollback guidance.
6. If artifact signing, SBOMs, or provenance are introduced, publish the
   corresponding verification instructions with those assets; do not claim
   them before then.
7. Update the compatibility matrix and CLI changelog before release.

## Incident releases

Security fixes and adapter regressions receive the smallest compatible release
practical. If a harness change invalidates a compatibility claim, mark that
entry unsupported or affected immediately, publish a mitigation, and restore
support only after a verified conformance run.
