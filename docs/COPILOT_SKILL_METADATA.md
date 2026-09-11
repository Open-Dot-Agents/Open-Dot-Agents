# Copilot skill metadata

This audit tests Copilot `1.0.83` on Linux with binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.
The [runner](../WORKBENCH/conformance/run_native_copilot_skill_metadata.py)
uses isolated native homes, project and user skill trees, and a local model
endpoint. It runs the public CLI plan and apply commands before native discovery
and invocation. User-scope cases also check import and reimport. It does not use
account credentials or install remote packages.

The [verifier](../WORKBENCH/conformance/verify_copilot_skill_metadata.py)
checks 28 cases. It requires native session and tool events, model requests,
matching approval identifiers, and file effects. A completed probe alone is
not a passed behavior check. This evidence does not establish full adapter
support or complete the native milestone.

## Sources and scope

The official command reference documents six skill frontmatter fields:
`name`, `description`, `argument-hint`, `allowed-tools`, `user-invocable`, and
`disable-model-invocation`. The [source record](../WORKBENCH/evidence/native-draft2-debug/copilot-skill-frontmatter.sources.json)
contains the reference URL, documentation index, captured files, and hashes.

Project apply keeps skills in `.agents/skills`. User apply copies valid skills to
the explicit native home's `skills` directory. The tests check definition hashes
and unchanged external `config.json` during apply. Native test setup provides
trust only for the isolated workspace; apply does not grant trust.

## Visibility and invocation

| Controls | Model catalog | ACP command list | Model skill call | Explicit ACP slash call |
| --- | --- | --- | --- | --- |
| Defaults | Present | Present | Body loaded | Body loaded |
| `user-invocable: false` | Present | Absent | Body loaded | Body not loaded |
| `disable-model-invocation: true` | Absent | Present | Native skill-not-found error | Body loaded |
| Both controls | Absent | Absent | Native skill-not-found error | Body not loaded |

These results apply to both scopes. Native `skill list --json` still lists all
six valid fixture skills as enabled. Thus an enabled discovery row does not
mean that a skill is available for every invocation method.

The model catalog can be in messages or in the `skill` tool description.
The verifier checks both locations. Initial matrix files ending in
`-verified.json` retain the false-negative test results from checking only
messages. The corrected observation matrix ends in `-checked.json`. Current
adapter-reporting cases end in `-reported-checked.json`.

## Argument hints and tool permissions

The fixture sets `argument-hint: "[fixture-arg]"`. ACP returns the generic
`instructions for the skill` hint. The tested terminal completion menu shows
the skill description but omits the custom hint, including after Tab selection.
This does not establish behavior in every picker, IDE, or other client version.

The fixture skill with `allowed-tools: ["*"]` loads successfully. Its native
`skill.invoked` event records `allowedTools: ["*"]`. Nevertheless:

- ACP requests approval for the fixture shell command. A denial produces a
  correlated rejected tool event and no file effect.
- Interactive slash invocation displays command approval. Escape cancels the
  turn and leaves no file effect. This is cancellation, not completed denial.
- Selecting one approval in the same terminal fixture produces a successful
  tool event and the expected file. The file does not exist before approval.
- Unattended `-p` model invocation fails because permission cannot be requested.

These tests use the shell built-in `printf` to write one file inside the
workspace. An earlier Python test requested directory access to the resolved
Python executable. That separate permission check cannot establish shell-tool
approval behavior. Both attempts are retained. Earlier `-p` slash-text probes
did not expand the skill and are not user-activation evidence.

The documented wildcard automatic allowance is not observed for this bounded
command and configuration. This finding does not prove that every tool pattern
is ignored. It does not establish mandatory portable `ask`, permission timeout,
disconnect, shell composition, or background-descendant behavior.

## Malformed and nonstandard metadata

The earlier adapter preserved every fixture definition byte during apply. Native
discovery accepts missing `name`, missing `description`, an underscore in the
name, and an unknown field. Missing values are derived from the directory or
body. Native discovery omits malformed YAML, quoted Boolean controls, and plain
Markdown without frontmatter.

Stable 1.0 requires UTF-8 Markdown and safe contained assets. It does not require
YAML frontmatter. Stable validation and projection keep their existing rules.

Draft.2 now refuses projection before writes for malformed or missing
frontmatter, duplicate fields, non-object frontmatter, and unverified field
types. It does not coerce quoted Boolean strings. Canonical import still
preserves the source; import does not establish activation. Current malformed
native probes run from a separate fixture setup after the verified adapter
refusal. That setup copy is not adapter projection evidence.

The [adapter](../CLI/internal/config/native_copilot_skill_metadata.go) reports
each selected skill and its explicit model and user invocation controls. It
reports known hint and tool-allowance limits in JSON and text output for plan,
apply, and sync. Unknown Markdown annotations retain their bytes and receive
an unmapped-annotation warning. Values are not copied into these warnings.
This is distinct from selection of settings in a native namespace profile.

The [regression tests](../CLI/internal/config/native_copilot_skill_metadata_test.go)
verify refusal, exact byte preservation, user-file permissions, scoped
destinations, type checks, redacted reports, and unchanged stable acceptance.
The [retained failing test](../WORKBENCH/evidence/native-draft2-debug/copilot-skill-activation-refusal-before.json)
shows that the earlier plan offered ignored native skills as applicable in
both scopes. The six coverage records now link to the bounded field mapping.
Their native limitations remain explicit. The 1,679 source rows are unchanged.

## Reuse of online skills

The [catalog review](PLUGIN_CATALOG_REVIEW.md#ai-hero-and-matt-pocock-skills)
records the AI Hero and Matt Pocock sources. The selected `diagnosing-bugs`
skill stays in the shared `.agents/skills` directory. Native discovery of that
skill does not prove that every optional metadata field or workflow from the
upstream collection has matching behavior in both clients.
