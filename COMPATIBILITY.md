# Codex and Copilot CLI compatibility audit

`oda` treats `.agents/` as canonical and supports repository-scoped import and
export for OpenAI Codex and GitHub Copilot CLI. Compatibility is measured
against the documented repository surface of each harness, not against
personal credentials, caches, sessions, or organization-managed policy.

## OpenAI Codex

| Repository artifact | Canonical representation | Import | Export | Verification |
|---|---|---:|---:|---|
| Root and nested `AGENTS.md` | `instructions/AGENTS.md`, `instructions/codex-project/**` | yes | yes | native prompt-input discovery |
| `AGENTS.override.md` and configured fallback filenames | `instructions/codex-project/**` | yes | yes | native prompt-input discovery |
| `.codex/config.toml` | category `codex.toml` fragments plus `settings/codex.raw.toml` | yes | yes | official schema, strict loading, doctor |
| `.codex/agents/*.toml` | `agents/*.md` plus raw TOML sidecars | yes | yes | native registration; child spawn currently blocked below |
| `.codex/rules/**/*.rules` | `permissions/codex-rules/**/*.rules` | yes | yes | native `execpolicy check` |
| `.agents/skills/**` | native canonical skills | native | native | native prompt-input discovery |
| Legacy `.codex/skills/**` | `skills/**` plus `settings/codex-legacy-skills.json` | yes | yes | byte-identical round trip |
| Codex MCP tables | `tools/mcp.json` with `codex` extensions | yes | yes | official schema, doctor, native MCP call |

The authenticated Codex CLI 0.147.0 gate passes all rows above and completes a
byte-identical `.codex` cycle. Its final custom-agent assertion is deliberately
still strict: the model currently emits a completed collaboration `wait` with
no spawned receiver thread, even though the generated agent is registered.
Until a real child thread is observed, the project does not claim complete live
Codex acceptance.

## GitHub Copilot CLI

| Repository artifact | Canonical representation | Import | Export | Verification |
|---|---|---:|---:|---|
| Root and nested `AGENTS.md` | `instructions/AGENTS.md`, `instructions/copilot-project/**` | yes | yes | authenticated native discovery |
| Root and nested `.github/copilot-instructions.md` | root canonical file plus `instructions/copilot-project/**` | yes | yes | authenticated native discovery |
| `.github/instructions/**/*.instructions.md` at documented discovery roots | portable `rules/**` or exact `rules/copilot-project/**` | yes | yes | authenticated exact-file reads and deterministic round trip |
| `.github/agents/*.agent.md` | `agents/*.md` | yes | yes | authenticated custom-agent invocation |
| `.github/hooks/*.json` | `hooks/*.json` | yes | yes | authenticated hook execution |
| `.github/skills/**` | `skills/**` | yes | yes | native skill inventory and invocation |
| `.github/mcp.json` and root `.mcp.json` | `tools/mcp.json` plus source provenance when needed | yes | yes | native MCP inventory and call; dual-source round trip |
| `.github/prompts/*.prompt.md` | `prompts/*.md` | yes | yes | deterministic round trip |
| `.github/plugin/**` | `plugins/copilot/**` | yes | yes | native `--plugin-dir` inventory |
| `.github/copilot/settings.json` | `settings/copilot.json` | yes | yes | authenticated CLI load and byte-identical round trip |
| `.github/allowed_models.txt` | `permissions/copilot-allowed-models.txt` | yes | yes | deterministic policy round trip |

The authenticated Copilot CLI 1.0.78 acceptance passes the complete matrix.

## Deliberate exclusions

`oda import` rejects inline credentials and symlinked harness inputs. It does
not import personal or machine state such as `~/.codex`, `~/.copilot`, Codex
sessions, Copilot permission caches, or `.github/copilot/settings.local.json`.
Canonical categories for which a harness has no shared repository artifact
remain explicitly `unsupported`; they are never reported as silently mapped.

Official references are maintained in `.agents/mappings.yaml`, including the
[Codex instruction hierarchy](https://learn.chatgpt.com/docs/agent-configuration/agents-md),
[Codex command rules](https://learn.chatgpt.com/docs/agent-configuration/rules),
[Copilot custom instructions](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions),
and the [Copilot configuration directory](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference).
