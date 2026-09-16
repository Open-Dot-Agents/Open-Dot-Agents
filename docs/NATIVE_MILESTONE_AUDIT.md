# Native milestone audit

This audit follows the universal configuration plan and the later requirements
for plugins, skills, tools, and global configuration. The milestone is not
complete. A zero count of unwritten mappings does not establish production
readiness or native enforcement.

The current inventory retains 1,679 original source rows, 1,396 coverage
records, and 1,379 distinct semantic features. The 17 syntax and alias records do not add
features. The Linux CLI milestone contains 1,355 semantic features.

| Requirement | Current evidence | Remaining work |
| --- | --- | --- |
| Separate stable, draft.1, and experimental draft.2 contracts | Separate schemas and fixtures; combined stable, draft, and Go checks | Keep these checks passing during the remaining work. |
| Source-linked feature dispositions | Reproducible coverage map; zero `mapping-pending`; frozen source hash retained | Review native behavior independently of these counts. |
| Native setting validation | 231 `validator-declared` semantic features | These are declarations of validators, not complete native behavior proofs. Audit activation, scope, precedence, reload, and effects by family. |
| Native security settings | 81 `security-evidence-required` features: 71 Codex and 10 Copilot | Establish safe native-only configurations and scope authority. Refuse conflicts with portable requirements. Do not treat missing tests as native limitations. |
| Scoped ownership and transactions | Go tests cover shared homes, locking, conflicts, private backups, removal, rollback, and unchanged state on refusal | Maintain coverage when adding new settings or assets. Retest affected native paths against the final source. |
| Skills, tools, and plugins | Direct discovery, package preservation, projection, selected native lifecycle receipts, and authenticated public GitHub MCP discovery through Codex and Copilot | Complete broader package families and authorized read-call behavior. Keep the required portable `mcp.envRef` refusal outside the tested package mapping. |
| Global defaults and project overrides | Explicit global source, fixed user core bindings, ownership checks, shared discovery, Codex model override and fallback | The real home still uses the older `dot-agents` format. Prepare explicit migration. Combined portable security remains refused when enforcement is unverified. |
| Mandatory portable approval | Trust-derived shell and image cases, an unattended execution counterexample, compound denials, a five-second pending window, disconnect before response, and detached descendant cleanup after approved execution | Keep mandatory `ask` refused and broaden built-in coverage. The observed pending window is not a universal native timeout. Native `on-request` is separate. |
| Local network denial | Copilot `1.0.83` denies tested host TCP and Unix-socket classes but permits sandbox-local TCP, UDP, and abstract Unix sockets | The exact version-bound native limitation is recorded. Keep the portable `network.local: deny` mapping refused. |
| External authority and operations | Trust, credentials, managed policy, account state, and live sessions remain external | Authentication, installation, scheduling, and execution require their own native prerequisites and evidence. Apply must not perform these operations. |
| Release support and ratification | No new adapter support promotion or ratification | Existing release, compatibility, governance, and pinned native gates remain separate. |

The 231 validator declarations comprise 164 Codex settings, 61 Copilot
settings, and six Copilot MCP fields (`oauthClientId`, `oauthPublicClient`,
`oauthGrantType`, `oidc`, `disableToolCache`, `deferTools`). All Copilot
agent-frontmatter fields have now left this backlog: `model` moved to a
bounded fixture-execution mapping and `reasoningEffort` moved to a native
limitation. The next audit
should separate settings that already have relevant bounded receipts from
settings whose effective behavior is still untested. Schema acceptance alone
must not become an activation claim. The keymap parser defect shows why this
distinction matters.

Eight semantic features currently have version-bound native limitation
dispositions: three Copilot sidekick fields, three Codex settings, Copilot
`sandbox.userPolicy.network.allowLocalNetwork`, and the Copilot
agent-frontmatter `reasoningEffort` field. These have source and native
evidence. This classification must not be extended to the 81 security gates
merely because those mappings need more work.

