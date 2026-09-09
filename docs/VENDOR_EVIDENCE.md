# Vendor Mapping Evidence

This document records the upstream facts used to design adapters. It is not a
compatibility guarantee: a mapping becomes supported only after a
version-pinned harness run passes the conformance suite and is recorded in the
[compatibility matrix](COMPATIBILITY.md).

| Harness | Project instructions | Project MCP | Project hooks | Project skills | Official sources |
| --- | --- | --- | --- | --- | --- |
| GitHub Copilot CLI | Discovers `AGENTS.md` and supports `.github/copilot-instructions.md` plus path-specific instruction files. | Committed `.github/mcp.json`; local `.mcp.json` is also supported. | `.github/hooks/*.json`; Copilot supports camelCase and VS Code-compatible PascalCase event names. | `.agents/skills/<name>/SKILL.md` is a supported project root. | [Instructions](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions), [MCP](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers), [hooks](https://docs.github.com/en/copilot/reference/hooks-reference), [skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| OpenAI Codex | Discovers `AGENTS.md` and `AGENTS.override.md` from project root to working directory. | `.codex/config.toml`, with `[mcp_servers.<name>]`. | `.codex/hooks.json`; inline TOML hook tables are also documented. | `.agents/skills/<name>/SKILL.md`. | [Instructions](https://developers.openai.com/codex/guides/agents-md/), [MCP](https://developers.openai.com/codex/extend/mcp/), [hooks](https://developers.openai.com/codex/hooks), [skills](https://developers.openai.com/codex/build-skills/) |
| Claude Code | Reads `CLAUDE.md` or `.claude/CLAUDE.md`; it can import `AGENTS.md` with `@AGENTS.md` but does not read it directly. | Root `.mcp.json`. | `.claude/settings.json` under `hooks`. | `.claude/skills/<name>/SKILL.md`. | [Memory](https://code.claude.com/docs/en/memory), [MCP](https://code.claude.com/docs/en/mcp), [hooks](https://code.claude.com/docs/en/hooks), [skills](https://code.claude.com/docs/en/skills) |
| OpenCode | Reads root `AGENTS.md`; `opencode.json` can add instruction paths. | Root `opencode.json` or `opencode.jsonc`, under `mcp`. | Not in the stable adapter set. | `.agents/skills/<name>/SKILL.md` is supported, as are compatible roots. | [Rules](https://opencode.ai/docs/rules/), [configuration](https://opencode.ai/docs/config/), [MCP](https://opencode.ai/docs/mcp-servers/), [skills](https://opencode.ai/docs/skills/) |

## Adapter design consequences

- The canonical instruction artifact projects to a root `AGENTS.md` for
  Copilot, Codex, and OpenCode. A Claude adapter must generate or maintain a
  thin `CLAUDE.md` that imports the root file; it must not claim native
  `AGENTS.md` discovery.
- The common skill source is `.agents/skills`. Claude uses a separate native
  `.claude/skills` projection. Adapters should copy or generate that tree
  rather than rely on symlinks unless symlink behavior is a documented,
  version-pinned guarantee.
- MCP requires native target files: `.github/mcp.json`, `.codex/config.toml`,
  `.mcp.json`, and `opencode.json` respectively. No adapter may assume a
  shared target format or put credentials into a committed configuration.
- Hooks require native target files: `.github/hooks/open-dot-agents.json`,
  `.codex/hooks.json`, and `.claude/settings.json` respectively. Copilot event
  names are projected to camelCase names such as `sessionStart` and
  `preToolUse`; Codex and Claude keep PascalCase event names. The 1.0 profile
  covers portable command hooks only; vendor-only hook types, handler fields,
  and event names are not support claims. No projection result is native
  execution evidence.

## Required harness verification

Before support is published, the adapter test record must include the exact
harness version and test date. It must cover nested-directory instruction
discovery, trusted and untrusted project MCP behavior, user/project name
collisions, native server startup, hook execution and blocking behavior, and
skill discovery/collisions. Native instruction precedence and reload behavior
differ by harness and must not be normalized without a documented, tested
adapter rule.

## Hook mapping review — 2026-09-09

The official OpenAI and GitHub `llms.txt` indexes were fetched. The Claude index
returned HTTP 403; the official hook page below was fetched directly.

| Source | Mapping used by this change |
| --- | --- |
| [Codex hooks](https://learn.chatgpt.com/docs/hooks.md) | `timeout` uses seconds; omitted command timeouts use native defaults. `UserPromptSubmit` and `Stop` ignore matchers. The page does not document `disableAllHooks`. |
| [Copilot hooks](https://docs.github.com/en/copilot/reference/hooks-configuration) | `disableAllHooks` skips hooks in the file. Native camelCase matchers filter tool events, `preCompact`, and `subagentStart`. Command timeout defaults to 30 seconds. |
| [Claude hooks](https://code.claude.com/docs/en/hooks.md) | `disableAllHooks` is a settings-level switch and can affect unrelated hooks. Matcher syntax is native. Removing one owned `hooks` field must preserve other settings. |

The reference CLI refuses catalogue-scoped disablement for Codex and Claude.
It does not set Claude's broader native switch. Hook import preserves Claude's
local disabled value in the canonical catalogue, so a later projection cannot
silently enable those commands. These current documentation checks do not
prove behavior in the pinned native versions.

## Latest native test targets — 2026-09-10 (Europe/Rome)

The official npm registry `latest` metadata resolves `@openai/codex` to
0.154.0 and `@github/copilot` to 1.0.83. The Workbench pins now contain these
exact versions and their registry `dist.integrity` values. CI reads this file
for installation and package-integrity verification. This updates test targets;
it does not establish native feature support.

Sources: [Codex package metadata](https://registry.npmjs.org/@openai%2fcodex/latest)
and [Copilot package metadata](https://registry.npmjs.org/@github%2fcopilot/latest).

## Existing-login native runs — 2026-09-09

[Codex authentication](https://learn.chatgpt.com/docs/auth.md) documents cached
CLI login reuse and `codex login status`. The installed Copilot CLI documents
saved credentials in `copilot login --help` and environment-token precedence
in `copilot help environment`.

[Copilot project MCP documentation](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers)
states that prompt-mode project MCP servers are skipped in untrusted folders.
`--allow-all` alone did not load the fixture's MCP server. Adding only the
fixture folders to the temporary `trustedFolders` setting resolved this.

Codex 0.153.4 and Copilot 1.0.83 passed the local native suite with saved logins.
The [Workbench report](../WORKBENCH/evidence/LATEST_NATIVE_TESTS.md) records the
scope and remaining gaps. These results do not promote full profile support.

## MCP cleanup cycle — 2026-09-10 (Europe/Rome)

The official Codex and GitHub documentation indexes and MCP pages were fetched
again for this cycle. The native target paths remain `.codex/config.toml` and
`.github/mcp.json`. The shared CLI cleanup also applies to Claude `.mcp.json`;
Claude native execution is deferred.

Sources: [Codex MCP](https://learn.chatgpt.com/docs/extend/mcp.md) and
[Copilot MCP](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers).

Codex 0.154.0 passed the baseline again with the final cleanup build and an
existing login. The earlier 0.153.4 results above are historical observations.
See the [baseline report](../WORKBENCH/evidence/LATEST_NATIVE_TESTS.md) and
[extended report](../WORKBENCH/evidence/EXTENDED_NATIVE_TESTS.md) for current evidence.
