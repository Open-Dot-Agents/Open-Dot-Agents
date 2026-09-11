# Copilot project skill import

Draft.2 import reads complete skill packages from `.github/skills/`,
`.agents/skills/`, and `.claude/skills/` under the supplied project root. It stores them in
`.agents/skills/` and selects the portable `skills` profile. The original
native packages remain unchanged. Stable import behavior is unchanged.

| Source package | Canonical destination |
| --- | --- |
| `.github/skills/<name>/` | `.agents/skills/<name>/` |
| `.agents/skills/<name>/` | Same files, selected in place |
| `.claude/skills/<name>/` | `.agents/skills/<name>/` |

Import keeps definition bytes, contained text and binary assets, and the
executable property of supporting files. New canonical files use private
permissions. Existing shared files keep their permissions and inodes. The native namespace's `import-report.json` records each source
package and its canonical destination. Its `external_fields` disposition
applies to excluded fields, not to the separately listed imported skills.

## Conflict and scope rules

The [importer](../CLI/internal/config/native_copilot_skill_import.go) compares
complete packages. It does not combine two different bundles that happen to
have the same `SKILL.md`. Different file sets, definition bytes, or executable
properties refuse import. Different package folders with the same parsed
native skill name also refuse. Identical complete packages are deduplicated.
These checks include existing canonical packages and cannot be bypassed with
`--force` or `--backup`.

Plan and apply also check for conflicting native packages in these project
locations. This prevents a successful canonical projection claim when another
definition with the same native identity remains present. The adapter does not
assume a native discovery order and does not delete unowned native packages.

An empty canonical `skills/.gitkeep` marker is removed in the same transaction
when import selects real skills. A changed or nonempty marker is not removed.
With `--backup`, its private backup goes to
`.agents/state/import-backups/skills.gitkeep.bak`. A backup inside the selected
skills directory would violate the package structure. The
[retained failure](../WORKBENCH/evidence/native-draft2-debug/copilot-project-skill-marker-backup-before.json)
shows that invalid placement. Native source markers remain unchanged. The [tests](../CLI/internal/config/native_copilot_skill_import_test.go)
check repeated and additive imports, exact assets, package conflicts, symlink
refusal, canonical conflicts, scope boundaries, and the marker's concurrent
change check. Existing transaction tests cover rollback of removals.

This import reads only the three locations under the supplied root. It does not
copy inherited parent skills, user homes, plugin stores, or directories added
through native invocation options. It does not grant trust or execute skill
scripts. Unknown native skill annotations keep the
[metadata reporting and refusal rules](COPILOT_SKILL_METADATA.md).

## Existing shared skill trees

Native Copilot can load `.agents/skills` without an Open-Dot-Agents manifest.
The earlier importer rejected that source. It also failed to select shared
packages in an existing draft.2 tree when there was no other native source.
The retained [Go failure](../WORKBENCH/evidence/native-draft2-debug/copilot-shared-skill-import-before.json)
records both cases. The retained
[native failure](../WORKBENCH/evidence/native-draft2-debug/copilot-project-skill-agents-before.json)
shows successful source discovery and execution, followed by the missing
manifest import error.

Import now selects those packages in place. A bare `.agents` tree can establish
draft.2 only if it contains recognized skill packages and an optional regular
`AGENTS.md`. Other unversioned content refuses. A malformed or older manifest
cannot use this exception. Existing canonical policy, binary assets, modes,
and source inodes remain intact. Force and backup do not bypass conflicts.
The [shared-source tests](../CLI/internal/config/native_shared_skill_import_test.go)
cover repeated import, private metadata, marker backups, malformed manifests,
symlinks, unknown content, duplicate identities, policy conflicts, and scope.

## Native evidence

The [runner](../WORKBENCH/conformance/run_native_copilot_skill_import.py) and
[verifier](../WORKBENCH/conformance/verify_copilot_skill_import.py) use Copilot
`1.0.83` on Linux, with binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.
The official discovery locations are in the captured
[command reference](../WORKBENCH/evidence/native-draft2-debug/copilot-skill-frontmatter-reference.source.txt),
with URLs and hashes in the [source record](../WORKBENCH/evidence/native-draft2-debug/copilot-skill-frontmatter.sources.json).
The [shared-source record](../WORKBENCH/evidence/native-draft2-debug/copilot-shared-skills.sources.json)
retains a new live index and command reference for this extension.

Each origin has a source and relocated native process. A local model endpoint
requests the skill, then its contained Python script. One ACP approval permits
that exact fixture command. The script reads the package's data file and writes
an observable workspace marker. Native skill, tool, approval, and session
events correlate with model requests and the marker. A successful file copy
alone cannot pass the verifier.

The retained [GitHub-origin failure](../WORKBENCH/evidence/native-draft2-debug/copilot-project-skill-github-before.json)
and [Claude-origin failure](../WORKBENCH/evidence/native-draft2-debug/copilot-project-skill-claude-before.json)
both show source execution followed by missing relocated discovery, a native
skill-not-found error, and no file effect. The fixed cases require source,
imported, projected, and reimported asset hashes to match. They also check
unchanged original packages and external `config.json` during adapter writes.
No native Claude test is involved; both cases run Copilot.

This is evidence for three supplied-root discovery locations and contained
assets. It does not establish all inherited discovery, duplicate precedence,
cross-package references, live reload, portable permission enforcement, or
full adapter support. Run the combined verifier with
`--with-copilot-skill-import` to check these records with the repository checks.

Current project-import receipts use the `-user-instructions.json` suffix. They include
six correlated native sessions across all three sources. Verifier tests reject
missing effects, unapproved execution, changed modes, and replaced shared
files. The combined record is `verification-copilot-user-instructions-final.json`.
Earlier receipts retain their original source hashes and remain historical.

```sh
python3 WORKBENCH/conformance/verify_plugin_selections.py \
  --repository-only --with-project-extensions --with-copilot-skill-import \
  --output /absolute/new-verification.json
```

This command checks recorded native evidence against current source hashes;
it does not start new native sessions.