Codex `mcp_servers.<name>.tools.<name>.output_token_limit` is a resource
limit, not an approval control. The adapter now selects it independently and
continues to block the adjacent `approval_mode`. Plugin marketplace entries
remain atomic because partial package configuration can change package
semantics.

Native receipts retain their exact binaries, fixtures, runners, and source
hashes. Older receipts remain useful historical evidence, but their hashes do
not match every later Go revision. The Workbench reports artifact integrity
separately from current support eligibility. Integrity requires the receipt to
match its captured runner and any declared captured source or fixture files.
Current eligibility also requires the tracked runner, implementation sources,
and helpers to match the receipt. Release gates keep the current-source
requirement; historical integrity cannot promote support. The current
parent-skill checkpoint is
`verification-copilot-parent-skills-final.json`; it is not a substitute for a
final cross-family native verification campaign.

All family verifiers now use this boundary. A regression test rejects direct
current-runner, current-helper, or current-implementation assertions in
historical verifier code. The retained Copilot local-network result has no
captured runner snapshot, so its runner integrity is explicitly unavailable
and it is not current support evidence.

Current eligibility requires a current runner comparison, a nonempty
implementation source map, and a comparison for each declared helper.
Missing comparisons leave historical integrity unchanged but prevent current
eligibility. An empty receipt set establishes neither result. Regression
tests cover these cases and changes to the compared helper files.

Baseline evidence must use the adapter result class. Each extended case must
include all source snapshots and all preflight checks required for its harness
and authentication mode. Tests reject empty and partial maps, missing
prerequisites, and a result class that would bypass adapter assertions. Complete
historical snapshots remain valid for checks outside release mode; release
mode still requires the current source hashes.

Reproduction scripts use the system temporary directory. They resolve native
executables from `PATH`, or from an absolute `CODEX_BIN` or `COPILOT_BIN`
override, and then enforce the recorded SHA256 pin. They do not depend on a
developer-specific workspace or home path.

Remote publication must be checked separately for the root and each component.
Component commits preserve existing user configuration
and historical evidence. Claude Code native testing remains skipped. No runtime
launcher or policy mediation service has been added. These commits do not
establish current native support or release eligibility.

## Draft.2 reproduction checkpoint

The clean source archive at root `0aeff47`, CLI `5b4f3b0`, SPEC `2604896`,
and Workbench `25eda52` passes the 26 stable specification checks, coverage
regeneration, compatibility checks, and pinned project-extension checks.
This archive contains committed files only. It is not a remote clone test.

The uncached Go run fails the alternate-mount check in this execution sandbox.
The 103 Workbench task tests have one error because local socket creation is
blocked. These runs do not establish a passing full suite.

The conformance verifier tests have a separate reproduction defect. They read
ignored receipts under `WORKBENCH/evidence/native-draft2-debug`. The local
workspace passes 126 tests; a clean archive reports missing receipt files.
Separate deterministic verifier fixtures from native evidence artifacts before
claiming that this suite works from a fresh checkout. Keep the existing
mutation assertions. Synthetic fixtures must never count as native evidence.

Of the 234 validator declarations, 37 have native or scope evidence links and
197 have neither. All 234 retain `native_status: unverified`. A link can cover
only one scope or an inactive field; it does not establish current behavioral
support. All 81 security requirements have no native or scope evidence link in
the coverage records. This does not mean that no related probes exist.

The next native run must check binary hashes before execution. The available
Codex command reports `0.154.0`; the available Copilot command reports
`1.0.84-4`, which differs from the required `1.0.83` pin. Do not substitute it
for the pinned campaign. Use an explicitly selected verified binary.

GitHub DNS access remains unavailable from this execution environment.
Remote component reachability and fresh recursive-clone validation remain
open. Claude native tests remain deferred, and no support status changes.

