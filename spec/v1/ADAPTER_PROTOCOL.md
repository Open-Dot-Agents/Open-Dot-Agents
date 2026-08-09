# Dota Adapter Protocol 1.0

Adapter plugins are executable processes speaking JSON-RPC 2.0 over standard
input and output. Messages use Language Server Protocol `Content-Length`
framing. Standard error is reserved for human diagnostics and MUST NOT contain
protocol frames.

## Lifecycle

The host calls methods in this order:

1. `initialize` with host and protocol versions. Major protocol versions MUST
   match; peers select the highest common minor version.
2. `describe` returns adapter identity, target name, category statuses, input
   patterns, and supported operations.
3. `validate`, `exportPlan`, or `importPlan` may be called repeatedly.
4. `shutdown` terminates the session cleanly.

Adapters MUST NOT write to the repository. Requests contain a workspace
snapshot of approved regular files. Responses contain diagnostics, losses, and
a deterministic plan of desired files. Content is UTF-8 text or base64 bytes.
The host alone validates paths, detects collisions, tracks ownership, writes
atomically, and removes stale owned files.

## Required methods

- `initialize(InitializeParams) -> InitializeResult`
- `describe() -> AdapterDescription`
- `validate(SnapshotParams) -> OperationResult`
- `exportPlan(SnapshotParams) -> OperationResult`
- `importPlan(SnapshotParams) -> OperationResult`
- `shutdown() -> null`

Unknown methods return JSON-RPC error `-32601`. Invalid parameters return
`-32602`. Adapter failures use `-32000` through `-32099` and SHOULD include a
stable diagnostic in `error.data`.

## Security

Execution of an adapter is execution of third-party code. A committed
`.agents/adapters.lock.json` explicitly selects the executable and pins each
platform artifact by SHA-256. `dota` never discovers adapters from `PATH`.
Local-path adapters are development-only and MUST be rejected by CI mode.

The host SHOULD use an empty working directory, a minimal environment, request
and response limits, and deadlines. Checksums establish integrity, not trust;
publishers SHOULD additionally provide a Sigstore bundle and provenance.
