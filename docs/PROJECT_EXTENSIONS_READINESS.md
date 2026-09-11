# Project extension readiness

The project GitHub package and shared skills are configured. Full production
support across Codex and Copilot is **not established**. Configuration, native
discovery, native execution, and release support remain separate checks.

This review uses Codex `0.154.0` and Copilot `1.0.83` on Linux. It does not change
the [compatibility status](COMPATIBILITY.md), ratify draft.2, or claim arbitrary
plugin conversion.

| Area | Verified result | Remaining limit |
| --- | --- | --- |
| Shared skills | Both clients discover `property-based-testing` and `diagnosing-bugs` from `.agents/skills`; no vendor copies are needed | Discovery does not prove that every skill instruction or native metadata field is followed |
| Copilot shared skill import | Draft.2 selects existing `.agents/skills` packages in place and accepts a bare shared skill tree; native source and relocated execution are verified | Other unversioned canonical content and policy conflicts refuse; inherited parent and user discovery remain separate |
| Skill input validation | Stable, draft.1, and draft.2 projection reject invalid UTF-8 definitions; stable and native user import refuse them before content writes or backups | Supporting assets can be binary; skill text is not rewritten |
| Portable tools | Existing round-trip, ownership, conflict, and refusal tests pass; selected project server startup evidence is retained | Required portable `mcp.envRef` still blocks full root sync; native missing-variable behavior does not meet that contract |
| OpenAI GitHub package | The unchanged pinned package is installed and enabled for this Codex project; all six cached files match the source | Its external HTTP MCP has not been exercised with a real token |
| GitHub app | The already connected app completed a read-only repository lookup | This is separate from activation of the new local package and its HTTP MCP server |
| GitHub bearer reference | Codex sends the named test token and exposes the probe tool after authenticated MCP discovery; without the variable, no MCP request is sent | The probe redirects only the test copy's URL to loopback; it does not test the public GitHub service |
| Copilot package conversion | Native installation and component listing were tested | Copilot ignores the package's bearer-token field, sends an unauthenticated request, and cannot expose the authenticated probe tool; OpenAI app mapping is not established |
| Source integrity | CI checks full commit pins, file hashes, package membership, assets, and symlink/path refusal | Integrity does not replace source review or native behavioral tests |

## Changes from this review

The requested plugin is under
[`.agents/plugins/com.openai.codex`](../.agents/plugins/com.openai.codex/README.md).
Its namespace also contains the local native marketplace, selection file, and
provenance. Its source revision is
`d416fd5a43426019986b1e489506db3db66dee3d`. The package version is `0.1.11`.
The source declares MIT but has no license file at this revision. This limits
the available license evidence; no replacement notice was added.

The native install uses the user's package cache. Its temporary user selection
was checked and removed by restoring the original configuration bytes and mode.
Only the project selection remains enabled. Other native settings are preserved.
The root stable manifest and its mandatory requirements remain unchanged. The
project setup is not represented as a successful portable sync.

Matt Pocock's
[`diagnosing-bugs`](../.agents/skills/diagnosing-bugs/SKILL.md) skill is pinned to
`3cca18b368ae95cdbdebbff572ccafa662551015`. Its MIT license, native display
metadata, and human reproduction template are included. The template is an
upstream skill asset; it is not installed or run as a maintenance command.
The [catalog review](PLUGIN_CATALOG_REVIEW.md) explains the selection.

The readiness audit reproduced invalid UTF-8 acceptance in all three schema
versions and both clients' projection paths. The CLI now checks the actual
definition bytes. Stable import validates skills before it creates backups or
writes canonical content. Native user import checks the same bytes that it
will write. Regression tests retain valid Unicode and binary supporting files.
The Python specification runner checks the same existing UTF-8 requirement.

