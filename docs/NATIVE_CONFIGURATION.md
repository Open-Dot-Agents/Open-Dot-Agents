# Native configuration implementation

The draft.2 implementation is incomplete. It adds a separate native profile and
scope path without changing stable 1.0 or draft.1 security semantics. It does not
mark Codex or Copilot fully supported.

Use `SPEC/examples/native-draft` as a project example:

```sh
cd CLI
go run ./cmd/agents plan --vendor codex --root ../SPEC/examples/native-draft --experimental --format json
```

Native import can create a canonical tree or merge into an existing draft.2 tree.
It adds disjoint values and preserves existing portable policy and required native
status. It refuses conflicting values, including with `--force`. It does not
migrate an existing stable or draft.1 tree:

```sh
agents import --vendor codex --root /absolute/repository --experimental --scope user --native-home /absolute/native-home
agents plan --vendor codex --root /absolute/repository --experimental --scope user --native-home /absolute/native-home --adopt --format json
agents apply --vendor codex --root /absolute/repository --experimental --scope user --native-home /absolute/native-home --adopt
```

Project scope is the default. User scope requires `--native-home`. Use an explicit
vendor for draft.2 sync. `--codex-home` remains a draft.1 security option.
`--adopt` can claim an unowned setting only when its value is equal to the desired
value. `--force` cannot override a draft.2 ownership conflict. `--backup` creates
private `.bak` files and refuses to replace an existing backup. Plan includes the
configuration and ownership backups before apply starts.

Codex `otel` settings require user scope. Pinned Codex `0.154.0` ignores these
settings in a trusted project. Required project telemetry refuses before writes;
optional project telemetry stays inactive. User import and apply preserve the
native exporter object. Known credentials in collector URLs and headers stay
external, and import reports their field paths without their values. TLS paths
remain references; apply does not copy certificates or private keys.
The [telemetry fixture](../WORKBENCH/evidence/native-draft2-debug/codex-otel-adapter-final.json)
verifies local OTLP HTTP JSON logs and traces, environment tags, static fixture
headers, and prompt export on and off. Each test starts a new native process.
CA-only TLS also has [native relocation evidence](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-ca-relative-final.json).
Import makes relative TLS paths absolute against the source configuration
directory. Absolute paths and native HOME expressions remain unchanged.
Codex `0.154.0` fails to build the tested HTTP client identities. Required
identity settings refuse activation; optional settings keep the complete
exporter inactive. The adapter does not remove authentication from an exporter.
gRPC, binary encoding, metrics delivery, and live reload need separate tests.

The user ownership registry is under
`$XDG_STATE_HOME/open-dot-agents/native/<vendor>/`, or
`~/.local/state/open-dot-agents/native/<vendor>/`. Its filename hashes the
absolute native-home path. Project ownership is under `.agents/state/`.
Transactions include settings, assets, ownership records, backups, and removals.
Lock files remain available for later transactions.
On Linux, native transactions pin parent directory handles. Reads, writes,
removals, and rollback do not follow a replacement symlink. Writes retain the
snapshot used by plan or import. Changed targets and late-created backups cause
refusal. Rollback restores unchanged transaction outputs and reports a conflict
if another writer has changed an output; it does not overwrite that newer edit.

Configuration files are merged by field. TOML and JSON values are preserved;
comments and formatting can change when a value changes. Duplicate JSON keys,
invalid value types, symlinks, and parent/child conflicts cause refusal.
Plan and capabilities output contain field names and dispositions, not values.
Unknown optional object fields stay in canonical sources and remain inactive;
known siblings can activate. Each omitted field has a JSON-pointer diagnostic.
Arrays stay atomic when a member is unmapped. Required unknown content refuses
activation. Parsing alone does not activate an unmapped field.
Native imports read recognized configuration files and instructions.
MCP, model-provider, and LSP definitions refuse import and projection when
credential exclusion would remove authentication. Optional profiles and forced
writes do not bypass this check. Codex `requires_openai_auth` remains a Boolean
setting; native account files stay external. See the
[authentication evidence and limits](NATIVE_AUTHENTICATION.md).

