# Reference adapter compatibility

Compatibility belongs to separately executed adapters, not to the `dota` host.
Each result below describes the repository-scoped behavior of the reference
adapter shipped by this repository at **1.0.0-rc.1**. Personal state,
credentials, sessions, caches, and organization policy are out of scope.

Vendor-specific source needed for a lossless projection is preserved under the
adapter namespace in `.agents/extensions/org.open-dot-agents.<target>/`.

## Category projection

| Category | Codex | Copilot | Claude |
| --- | --- | --- | --- |
| `agents` | mapped | mapped | mapped |
| `instructions` | supported | supported | supported |
| `rules` | unsupported | supported | supported |
| `hooks` | mapped | supported | supported |
| `tools` | mapped | mapped | mapped |
| `skills` | supported | mapped | mapped |
| `guardrails` | mapped | unsupported | unsupported |
| `memories` | mapped | unsupported | unsupported |
| `permissions` | mapped | mapped | unsupported |
| `plugins` | mapped | mapped | unsupported |
| `profiles` | mapped | unsupported | unsupported |
| `prompts` | unsupported | mapped | unsupported |
| `settings` | mapped | mapped | unsupported |

`supported` means a direct projection, `mapped` means an adapter transform,
`partial` would require an explicit loss report, and `unsupported` produces no
output. These statuses do not claim native harness feature parity.

## Native acceptance boundary

The Codex probe validates generated TOML against the official schema, exercises
strict config loading, instructions and fallback discovery, command policy,
skills, MCP, custom-agent registration and a real collaboration child thread,
then proves import/export drift and cleanup behavior.

The Copilot probe exercises instructions, nested scoped rules, custom agents,
hooks, skills, MCP, plugins, settings, import/export drift and cleanup through
the authenticated native CLI.

Claude is covered by deterministic fixture and external-adapter tests. It does
not yet have the authenticated end-to-end matrix used for Codex and Copilot.

## Deliberate exclusions

The adapters reject inline credentials and symlinked inputs. They do not import
machine or personal state such as `~/.codex`, `~/.copilot`, sessions, permission
caches, or local-only settings. An unrepresentable source field must be reported
as a loss or preserved in the adapter extension namespace; silent loss is a
protocol conformance failure.