A further [stable import audit](NATIVE_DEBUG_RESEARCH.md#stable-import-credentials-policy-and-rollback)
fixed credential writes before refusal, silent loss of native MCP controls,
removal of required capabilities during forced import, and incomplete import
rollback. The importer now validates the proposed selected tree first and
includes backups in rollback. Existing requirements and file modes remain
intact. These fixes improve the adapter; they do not change native support
status or clear the external prerequisites below.

The [OTLP transport review](CODEX_OTEL_TRANSPORTS.md) then added native gRPC and
HTTP binary logs, traces, and metrics evidence. It found and fixed a related
native credential-exclusion defect: the complete exporter now stays explicitly
disabled instead of sending unauthenticated telemetry or restoring a default.
This result does not clear the plugin authentication or portable environment
reference limits in the table.

The [runtime authentication audit](NATIVE_AUTHENTICATION.md) found another
loss: filtering could remove an MCP or provider credential while the definition
stayed active. Import and projection now refuse this loss, including optional
forced apply and Copilot LSP definitions. Codex account-authentication selection
now survives relocation. Native tests retain the earlier failed request and
the fixed on/off controls. These results do not establish full production
support across both clients.

## Evidence and reproduction

- [Source-integrity check](../CLI/scripts/check_project_extensions.py) and
  [refusal tests](../CLI/scripts/check_project_extensions_test.py).
- [Skill regression before the fix](../WORKBENCH/evidence/project-tools/skills-encoding-before.json).
- [Native setup](../WORKBENCH/evidence/project-tools/github-project-setup.json)
  and [Codex project discovery](../WORKBENCH/evidence/project-tools/github-codex-project-discovery-summary.json).
- [GitHub connector lookup](../WORKBENCH/evidence/project-tools/github-connector-read.json)
  and [shared skill discovery](../WORKBENCH/evidence/project-tools/diagnosing-bugs-native-discovery.json).
- [Authentication runner](../WORKBENCH/conformance/run_github_plugin_auth.py).
  Each result has a frozen runner, exact native binary hash, source hashes,
  native events, local model requests, and HTTP authentication observations.
- Codex with [token present](../WORKBENCH/evidence/project-tools/github-codex-auth-present-first.json)
  and [token missing](../WORKBENCH/evidence/project-tools/github-codex-auth-missing.json).
- Copilot with [token present](../WORKBENCH/evidence/project-tools/github-copilot-auth-present-first.json)
  and [token missing](../WORKBENCH/evidence/project-tools/github-copilot-auth-missing.json).
  A passed observation check records the native limitation; it is not a passed
  authentication or conversion claim.
- [Initial marketplace attempts](../WORKBENCH/evidence/project-tools/github-initial-marketplace-attempts.json)
  and [upstream validator refusal](../WORKBENCH/evidence/project-tools/github-upstream-validator.json).
- [Official source captures](../WORKBENCH/evidence/project-tools/github-official-sources.json).

Run a new local authentication probe with a new output name:

```sh
python3 WORKBENCH/conformance/run_github_plugin_auth.py \
  --vendor codex --token present --output /absolute/new-result.json
```

Use `--vendor copilot` and `--token missing` for the other cases. The runner uses
isolated homes, a synthetic token, and local services. It does not log in,
copy account credentials, call GitHub, or execute a plugin tool. Successful
authentication is correlated with MCP initialization, tool listing, tool
visibility in a model request, and a completed native turn.

Run repository checks and validate this evidence together:

```sh
python3 WORKBENCH/conformance/verify_plugin_selections.py \
  --repository-only --with-project-extensions --output /absolute/new-verification.json
```

The remaining production gates are safe required environment-reference
semantics, an established Copilot mapping for this package's authentication and
optional apps, and authenticated public-service evidence for external MCP use.
Native trust and account access are external prerequisites. A configuration
write, source pin, or successful install cannot clear these gates.

The [project-scope audit](CODEX_PROJECT_SCOPE.md) also fixes false activation
claims for settings that Codex ignores in project files. Required profiles
refuse them; optional source stays preserved and inactive. The
[role override audit](CODEX_ROLE_OVERRIDES.md) now has native evidence for the
bounded child contract. The [reference audit](CODEX_ROLE_REFERENCES.md) also
corrects role-file and skill-selector changes after relocation. These cases
do not clear the remaining production gates above.

The [Copilot metadata audit](COPILOT_SKILL_METADATA.md) fixes a false activation
claim: draft.2 now refuses skill definitions that fail its native discovery
checks and reports per-skill invocation controls and losses. Stable Markdown
rules remain unchanged. The tested wildcard tool allowance and custom hint do
not have their documented effects. Shared discovery does not establish complete
metadata portability, and these mappings do not clear the production gates.

The [Copilot project import audit](COPILOT_SKILL_IMPORT.md) also fixes lost
packages from `.github/skills/` and `.claude/skills/`. The native tests confirm
skill loading and contained script execution after relocation. Complete-package
conflicts refuse before writes, and private marker backups stay outside skill
discovery. The combined verification passes 39 checks. These results leave
the production gates above open.