Excluded authority and credential field names are recorded without values in
`import-report.json`; their source files stay unchanged. They move
established MCP fields to the portable core and keep other fields native.

Copilot user import also reads recognized legacy preferences from `config.json`.
Pinned Copilot `1.0.83` migrates these preferences at startup: a legacy root
replaces a matching `settings.json` root, including a whole object. Import uses
that precedence without changing either native file. Authentication, trust,
installed-package state, and unrecognized legacy fields stay external.
Documented preferences without activation mappings remain optional and inactive.

User apply refuses a setting write or removal that pending legacy migration
would undo. Use the pinned native client to migrate those preferences before
retrying; force and adopt do not bypass this check. Apply does not migrate or
edit `config.json`. Project operations do not inspect user application state.
See the [legacy verification](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-legacy.json).

Copilot user preferences include `defaultMode`, `inlineImages`,
`inlineImageLiveWindow`, `notifications`, and the `statusLine` fields.
`commandHistoryMaxSize` must be an integer from 1 to 1000. Image-window and
status-padding values must be nonnegative integers within the exact JavaScript
integer range. Status refresh accepts integers from 1 to 2147483 seconds.
The status command must be nonempty and contain no NUL. `statusLine.type` is
optional; its only mapped value is `"command"`.

Tab names are case-insensitive. An unknown tab name or a request to hide the
Session tab keeps its array inactive because native ignores that request.
A required inactive preference blocks apply before writes. These preferences
are user-scoped; project scope does not activate them. `defaultPermissionMode`
remains blocked by the native security gate.

The [terminal fixture](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-final.json)
verifies initial plan/interactive mode, tab order and visibility, status command
input/output, padding, timer refresh, and removal after restart. Apply writes
the settings; the native terminal executes the status command. Notification
delivery, image rendering, autopilot execution, command history behavior,
status-command failures, and event-only refresh still require native tests.

Copilot user settings also map `subagents.agents.<name>.model`, `effortLevel`,
and `contextTier`. `inherit` keeps the parent choice. Model names remain native
identifiers. Agent names are selectors, so a name such as `permissions` does
not become an authority field; unmapped nested fields remain inactive.
The `disabledSubagents` array is atomic. A request to disable `rubber-duck`
stays inactive because Copilot explicitly ignores it through this setting.

The implemented depth and concurrency values are positive integers, with upper
bounds 256 and 32 respectively. Other lower-bound behavior needs a native
usage-based billing fixture. These limits only take effect with native
usage-based billing. The isolated BYOK test still starts two nested agents
when both settings are one. Plan output reports the billing prerequisite.
Apply changes neither account state nor runtime enforcement.

The [dispatch evidence](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-verified-override.json)
correlates per-agent model and effort choices with provider requests, native
configuration events, approved child execution, and a file effect. The
[disabled control](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-verified-disabled.json)
confirms removal from discovery and refusal of a direct dispatch attempt.
The context-tier event confirms configuration selection; actual context
capacity and billing behavior are not measured.

## Shared project skills and Codex skill controls

