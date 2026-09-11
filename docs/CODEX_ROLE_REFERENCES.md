# Codex role and skill references

Draft.2 import now preserves the target of relative references when native
configuration moves. This corrects two observed failures in Codex `0.154.0`:

- An imported `agents.<name>.config_file` pointed at a missing file in the new
  native home. The CLI reported success, but no child agent ran.
- A relative skill selector inside a copied agent file pointed at a different
  location. The child listed a skill that the source role disabled.

## Import rules

Codex resolves a declared role file against the containing native configuration
directory. Project configuration uses `.codex/`; user configuration uses the
explicit native home. Skill selectors inside a role file use that file's
directory. These origins differ from the canonical artifact directory.

Import handles each reference as follows:

| Reference | Canonical result |
| --- | --- |
| Mapped standalone file in the native `agents/` directory | Relative `agents/<file>.toml`, including when the original reference was absolute |
| Role file outside the registered asset directory | Absolute reference to the original file; the external library is not copied |
| Declared role file without required standalone metadata | Absolute reference to the original file; its preserved canonical asset remains inactive |
| Role selector for a copied user skill or canonical project skill | Relative path from the native agent file to that skill |
| Role selector for an external skill | Absolute reference to the original skill definition |
| HOME-relative reference | Preserved for native resolution; a changed native HOME can change its meaning |

Invalid skill path selectors remain invalid. Import does not remove a trailing
slash or add `SKILL.md` to make an ignored selector active. Unknown and malformed
fields still use the existing validation and refusal rules.

The importer reads registered assets only. It does not traverse an external
role or skill library, copy credentials, or grant runtime authority. External
files must remain available to the native client. Plans identify that action.
Absolute references preserve a local source location; they do not make an
external library portable to another machine.

## Native evidence

Linux Codex `0.154.0` is pinned to SHA256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
The [pinned loader source](../WORKBENCH/evidence/native-draft2-debug/codex-agent-role-scope.sources.json)
is revision `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`.

The native matrix has five cases in each scope: external relative reference,
mapped standalone asset, external absolute reference, declared-only file, and
an external skill selector inside a mapped agent file. Each case imports,
applies to a different native location, runs Codex there, and imports again.
The skill cases move the destination to a different directory depth so a
broken relative path cannot succeed by coincidence.

Ten cases contain 20 native processes and 40 completed parent or child turns.
A unique prompt, completed spawn, child model and effort, completed child turn,
and completed parent turn are correlated. The skill cases also require the
parent to list the fixture skill and the child to omit it. Source configuration
and role files stay unchanged. No public model or account credential is used.

The [verifier](../WORKBENCH/conformance/verify_codex_role_references.py) checks
the `codex-role-reference-<case>-<scope>-checked.json` records, frozen runners,
current Go source hashes, round-trip values, and retained failures. The
[missing-role failure](../WORKBENCH/evidence/native-draft2-debug/codex-role-reference-external-before.json)
and [changed-selector failure](../WORKBENCH/evidence/native-draft2-debug/codex-agent-skill-reference-user-before.json)
remain available with native requests and events.

```sh
python3 WORKBENCH/conformance/verify_codex_role_references.py
```

These cases establish local reference preservation and child configuration.
They do not establish all filesystem layouts, changed HOME behavior, external
library availability on another machine, or security enforcement. Plugin
package contents and installation remain separate native operations. Full
production readiness remains unproven.
