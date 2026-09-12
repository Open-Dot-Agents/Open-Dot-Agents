# Native milestone audit

This audit follows the universal configuration plan and the later requirements
for plugins, skills, tools, and global configuration. The milestone is not
complete. A zero count of unwritten mappings does not establish production
readiness or native enforcement.

The current inventory retains 1,679 original source rows, 1,397 coverage
records, and 1,379 distinct semantic features. The 18 alias records do not add
features. The Linux CLI milestone contains 1,355 semantic features.

| Requirement | Current evidence | Remaining work |
| --- | --- | --- |
| Separate stable, draft.1, and experimental draft.2 contracts | Separate schemas and fixtures; combined stable, draft, and Go checks | Keep these checks passing during the remaining work. |
| Source-linked feature dispositions | Reproducible coverage map; zero `mapping-pending`; frozen source hash retained | Review native behavior independently of these counts. |
| Native setting validation | 234 `validator-declared` semantic features | These are declarations of validators, not complete native behavior proofs. Audit activation, scope, precedence, reload, and effects by family. |
| Native security settings | 81 `security-evidence-required` features: 71 Codex and 10 Copilot | Establish safe native-only configurations and scope authority. Refuse conflicts with portable requirements. Do not treat missing tests as native limitations. |
| Scoped ownership and transactions | Go tests cover shared homes, locking, conflicts, private backups, removal, rollback, and unchanged state on refusal | Maintain coverage when adding new settings or assets. Retest affected native paths against the final source. |
| Skills, tools, and plugins | Direct discovery, package preservation, projection, selected native lifecycle receipts, and authenticated public GitHub MCP discovery through Codex and Copilot | Complete broader package families and authorized read-call behavior. Keep the required portable `mcp.envRef` refusal outside the tested package mapping. |
| Global defaults and project overrides | Explicit global source, fixed user core bindings, ownership checks, shared discovery, Codex model override and fallback | The real home still uses the older `dot-agents` format. Prepare explicit migration. Combined portable security remains refused when enforcement is unverified. |
| Mandatory portable approval | Trust-derived shell and image cases, an unattended execution counterexample, compound denials, a five-second pending window, disconnect before response, and detached descendant cleanup after approved execution | Keep mandatory `ask` refused and broaden built-in coverage. The observed pending window is not a universal native timeout. Native `on-request` is separate. |
| Local network denial | Copilot `1.0.83` denies tested host TCP and Unix-socket classes but permits sandbox-local TCP, UDP, and abstract Unix sockets | The exact version-bound native limitation is recorded. Keep the portable `network.local: deny` mapping refused. |
| External authority and operations | Trust, credentials, managed policy, account state, and live sessions remain external | Authentication, installation, scheduling, and execution require their own native prerequisites and evidence. Apply must not perform these operations. |
| Release support and ratification | No new adapter support promotion or ratification | Existing release, compatibility, governance, and pinned native gates remain separate. |

The 234 validator declarations comprise 164 Codex settings, 61 Copilot
settings, seven Copilot MCP fields, and two Copilot agent-frontmatter fields.
The next audit should separate settings that already have relevant bounded
receipts from settings whose effective behavior is still untested. Schema
acceptance alone must not become an activation claim. The keymap parser defect
shows why this distinction matters.

Seven semantic features currently have version-bound native limitation
dispositions: three Copilot sidekick fields, three Codex settings, and Copilot
`sandbox.userPolicy.network.allowLocalNetwork`. These have source and native
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
passes 127 tests. Other verifier families still need independent fixtures.
The unchanged global native verifier accepts the retained historical receipts
but reports both vendors ineligible for current support because source hashes
differ. Synthetic unit-test inputs do not change that result.
