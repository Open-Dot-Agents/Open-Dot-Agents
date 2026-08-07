# Tools

Connections to external tools/services the agent can call — most commonly
Model Context Protocol (MCP) servers, but any external tool integration
belongs here.

Maps to: MCP server configuration across all major harnesses (GitHub Copilot
CLI, Claude Code `.mcp.json`/`.claude/settings.json`, OpenAI Codex CLI
`[mcp_servers.*]` in `config.toml`).

Expected content: tool/server definitions (command, args, env, transport).

## `oda` adapter input

The `oda` adapter reads `mcp.json` from this directory. Its content uses the
portable MCP `mcpServers` object and is projected to GitHub Copilot CLI's
`.github/mcp.json`, OpenAI Codex's `.codex/config.toml`, or Claude Code's
`.mcp.json`.

Do not put credentials in this file. Reference environment variables or use
the harness's local secret configuration instead.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`, including
the `tools/mcp.json` portable MCP schema and supported transports.
