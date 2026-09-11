# Native authentication preservation

The draft.2 importer and projector refuse a runtime definition when credential
exclusion would remove its authentication. This check covers Codex MCP servers
and model providers, and Copilot MCP and LSP servers. It also runs before import
of standalone Codex agent TOML. Optional profiles, `--force`, and `--backup` do
not bypass the check. Refusal occurs before configuration and backup writes.

These definitions have no common disabled value that prevents native fallback
or connection attempts. The adapter therefore refuses the operation. It does
not remove a credential and leave the server or provider configured. Supported
environment references remain configuration; native account files remain
external. The check also refuses malformed authentication controls and
credential-bearing URLs. It reports field names without credential values.

Codex `model_providers.<id>.requires_openai_auth` is a Boolean configuration
setting. It selects native account authentication. Both `true` and `false`
remain intact through import, apply, and reimport. The earlier classifier
incorrectly excluded this field because its name contains `auth`.

## Evidence

Tests use Codex `0.154.0` on Linux, with binary SHA-256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
The model server runs on loopback. Both isolated native homes contain the same
synthetic account file before the test starts. The adapter does not create,
import, change, or remove that file.

| Case | Observed result |
| --- | --- |
| [Earlier CLI](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth-baseline.json) | The source request had authentication. After import and apply, the request had none. Both native turns completed. The test failed. |
| [Authentication on](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth-on-verified.json) | Source and relocated requests had the expected header. The Boolean survived reimport. Both account files stayed unchanged. |
| [Authentication off](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth-off-verified.json) | Source and relocated requests had no account authentication header. The Boolean survived reimport. Both account files stayed unchanged. |

The [runner](../WORKBENCH/conformance/run_native_provider_auth.py) records native
completion events and model requests with a unique prompt marker. It records
only the authentication match result, not the header value. Each result has a
frozen runner and source hashes. The [verifier](../WORKBENCH/conformance/verify_native_provider_auth.py)
checks the current source, native pin, observations, and retained failure.

The [official source receipt](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth.sources.json)
pins Codex source revision `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`.
The [provider definition](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth-source.source.txt)
and [configuration reference](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth-reference.source.txt)
describe the Boolean and its default of `false`.

The [Go regression tests](../CLI/internal/config/native_auth_units_test.go)
cover credential and malformed-control refusals in project and user scope,
unchanged files and backups, optional forced apply, supported references,
context-sensitive LSP initialization data, and standalone agent import.
The [initial failed tests](../WORKBENCH/evidence/native-draft2-debug/native-auth-units-before.json)
retain the authentication-loss regressions. Early atomicity-test attempts used
an incomplete canonical fixture and failed before the authentication check;
the corrected fixture uses a valid draft.2 tree.

## Limits and reproduction

These tests do not establish login behavior, missing-credential enforcement,
remote-provider behavior, or full adapter support. They do not change portable
security requirements or the [compatibility status](COMPATIBILITY.md).
The [telemetry rule](CODEX_OTEL_TRANSPORTS.md) still uses an explicit disabled
exporter value, because that native format has a verified disabled form.

Use a new output name for each run:

```sh
python3 WORKBENCH/conformance/run_native_provider_auth.py --auth on --output /absolute/new-result.json
python3 WORKBENCH/conformance/run_native_provider_auth.py --auth off --output /absolute/new-control.json
python3 WORKBENCH/conformance/verify_native_provider_auth.py
```

Codex also provides [command-based provider authentication](CODEX_COMMAND_AUTHENTICATION.md).
Its `auth` table is configuration, while the helper and its returned token stay
external. The adapter validates the complete table and rejects conflicting
authentication modes. Native helper failure can still send an unauthenticated
request; the new evidence records that limitation before and after projection.
The account-selector verifier now selects `*-command-current.json` records
from a repeat run against the current source. The earlier records remain intact.

The project-scope correction was followed by a repeat run against the current
Go source. The verifier selects `*-scope-current.json` by default. Earlier
records remain unchanged. See the [project-scope report](CODEX_PROJECT_SCOPE.md).
