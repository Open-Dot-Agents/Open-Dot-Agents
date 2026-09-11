# Codex project configuration scope

Draft.2 now refuses required Codex project settings that the pinned native
loader ignores. Optional profiles keep those fields in canonical source but
do not project or claim ownership of them. User configuration stays unchanged.

The earlier adapter wrote project provider settings and reported success.
Codex loaded the project model setting but kept the provider from user
configuration. A completed native model request went to the user endpoint.
Successful parsing and a configuration write did not establish activation.

## Verified restrictions

The pinned loader removes 12 root keys from project configuration, including
trusted projects. It also removes one nested feature flag unconditionally.

| Group | Ignored project fields |
| --- | --- |
| Provider selection and definitions | `model_provider`, `model_providers` |
| Provider endpoints | `openai_base_url`, `chatgpt_base_url` |
| Host request metadata | `apps_mcp_product_sku`, `responses_api_metadata` |
| Notifications | `notify` |
| Configuration profiles | `profile`, `profiles` |
| Realtime endpoints | `experimental_realtime_webrtc_call_base_url`, `experimental_realtime_ws_base_url` |
| Telemetry | `otel` |
| System proxy preference | `features.respect_system_proxy` |

The CLI applies this scope rule before field selection. Required profiles
refuse before any configuration, ownership, or backup write. Optional profiles
project the remaining supported fields and retain the ignored content in the
canonical artifact. Plans and capabilities identify the ignored scope and
link the native evidence. Setting declarations no longer advertise these
fields as available in project scope.

Unchanged settings owned by an earlier projection can be removed when the
profile becomes optional. Unowned settings and file permissions stay intact.
Modified owned values still cause a conflict, including with `--force`.
This change does not alter draft.2 ownership rules or authorize a user-scope
write. User scope still requires an explicit absolute `--native-home`.

## Evidence

Tests use Codex `0.154.0` on Linux with SHA-256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
The fixture creates an isolated native home, a trusted project, and two local
model endpoints. Trust is test setup in the isolated native home. Apply does
not create or change trust entries.

The [provider case](../WORKBENCH/evidence/native-draft2-debug/codex-project-provider-verified.json)
tests provider selection and definitions. The
[complete unconditional-key case](../WORKBENCH/evidence/native-draft2-debug/codex-project-allkeys-verified.json)
tests the table above. Each case has two completed native turns: direct native
configuration, then optional adapter projection. Effective configuration
retains user values. The changed project model proves that the trusted project
was loaded. The observed model request uses that project model and the user
provider endpoint. The user configuration remains byte-for-byte unchanged.

Both cases also test required-profile refusal with `--force --backup`.
File hashes and permissions stay unchanged on refusal. Optional projection
writes only the supported project model while retaining the full canonical
source. The [Go tests](../CLI/internal/config/native_codex_project_scope_test.go)
check each field, source preservation, declarations, ownership cleanup,
unowned settings, file permissions, and refusal of modified owned settings.

The [earlier native failure](../WORKBENCH/evidence/native-draft2-debug/codex-project-provider-before.json)
and its [retained projected configuration](../WORKBENCH/evidence/native-draft2-debug/codex-project-provider-before-projection.json)
show that an optional profile wrote ignored provider settings. An early
follow-up fixture incorrectly assumed that import creates a required profile.
Its failed result remains as `codex-project-allkeys-first.json`. The corrected
fixture explicitly sets `required: true` before testing refusal, then sets it
to `false` before testing inactive optional content.

The [source receipt](../WORKBENCH/evidence/native-draft2-debug/codex-provider-scope.sources.json)
pins revision `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`. It includes the
official documentation index, configuration reference, and native loader.
The [verifier](../WORKBENCH/conformance/verify_codex_project_scope.py) checks
that the native fixture covers the exact root-key list in that source. It
also checks source hashes, native events, model requests, and the retained
failure. The [runner](../WORKBENCH/conformance/run_native_codex_project_scope.py)
stores a frozen copy beside each new result.

## Limits and reproduction

The native loader has more restrictions that depend on credential-broker
state. Those conditions were not tested here. Agent-role overrides use a
separate bounded mechanism and need separate evidence. This project-scope
result does not establish user-scope support for every field in the table,
all instruction or configuration precedence, or full adapter support.

The coverage map records project restrictions separately from user mappings.
The original 1,679 evidence entries and semantic identifiers stay intact.
`features.respect_system_proxy` is absent from that frozen inventory; its
scope evidence is in the native artifact declarations rather than a new
invented source entry. The map remains incomplete and its counts are not
completion claims.

```sh
python3 WORKBENCH/conformance/run_native_codex_project_scope.py --all-keys --output /absolute/new-scope-result.json
python3 WORKBENCH/conformance/verify_codex_project_scope.py
```

The [compatibility status](COMPATIBILITY.md) and release gates are unchanged.