Eight verifier test families now use explicit synthetic inputs: global
configuration, Codex setting gaps, Copilot sidekick, and Copilot root, user,
canonical, GitHub, and agent instructions. All 40 tests pass in the clean
archive without ignored receipts. The existing mutation checks remain, and
the global tests add a valid Copilot case. The full workspace verifier suite
passes 127 tests. At that checkpoint, other verifier families still needed
independent fixtures.
The unchanged global native verifier accepts the retained historical receipts
but reports both vendors ineligible for current support because source hashes
differ. Synthetic unit-test inputs do not change that result.

## Deterministic verifier follow-up

All verifier test families now use explicit synthetic inputs. The remaining
keymap, settings, skill, recursive instruction, trust-boundary, local-network,
and public MCP tests no longer read ignored receipts. Existing mutation checks
remain. The retained skill-refusal regression is checked by the native evidence
verifier instead of the deterministic suite. Production verifier discovery
excludes `_test.py` files and still checks production eligibility reporting.

Run `python3 WORKBENCH/conformance/check_clean_source.py` from the superproject
to copy tracked and new non-ignored source files into a temporary tree and run
the deterministic checks there. The Verify workflow uses this command. It
does not copy Git metadata or ignored local evidence. Before commit, this is
a working-source reproduction check, not a remote checkout check.

The permissions coverage correction uses canonical dotted paths for table-key
syntax and retains old IDs in `previous_ids`. One duplicate syntax record is
merged. The 1,679 source rows, 1,379 semantic features, 234 validator declarations,
and 81 security requirements remain unchanged. No native support is promoted.

On 2026-09-15, the maintainer ran coverage regeneration and the clean-source
check in the host terminal. The supplied output reports all checks passing:
26 stable specification checks, 59 security draft cases, 40 native draft cases,
both Go packages, repository validation, 56 coverage tests, compatibility
consistency, 104 Workbench task tests, and 129 verifier tests. The host run
passes the mount and TLS socket checks that the execution sandbox blocked.
The regenerated coverage file was then checked locally against the reviewed
output; both have SHA-256
`e7bdbd89af11412aa1e70d234a756e95041655dd2456039673a1945bb66799a7`.
Whitespace checks pass for the root, CLI, and Workbench changes. This is local
working-source validation; remote CI and native support were not tested.

## Copilot agent-frontmatter `model` field audit (2026-09-16)

The 234-item validator-declared backlog's smallest family, the two Copilot
agent-frontmatter fields, was reviewed for activation instead of schema
acceptance alone. A new `model-override` scenario in
`WORKBENCH/conformance/run_native_copilot_agent.py` set the parent session's
model and the fixture agent's frontmatter `model` to distinct, unambiguous
values and captured every request the native Copilot 1.0.84-9 binary sent to a
local synthetic OpenAI-compatible provider. The parent turn's request used the
session default; the delegated child turn's request used the agent's own
`model` field, confirming the native binary honors this documented override
rather than only accepting it as valid YAML. `CLI/internal/config/native_copilot_agent.go`
now records the `model` field as `artifact-field-mapping` /
`bounded-fixture-execution` with evidence at
`WORKBENCH/evidence/native-draft2-debug/copilot-agent-model-override-project.json`
and `-user.json` (local, gitignored receipts; not committed). Regenerating
`.agents/features/coverage.json` drops the validator-declared count from 234
to 233 and both `--check` reproducibility and `check_compatibility.py` still
pass with no other change.

`reasoningEffort` remains `validator-declared`/`unverified`. The same
experiment with `reasoningEffort: high` and a model name recognized as
reasoning-capable did transmit a `reasoning_effort` request field, so the
control reaches the wire protocol, but the transmitted value was `medium`
instead of the declared `high` in this synthetic single-turn fixture. This
does not establish a native defect: the documented precedence chain (explicit
per-call value, a `subagents` override in `~/.copilot/settings.json`, the
agent's own field, then the parent session's value, with a policy-dependent
fallback when a declared value "can't be honored") gives several legitimate
reasons an unrecognized synthetic model could fall back to a default effort.
Closing this field needs a fixture built against a model name the harness
recognizes as reasoning-capable with a known default, plus coverage of the
`models` and `modelPolicy` fields the same documentation section defines but
the current inventory does not yet track. This is left open for a follow-up
audit slice rather than claimed prematurely.

