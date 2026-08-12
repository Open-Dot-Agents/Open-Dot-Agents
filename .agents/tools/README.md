# MCP tool dependencies

Official vendor documentation sources live in
[`llms-sources.json`](llms-sources.json). Use `$agent-docs` before vendor,
API, MCP, skills, tools, or adapter research. The catalog stores unique
`llms.txt` and `llms-full.txt` URLs, effective URLs after redirects, hashes,
sizes, and duplicate removals. Large full-text files are referenced, not
vendored.

The portable MCP catalogue currently includes `codegraph`, `engram`,
`firecrawl`, `context7`, `openai-developer-docs`, and `semgrep`. Keep this
catalogue secret-free: configurable values must use
`urn:open-dot-agents:env:*` references, and no GitHub MCP server should be
added until the repository documents the selected OAuth or PAT model.

The Firecrawl entry invokes `firecrawl-mcp` from `PATH`. The portable tree
declares the server and its environment contract; it does not own package
installation or package-manager state. Pinned installation, audit, and startup
testing live in the Workbench MCP-server conformance harness.

Set `FIRECRAWL_API_URL` in the environment before using the Firecrawl MCP
server. The portable catalogue stores only
`urn:open-dot-agents:env:FIRECRAWL_API_URL`, not the URL value.

Other catalogued servers are intentionally not lockfile-backed here:

- `codegraph` uses the contributor's local `codegraph` binary and should be
  validated with `codegraph --help` plus a repository-level `codegraph explore`
  smoke test when `.codegraph/` is present.
- `engram` uses the local-first `engram mcp` server and should be validated
  with `engram doctor` before it is required in a workflow.
- `context7` uses `npx -y @upstash/context7-mcp` for current library and
  framework documentation lookup. Pin it only after package identity and stdio
  startup are checked for the exact version being committed.
- `openai-developer-docs` is a remote MCP endpoint for official OpenAI and
  Codex documentation. Validate reachability separately from local stdio server
  startup.
- `semgrep` uses `uvx --from semgrep==1.172.0 semgrep mcp`; keep the version
  pinned and run probes with a writable `UV_CACHE_DIR`.

## Validation workflow

Run this check after changing `.agents/tools/mcp.json`:

```sh
python3 -m json.tool .agents/tools/mcp.json >/dev/null
```

Run the pinned server installation and startup checks from `WORKBENCH` with
`task mcp-servers`.

For live startup validation, use isolated writable caches and record the exact
package or binary versions in `docs/VENDOR_EVIDENCE.md` or Workbench evidence before
pinning a server:

```sh
npm_config_cache=/tmp/open-dot-agents-context7-cache npx -y @upstash/context7-mcp --help
UV_CACHE_DIR=/tmp/open-dot-agents-uv-cache uvx --from semgrep==1.172.0 semgrep mcp --help
codegraph --help
engram doctor
```

Remote MCP validation must check the endpoint independently. Do not treat a
successful JSON parse, package install, CLI unit test, or projection test as
native harness compatibility evidence.

Do not add broad filesystem, GitHub, browser, or database servers without a
specific project gap and an explicit authentication or trust model.
