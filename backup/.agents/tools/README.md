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

Vendor-only server fields are retained under `../extensions/` so they survive
import/export even when the portable MCP shape has no equivalent.

## V1 contract

`mcp.json` uses strict JSON and portable `stdio` or `streamable-http`
transports. Credential values must be environment references. See
`../../spec/v1/schema/mcp.schema.json`.
