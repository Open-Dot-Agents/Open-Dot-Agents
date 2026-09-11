# Project GitHub plugin

The project contains the unchanged OpenAI GitHub plugin `0.1.11` at
[`plugins/github`](plugins/github). Its source is
[`openai/plugins` at `d416fd5a43426019986b1e489506db3db66dee3d`](https://github.com/openai/plugins/tree/d416fd5a43426019986b1e489506db3db66dee3d/plugins/github).
[`provenance.json`](provenance.json) records all six source files and their
hashes. Upstream declares MIT in the manifest but supplies no license file in
the pinned package or repository root. No license notice was invented.

[`profile.json`](profile.json) describes the draft.2 native selection source.
[`config.toml`](config.toml) enables `github@open-dot-agents` and identifies the
local marketplace. The absolute marketplace source applies to this checkout.
Update it to the namespace directory when the checkout moves.

The native marketplace file is
[`.agents/plugins/marketplace.json`](.agents/plugins/marketplace.json). Its
package source is relative to this namespace directory. The nested native
location is required by Codex; a root `marketplace.json` was not discovered by
Codex `0.154.0` in the retained test.

## Activation and scope

The 2026-09-11 native setup installed the package in the Codex user cache and
merged these selection settings into the ignored project `.codex/config.toml`.
The native install command temporarily enabled a user selection. The setup
verified that exact change, then restored the original user configuration and
permissions. The project selection remains enabled. Native `config/read` and
`plugin/list` confirm the project source, version, installed state, and enabled
state. Private backups are identified in the setup evidence.

The root manifest is still stable `1.0.0`. It does not select the draft.2
`plugins` profile. This native setup is separate from portable CLI sync. Full
sync still refuses the required `mcp.envRef` capability. Do not remove that
requirement to make sync pass. Package installation and its cache remain native
operations; reference CLI apply does not perform them.

The package has no skills. Its two optional app declarations are GitHub and
GitHub Enterprise. The existing GitHub connector is connected in the current
session; a read-only lookup of this repository passed. That connector result
does not prove that the newly installed package made an external MCP call.

Direct HTTP MCP uses `GITHUB_PAT_TOKEN` from the process environment. It is
currently absent. No token is stored here. Use the connected GitHub app, or
provide a token through the native environment before a new session. Native
authentication, project trust, and tool approvals remain in effect. This setup
does not authorize publication, issue comments, merges, or other remote writes.

## Copilot limit

Copilot `1.0.83` can accept the package and list its HTTP MCP component while
ignoring `bearer_token_env_var`. The local authentication probe received an
unauthenticated request even when the test token was present. Native install
success is insufficient to activate this package as a working conversion.
The OpenAI app file also has no established Copilot mapping. No Copilot
selection or replacement GitHub MCP server was enabled by this setup.

The bundled OpenAI plugin validator rejects the upstream app `required` fields.
The failure is retained. The package was not changed to satisfy that validator;
native installation and discovery have separate evidence.

## Checks

From the repository root:

```sh
python3 CLI/scripts/check_project_extensions.py
python3 CLI/scripts/check_project_extensions_test.py
```

These checks validate source integrity. The
[readiness report](../../../docs/PROJECT_EXTENSIONS_READINESS.md) records
native checks, prerequisites, and current release limits. Start a new Codex
session to load the installed project plugin.
