---
name: agent-docs
description: Use official llms.txt documentation indexes for OpenAI, ChatGPT/Codex, GitHub, Claude Code, Anthropic API, and Model Context Protocol research before changing skills, tools, adapters, MCP behavior, or compatibility claims.
---

# Agent Docs

## Workflow

Use this skill when work depends on current vendor documentation for agent
tools, skills, MCP, adapters, APIs, or compatibility evidence.

1. Read `.agents/tools/llms-sources.json` and select the smallest source that
   matches the domain:
   - ChatGPT and Codex docs: `chatgpt-codex-docs`.
   - OpenAI API docs: `openai-api-docs`; use `openai-api-docs-full` only when
     the index is not sufficient.
   - GitHub, Copilot, GitHub APIs, and Actions: `github-docs`.
   - Claude Code: `claude-code-docs`.
   - Anthropic API: `anthropic-api-docs`; use `anthropic-api-docs-full` only
     when the index is not sufficient.
   - MCP docs and specification: `modelcontextprotocol-docs`; use
     `modelcontextprotocol-docs-full` only when the index is not sufficient.
2. Fetch the selected URL live before relying on it, unless the task is an
   offline-only edit. Compare the fetched hash or effective URL with the
   catalog when exact evidence matters.
3. Prefer index files before full files. Full files are large and should not be
   loaded unless the needed page is absent from the index or exact full-page
   wording is required.
4. Use official documentation as upstream evidence only. Do not convert docs
   research, CLI unit tests, projection tests, or successful package startup
   into native harness support claims.
5. When documentation changes affect Open-Dot-Agents support status, update
   `docs/VENDOR_EVIDENCE.md`, `compatibility.json`, `docs/COMPATIBILITY.md`, and related
   skills or tests together.

## Validation

For catalog changes, validate JSON and repository conformance:

```sh
python3 -m json.tool .agents/tools/llms-sources.json >/dev/null
GOCACHE=/tmp/agents-gocache GOPATH=/tmp/agents-gopath go run ./cmd/agents validate --root .. --format json
```
