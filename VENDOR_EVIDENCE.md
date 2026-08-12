# Vendor Mapping Evidence

This document records the upstream facts used to design adapters. It is not a
compatibility guarantee: a mapping becomes supported only after a
version-pinned harness run passes the conformance suite and is recorded in the
[compatibility matrix](COMPATIBILITY.md).

| Harness | Project instructions | Project MCP | Project skills | Official sources |
| --- | --- | --- | --- | --- |
| GitHub Copilot CLI | Discovers `AGENTS.md` and supports `.github/copilot-instructions.md` plus path-specific instruction files. | Committed `.github/mcp.json`; local `.mcp.json` is also supported. | `.agents/skills/<name>/SKILL.md` is a supported project root. | [Instructions](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions), [MCP](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers), [skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| OpenAI Codex | Discovers `AGENTS.md` and `AGENTS.override.md` from project root to working directory. | `.codex/config.toml`, with `[mcp_servers.<name>]`. | `.agents/skills/<name>/SKILL.md`. | [Instructions](https://developers.openai.com/codex/guides/agents-md/), [MCP](https://developers.openai.com/codex/extend/mcp/), [skills](https://developers.openai.com/codex/build-skills/) |
| Claude Code | Reads `CLAUDE.md` or `.claude/CLAUDE.md`; it can import `AGENTS.md` with `@AGENTS.md` but does not read it directly. | Root `.mcp.json`. | `.claude/skills/<name>/SKILL.md`. | [Memory](https://code.claude.com/docs/en/memory), [MCP](https://code.claude.com/docs/en/mcp), [skills](https://code.claude.com/docs/en/skills) |
| OpenCode | Reads root `AGENTS.md`; `opencode.json` can add instruction paths. | Root `opencode.json` or `opencode.jsonc`, under `mcp`. | `.agents/skills/<name>/SKILL.md` is supported, as are compatible roots. | [Rules](https://opencode.ai/docs/rules/), [configuration](https://opencode.ai/docs/config/), [MCP](https://opencode.ai/docs/mcp-servers/), [skills](https://opencode.ai/docs/skills/) |

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

## Required harness verification

Before support is published, the adapter test record must include the exact
harness version and test date. It must cover nested-directory instruction
discovery, trusted and untrusted project MCP behavior, user/project name
collisions, native server startup, and skill discovery/collisions. Native
instruction precedence and reload behavior differ by harness and must not be
normalized without a documented, tested adapter rule.