## Copilot agent-frontmatter `reasoningEffort` field audit (2026-09-16)

Following the prior entry's open item, a `reasoning-effort` scenario was
added to `WORKBENCH/conformance/run_native_copilot_agent.py` using `gpt-5`
as the agent's frontmatter `model`, a name Copilot 1.0.84-9 recognizes as
reasoning-capable in its offline fixture harness (it is the only tested
name, alongside `gpt-5-codex`, that causes a `reasoning_effort` request
field to be transmitted at all; `o3`, `o3-mini`, `o4-mini`, and tested
Claude model names transmit no such field in this harness). With
`reasoningEffort: high` declared, the delegated child turn's request still
transmitted `reasoning_effort: "medium"`. Repeating the probe with
`reasoningEffort` set to `low`, `medium`, `xhigh`, `minimal`, `none`, and
with the field omitted entirely all produced the identical transmitted
value of `medium`. This is conclusive rather than inconclusive: the
declared value has no observed effect on the transmitted request for this
harness-recognized model, across every declared value and its absence.

`CLI/internal/config/native_copilot_agent.go` now records the
`reasoningEffort` field as `native-limitation` /
`bounded-reasoning-effort-ignored`, matching the vocabulary used for other
confirmed negative findings (for example `bounded-acp-native-ignored` and
`bounded-local-network-mismatch`), with evidence at
`WORKBENCH/evidence/native-draft2-debug/copilot-agent-reasoning-effort-project.json`
and `-user.json` (local, gitignored receipts; not committed). Regenerating
`.agents/features/coverage.json` drops the validator-declared count from
233 to 232 and raises the native-limitation count from 7 to 8; both
`--check` reproducibility and `check_compatibility.py` still pass with no
other change. This closes the Copilot agent-frontmatter field family: both
`model` and `reasoningEffort` now have durable native evidence, one
confirming activation and one confirming non-activation. The `models`
(array) and `modelPolicy` fields remain outside the current inventory and
are a separate, deliberate inventory-completeness gap, not part of this
closure.

## Copilot MCP `timeout` field audit (2026-09-16)

Moving to the next-smallest validator-declared family, the seven Copilot MCP
server fields, a `timeout-exceeded`/`timeout-tolerated` pair of scenarios was
added to `WORKBENCH/conformance/run_native_copilot_mcp.py`. Both scenarios
pair an identical artificial delay (4 seconds) in a local stdio MCP server's
`tools/call` response with a different configured `timeout`: 500ms
(`timeout-exceeded`) or 8000ms (`timeout-tolerated`). The short timeout ends
the call as a failed tool call with `MCP error -32001: Request timed out`;
the long timeout tolerates the identical delay and completes normally with
the tool's result reaching the model. This isolates the `timeout` field's
effect from the server's own behavior and rules out an unrelated failure
cause, since the server still receives and would have answered the call.

`CLI/internal/config/native_copilot_mcp.go` now records the `timeout` field
as `artifact-field-mapping`/`bounded-fixture-execution`, with evidence at
`WORKBENCH/evidence/native-draft2-debug/copilot-mcp-timeout-exceeded-{scope}.json`
and `copilot-mcp-timeout-tolerated-{scope}.json` (local, gitignored
receipts; not committed). Regenerating `.agents/features/coverage.json`
drops the validator-declared count from 232 to 231 and raises the
artifact-field-mapping count from 56 to 57; both `--check` reproducibility
and `check_compatibility.py` still pass with no other change.

Six Copilot MCP fields remain validator-declared: `oauthClientId`,
`oauthPublicClient`, `oauthGrantType`, and `oidc` (all OAuth-related, needing
a mock OAuth-capable HTTP MCP server to test) and `disableToolCache` and
`deferTools` (needing a fixture that changes the server's tool list across
repeated CLI invocations against the same `COPILOT_HOME`, and one that
proves a deferred tool is hidden from initial discovery). These are left
open for follow-up audit slices.