Codex and Copilot already discover project skills in `.agents/skills`. Keep one
canonical directory. No new loader or duplicate skill tree is required. See
[skill distribution with skills.sh](PLUGIN_STANDARD.md#skill-distribution-with-skillssh).

Codex `skills.config` selectors are a separate native user preference. Paths must
name `SKILL.md`; the pinned native client ignores folder selectors. Imported
absolute paths to copied user skill packages become paths relative to the new
native home. Relative references outside those packages become absolute references
to their original targets. Native `~` expressions retain their runtime meaning.
External files and system packages are not copied or owned through those references.

Codex `0.154.0` reports project selectors in `config/read`, but ignores them in
both skill discovery and the model's initial catalog. Optional project selectors
remain inactive; required selectors refuse before writes. This does not prevent
ordinary discovery of `.agents/skills`. The
[user fixture](../WORKBENCH/evidence/native-draft2-debug/codex-skills-verified-user.json)
tests import, relocation, enable/disable, model context, and reimport. The
[project fixture](../WORKBENCH/evidence/native-draft2-debug/codex-skills-verified-project.json)
tests refusal and the native limitation while preserving user configuration.

Native version constraints currently accept only `=0.154.0` for Codex and
`=1.0.83` for Copilot, with an optional omitted `=`. These are test pins, not a
supported version range. Codex values use the vendored schema from the exact
`rust-v0.154.0` source tag, with integer bounds and two source-confirmed aliases.
Copilot types use the source-linked reference corpus. The documentation can be
newer than the binary. A registry mapping does not prove native activation.
Known Copilot object containers are derived from their declared fields. This
permits `ide`, `subagents`, and `tabs` without accepting unknown descendants.
The selector accepts indexed members of declared `string[]` fields. The `memory`
preference is boolean; its permission-operation namesake has a separate semantic
context. `bannerStyle` accepts the documented `mona` and `classic` values.

The agent limit alias `agents.max_threads` and the memory alias
`memories.no_memories_if_mcp_or_web_search` retain their original spelling.
Duplicate alias and canonical assignments are refused, including assignments
from separate source and destination files. JSON numbers retain exact decimal
values. Versioned ownership hashes compare equal numeric values without rounding.

The current implementation includes portable MCP and hooks in draft.2 merging.
Project skills retain canonical native discovery; selected user skills are copied
as owned files. User skill import excludes native `.system` packages and preserves
executable assets with private mode `0700`. Import also reads recognized agents,
scoped instructions, and hooks. Codex standalone agents require non-empty `name`,
`description`, and `developer_instructions` fields. Their configuration fields
pass the pinned validator and independent authority checks. Unknown or blocked
agent content keeps the complete asset inactive. Duplicate normalized agent
names refuse, including names in existing unowned files. An unchanged owned
agent can move to a new registered filename in one transaction.

Pinned native fixtures verify Codex agent discovery, instructions, model and
reasoning selection, delegation, and a child command in both scopes. They use a
local deterministic model provider. Copilot agent and LSP mappings are described
below. Native hook fields have the bounded mapping described below. Other asset
families remain incomplete. Combined portable
security and draft.2 native configuration are refused until their behavior is
verified. This refusal keeps the existing mandatory security requirements intact.

The coverage map is `.agents/features/coverage.json`. Run
`python3 CLI/scripts/native_coverage.py` to rebuild it. The original 1,679-entry
inventory stays unchanged. Counts refer to semantic features, not source rows or
verified native behavior. `complete: false` is intentional: the implementation
still has missing mappings, incomplete asset import, and incomplete native
behavior tests. These are outstanding work, not evidenced native limitations.

## Verification for this implementation

The following checks passed on 2026-09-10:

- Go unit tests with the race detector, plus `go vet ./...`.
- Stable conformance: 25 checks.
- Draft.1 schemas: 59 cases; draft.2 native and plugin metadata: 21 cases,
  with separate plugin rejection checks for stable and draft.1.
- Workbench deterministic suite: 74 tests, including security, approvals, and MCP projections.
- Compatibility registry, Markdown, CLI capability, and evidence-rule agreement.
- Canonical repository validation and experimental native example validation.
- JSON parsing and whitespace checks for changed files.

The coverage map retains exactly 1,679 unique source-row references and the
unchanged inventory hash. Native tests retain all inconclusive attempts. See
`WORKBENCH/evidence/NATIVE_DRAFT2.md`; unit test success does not replace those
native results. The [debug review](NATIVE_DEBUG_RESEARCH.md) records further
defects, fixes, source evidence, and remaining completion requirements.

## Existing plugin selections

The experimental `plugins` profile stores native package selections under
`.agents/plugins/com.openai.codex/` or `.agents/plugins/com.github.copilot/`.
Each namespace has `profile.json` and a native config artifact. Use the existing
import, plan, apply, and sync commands with an explicit vendor and scope. Import
places new selections in this profile and preserves selections already declared
in native profiles, including custom source filenames and required status.

The package remains at its native source or installation location. Apply does
not install packages, grant trust, update caches, or copy plugin runtime data.
It reports that native startup or refresh can fetch enabled packages. User
ownership, private backups, removal, and rollback also apply to plugin settings.
A second repository cannot replace another repository's selection with force.

Codex `0.154.0` and Copilot `1.0.83` pass the same local Agent Plugins 1.0.0
skill fixture in project and user scope. The tests check exact discovery records,
package bytes, reimport, disable, and project isolation. Copilot loads a local
marketplace package directly; Codex caches a copy. Separate Git tests verify
HTTP fetching over loopback, a pinned Codex tag, the Copilot default branch,
reimport, and disable in both scopes. Public services, credentials, automatic
updates, and other plugin components remain unverified. A shared stdio MCP
package has separate execution, approval, denial, and disable evidence in both
scopes. Copilot does not meet the tested unknown-placeholder and PLUGIN_DATA
environment-value rules. Plan and capabilities report these native limits.
Copilot marketplace
`autoUpdate` is inactive in project scope because the client ignores it there.
Plan reports Codex's separate Git marketplace-add prerequisite. See the [plugin guide](PLUGIN_STANDARD.md) and
[verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-selections.json).
The [Git follow-up](../WORKBENCH/evidence/plugin-standard/verification-plugin-git.json)
records the current implementation hashes and all eight native cases.
The [MCP follow-up](../WORKBENCH/evidence/plugin-standard/verification-plugin-mcp.json)
supersedes those hashes with 22 native cases for the current implementation.
This result does not close the full universal configuration milestone.

## Copilot LSP configuration

Use a native artifact with `"kind": "lsp"` and a JSON source file. The adapter
writes `.github/lsp.json` in project scope or `lsp-config.json` in the explicit
user native home. It imports only these recognized paths. The adapter owns each
JSON setting and preserves unrelated servers. Installation and server execution
remain native operations.

The `lspServers` object contains named servers. Each server requires `command`
and `fileExtensions`. The adapter also maps `args`, `env`, `rootUri`,
`initializationOptions`, and `requestTimeoutMs`. Initialization options retain
server-specific JSON semantics. Unknown optional fields remain in the source
and inactive. Unknown required fields block apply. Invalid server configuration
cannot launch a partial server definition.

Credential values remain excluded. A sensitive environment variable can use an
explicit `${VARIABLE}` reference; apply does not resolve that reference. Copilot
expands the command, arguments, and environment when it starts the server. The
native fixture verifies this behavior in both scopes with Copilot `1.0.83`.
It also verifies root URI selection, initialization options, file-extension
matching, and a completed hover tool call. A one-second request timeout suppresses
a delayed hover response. Copilot reports an empty result without a timeout
diagnostic. Other LSP operations need separate tests.
Start a new Copilot session after apply. See the
[debug report](NATIVE_DEBUG_RESEARCH.md#copilot-lsp-configuration-and-native-execution)
for evidence and limits.

## Copilot agents

The `agent` artifact maps Markdown files under `.github/agents/` in project
scope and `agents/` in the explicit user home. Both `.md` and `.agent.md` names
are accepted. Import and apply preserve the source bytes. YAML validation
refuses duplicate fields, unknown fields, malformed values, and unsupported
tags. The whole agent stays inactive if any field has no mapping.

Native tests with Copilot `1.0.83` verify discovery, prompt loading, delegation,
and an approved child command in both scopes. `tools` accepts a string or a
string array. An empty list blocks a requested Bash call. `infer: false`
removes the agent from automatic task selection. These native settings do not
replace portable permission enforcement.

Model and reasoning overrides, target selection, metadata, and newer invocation
controls have typed validators; their effective native behavior remains
unverified.

Agent-local `mcp-servers` uses the typed MCP validator. A required agent with an
unknown nested field refuses apply. An optional agent with such a field remains
preserved and inactive as a whole file. Local MCP tests verify child discovery,
tool execution, and parent isolation in both scopes. A child definition can
replace a same-name user server for that child; the parent continues to use its
original definition after delegation. Other source precedence and remote agent
transports need separate native tests.

Import and projection refuse explicit credential fields and credentials in MCP
URL user information. Import refuses before it writes canonical files. It does
not rewrite the agent or silently remove credential fields. Native environment
references remain native values. YAML maps require string keys; merge keys and
unsupported tags refuse. These checks do not scan arbitrary prompt text or
command arguments for secrets.
The [debug report](NATIVE_DEBUG_RESEARCH.md#copilot-agent-projection-and-native-delegation)
contains the source documents, native results, and remaining work. The
[agent MCP review](NATIVE_DEBUG_RESEARCH.md#copilot-agent-local-mcp-and-parent-isolation)
records the separate MCP tests.

## Copilot MCP configuration

The draft.2 `mcp` artifact writes `.mcp.json` in project scope and
`mcp-config.json` in the explicit user home. Project import reads both
`.github/mcp.json` and `.mcp.json`. It keeps disjoint servers and uses the
`.mcp.json` definition when a server name repeats, as Copilot does. Existing
owned settings in the former draft.2 `.github/mcp.json` destination migrate
through the same ownership and rollback code. Unowned settings remain intact.
Stable 1.0 and draft.1 destinations are unchanged.

The adapter validates the documented transport, command, arguments, environment,
working directory, tool list, timeout, URL, headers, OAuth declarations, and
cache controls. Literal commands and arguments can move to the portable core.
Native expressions, transport spelling, and extended fields remain in the native
namespace. The adapter merges them once per server. It refuses incompatible
portable and native transports. Native variable expansion remains separate from
the portable requirement to fail when a referenced environment value is missing.
The portable environment-reference refusal remains in effect.

Project MCP requires native folder trust. Apply does not write trust, credential,
or account state. Authentication and installation remain native operations.
Pinned local and HTTP fixtures verify loading, context, expansion, filtered tool
calls, and returned results. OAuth, SSE, timeout enforcement, and cache behavior
still need separate native evidence. See the
[debug report](NATIVE_DEBUG_RESEARCH.md#copilot-mcp-import-projection-and-source-priority).


## Copilot native hooks

Draft.2 `hooks` artifacts map to `.github/hooks/<name>.json` in project scope
and `hooks/<name>.json` in the explicit user home. Files require `version: 1`.
The `config` artifact also accepts the inline `hooks` field in the registered
project or user `settings.json` target. Import preserves native hook values;
it does not promote direct-executable arguments or environment expressions into
portable command or environment-reference semantics.

Typed fields cover native command, HTTP, and prompt hook entries. Unknown events
stay inactive. If an entry contains an unknown field, its entire ordered event
array stays inactive. A required profile refuses that loss. Known malformed
values refuse projection. Windows-only commands remain inactive on Linux.
Credentials in explicit fields or URL user information refuse import before
any canonical write. An import does not remove part of an ordered hook array.
These checks do not scan arbitrary commands or prompt text for secrets.

The native tests use Copilot `1.0.83`. They verify local command and HTTP hooks,
working directories, literal environment values, direct arguments, tool
matchers, injected context, explicit denial, and timeouts. `${VAR}` remains
literal in the tested direct-executable environment. The shell form and HTTP
header expansion need separate checks before any portable reference mapping.
HTTP requires HTTPS outside localhost. Local HTTP needs the separate native
`COPILOT_HOOK_ALLOW_LOCALHOST=1` opt-in. HTTP permission responses require HTTPS.

Project hook evidence uses CLI prompt mode with preconfigured folder trust.
The tested ACP sessions did not load project hooks; this remains an unresolved
interface difference. User hook evidence uses ACP. The current prompt-mode fixture grants `shell(/usr/bin/python3)` and adds
`/usr/bin` as an allowed directory. It uses a deterministic local provider that
requests only its marker script. Earlier evidence used broader grants. These
hook fixtures do not establish approval enforcement.

`timeoutSec` takes precedence over `timeout`. The alias also works alone.
A timed-out command hook allows the normal permission flow to continue; it
does not enforce mandatory denial. Explicit denial blocks the tested tool.
The file-level `disableAllHooks` switch leaves neighbouring files active.
Changing or removing that switch cannot affect unowned events. Removing owned
settings cannot remove the version required by remaining unowned hook events.
An empty hook file is removed after all its owned settings are removed.
The global settings switch remains blocked pending separate authority checks.

The [hook review](NATIVE_DEBUG_RESEARCH.md#copilot-native-hook-fields-and-authority)
records native results, failed attempts, and remaining event and protocol work.


## Copilot shell-rule interface checks

Pinned Copilot `1.0.83` uses different shell allowance matching in the tested
CLI prompt and ACP paths. Prompt mode grants the absolute executable stem,
such as `shell(/usr/bin/python3)`, but not that executable plus the script
argument. The matching full-command rule grants the ACP call. A stem-only ACP
rule still requests client approval in the fixture. These results do not
establish every form of native shell matching.

A basename is not a substitute for the absolute executable spelling in the
tested rules. Prompt-mode deny rules match the tested full command as well as
the executable stem; a full-command deny rule for a different argument did not
match. Explicit matching denial takes precedence over a general shell grant
in both tested interfaces. File and directory access is a separate permission:
the absolute interpreter needed access to `/usr/bin` in the prompt fixture.

The [shell-rule review](NATIVE_DEBUG_RESEARCH.md#copilot-shell-rule-matching-and-interface-boundaries)
records the exact commands, rule strings, native events, and observable effects.
These are native invocation tests. The adapter does not translate portable
argument-sensitive rules from these results, and mandatory permission refusals
remain unchanged.


## Codex native hooks

Draft.2 maps Codex hook files to `.codex/hooks.json` in project scope and
`hooks.json` in the explicit user home. It also maps inline `[hooks]` in the
registered `config.toml` target. The mapper uses an explicit event and handler
registry plus the pinned Codex `0.154.0` schema.

Command and MCP-tool fields have typed mappings. Native `prompt` and `agent`
handlers remain inactive because the pinned native client parses and skips them.
The same rule applies to MCP handlers for `SessionEnd`. MCP input cannot contain
`null`, including inside arrays or objects: the native parser rejects the whole
file because it converts input through TOML.
An unknown field or skipped handler keeps its entire event array inactive.
A required profile refuses this loss. Unknown root fields keep the entire JSON
file inactive because the native file parser rejects them. The same applies to
`disableAllHooks`, which is not a Codex JSON-file field. Draft.2 refuses the
portable disabled-hooks mapping for Codex; it does not replace it with a global
feature switch. Stable 1.0 and draft.1 projection semantics are unchanged.

Import excludes `hooks.state`. Native trust hashes and per-hook enable state
remain external. Credential-bearing event arrays refuse import before any
canonical write. The importer does not remove one handler from an event array.
It does not scan arbitrary command or prompt strings for secrets.

Apply writes configuration but does not trust hooks. The user must review exact
hook definitions through native controls. Project sources also require native
project trust. Referenced MCP servers must be configured separately. A changed
trusted definition requires a new review. In the pinned app-server fixture,
the top-level `--dangerously-bypass-hook-trust` flag did not cause unreviewed
hooks to execute. Do not treat that flag as proof of effective hook trust across
all interfaces.

Native tests verify project and user file and inline loading, command matchers,
context delivery, denial, and preservation of external trust state. Changed
trusted definitions are marked `modified` and skipped. A user-scope timeout test
verifies a one-second limit and process termination; tool execution then
continues under the fixture's native policy. These hook fixtures use a local
model and preconfigured native policy. They do not prove portable permission
enforcement.

MCP `PreToolUse` and `PostToolUse` tests also pass for files and inline settings
in both scopes. They verify recursive event references, preserved JSON types,
native status metadata, context delivery, and explicit denial. Missing servers,
server errors, missing event references, and timeouts let the operation continue
under the fixture policy. An unlisted tool name reaches the server, which must
handle an unavailable tool. A timeout does not prove that the MCP tool stopped;
the fixture server can complete work after Codex stops waiting.

The MCP connection in these tests is a separate native prerequisite. Remote
transports, elicitation, server-timeout precedence, Windows commands, additional
events, and other interfaces still need tests. See the
[MCP hook review](NATIVE_DEBUG_RESEARCH.md#codex-mcp-hook-execution-and-parser-limits).

### Codex background hooks and context limits

Pinned `0.154.0` tests verify `async` for `PreToolUse` command handlers. The
operation completes while the hook waits. Returned context can reach a later
model step in the same turn or the next user turn. A background denial cannot
block the completed operation. This app-server path emits no background hook
start/completion notifications; the tests correlate native definitions, command
events, hook input, process state, and model context.

A one-second timeout stops the tested hook and its attached child. A child in a
separate process group survives and writes its marker. A timeout does not prove
that all descendant work stops. Thread archive and graceful app-server shutdown
stop the tested attached processes. Unsubscribe leaves the thread loaded during
the native inactivity grace period; it does not mean session shutdown.

`additionalContextLimit` tests cover omitted, positive, and zero values. Large
context is saved under the native temporary directory. The model receives a
shortened message and the saved-file path. A tiny limit can leave only the file
notice. Zero keeps all context in the model input. Each handler has its own
limit. The same spill and unlimited behaviors pass with delayed background
context. Apply does not manage these runtime files, their permissions, or their
cleanup.

Other events, additional native interfaces, spill-write failures, and wider
process topologies still need tests. See the
[background and context review](NATIVE_DEBUG_RESEARCH.md#codex-background-hooks-and-context-spill-limits).

## Provider token commands

Codex native user configuration can use `model_providers.<id>.auth`. Draft.2
preserves this table after complete validation. Apply does not run the helper
or store its output. Native failures can send unauthenticated model requests.
The plan, capabilities, and [command-authentication report](CODEX_COMMAND_AUTHENTICATION.md)
state this limit. No portable authentication guarantee or project behavior is
claimed by this user-scope evidence.

## Codex project restrictions

The [project-scope audit](CODEX_PROJECT_SCOPE.md) verifies 12 root keys and one
feature flag that Codex `0.154.0` removes from trusted project configuration.
Required native profiles refuse these keys before writes. Optional profiles
keep their source but project only the supported fields. Plans, capabilities,
and coverage identify this scope limit. Provider token-command evidence remains
user-scoped; project files cannot select or redefine a provider.

## Codex role files

Agent files use a bounded child override contract. They cannot replace the
parent provider, MCP definitions, or arbitrary session settings. Required
ignored content refuses before writes; optional content stays intact and
inactive as a whole file. Project and user role files can reduce selected
features and skills. These role controls are separate from project
`config.toml` restrictions. See the [native role tests and limits](CODEX_ROLE_OVERRIDES.md).

Import also preserves declared role-file paths and skill selectors inside
agent files across relocation. Mapped assets keep relative references.
External libraries keep absolute references to their source location and
remain outside ownership. See the [reference audit](CODEX_ROLE_REFERENCES.md).

## Copilot skill metadata

Draft.2 checks selected skills separately from canonical Markdown validation.
Plan reports the source, destination, ownership, invocation controls, and known
losses for each skill. Unverified field types and malformed discovery metadata
refuse projection before writes. Known native limitations appear in JSON and
text warnings. Existing project discovery remains outside apply control.
See the [metadata audit](COPILOT_SKILL_METADATA.md). Stable validation rules
and shared `.agents/skills` discovery paths remain unchanged.

Draft.2 project import also reads complete packages from the supplied root's
`.github/skills/` and `.claude/skills/`. It stores them in `.agents/skills/`
and selects `skills`. Original packages remain unchanged. Import, plan, and
apply refuse conflicting packages, including different assets or executable
properties, even with `--force`. An empty canonical skill marker is removed
transactionally; its backup stays under `.agents/state/import-backups/`.
See the [project skill import report](COPILOT_SKILL_IMPORT.md) for scope,
ownership, and native execution evidence.

## Copilot recursive instruction files

Draft.2 preserves nested `*.instructions.md` paths under the registered project
or user instruction directory. Each file retains its bytes and has separate
ownership. The native tests confirm recursive loading with `applyTo: "**"`.
Path-specific files appear in a native catalog that tells the model to read
them. Reading a matching source file does not automatically load their bodies.
User instruction reads outside trusted directories can require native approval;
plan reports this action and apply does not grant trust. See the
[recursive instruction audit](COPILOT_RECURSIVE_INSTRUCTIONS.md).
