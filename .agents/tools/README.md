# Tools

Tools define external integrations the agent can call, most commonly Model
Context Protocol (MCP) servers.

## Purpose

- register external services and commands
- configure invocation details (`command`, `args`, `env`, transport)
- establish safe, explicit tool wiring

## Vendor mappings

- GitHub Copilot CLI (`.github/mcp.json`)
- Claude Code (`.mcp.json` / `.claude/settings.json`)
- OpenAI Codex CLI (`[mcp_servers.*]` in `config.toml`)

## Expected content

- MCP server definitions and related transport configuration
- secure connection metadata (excluding secrets)
- environment-variable placeholders

`oda` adapter input expects a portable MCP definition in `mcp.json`.
Codex-only server fields are retained under each server's `codex` extension so
they survive import/export even when the portable MCP shape has no equivalent.

## Contract

Machine metadata is in `../schema/v0.0.1/agents.schema.json`, including expected
`tools/mcp.json` shape and supported transports.

See `../mappings.yaml` for vendor references.
