# Codex provider token commands

Draft.2 can preserve Codex `model_providers.<id>.auth` as native configuration.
The CLI writes the command, arguments, working directory, timeout, and refresh
interval. It does not run the command, copy its executable, or store its output.
The executable and returned token stay external.

**This setting does not enforce authentication.** Codex `0.154.0` sends an
unauthenticated model request when the token command fails. Tests observed
this after timeout, empty output, nonzero exit, invalid UTF-8 output, and a
missing executable. The same behavior occurs before and after import/apply.
The adapter preserves this native setting; it does not change the native
failure behavior or claim a portable authentication guarantee.

## Configuration and validation

The earlier field-name classifier treated the whole `auth` table as credential
material. A native baseline ran the command and sent an authenticated request,
but the CLI then refused to import the valid configuration. The fix uses the
pinned `ModelProviderAuthInfo` schema to validate the complete table before
field selection. The exact `auth` field is configuration. Explicit credential
fields inside it remain excluded.

The adapter refuses incomplete, malformed, or unknown token-command fields,
even for optional profiles and forced writes. It also refuses an empty command
and the combinations that the pinned native source rejects: `env_key`, an
inline bearer token, `requires_openai_auth=true`, or AWS authentication. An
explicit `requires_openai_auth=false` is compatible. No configuration or backup
write occurs after refusal. Required portable policy still takes precedence.

The import/apply plan and capabilities output identify the native fallback
behavior. Install and check the token command separately. Start a new native
session to use changed configuration. This review covers user scope. It does
not establish project-scope behavior, public-provider behavior, arbitrary
credential helpers, background descendants, or full adapter support.

## Native evidence

Tests pin Codex `0.154.0` on Linux with SHA-256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
The model endpoint and token command are local fixtures. There are no account
files. The token command records its working directory and literal arguments
in a separate file. Model-request records contain a token-match Boolean and a
revision number, not the token value. Native completion events correlate with
unique prompt markers in the model requests.

| Case | Before and after import/apply |
| --- | --- |
| [Successful command](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-success-verified.json) | The request uses the returned token. The relative script path, working directory, and literal shell characters in arguments remain intact. |
| [Caching](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-cache-verified.json) | With refresh interval `0`, the second turn in the same native process does not rerun the command. |
| [Timed refresh](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-refresh-verified.json) | With a 50 ms interval, a second turn after 150 ms runs the command again. |
| [401 retry](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-retry-verified.json) | A 401 response causes another command invocation. The next model request uses the new token revision. |
| [Timeout](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-timeout-verified.json) | A 100 ms timeout logs a native error. The model request still runs without authentication. |
| [Empty output](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-empty-verified.json) | Native error, followed by an unauthenticated request. |
| [Nonzero exit](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-exit-verified.json) | Exit status 7 is reported, followed by an unauthenticated request. |
| [Invalid UTF-8](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-invalid-utf8-verified.json) | Native output-decoding error, followed by an unauthenticated request. |
| [Missing executable](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-missing-verified.json) | No helper invocation. Native launch error, followed by an unauthenticated request. |

The nine cases cover 18 native processes and 22 completed turns. Five cases
are verified observations of a native limitation. They are not authentication
acceptance results. Import/apply and reimport leave the external helper and
source configuration unchanged and do not create account files.

The [initial Go failure](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-before.json)
and [native import refusal](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth-native-before.json)
remain in the repository. The first five failure tests also remain as
`codex-command-auth-<case>-first.json`: they failed the expectation that a token
error would prevent a model request. Later observation runs use an explicit
`--observe-failure-fallback` option. They do not replace those failed tests.
The observation runs also return a valid empty model catalog, to separate
token-command failure from catalog lookup errors in the first fixture.

## Sources and reproduction

The [source receipt](../WORKBENCH/evidence/native-draft2-debug/codex-command-auth.sources.json)
pins revision `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`. It includes the
configuration types, provider authentication selection, command execution,
and account-manager error handling. The latter returns no authentication after
a transient token-command error. The
[official reference](../WORKBENCH/evidence/native-draft2-debug/codex-provider-auth-reference.source.txt)
documents the fields and defaults.

The [runner](../WORKBENCH/conformance/run_native_command_auth.py) preserves a
frozen runner and token-helper source beside each new result. The
[verifier](../WORKBENCH/conformance/verify_native_command_auth.py) checks native
pins, current Go hashes, field preservation, native events, request effects,
the fallback limit, and the retained failures. The
[Go tests](../CLI/internal/config/native_provider_command_auth_test.go) cover
round-trip preservation, no command execution by the adapter, malformed units,
conflicts, excluded credential fields, and optional forced apply.

```sh
python3 WORKBENCH/conformance/run_native_command_auth.py --case success --output /absolute/new-success.json
python3 WORKBENCH/conformance/run_native_command_auth.py --case timeout --observe-failure-fallback --output /absolute/new-observation.json
python3 WORKBENCH/conformance/verify_native_command_auth.py
```

For provider account selection, see [native authentication](NATIVE_AUTHENTICATION.md).
The [compatibility status](COMPATIBILITY.md) and release gates are unchanged.

The project-scope correction was followed by a repeat run against the current
Go source. The verifier selects `*-scope-current.json` by default. Earlier
records remain unchanged. See the [project-scope report](CODEX_PROJECT_SCOPE.md).
