# Dota v1 diagnostic registry

Diagnostic identifiers are stable within specification major version 1. New
codes may be added in a minor release; an existing code MUST NOT be reused for a
different condition. Human-readable messages may become more specific.

| Code | Meaning |
| --- | --- |
| `DOTA0002` | Invalid command or option usage |
| `DOTA1000` | Required manifest is missing |
| `DOTA1001` | Legacy manifest is unsupported |
| `DOTA1002` | Manifest is invalid |
| `DOTA1100` | Unknown `.agents/` root entry |
| `DOTA1101` | Invalid or unsafe extension namespace |
| `DOTA1200` | Invalid instruction metadata |
| `DOTA1201` | Invalid or duplicate rule metadata |
| `DOTA1202` | Invalid agent metadata |
| `DOTA1203` | Invalid prompt metadata |
| `DOTA1204` | Invalid skill directory |
| `DOTA1205` | Invalid skill contract |
| `DOTA1206` | Symlinked narrative or skill entry |
| `DOTA1207` | Invalid narrative file type |
| `DOTA1208` | Invalid YAML front matter |
| `DOTA1300` | Symlinked structured category entry |
| `DOTA1301` | Invalid structured file type |
| `DOTA1302` | Structured category document is invalid |
| `DOTA1303` | MCP document or transport is invalid |
| `DOTA1304` | MCP secret is not an environment reference |
| `DOTA1305` | Profile definition or inheritance is invalid |
| `DOTA3000` | Adapter snapshot input is a symlink |
| `DOTA3001` | Adapter snapshot exceeds its byte limit |
| `DOTA3002` | Adapter returned no plan |
| `DOTA3003` | Adapter plan contains duplicate or colliding output |
| `DOTA3004` | Output exists but is not adapter-owned |
| `DOTA3005` | Adapter-owned output was modified |
| `DOTA3006` | Ownership manifest is missing |
| `DOTA3007` | Adapter output is stale or modified |
| `DOTA3008` | Adapter output path violates operation policy |
| `DOTA3009` | Adapter output encoding is invalid |
| `DOTA3010` | Ownership manifest is invalid |
| `DOTA3011` | Output parent is a symlink |
| `DOTA3012` | Existing output is not a regular file |
| `DOTA4000` | Adapter lock is invalid |
| `DOTA4001` | Adapter ID is duplicated in the lock |
| `DOTA4002` | Adapter ID is invalid |
| `DOTA4003` | Requested adapter is not locked |
| `DOTA4004` | Local adapter is forbidden by CI policy |
| `DOTA4005` | Adapter artifact integrity or size check failed |
| `DOTA4006` | Runtime adapter description differs from its lock |
| `DOTA4007` | Publisher manifest, URL, or release metadata is invalid |
| `DOTA4008` | Adapter conformance or determinism check failed |

Command exit status is `2` for usage, `3` for drift/ownership state, `4` for
adapter trust/protocol failures, and `1` for other validation or runtime errors.
