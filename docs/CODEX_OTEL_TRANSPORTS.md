# Codex telemetry transports and credential exclusion

This Linux review uses Codex `0.154.0`, binary SHA-256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
It adds eight bounded native cases with 16 completed native sessions. The
adapter support status and release gates do not change.

| Case | Source configuration | After import, apply, and reimport |
| --- | --- | --- |
| OTLP gRPC | Logs, traces, and metrics delivered | All three signals delivered |
| gRPC with CA | All signals reach the local TLS collector | Same CA reference works from the new native home |
| gRPC with client certificate | Collector requires and verifies the client certificate | Same certificate serial observed; no TLS file copied |
| HTTP binary | Logs, traces, and metrics delivered | All three signals delivered |
| HTTP binary with CA | All signals reach the local TLS collector | Same CA reference works from the new native home |
| gRPC with analytics disabled | Logs and traces delivered; metrics absent | Same result with the retained analytics setting |
| HTTP binary with inline authentication | Source sends authenticated telemetry | Imported exporters are explicitly disabled; no delivery |
| gRPC with inline authentication | Source sends authenticated telemetry | Imported exporters are explicitly disabled; no delivery |

Every enabled case uses a local model and collector. The records correlate a
completed native turn, the model request, log conversation IDs, trace IDs,
resource environment, and configured headers. Metrics include the exact
`codex.api_request` count for successful requests to the model's Responses
endpoint. Other native HTTP requests, such as feedback, are recorded separately.
Prompt text is redacted in exported logs. TLS assets are generated in the test
fixture and remain unchanged at their original paths.

## Defect and correction

Before the fix, native import removed an external authentication header but kept
the rest of the exporter. Native execution from the applied configuration sent
telemetry without the original authentication. The collector observed this in
the [retained failing run](../WORKBENCH/evidence/native-draft2-debug/codex-otel-auth-preservation-before.json).
No real credential or remote collector was used.

The complete exporter is now one configuration unit for exclusion. If an
exporter contains an excluded credential or invalid collector URL, import
records the exclusion and writes the explicit native value `none`. The import
report lists `disabled_exporters` and explains the change. Other settings and
the source configuration remain unchanged.

Apply also uses `none` for optional native exporters with excluded credentials
or the pinned client's unsupported HTTP identity fields. Required content still
blocks apply before writes. The explicit value matters: deleting a metrics
exporter can restore a native default. A disabled exporter is not a successful
mapping of its authentication. The draft.2 specification now states this
[authentication requirement](../SPEC/spec/1.1-draft.2/SPECIFICATION.md#activation-and-security).

Unit tests cover all three signal exporters and both transports. They check
required refusal, optional disablement, import reports, redaction, and unchanged
source values. The HTTP binary and gRPC native cases then verify authenticated
source delivery and no delivery from the applied disabled configuration.

## Analytics dependency and retained attempts

The effective analytics setting must be enabled for custom metrics export.
Codex's pinned `otel_init.rs` changes the metrics exporter to `None` when
analytics is disabled. It does not disable configured logs or traces. The test
now sets analytics explicitly and checks the disabled state separately.

The first gRPC attempt incorrectly expected metrics while analytics was
disabled. The first HTTP binary CA attempt also counted a native feedback POST
as a model request. The corrected runner records request paths and compares
metrics only with requests to `/responses`. Both failed attempts remain in the
evidence directory. The initial Go regression command matched no tests; it is
retained and is not counted as a regression result. The corrected command
reproduced the defect before the fix.

## Evidence and reproduction

The [native runner](../WORKBENCH/conformance/run_native_codex_otel_transports.py)
uses these exact Python dependencies in an external test environment:
`grpcio==1.83.1`, `opentelemetry-proto==1.44.0`, `protobuf==7.36.1`, and
`cryptography==46.0.5`.

Run it with a new output path:

```sh
python WORKBENCH/conformance/run_native_codex_otel_transports.py \
  --transport grpc --tls mutual --adapter --output /absolute/new-result.json
```

Use `--transport http-binary` for the other protocol, `--tls ca` for server
trust, `--analytics off` for the metrics control, or `--credentials inline` for
the credential-exclusion regression. Inline credentials are synthetic test data.
The runner does not copy account credentials or use a remote model or collector.
Each record includes a frozen runner, helper hashes, implementation hashes,
native events, wire observations, dependency versions, and native binary pin.

The [evidence verifier](../WORKBENCH/conformance/verify_codex_otel_transports.py)
checks all eight final records against the current implementation and retained
failure. It also checks the
[official source captures](../WORKBENCH/evidence/native-draft2-debug/codex-otel-transports.sources.json)
and upstream source revision `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`
(`rust-v0.154.0`).

```sh
python3 WORKBENCH/conformance/verify_plugin_selections.py \
  --repository-only --with-project-extensions --with-otel-transports \
  --output /absolute/new-verification.json
```

The earlier HTTP JSON, absolute-path, native HOME-expression, and HTTP identity
failure records remain historical evidence. This review does not establish
remote collector operation, certificate rotation, live reload, arbitrary
identity formats, or project-scope telemetry. Project telemetry remains
inactive. Required portable environment references and the broader security
matrix still have separate gates.

The runtime-authentication audit reran all eight cases against the current Go
source. The current records use `*-auth-current.json`; the earlier
`*-final.json` records remain unchanged. The verifier selects the current set
by default and accepts `--evidence-suffix` for an explicit set. Every selected
record must match the current implementation hashes. See the
[related provider-authentication report](NATIVE_AUTHENTICATION.md).

After the token-command mapping change, all eight transport cases were rerun
as `*-command-current.json`. The verifier now selects that set by default.
The earlier sets remain unchanged and can be selected explicitly; current
implementation hashes are required for any current-source verification.

The project-scope correction was followed by a repeat run against the current
Go source. The verifier selects `*-scope-current.json` by default. Earlier
records remain unchanged. See the [project-scope report](CODEX_PROJECT_SCOPE.md).
