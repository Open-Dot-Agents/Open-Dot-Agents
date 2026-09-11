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

Other server prerequisites are:

- `codegraph` uses the contributor's local `codegraph` binary and should be
  validated with `codegraph --help` plus a repository-level `codegraph explore`
  smoke test when `.codegraph/` is present.
- `engram` uses the local-first `engram mcp` server and should be validated
  with `engram doctor` before it is required in a workflow.
- `context7` uses `npx -y @upstash/context7-mcp@3.2.2` for current library and
  framework documentation. The package identity, MCP initialization, and tool
  list were checked for this version. Node.js and npm must be on `PATH`.
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
npm_config_cache=/tmp/open-dot-agents-context7-cache npx -y @upstash/context7-mcp@3.2.2 --help
UV_CACHE_DIR=/tmp/open-dot-agents-uv-cache uvx --from semgrep==1.172.0 semgrep mcp --help
codegraph --help
engram doctor
```

Remote MCP validation must check the endpoint independently. Do not treat a
successful JSON parse, package install, CLI unit test, or projection test as
native harness compatibility evidence.

Do not add broad filesystem, GitHub, browser, or database servers without a
specific project gap and an explicit authentication or trust model.

## Project setup

The 2026-09-11 setup merged `context7`, `openai-developer-docs`, and `semgrep`
from this catalog into `.codex/config.toml` and `.github/mcp.json`. These local
files are ignored by Git. Copilot uses native `type: "http"` for the remote
documentation server. Existing model, security, and other server settings
were preserved. Private backups were made before the merge.

This was a targeted file merge. A full reference CLI plan still refuses the
root manifest's required `mcp.envRef` capability for both clients. That
requirement remains in place. The Firecrawl entry was not changed; its
environment-reference behavior still needs a safe mapping. Do not remove the
requirement or force a full sync to bypass this refusal.

Direct MCP checks and merge hashes are in
[`WORKBENCH/evidence/project-tools`](../../WORKBENCH/evidence/project-tools).
The checks cover initialization and tool lists, not native tool execution.
Start a new client session to load changed configuration. Native project
trust rules still apply; this setup does not change trust stores.

The shared
[`property-based-testing` skill](../skills/property-based-testing/SKILL.md)
comes from Trail of Bits. It helps test round trips, normalization, and
state invariants. Both clients can discover it in `.agents/skills`; no vendor
copy is needed. Its `provenance.json` records the source commit and file
hashes. Its upstream CC-BY-SA-4.0 license is included. Test libraries are not
installed by the skill setup.

The upstream static-analysis and differential-review plugins are further
candidates. Their complete workflows need a separate client compatibility
check. They are not installed by this setup.

The [plugin catalog review](../../docs/PLUGIN_CATALOG_REVIEW.md) compares the
user-supplied Codex, Copilot, wshobson, and DSH sources. It records candidate
packages, native-format differences, and missing package paths. Catalog review
does not activate those packages.

The requested [OpenAI GitHub plugin](../plugins/com.openai.codex/README.md) now
has a pinned project source and a Codex selection. It uses the already connected
GitHub app, or `GITHUB_PAT_TOKEN` for its separate native HTTP MCP endpoint.
No credential is stored in the portable tool catalog. Copilot activation is
blocked by the tested package authentication limit.

The [project readiness report](../../docs/PROJECT_EXTENSIONS_READINESS.md)
records source checks, shared skill discovery, the UTF-8 validation fix, and
remaining production gates. The new `diagnosing-bugs` skill is shared directly
from `.agents/skills`. Maintenance checks remain under `CLI/scripts/`.
