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

## Core refusal review — 2026-09-10

The official OpenAI and GitHub indexes were fetched again. The Claude index
and skills page returned HTTP 403; this review makes no new Claude behavior
claim from those pages.

- [Codex MCP](https://learn.chatgpt.com/docs/extend/mcp.md) describes `env_vars`
  as variables to allow and forward. It does not establish the required
  missing-variable failure. The recorded native failure remains the reason
  to refuse stdio references.
- [Codex configuration](https://learn.chatgpt.com/docs/config-file/config-reference.md)
  lists per-skill configuration controls. This change does not claim that
  those controls preserve complete profile selection in the pinned harness.
- [Copilot skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)
  lists `.agents/skills` as a project discovery root. The adapter now refuses
  an unselected non-empty canonical skills directory.

Direct native invocation is outside the CLI refusal boundary. No skill files
are moved or deleted, and no new runtime launcher is introduced.

## Security research, 2026-09-10

The [feature inventory](FEATURE_INVENTORY.md) records live official references
and native help from Codex 0.154.0 and Copilot 1.0.83. Source hashes and versions
are separate. Codex legacy sandbox keys can supersede newer permission profiles;
domain policy needs an active native proxy. Copilot tool path checks and OS
process isolation have separate scope. Its cooperative proxy requires tools
to honor proxy environment variables. Neither vendor preset proves the draft
contract. No native security adapter is promoted. See the
[draft and limits](SECURITY_PROFILES.md).

## Security mapping review, 2026-09-10

The official indexes and native security pages were fetched for this phase.
The source files and hashes are retained with the local security evidence.

- [Codex permissions](https://learn.chatgpt.com/docs/permissions.md) and
  [managed configuration](https://learn.chatgpt.com/docs/enterprise/managed-configuration.md)
  define the named permissions and authority surfaces used by the shell subset.
  Unknown authority and legacy security keys cause refusal.
- [Copilot local sandbox configuration](https://docs.github.com/en/copilot/how-tos/cloud-and-local-sandboxes/configuring-local-sandbox-settings)
  and [local sandboxing](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/understanding-local-sandboxing)
  distinguish tool checks, process isolation, automatic grants, authentication,
  and MCP/LSP scope. The Linux test used locally extracted slirp4netns 1.3.3.

Codex evidence now follows canonical policy through apply to direct native
sandbox commands. It does not cover model tools or approvals. Copilot reached
a live host Unix socket with local network access disabled. Its projection
remains refused. An inside-sandbox temporary write with explicit `/tmp` denial
did not create the host marker; that result does not prove a shared-temp bypass.
See [Codex evidence](../WORKBENCH/evidence/CODEX_SECURITY_SUBSET.md) and the
[Copilot assessment](../WORKBENCH/evidence/COPILOT_SECURITY_ASSESSMENT.md).

## Native isolation and approval follow-up, 2026-09-10

Official Codex and GitHub indexes, security pages, and latest package metadata
were fetched again. Installed versions remain Codex 0.154.0 and Copilot 1.0.83.
The [Codex app-server protocol](https://learn.chatgpt.com/docs/app-server.md)
and schemas generated by the pinned binary define the native approval channel.
The generated schema still lists `untrusted`, but that binary rejects it at
startup. Tests use the accepted `on-request` policy.

Copilot native help exposes outbound/local network switches and path denial.
The upstream [MXC network implementation](https://github.com/microsoft/mxc/blob/d3d57c38a6c3edacd31a46fbf2f410439d82fa00/src/backends/bubblewrap/common/src/network_rules.rs)
exempts the sandbox's own loopback before network rules. That source is an
upstream explanation, not a build identity for the bundled runtime. Native
Copilot tests separately confirmed TCP, UDP, and IPv6 loopback access with
both network switches false. Denying the socket directory blocked the named
host Unix socket. No exposed control tested here preserves the full draft rule.

Native app-server/ACP clients observed actual model shell calls and explicit
approval decisions. Copilot's unattended JSON events show a native denial.
Codex's unattended events include execution under a read-only sandbox with
`on-request`; that setting is not equivalent to portable mandatory `ask`.
See the [report](../WORKBENCH/evidence/ISOLATION_APPROVALS.md). No projection or
support status is promoted.

## Draft.2 scope and trust-derived tests, 2026-09-10

The live Codex configuration reference distinguishes an omitted approval policy
in an untrusted workspace from an explicit native `on-request` setting. The new
Codex test records effective `untrusted` policy, explicit-rule acceptance and
denial, and inconclusive unattended and built-in-tool attempts. Copilot tests now
separate host and sandbox abstract Unix sockets from filesystem sockets and
TCP/UDP loopback. See `WORKBENCH/evidence/NATIVE_DRAFT2.md` for exact pins and
results. These tests do not promote an adapter or expand mandatory security
mappings.

Official sources:

- https://learn.chatgpt.com/docs/llms.txt
- https://learn.chatgpt.com/docs/config-file/config-reference.md
- https://docs.github.com/llms.txt
- https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference

The separate draft.2 projection path uses a typed native setting registry.
Configuration writes and native activation remain separate claims.

## Native configuration debug review, 2026-09-10

The Codex validator now uses the exact `rust-v0.154.0` schema, with integer bounds
and source-confirmed aliases. The live official schema had the same SHA-256 when
checked. A native probe loaded the agent-limit and memory aliases. A duplicate
agent-limit assignment produced a warning while the client continued to start.
The first probe's failed expectation is retained with the corrected assessment.

The new evidence covers configuration load only. It does not establish resource
limit enforcement, approval behavior, or full adapter support. See the
[debug review](NATIVE_DEBUG_RESEARCH.md) and
[raw alias results](../WORKBENCH/evidence/native-draft2-debug/codex-alias-load-v2.json).

The follow-up uses a loopback Responses fixture based on the pinned upstream
test protocol. Codex accepts the CLI-projected agent in user and trusted-project
scope. Native spawn and command events correlate with the child thread, and the
child request contains the configured instructions, model, and reasoning value.
The child creates the expected fixture file. User configuration stays unchanged
by project projection. These results cover the recorded agent fixture only.

The same deterministic provider resolves the earlier safe-command and built-in
approval uncertainty. In an isolated untrusted workspace, with the approval key
omitted and project configuration disabled, `cat` requests approval. A built-in
image read completes without approval. Explicit-rule acceptance and denial have
correlated terminal events and the expected file effect. This excludes a
mandatory portable `ask` mapping for the tested trust-derived native mode.

The refreshed configuration references match the inventory hashes:

- Codex: `f432ae52ed50ad88998c396c8da88cba74ef022f51f2def7ca494fd08c94ae58`.
- Copilot: `da7e43c09e769e9e294cf57eb27cb80bb7b12e47fbc6be4c70bb98cbd5a5875f`.


The Copilot CLI reference was refreshed for agent-local MCP. It documents
`mcp-servers` in agent YAML with the user MCP schema. The
[archived source](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-docs.md)
and [fetch record](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-docs.source.json)
retain the source URL and hashes. Pinned Copilot `1.0.83` tests verify local
agent MCP calls, parent isolation, and a child override of a same-name user
server in both scopes. See the
[agent MCP review](NATIVE_DEBUG_RESEARCH.md#copilot-agent-local-mcp-and-parent-isolation).
These fixture results do not change adapter support or release gates.


The [refreshed Copilot hook reference](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-docs.md)
and [fetch record](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-docs.source.json)
document hook directories, inline settings, native fields, and timeout behavior.
Copilot `1.0.83` fixtures verify the bounded command and local HTTP mappings.
Project results use CLI prompt mode; the ACP project attempts did not load hooks.
Direct-executable environment expressions stayed literal. A command-hook timeout
allowed tool execution to continue. See the
[hook review](NATIVE_DEBUG_RESEARCH.md#copilot-native-hook-fields-and-authority).
These results do not establish portable permission enforcement or adapter support.


The [allowing-tools article](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-rules-docs.md),
[fetch metadata](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-rules-docs.source.json),
and [pinned CLI help](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-rules-help.txt)
are the sources for the follow-up
[shell-rule review](NATIVE_DEBUG_RESEARCH.md#copilot-shell-rule-matching-and-interface-boundaries).
Native tests distinguish command stems, full arguments, exact executable
spelling, denial precedence, path grants, and prompt/ACP behavior. The tests do
not establish a portable permission mapping.


The OpenAI Codex index now routes hook documentation to
[the official hook page](https://learn.chatgpt.com/docs/hooks). The
[archived document](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-docs.md)
and [fetch metadata](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-docs.source.json)
record that route. Codex `0.154.0` native tests verify command hooks only after
exact native hash trust is configured. Unknown JSON root fields cause native
file rejection; `prompt` handlers are skipped. See the
[Codex hook review](NATIVE_DEBUG_RESEARCH.md#codex-native-hooks-and-external-trust).
Schema acceptance alone does not establish execution or hook-trust authority.

## Codex MCP hook execution and parser limits

The [official hook page](https://learn.chatgpt.com/docs/hooks.md), fetched again
from the official index, matches the archived document hash. It states that MCP
hooks use existing connections, run synchronously, and do not support
`SessionEnd`. Errors and timeouts do not establish an operation denial.

Pinned Codex `0.154.0` tests now verify projected `PreToolUse` and `PostToolUse`
MCP hooks in project and user scope, for files and inline settings. A native
parser matrix confirms that `SessionEnd` MCP handlers are skipped and any
nested `null` input rejects the hook file. A local MCP fixture verifies type
preservation, event references, returned context, explicit denial, error paths,
and timeout behavior. The server can finish after the native wait times out.
See the [MCP hook review](NATIVE_DEBUG_RESEARCH.md#codex-mcp-hook-execution-and-parser-limits)
for evidence and limits. Full adapter support remains unchanged.

## Codex background hooks and context output

The official [hooks](https://learn.chatgpt.com/docs/hooks.md) and
[app-server](https://learn.chatgpt.com/docs/app-server.md) pages describe
background execution, per-handler context limits, and the unsubscribe grace
period. The [source record](../WORKBENCH/evidence/native-draft2-debug/codex-background-docs.source.json)
contains the archived paths and hashes.

Pinned Codex `0.154.0` evidence now covers delayed background context, ignored
background denial, ordinary and detached descendants at timeout, archive,
graceful shutdown, unsubscribe, and context spilling. A detached child can
produce an effect after the hook timeout. Unsubscribe keeps the native thread
loaded. These results do not establish portable mandatory approval or universal
process termination. See the
[background and context review](NATIVE_DEBUG_RESEARCH.md#codex-background-hooks-and-context-spill-limits)
for the exact cases and limits. Adapter support gates are unchanged.

## Shared plugin package and native selection

The [Agent Plugins review](PLUGIN_STANDARD.md) pins upstream commit
`ff8ab5e392cc87bd88d87c060815a87490e51003`. Published version 1.0.0 defines
root `plugin.json`, skills, and MCP. Marketplace catalogs and installation
commands remain native. OpenAI's [builder page](https://learn.chatgpt.com/docs/build-plugins.md)
and GitHub's [plugin reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-plugin-reference)
describe the shared format and separate native extensions.

Codex `0.154.0` and Copilot `1.0.83` install and discover the same unchanged
local skill package in isolated project and user fixtures. Codex exposes
`fixture:fixture` with its plugin ID and versioned cache path. Copilot's
`skill list --json` exposes the live package skill; its broader `plugins list`
output omits that skill in this fixture. Both stop exposing the skill after
disable. The [verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-selections.json)
retains exact hashes, correlated native output, and failed attempts. These
results cover local installation and discovery, not remote marketplaces,
plugin MCP execution, or all package loading rules. Support gates are unchanged.

## Git marketplace fetching and Copilot update scope

The [archived Copilot settings reference](../WORKBENCH/evidence/plugin-standard/copilot-marketplace-settings.md)
and [source record](../WORKBENCH/evidence/plugin-standard/copilot-marketplace-settings.source.json)
state that marketplace `autoUpdate` in project settings is ignored. Its opt-in
requires user or managed scope, and applies to interactive or prompt sessions.
The adapter now keeps project entries with this field inactive and refuses
required activation. It does not claim that writing a user opt-in proves an
automatic update.

An isolated HTTP Git fixture verifies native fetch requests and exact installed
bytes in both scopes. Codex `0.154.0` requires native marketplace installation
before package installation when its cache is absent; its `ref` selects the
tested tag instead of the default branch. Copilot `1.0.83` fetches a configured
Git source during plugin install. Both clients discover the package skill and
stop exposing it after disable. See the
[Git verification](../WORKBENCH/evidence/plugin-standard/verification-plugin-git.json).
This is loopback HTTP Git evidence. Public HTTPS/SSH, GitHub API behavior,
credentials, and automatic updates remain separate checks.

## Shared stdio MCP package execution

The [MCP verification](../WORKBENCH/evidence/plugin-standard/verification-plugin-mcp.json)
records 14 MCP cases and eight repeated marketplace cases with Codex `0.154.0`
and Copilot `1.0.83`. An identical Agent Plugins 1.0.0 package executes through
both clients. Approval acceptance and denial, native completion, server calls,
file effects, returned model input, disable, and data retention are correlated.
All servers and the deterministic model run locally. Native approval settings
are fixture setup; no portable security mapping is added.

Codex's unapproved MCP tool fails under `approval_policy=never`. An explicit
native tool approval allows it. Prompt mode uses `mcpServer/elicitation/request`
with `codex_approval_kind=mcp_tool_call`. Copilot's ACP permission request uses
`session/request_permission`. Both reject a denied call before it reaches the
server. Copilot ends the turn after denial without a model follow-up.

The [standard environment requirements](../WORKBENCH/evidence/plugin-standard/agent-plugins-environment-requirements.md)
and [source hashes](../WORKBENCH/evidence/plugin-standard/agent-plugins-environment-requirements.source.json)
establish two pinned Copilot limits: it expands the unknown `${ODA_AMBIENT}`
placeholder, and a `${PLUGIN_DATA}` environment value remains literal during the
ACP session. Argument/cwd expansion and reserved variables work in this fixture.
The native limits remain visible in capabilities. Full Agent Plugins conformance
and adapter support are not claimed.

## Copilot legacy preference migration

The [native config help](../WORKBENCH/evidence/native-draft2-debug/copilot-legacy-config-help.txt),
[settings reference](../WORKBENCH/evidence/native-draft2-debug/copilot-legacy-settings-reference.md),
and [source record](../WORKBENCH/evidence/native-draft2-debug/copilot-legacy-sources.json)
identify the modern preferences file and separate application state. Three native
`1.0.83` fixtures establish precedence: legacy preferences replace matching
modern roots, including complete objects. This was measured in an independent
control home, not inferred from the migration note.

The [verification](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-legacy.json)
checks that read-only import matches native migration, preserves source bytes
and modes, excludes application state, writes private user preferences, survives
native startup, and reimports unchanged. It also checks refusal when pending
migration would overwrite a requested setting. The eight local/Git marketplace
cases pass again with the updated user importer.

This proves preference migration and configuration preservation. It does not
prove memory-service behavior or graphical rendering of UI preferences.

## Copilot terminal preferences

The pinned [native help](../WORKBENCH/evidence/native-draft2-debug/copilot-legacy-config-help.txt)
documents user preference names, tab matching, history bounds, and status-line
commands. It explicitly permits an omitted `statusLine.type`. The online
settings table describes `type` as `"command"`; the native fixture tests both
the explicit and omitted forms.

The [preference runner](../WORKBENCH/conformance/run_native_copilot_preferences.py)
applies canonical user settings, then starts fresh Copilot `1.0.83` sessions in
an isolated terminal and native home. Its [result](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-final.json)
records initial plan and interactive mode indicators, tab order, case-insensitive
names, hide/restore, and disabled tabs. Native status JSON correlates with command
file events and terminal output. The one-second fixture produced nine calls;
the two-second fixture produced five. Removing the status setting produced no
further calls after restart. Padding values three and zero are observed.

The adapter preserves native state during apply, creates private user settings,
and preserves preference values through reimport. The first terminal assertion
failed because column one was reached with CRLF instead of an explicit cursor
command. The [failed attempt](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-first.json)
and its runner remain available. No mode-change event, model turn, notification,
image, autopilot, or full adapter conformance claim follows from these checks.

## Copilot dispatch preferences

The [source manifest](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-sources.json)
records the settings reference, subagent-limit reference, and pinned native BYOK
help. Per-agent configuration uses native model names, effort values, and context
tiers. The limit reference states that depth and concurrency overrides require
usage-based billing. The pinned config help states that `disabledSubagents`
cannot disable `rubber-duck`.

The [native runner](../WORKBENCH/conformance/run_native_copilot_subagents.py)
uses Copilot `1.0.83`, an isolated native home, and a deterministic local provider.
It applies canonical settings and the agent asset before each test. Its controls
establish these observations:

- [Inherit](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-verified-inherit.json):
  the parent and child both request `gpt-5.4` with low effort.
- [Override](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-verified-override.json):
  the parent requests `gpt-5.4` with low effort; the child requests `gpt-5-mini`
  with high effort. The native configured event records `long_context`.
- [Disabled](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-verified-disabled.json):
  the custom agent is absent from the task schema. A direct dispatch request
  fails and creates no child effect.
- [Limits](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-verified-limits.json):
  both limits are one, but two nested agents start and complete. This BYOK
  fixture does not satisfy the native usage-based billing prerequisite.

Successful child runs have correlated native dispatch events, fixture approval,
and an observable file effect. Import preserves native configuration and state.
No remote model runs, context-capacity measurement, paid-plan limit enforcement,
or full adapter support claim follows from these fixtures.

## Codex telemetry scope

The [source manifest](../WORKBENCH/evidence/native-draft2-debug/codex-otel-sources.json)
retains the current official reference and advanced configuration page. The
reference states that project-local `otel` is ignored. Its `<id>` exporter
notation denotes `otlp-http` or `otlp-grpc`; it does not define arbitrary names.
The embedded `0.154.0` schema contains these concrete variants.

The [native adapter fixture](../WORKBENCH/evidence/native-draft2-debug/codex-otel-adapter-final.json)
uses Codex `0.154.0`, synthetic prompts, and local model and collector services.
It correlates completed turns, prompt log events, and exported trace IDs.
User environment tags and headers reach the collector. Prompt export emits the
synthetic prompt when enabled and `[REDACTED]` when disabled. A trusted project
cannot override the user collector or prompt-export preference. The fixture
writes the ignored project control directly after the adapter refuses it.
No TLS, gRPC, binary protocol, metrics delivery, live reload, or full adapter
support claim follows from this evidence.

The later [TLS review](NATIVE_DEBUG_RESEARCH.md#codex-telemetry-tls-references-and-client-identities)
adds CA-only HTTP TLS delivery and reference relocation evidence. Six separate
native controls reproduce HTTP identity builder failures for logs, traces,
and metrics with EC and RSA keys. Independent TLS controls verify the generated
certificates. HTTP identity settings therefore keep the complete exporter
inactive; the adapter does not remove authentication. gRPC, metrics delivery,
certificate rotation, and live reload remain unverified.
