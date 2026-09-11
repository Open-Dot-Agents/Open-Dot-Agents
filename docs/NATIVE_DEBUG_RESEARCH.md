# Native configuration debug review

The experimental draft.2 implementation has defects that can change native data,
misclassify configuration, or apply invalid ownership records. The debug changes
address the reproduced defects listed below. They also add a pinned Codex schema,
source-confirmed aliases, safe import merging, and a reproducible coverage map.
The full universal-configuration milestone remains incomplete. Neither native
adapter is promoted by this review.

This review covers the Linux Codex and Copilot CLI configuration path. It keeps
stable 1.0 and draft.1 semantics separate from draft.2. It treats a configuration
write, native configuration load, native discovery, and action enforcement as
different evidence. The distinction is necessary: the pinned Codex client can
start after it warns that it ignored part of a malformed configuration.[^1][^2]

## Findings and corrections

| Finding | Reproduced effect | Correction | Evidence boundary |
| --- | --- | --- | --- |
| JSON numbers passed through floating-point decoding | `9007199254740993` became `9007199254740992` | Decode JSON numbers exactly; use versioned, normalized ownership hashes | Full apply preserves large unknown numbers; this is data preservation, not native support for those fields |
| Dotted TOML keys used string-based path matching | A literal `"agents.enabled"` key could be treated as a nested setting | Match path components and validate Codex structure against the pinned schema | Unknown literal keys cannot activate |
| Weak numeric type checks | An integer declaration accepted `1.5` | Validate integral values, positive bounds, unions, and structured arrays | Go regression tests cover invalid and valid values |
| Incomplete Codex type declarations | Valid aliases failed schema validation | Use the pinned schema plus source-confirmed alias normalization | Two aliases load in the pinned native client |
| Permission filters missed native controls | Default permission and approval fields could pass validation | Expand independent permission checks | Selected portable security still blocks combined draft.2 projection |
| Filters treated names as authority fields | A role name or co-author setting could be excluded | Use field context for resource names, UI bindings, and environment references | Literal authorization header values remain excluded |
| Malformed ownership could represent configuration as an asset | A configuration file could become a whole-file removal target | Validate every record against the target registry and field-pointer contract | Includes records owned by another repository |
| User skill import omitted recognized assets | Native user skills were absent after import | Read recognized skill packages and preserve executable assets | Native skill discovery is separately recorded |
| Backup operations appeared only during apply | Plan did not describe the complete write transaction | Include configuration and ownership backups in plan | Plan/apply action comparison and private file modes are tested |
| Duplicate transaction destinations were not rejected first | Rollback snapshots could refer to an intermediate write | Refuse duplicate destinations before writes | Regression test checks that no target is created |
| Existing canonical roots always blocked import | A second native import could not add disjoint configuration | Merge into existing draft.2 roots under the import lock | Conflicts still refuse; stable and draft.1 trees are not migrated |
| Coverage searched Go source text | Names in unrelated configuration contexts counted as mappings | Read declarations from the compiled CLI | Declaration counts are explicitly not completion counts |

These corrections are implemented in the native projection files under
`CLI/internal/config/`. The regression tests are in `native_debug_test.go`,
`native_test.go`, and the Linux lock tests. Each claim above is narrower than full
adapter conformance. For example, an exact JSON round trip does not prove that
Copilot uses an unknown field.[^3]

## Native schema and source agreement

The vendored Codex schema comes from the `rust-v0.154.0` source tag. Its SHA-256 is
`2e1fcf1cbb20f255c3baca2e174b4a3c954cef577a130587b8935e2d12c8ade6`.
The live official schema had the same hash when checked on 2026-09-10. The schema,
source metadata, and upstream Apache 2.0 license are stored together under
`CLI/internal/config/native_schemas/`. Validation does not fetch remote schema
resources at runtime.[^4][^5]

A generated schema is useful but does not express the complete native loader.
The pinned Rust definition gives `agents.max_concurrent_threads_per_session` the
Serde alias `max_threads`. The memory field `disable_on_external_context` has the
alias `no_memories_if_mcp_or_web_search`. Those aliases are absent from the schema.
The adapter normalizes them in a private validation copy. It retains the original
spelling in source artifacts, target configuration, and ownership records.[^6][^7]

The agent limit has a minimum of one. The schema also records Rust integer widths
as format annotations. The validator applies numeric bounds for these widths;
it does not assume that a generic JSON Schema validator enforces Rust formats.
Structured arrays retain their required fields. For example, a skill control
entry requires `enabled`. A parse result without that required field does not
qualify as a mapped value.[^4]

Reasoning effort is a non-empty, model-defined string in this pinned schema.
The adapter therefore does not impose a shared list of reasoning levels. A test
that initially assumed a closed enumeration was corrected after reading the
schema. Empty strings still fail. This keeps native identities separate from
portable concepts, as the milestone requires.[^4]

The pinned source also marks some old fields as no-ops. For example,
`agents.job_max_runtime_seconds` is retained for compatibility but omitted from
the schema. The adapter does not count this as an active resource-limit mapping.
Other schema/source differences, including MCP legacy fields and flattened
structures, still need a systematic audit. The schema is a validation source,
not a statement that every documented configuration feature is implemented.[^6]

## Native alias load evidence

The new native probe runs Codex `0.154.0` with binary SHA-256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
Each case has an isolated native home and workspace. The probe initializes the
app server, reads configuration with layer information, and creates an ephemeral
thread. It does not copy account credentials or start a model turn.[^2]

| Case | Effective configuration | Native warning | Assessment |
| --- | --- | --- | --- |
| `agents.max_threads = 2` | Canonical limit is `2` | None | Alias loaded |
| Canonical agent limit `= 2` | Canonical limit is `2` | None | Canonical field loaded |
| Both names assigned | Client starts and returns a limit | Duplicate-field warning; malformed role definition ignored | Invalid duplicate must not be silently projected |
| Memory legacy alias `= true` | `disable_on_external_context` is `true` | None | Alias loaded |

The first probe expected the duplicate case to stop startup. It did not. That
failed attempt and its exact runner are retained in `codex-alias-load.json` and
`codex-alias-load.runner.py`. The second probe checks the effective values and the
warning. Its results are in `codex-alias-load-v2.json`. Both attempts remain part
of the evidence. Starting a client does not establish correct configuration.[^2]

The adapter refuses duplicate alias/canonical assignments even when the values
are equal. It checks the merged target as well as each source. This prevents a
source alias from colliding with a canonical field already present in a shared
native home. The refusal follows the draft.2 duplicate-assignment contract; it
does not depend on whether the native loader stops or recovers.[^1][^3]

## Exact values and ownership

Native JSON can contain unowned values that are outside the mapped setting set.
Updating an owned model setting must preserve those values. Floating-point
conversion violated this requirement before native activation was involved.
Exact-number decoding now preserves both large integers and exponent notation.
The apply regression includes an unowned integer above the exact binary64 range
and the valid JSON number `1e999`.[^3]

New ownership records use `json-v2:` or `toml-v2:` hash prefixes. Decimal
normalization compares equivalent numeric values without expanding exponents.
This prevents unnecessary conflicts between forms such as `1`, `1.00`, and
`1e0`, while keeping distinct large integers distinct. Legacy unprefixed hashes
remain readable when the exact legacy value matches. Serialization failures do
not silently produce one common empty hash.[^3]

Ownership validation covers every record in a shared registry. A foreign record
is not exempt from validation because the current repository does not own it.
Targets must be canonical absolute paths in the adapter registry. Configuration
records need valid field pointers; standalone assets use file ownership.
Malformed escapes, external credential fields, invalid hashes, whole-file
configuration ownership, and overlapping parent/child records are refused.[^3]

The same target lock continues to protect ownership reload and apply checks.
Backups now form part of the proposed plan and the write transaction. New backups
and ownership files use mode `0600`. Existing target modes remain unchanged.
Duplicate transaction paths fail before the first write. The static path checks
also reject special files, so a named pipe cannot be used as an ordinary native
configuration file.[^3]

The filesystem follow-up reproduced directory redirection in the old writer.
Native transactions now pin directory handles and use descriptor-relative
operations with no symlink following. Snapshot preconditions detect changed
configuration, ownership, and backup destinations. The injected race and rollback
cases are described below. These tests do not claim atomic visibility across
multiple files or recovery after an uncatchable process failure.[^20]

## Import and preservation

The native importer reads recognized configuration and instruction files. It now
also reads user skill packages from the selected native home. Each package needs
a regular `SKILL.md`; assets must be regular files with safe relative paths.
Native `.system` packages are excluded. Imported scripts retain their executable
bit with private mode `0700`; ordinary new private files use `0600`.[^3]

The separate Codex skill probe confirms discovery of a fixture under
`CODEX_HOME/skills` and a fixture under `HOME/.agents/skills`. The response reports
both as enabled user skills. It also lists native system packages. This supports
the selected user skill path and the decision to keep system packages outside
repository ownership. It does not prove execution of every skill asset or
Copilot discovery behavior.[^8][^9]

Import into an existing draft.2 tree is additive. Equal content is idempotent;
disjoint object fields are merged. Conflicting scalar or array values are
refused, including with `--force`. Existing instructions, portable security
files, manifest requirements, and required native status remain in place.
The importer checks the canonical tree again under its lock. Stable and draft.1
trees are refused rather than implicitly migrated.[^3]

A regression imports a model setting into a draft.2 tree that already selects
mandatory portable permissions and sandbox policy. The policy bytes and
instructions remain unchanged. The resulting tree validates. Applying the
combined native/security configuration still refuses, so successful import does
not remove the enforcement boundary.[^3]

Recognized agent, scoped-instruction, and hook files are now imported. Codex
standalone agent files also have a validated projection path. Later sections
record the Copilot agent and LSP mappings. Native hook fields, agent-local MCP,
automation assets, and other documented files remain incomplete. Optional unknown object fields now
have separate JSON-pointer diagnostics; they no longer suppress known siblings.
Arrays remain atomic if any member is unmapped. Unknown source content is retained,
and required unknown content still refuses the complete projection.[^14]

## Coverage map reliability

The original inventory remains unchanged at 1,679 rows, with SHA-256
`f41f70b081e3920822c125f90a1c27b5bea920061f74510ed20dc49c419ed310`.
The coverage generator checks both the count and the hash. Every source row is
linked exactly once from the semantic coverage map. Tests check these properties
and compare generated output with the checked-in map.[^10]

The first generator classified every Copilot native-help setting as external
state. This included `autoUpdate` and `banner`, which are user settings.
It also searched Go source text for names. A root setting named `model` could
therefore provide false mapping evidence for an agent frontmatter field with the
same name. The corrected generator uses compiled declarations and separates the
configuration contexts.[^10][^11]

Semantic IDs now depend on vendor, context, and canonical name. Category and
implementation disposition are not part of the identity. Source-confirmed
aliases share one feature. Managed model policy and a user model preference have
separate identities. This replaces the initial draft.2 ID rule; consumers of
that early map must account for the format change.[^10]

The map uses `validator-declared` to identify available value validators.
It records native behavior separately as unverified unless separate evidence is
attached. It does not call a schema entry a completed mapping. The current
counts therefore measure dispositions in a bounded corpus. They do not measure
native conformance, full documented coverage, or the fraction of the milestone
that is complete.[^10]

Some documentation rows have weak source context, notably rendered command
reference rows. Further command-context and alias review is needed before using
the map as a complete public compatibility catalogue. A source-linked disposition
is necessary for coverage, but it is not sufficient for implementation completion.

## Security evidence and limits

The retained trust-derived Codex tests omit `approval_policy` and use an isolated,
preconfigured untrusted workspace. Their effective thread reports the untrusted
approval mode; the project model sentinel does not take effect. An explicit
execution-rule test records correlated acceptance and denial. The unattended
attempt does not have a correlated terminal execution event and remains
inconclusive.[^12]

The earlier safe-command and built-in attempts are retained as inconclusive.
A later local-model fixture gives direct evidence for those two classes: safe
`cat` requests approval, but built-in `view_image` reads the image and returns it
to the model without an approval request. The native thread reports `untrusted`,
the approval key is omitted, and project configuration is disabled. Mandatory
portable `ask` therefore remains refused for this tested mode.[^15] Timeout,
disconnect, composed shell commands, and background descendants still require
complete scenarios. Native `on-request` remains a separate native setting.[^12]

The Copilot local-network probe distinguishes host TCP, host filesystem Unix
sockets, host abstract Unix sockets, sandbox abstract Unix sockets, and sandbox
TCP/UDP loopback. The recorded host cases are denied, but sandbox-local transfers
succeed. Those results do not implement portable `network.local: deny`, which
covers all of those classes. This is an evidenced native mismatch for the tested
configuration, not proof that every possible native configuration has the same
limit.[^1][^12]

The native extension path still refuses combination with selected portable
security profiles. Broader Codex profile combinations need native execution
proof before that refusal can be relaxed. Permission and credential filters also
need a full schema-aware audit; passing the targeted regressions does not prove
that every security-sensitive vendor field has been classified correctly.

## Remaining completion requirements

| Requirement | Current evidence | Remaining work or prerequisite |
| --- | --- | --- |
| Stable and draft isolation | Deterministic regression checks | Continue to run these checks with each contract change |
| Native core setting validation | Pinned Codex schema and Copilot declarations | Audit schema/source differences and all configuration families |
| Shared user-home ownership | Conflict, malformed-record, lock, pinned-directory, stale-plan, backup-race, and post-write rollback tests | Continue validation across the remaining artifact mappings |
| Native import | Config, instructions, MCP extraction, user skills, agents, scoped instructions, hooks, draft.2 additive merge | LSP, remaining asset families, and explicit migration workflow |
| Unknown optional content | Object field selection preserves known siblings; arrays stay atomic | Broader configuration-family and schema/source coverage |
| Agents and delegation | Projected Codex agents pass discovery, instruction/model/reasoning selection, delegation, and child-command fixtures in both scopes | Copilot agent-local MCP and remaining controls, resource-limit and lifecycle matrix |
| MCP, hooks, and LSP | Portable MCP/hooks merge; some native fields validate | Native field registries, precedence, reload, execution, and conflict tests |
| Plugins, automation, remote settings | Selected settings have validators | Complete asset mapping and distinguish declaration from installation/runtime state |
| Mandatory approval | Trust-derived shell and image cases are correlated; image read has no approval | Timeout, disconnect, unattended, composition, and descendant matrix; broader built-in coverage |
| Local network denial | Separate Copilot connection classes observed | A mapping that denies every required class, or a precise native limitation |
| Feature coverage | Frozen row links, contextual IDs, compiled declarations | Resolve mapping gaps with implementation and native evidence |
| Adapter promotion | No new promotion | Existing pinned native conformance and release gates still apply |

The unresolved items above are not all external blockers. Most are remaining
adapter implementation and test work. Authentication is an external prerequisite
for some native model-driven scenarios, but not for schema validation, import,
ownership, configuration reads, or discovery probes. Those independent tasks
can continue without a new account operation.

## Deterministic native model follow-up

The new fixture serves Responses events from a loopback HTTP endpoint. It uses
the event format in the pinned upstream Codex test support. Each run has a new
native home and workspace, a fixed binary hash, and no copied credentials. Native
Codex still performs configuration loading, tool routing, delegation, approval
checks, sandbox execution, and event reporting. The fixture supplies predictable
model output; it does not replace those native components.[^16]

This method removed model-response uncertainty from several tests. The first
transport run completed a native turn and exposed the custom agent discovery
marker in the outgoing model request. The subsequent run applied a canonical
agent through the reference CLI before starting Codex. In both user scope and a
preconfigured trusted project, the root thread spawned the named agent. The
child request contained the agent's developer instructions, selected model, and
reasoning value. A command-completion event from that child had exit code zero,
and the fixture file contained the expected marker. The project projection did
not change user configuration.[^17]

The agent validator checks the three documented required fields and uses the
pinned configuration schema for the remaining fields. Independent permission
and authority checks still apply. Unknown or blocked agent fields leave the
whole standalone asset inactive. Native name normalization trims whitespace;
duplicate identities are refused across selected files and existing native files.
An unchanged file owned by the same repository can move during one transaction.
A regression caught and fixed an initial refusal of that valid move.[^14][^18]

The untrusted security cases use the same local provider with the approval-policy
key omitted. A disabled project configuration contains both a model sentinel
and `approval_policy = "never"`; neither becomes effective. The native thread
instead reports the global fixture model and `untrusted` approval policy.
The safe shell case gets an approval request and a correlated declined command
when the client refuses. Explicit-rule cases get the expected approval decision,
terminal command event, and file effect. The built-in image case emits a completed
`imageView` item and includes the fixture image in the next model request, without
an approval request.[^15]

The first safe-command probe expected execution without approval. That expectation
was false, and the failed assessment is retained. A corrected assessment records
the actual requested approval and denial. The built-in case supplies the concrete
counterexample to universal mandatory approval. These results do not establish
behavior for every built-in tool, unattended mode, disconnect, or descendants.
They do not promote either adapter.[^15]

## Filesystem mutation follow-up

A deterministic race test first checks the native target path, then moves its
parent directory and replaces that parent with a symlink. The old atomic writer
followed the new path and changed the outside fixture file. The regression now
requires refusal and unchanged outside content. An isolated reproduction of the
pre-change writer records the outside file changing from `outside` to
`projected`.[^20][^24]

The Linux backend opens each parent with `O_DIRECTORY`, `O_NOFOLLOW`, and
`O_CLOEXEC`. It retains those handles through the transaction. File reads use
`openat` with `O_NOFOLLOW` and `O_NONBLOCK`, followed by a regular-file check.
Writes create private temporary files relative to the pinned parent and replace
targets with `renameat`. Removals and rollback also use the pinned parent. A
fresh directory lookup detects identity changes before writes and completion.
A symlink cannot redirect the rollback to an outside directory.[^21]

Configuration and ownership changes carry the snapshot used to construct the
plan. Import merges carry the canonical snapshot they read. New files and
backups require absence. The transaction checks all preconditions before writing,
rechecks each target before its write, and checks the resulting state before
success. An external editor's newer value therefore cannot silently become the
input to a stale plan. A backup created after planning is not overwritten.[^20]

The failure-injection tests cover completed configuration writes, ownership
writes, backup creation, asset removal, and directory cleanup. One test changes
a parent after the first write: rollback restores the original file through its
pinned directory, leaves the outside sentinel unchanged, and removes the new
state directory. Another test changes a target after the transaction writes it.
The transaction reports failure, restores unaffected operations, and refuses to
overwrite that concurrent edit. It reports the conflicting rollback target;
it does not claim complete rollback in that case.[^20]

Existing mode and ownership identifiers are retained when replacing a file.
If the process cannot preserve the ownership identifiers, replacement fails.
New private files retain their requested private mode. Lock files are opened
relative to checked directory handles, must be regular files, and do not pass
their descriptors through exec. FIFO and symlink-parent lock tests confirm
refusal without creating an outside lock. The transaction also checks the
acquired lock file and directory identities throughout its writes. Replacing the
lock file or its parent does not turn an old lock into authority over a new
unlocked target.[^20][^21]

The projected Codex agent fixtures were rerun after this filesystem change.
Both user and project scope still produce the configured child model request,
a correlated native child-command completion, and the expected fixture file.
The full deterministic validation is recorded separately for this filesystem
follow-up.[^22][^23]

## Verification record

The deterministic verification covers Go tests with the race detector, Go vet,
25 stable conformance checks, 59 draft.1 schema cases, 18 draft.2 schema cases,
74 Workbench tests, compatibility agreement, repository and native-example
validation, coverage tests, JSON parsing, and whitespace checks. Exact commands,
exit status, and output are stored in the accompanying verification record.
Native alias loading, skill discovery, projected agent execution, and approval
observations remain separate evidence. The latest follow-up checks are recorded
separately from the initial debug verification.[^13][^19]

No commit, push, release, or adapter support promotion is part of these changes.
The source inventory and previous failed or inconclusive native attempts remain
available. The implementation state continues to say `complete: false`.

## Sources

[^1]: Open-Dot-Agents. [Native configuration draft.2](../SPEC/spec/1.1-draft.2/SPECIFICATION.md). Current working-tree proposal, reviewed 2026-09-10. Contract and evidence boundaries.
[^2]: Open-Dot-Agents Workbench. [Codex alias load, first attempt](../WORKBENCH/evidence/native-draft2-debug/codex-alias-load.json) and [corrected assessment](../WORKBENCH/evidence/native-draft2-debug/codex-alias-load-v2.json). Native Codex 0.154.0, 2026-09-10. Raw events and runner snapshots are local evidence.
[^3]: Open-Dot-Agents CLI. [Native debug regressions](../CLI/internal/config/native_debug_test.go), [native projection tests](../CLI/internal/config/native_test.go), and [native projection implementation](../CLI/internal/config/native.go). Current uncommitted implementation, reviewed 2026-09-10.
[^4]: OpenAI. [Codex configuration schema, rust-v0.154.0](https://raw.githubusercontent.com/openai/codex/rust-v0.154.0/codex-rs/core/config.schema.json). Retrieved 2026-09-10. [Vendored source metadata](../CLI/internal/config/native_schemas/codex-0.154.0.source.json).
[^5]: OpenAI. [Official configuration schema](https://learn.chatgpt.com/docs/config-schema.json) and [Configuration Reference](https://learn.chatgpt.com/docs/config-file/config-reference.md). Retrieved 2026-09-10. Live documentation can change after this review.
[^6]: OpenAI. [Config TOML definitions, rust-v0.154.0](https://github.com/openai/codex/blob/rust-v0.154.0/codex-rs/config/src/config_toml.rs). `AgentsToml`, `AgentRoleToml`, and legacy no-op definitions. Retrieved 2026-09-10.
[^7]: OpenAI. [Configuration types, rust-v0.154.0](https://github.com/openai/codex/blob/rust-v0.154.0/codex-rs/config/src/types.rs). `MemoriesToml` alias definition. Retrieved 2026-09-10.
[^8]: Open-Dot-Agents Workbench. [Codex skill discovery](../WORKBENCH/evidence/native-draft2-debug/codex-skill-discovery.json). Native Codex 0.154.0, 2026-09-10. Discovery only; no model turn.
[^9]: OpenAI. [Build skills](https://learn.chatgpt.com/docs/build-skills.md). Retrieved 2026-09-10. Native skill layout and discovery guidance.
[^10]: Open-Dot-Agents. [Frozen source inventory](../.agents/features/codex-copilot.json), [semantic coverage](../.agents/features/coverage.json), [generator](../CLI/scripts/native_coverage.py), and [coverage tests](../CLI/scripts/native_coverage_test.py). Current working tree, 2026-09-10.
[^11]: GitHub. [Copilot CLI configuration directory reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference). Retrieved 2026-09-10. User settings, native assets, saved permissions, managed policy, and runtime state.
[^12]: Open-Dot-Agents Workbench. [Draft.2 native observations](../WORKBENCH/evidence/NATIVE_DRAFT2.md), with links to retained raw security runs. Native Codex 0.154.0 and Copilot 1.0.83, 2026-09-10. Earlier runs inspected in this review; not rerun as part of the alias probe.
[^13]: Open-Dot-Agents Workbench. [Native debug verification](../WORKBENCH/evidence/native-draft2-debug/verification.json). Local deterministic commands and results, 2026-09-10.

[^14]: Open-Dot-Agents CLI. [Field selection](../CLI/internal/config/native_selection.go), [selection regressions](../CLI/internal/config/native_selection_test.go), [native agent validation](../CLI/internal/config/native_agents.go), and [agent/import regressions](../CLI/internal/config/native_agents_test.go). Working-tree follow-up, 2026-09-10.
[^15]: Open-Dot-Agents Workbench. [Safe-command failed expectation](../WORKBENCH/evidence/native-draft2-debug/codex-local-trust-safe.json), [corrected safe-command assessment](../WORKBENCH/evidence/native-draft2-debug/codex-local-trust-safe-v2.json), [built-in image read](../WORKBENCH/evidence/native-draft2-debug/codex-local-trust-image.json), [explicit rule acceptance](../WORKBENCH/evidence/native-draft2-debug/codex-local-trust-rule-allow.json), and [explicit rule denial](../WORKBENCH/evidence/native-draft2-debug/codex-local-trust-rule-deny.json). Codex 0.154.0, 2026-09-10. Isolated untrusted workspace and local deterministic provider.
[^16]: OpenAI. [Responses test support, rust-v0.154.0](https://github.com/openai/codex/blob/rust-v0.154.0/codex-rs/core/tests/common/responses.rs). Retrieved 2026-09-10. Open-Dot-Agents [native local-model runner](../WORKBENCH/conformance/run_native_local_model.py) retains each executed runner version beside its result.
[^17]: Open-Dot-Agents Workbench. [User agent execution](../WORKBENCH/evidence/native-draft2-debug/codex-projected-agent-execution-v2.json) and [project agent execution](../WORKBENCH/evidence/native-draft2-debug/codex-projected-project-agent-execution.json). Codex 0.154.0, 2026-09-10. Correlated native child events, effective model request, and file effect.
[^18]: OpenAI. [Custom subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents.md) and [pinned agent-role parser](https://github.com/openai/codex/blob/rust-v0.154.0/codex-rs/agent-roles/src/agent_role_config.rs). Retrieved 2026-09-10. Documented required fields, native name normalization, and config-layer semantics.
[^19]: Open-Dot-Agents Workbench. [Follow-up verification](../WORKBENCH/evidence/native-draft2-debug/verification-followup.json). Local deterministic checks, 2026-09-10.

[^20]: Open-Dot-Agents CLI. [Filesystem race and rollback regressions](../CLI/internal/config/native_fs_linux_test.go). Working-tree follow-up, 2026-09-10. Tests include directory replacement, post-write rollback, stale plans, backup creation races, concurrent edits, special files, and import preconditions.
[^21]: Open-Dot-Agents CLI. [Linux filesystem backend](../CLI/internal/config/native_fs_linux.go), [snapshot checks](../CLI/internal/config/native_fs.go), and [native lock handling](../CLI/internal/config/native_lock_linux.go). Current implementation; no full crash-recovery claim.
[^22]: Open-Dot-Agents Workbench. [User agent after filesystem changes](../WORKBENCH/evidence/native-draft2-debug/codex-pinned-fs-user-agent.json) and [project agent after filesystem changes](../WORKBENCH/evidence/native-draft2-debug/codex-pinned-fs-project-agent.json). Codex 0.154.0, 2026-09-10.
[^23]: Open-Dot-Agents Workbench. [Filesystem follow-up verification](../WORKBENCH/evidence/native-draft2-debug/verification-filesystem.json). Local deterministic commands, output, implementation hashes, and native evidence links.

[^24]: Open-Dot-Agents Workbench. [Pre-change writer reproduction](../WORKBENCH/evidence/native-draft2-debug/filesystem-race-before.json) and [isolated writer source](../WORKBENCH/evidence/native-draft2-debug/filesystem-race-before.writer.go). This reproduces the old writer algorithm and retains the initial failed regression result; it is not a native CLI support test.


## Copilot LSP configuration and native execution

The official [LSP guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/add-lsp-servers)
was checked through the GitHub documentation index on 2026-09-10. It specifies
`lsp-config.json` in the user home and `.github/lsp.json` in the project. The
new `lsp` artifact maps these fixed paths and validates the documented server
fields. It uses the existing setting ownership, lock, snapshot, backup, and
rollback code. Unknown optional fields remain inactive. Required unknown fields
and malformed server definitions refuse activation. Environment references do
not expose literal credentials. Selected portable security requirements still
refuse the unverified combined native path.

The original 1,679-row inventory does not list these LSP server fields. It is
unchanged. The compiled capability report now includes separate LSP artifact
field declarations. The coverage file includes these declarations under
`native_artifacts`; it does not add them to the original semantic completion
count. Request timeout behavior is tested separately below.

The [project fixture](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-projected-project-v3.json)
and [user fixture](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-projected-user.json)
passed with Copilot `1.0.83`, SHA256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.
The [runner](../WORKBENCH/conformance/run_native_copilot_lsp.py) uses an isolated
native home, a local OpenAI-compatible provider, and a fixture stdio server.
It copies no credentials. The provider selects one deterministic LSP hover
call. Copilot performs discovery, command and argument expansion, process
launch, LSP initialization, and the tool call.

Both fixtures correlate the tool-call ID with a completed native event and the
returned hover marker. Server logs record the expanded environment, configured
root URI, exact initialization options, opened file URI and language ID, and
hover request. The result appears in the next model request. Project apply
leaves user configuration hashes unchanged. These observations prove the
listed fixture behavior. They do not prove arbitrary language-server behavior,
timeout behavior for other methods, hot reload, precedence between competing native sources,
or combined portable sandbox enforcement.

Failed and incomplete attempts remain available:

- [ACP slash command](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-initial.json):
  session startup succeeded; `/lsp` was absent from the advertised commands and
  the prompt reached its deadline.
- [Local-provider discovery](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-local-initial.json):
  confirmed the native `lsp` tool schema; did not issue an LSP call.
- [Direct native hover](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-local-hover.json):
  confirmed initialization and hover before adapter projection tests.
- [First projection attempt](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-projected-project.json):
  runner used unsupported positional CLI arguments; apply refused before writes.
- [Second projection attempt](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-projected-project-v2.json):
  runner used `harnessVersion` instead of `harness_version`; profile validation
  refused before writes.

The Go regression tests cover both scopes, additive import, exact JSON number
preservation, unknown fields, malformed configuration, literal credential
exclusion, environment references, portable policy refusal, shared ownership,
`--force` conflict refusal, removal, and private new-file permissions.

The official document is retained as a [source snapshot](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-docs.md)
with [URL, time, and hash](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-docs.source.json).

A second fixture compares the same five-second hover delay with two timeout
settings. The project [30-second control](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-delayed-project.json)
returns the marker after about 5.05 seconds. The
[one-second limit](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-timeout-project-v2.json)
returns an empty hover result after about 1.06 seconds. Both events use the
same expected tool-call ID. The [user control](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-delayed-user.json)
and [user limit](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-timeout-user.json)
also pass. Each run records the native request time and prompt completion time.
The delayed marker is absent from the model request when the short limit applies.

Copilot reports `No hover information available at this position.` with a
completed status when this request times out. It does not expose a specific
timeout event through ACP. The
[first timeout expectation](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-projected-timeout-project.json)
failed because it required such an event. The comparative result proves bounded
request timeout behavior, not an explicit native timeout diagnostic.


The final hover reruns use the current implementation in
[project scope](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-final-project-v2.json)
and [user scope](../WORKBENCH/evidence/native-draft2-debug/copilot-lsp-final-user-v2.json).
The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-lsp-final.json)
contains the current file hashes and check output. The earlier
[verification record](../WORKBENCH/evidence/native-draft2-debug/verification-lsp.json)
retains an incorrect test expectation of eight LSP fields; the documented
configuration has seven fields. The corrected coverage test passes.

This step completes the LSP artifact mapping and its bounded native tests.
Remaining Copilot agent fields, broader hooks and MCP fields, other native configuration
families, and the full security interaction matrix remain implementation work.
The universal milestone and adapter promotion gates remain incomplete.

The LSP initialization payload also exposed a filter ambiguity: compiler fields
such as `tokenTypes` are not credentials. The LSP payload filter now excludes
explicit credential keys, including `apiKey`, without rejecting ordinary fields
that contain similar words. Go tests cover credential refusal and preservation
of `tokenTypes` and `authenticationMode`. These fields are also included in the
final native fixture. This payload exception does not change settings, authority
stores, or portable security checks.


## Copilot agent projection and native delegation

The [CLI reference snapshot](../WORKBENCH/evidence/native-draft2-debug/copilot-command-reference-agent-docs.md)
and [common agent reference snapshot](../WORKBENCH/evidence/native-draft2-debug/copilot-agents-docs.md)
were checked through the official GitHub index on 2026-09-10. Their corresponding
`.source.json` files record URLs, retrieval times, and hashes. The common
reference permits `.md` and `.agent.md` names and a string or array for `tools`.
The CLI reference also defines `reasoningEffort` and documents fallback when a
model or reasoning override cannot be honored.

The adapter now imports and projects Copilot agent files in both scopes. It
preserves Markdown bytes and validates YAML with `go.yaml.in/yaml/v3` version
`3.0.5`. Unknown fields, unsupported tags, duplicate fields, invalid types,
and agent-local MCP configuration keep the whole artifact inactive. Required
artifacts refuse apply in these cases. The ownership registry now recognizes
both Markdown suffixes. Collision checks compare display names and normalized
filename stems. They preserve unrelated native files with unknown nonidentity
fields. Existing portable security checks remain in place.

The [project execution](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-final-project.json)
and [user execution](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-final-user.json)
fixtures pass with Copilot `1.0.83` and the same binary SHA256 used by the LSP
fixtures. The [runner](../WORKBENCH/conformance/run_native_copilot_agent.py)
projects the configuration before it starts Copilot. It uses an isolated home
and a local deterministic provider; it copies no credentials.

Copilot adds the display name and description to the `task` tool schema. The
child model request includes the agent prompt marker and only the Bash tool
group. The fixture approval client permits only `/usr/bin/python3` with the
exact fixture script path. A completed command event includes the native child
agent ID. The script writes `ODA_NATIVE_CHILD_EFFECT`, and the completed parent
task returns the child result. Project apply leaves user configuration hashes
unchanged.

Additional bounded checks pass:

- [`infer: false`, project](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-no-infer.json)
  and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-no-infer-user.json):
  the agent is absent from automatic task selection and no child runs.
- [`tools: []`](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-tool-none.json):
  the provider deliberately requests the absent Bash tool. Copilot emits a
  correlated failed tool event and the marker is absent.
- [String-form tools](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-tool-string.json):
  `tools: bash` exposes the Bash group and the child command completes.

The [first direct probe](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-initial.json)
and [second probe](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-allow.json)
retain rejected commands. The first used the wrong approval decision label;
the second used a shell command that did not match the client's exact Python
probe allowlist. The [corrected probe](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-exact-allow.json)
uses that allowlist and produces the correlated marker and completion events.

These results do not prove model or reasoning override behavior. Both the parent
and child use the same fixture model, and the provider requests contain no
reasoning option. The coverage map therefore keeps model, reasoning, metadata,
target selection, and newer invocation controls as validator declarations with
unverified native status. Agent-local MCP remains mapping work. The model and
reasoning defaults are not shared portable identities or scales.

The Go tests cover field and YAML errors, unchanged bytes, both scopes,
import/apply preservation, idempotence, `.md` ownership, display-name and filename
collisions, and preservation of unrelated native agents. The final check record
is [verification-copilot-agent.json](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-agent.json).
It supersedes the earlier LSP-only file hashes. The complete universal milestone
and adapter support gates remain incomplete.


## Copilot MCP import, projection, and source priority

The [MCP guide snapshot](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-docs.md)
and the earlier [full CLI reference snapshot](../WORKBENCH/evidence/native-draft2-debug/copilot-command-reference-agent-docs.md)
provide the field definitions and source order. The
[native config help](../WORKBENCH/evidence/native-draft2-debug/copilot-native-config-help.txt)
also records the installed CLI's configuration guidance. The binary remains
Copilot `1.0.83`, SHA256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.

The new regression first reproduced
[rejection of the documented `type: local`](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-local-import-before.json).
The importer previously sent that native spelling through the older portable
converter. It now extracts literal commands, arguments, and applicable HTTPS
URLs into the portable core, while it retains the explicit native transport and
extended fields. The projection merges these parts and refuses incompatible
transports. Empty argument arrays, exact numbers, native expressions, and unknown
optional fields survive import. Unknown optional fields stay inactive; required
unknown fields refuse apply. Malformed known fields and literal credential fields
also refuse projection.

Native command and argument expressions stay in the native namespace because
portable commands do not have those expansion semantics. Native environment and
header values stay there for the same reason: successful expansion does not prove
that a missing source blocks activation. The existing portable reference refusal
is unchanged and has a dedicated regression test. MCP content in `settings.json`
remains unmapped and inactive; import no longer promotes that ignored root into
an active portable server.

### Native source and trust checks

The initial [plugin listing](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-paths.json)
and [listing with misplaced trust fields](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-paths-trusted.json)
found only the user source. ACP attempts with
[`.github/mcp.json`](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-session-github.json)
and [`.mcp.json`](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-session-root.json)
also omitted the project tools, including the
[experimental `.github` attempt](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-session-github-experimental.json)
and [experimental root attempt](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-session-root-experimental.json).
Those fixtures put `trustedFolders` in `settings.json`. The native trust state
belongs in `config.json`.

With trust preconfigured in the isolated native state, both
[`.github/mcp.json`](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-session-trusted-state-github.json)
and [`.mcp.json`](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-session-trusted-state-root.json)
load and complete the MCP call. Two intervening recording attempts ran to
completion but refused to replace the existing evidence filenames; the retained
reruns use distinct filenames. The
[precedence fixture](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-precedence-github.json)
then defines the same server in both files. Only the root `.mcp.json` server gets
the handshake and tool call. The lower-priority server has no process events.

This result changes the draft.2 destination to `.mcp.json`. The old destination
was a valid source, but a root source could hide changes written there. Project
import now reads the two fixed paths and merges definitions in native priority
order. The ownership registry still accepts the former `.github/mcp.json` path
for transactional removal of owned settings. The migration test preserves an
unowned server in that file and verifies an empty follow-up plan. Stable 1.0 and
draft.1 semantics remain unchanged. No recursive native-home copy was added.

### Projected native execution

The [runner](../WORKBENCH/conformance/run_native_copilot_mcp.py) imports from an
isolated source, applies the canonical result to a fresh target, and then starts
Copilot with a local deterministic provider. It configures trust only as separate
native test setup. It checks that apply leaves the trust file unchanged. It also
checks that project apply leaves all user configuration hashes unchanged.

The [final project run](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-final-project.json)
and [final user run](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-final-user.json)
pass. The MCP server logs the configured working directory, literal and expanded
environment values, protocol initialization, tool discovery, and the tool call's
marker. A completed native event has the expected tool-call ID and returned
marker. The next model request contains that result. The initial projected
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-projected-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-projected-user.json)
runs remain available.

The [tool filter probe](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-blocked-project.json)
deliberately requests an excluded tool. Copilot emits a correlated failed event,
and no tool call reaches the MCP server. The
[command and argument expansion probe](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-expansion-user.json)
also passes. Local HTTP fixtures pass in
[project scope](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-http-project.json)
and [user scope](../WORKBENCH/evidence/native-draft2-debug/copilot-mcp-http-user.json).
They verify the handshake, filtered call, and literal and expanded headers on
all recorded HTTP requests. They do not test TLS, SSE, OAuth, or external services.

### Checks and remaining work

Go regressions cover local and remote import, both scopes, native field
preservation, malformed values, credential exclusion, unknown required fields,
portable reference refusal, source priority, old-path migration, shared-home
ownership, `--force` refusal, and removal. Self-review also found that an artifact
with no selected values could claim `projected` status. Such artifacts now report
`inactive`, including the unknown-only MCP regression case.

The compiled capability report and coverage map distinguish these bounded
native results from validator-only fields. OAuth declarations, SSE, timeout
enforcement, and cache/defer controls need separate native tests. Agent-local
MCP still needs its own projection and isolation evidence. The larger permission,
approval, and native feature matrices remain incomplete. The original 1,679
source entries and their SHA256 remain unchanged.

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-mcp.json)
contains the implementation hashes at the end of that MCP step, native evidence references, and
check outputs. Earlier records remain historical evidence.


## Copilot agent-local MCP and parent isolation

The follow-up resolves the inactive agent-local MCP mapping. The
[refreshed CLI reference](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-docs.md)
and [fetch metadata](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-docs.source.json)
confirm the documented `mcp-servers` field and native MCP schema. Native tests
use Copilot `1.0.83`, SHA256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.

Before changing activation, the
[direct isolation probe](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-direct-project.json)
verified that the child loads and calls its own local MCP tool while the parent
has no such tool. The
[direct override probe](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-direct-override-project.json)
verified that a child can override a same-name user server. The parent calls its
original server before and after delegation. These direct probes test native
behavior; they do not test the adapter.

### Validation and import fixes

The agent validator now uses the typed native MCP selector and complete-server
checks. The whole agent stays inactive if any nested MCP field has no mapping.
A required agent refuses apply in that case. The adapter never copies an
unfiltered agent after validating only a selected subset. Unknown optional
noncredential content remains preserved on import.

Agent import previously copied Markdown without checking credential fields.
Import now reads the YAML frontmatter and refuses explicit credential fields,
including nested MCP environment and header credentials and URL user
information. Projection uses the same check. The error omits credential values.
Import stops before the transaction writes files; it does not remove selected
fields or rewrite the agent. Native references such as `${KEY}` remain native
values. Arbitrary prompt text and command arguments are not secret-scanned.

YAML map keys must be strings. Merge keys refuse. This prevents non-string maps
from bypassing recursive field checks. Tests also cover credential fields in
metadata and a second `mcpServers` root, so normalization cannot hide excluded
values. Malformed YAML, duplicate fields, wrong types, unsupported tags,
transport conflicts, and unknown nested fields have regression tests.

An initial unit fixture attempted to project over an unowned source agent and
met the existing ownership refusal. The round-trip test now uses a fresh target,
as the native runner does. It does not relax ownership checks to accept import.

### Projected native results

The [runner](../WORKBENCH/conformance/run_native_copilot_agent_mcp.py) imports a
native agent, applies it to a fresh target, imports the result again, and checks
exact source bytes. It then starts the native session with isolated local model
and MCP fixtures. Trust is preconfigured as native test state; apply leaves that
file unchanged. Project apply also leaves all user configuration unchanged.

The [project isolation run](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-final-project.json)
and [user isolation run](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-final-user.json)
pass. Child instructions, exposed tool schema, child tool-call ID, native child
agent ID, server request marker, completed event, and returned model input agree.
The agent-only tool is absent from parent model requests before and after the
child. The initial
[projected project run](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-projected-project.json)
remains available as historical evidence.

The [project override run](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-override-project.json)
and [user override run](../WORKBENCH/evidence/native-draft2-debug/copilot-agent-mcp-override-user.json)
also pass. The child server receives only the child call. The same-name user
server receives the parent calls before and after delegation. Parent and child
return different fixture markers. This establishes the tested user-server
precedence and isolation only. It does not establish precedence against every
project, plugin, or managed source.

### Verification and remaining work

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-agent-mcp.json)
records the current source hashes and executed checks. It covers Go race tests,
vet, stable and both draft conformance suites, Workbench tests, compatibility,
coverage, repository validation, JSON, report links, and whitespace. The compiled
capabilities and coverage map now identify the separate agent-local MCP evidence.
The frozen 1,679 source entries remain unchanged.

Remote agent MCP, OAuth, timeout/cache/defer behavior, broader hook fields,
agent model and reasoning overrides, invocation controls, and the full approval
matrix still need work. The full universal configuration milestone remains
incomplete. Adapter support and release gates remain unchanged.


## Copilot native hook fields and authority

The [refreshed source](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-docs.md)
and [fetch metadata](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-docs.source.json)
cover file and inline hooks, direct executables, command fields, HTTP fields,
matchers, native decisions, and timeout behavior. Tests use Copilot `1.0.83`,
SHA256 `a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.

### Failed probes and interface differences

The initial [project ACP file probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-project.json)
and [project ACP inline probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-inline-project.json)
loaded no hooks. Neither
[tracking the project file](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-tracked-project.json)
nor [the prompt-mode repository-hook opt-in](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-opt-in-project.json)
changed that result. Each fixture had isolated native folder trust. This is an
unresolved ACP difference, not proof that every ACP path lacks project hooks.

The first [user file probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-user.json)
loaded hooks but failed its environment-expansion expectation: `${ODA_HOOK_INPUT}`
remained literal in the direct executable's environment. Direct arguments also
remained literal. The corrected
[user inline probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-inline-user.json)
passed. The adapter preserves these native values without claiming portable
fail-on-missing environment-reference semantics.

The [first prompt-mode project probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-prompt-project.json)
loaded project hooks, but its tool was denied. Moving the script
[inside the workspace](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-prompt-workspace-project.json),
adding an [exact normalized command rule](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-prompt-normalized-project.json),
and [granting paths](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-prompt-paths-project.json)
still did not grant the command. The cause of those exact-rule failures needs a
separate permission test. A
[shell-granted prompt fixture](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-direct-prompt-granted-project.json)
passed. It uses a deterministic local provider that requests only the isolated
marker script. That fixture tests hooks, not native approval acceptance.

### Implementation and safety fixes

A dedicated validator now maps native hook files and inline `hooks` settings.
Command, HTTP, and prompt fields have explicit type and compatibility checks.
Unknown events stay inactive. An unknown member or field keeps its entire
ordered event array inactive; the adapter never removes one member and activates
the rest. Required profiles refuse unknown content. Known malformed values
refuse projection. Windows-only commands stay inactive on Linux. HTTP requires
HTTPS outside localhost, and HTTPS is required for permission decisions or
`allowedEnvVars`. Local HTTP remains a separate native opt-in.

Hook import now refuses explicit credentials before the transaction writes.
The previous generic filter could remove a credential-bearing event array.
Refusal preserves the source and avoids changing hook order or decisions.
Native references remain unchanged. Arbitrary command and prompt strings are
not secret-scanned.

The review found that inactive fields could retain `pending native reload` and
inherit a positive native status from their parent artifact. Inactive fields
now report `inactive` and `unverified`, without inherited positive evidence.
The same correction applies to refused agent content.

A hook-file switch affects every event in its file. Owning only
`disableAllHooks` does not grant authority over unowned events. Changes and
removal now check that dependency. The global settings switch remains blocked.
Removal also refuses to leave remaining hook events without their required
version. Files are removed when all owned settings are removed and no data
remains. Portable and native events merge once per target; the adapter reuses
the explicit native format version and refuses a native disable switch that
would disable selected portable hooks.

### Projected native results

The [runner](../WORKBENCH/conformance/run_native_copilot_hooks.py) imports from a
separate native source and applies to a fresh target. Final runs import again
and compare the hook configuration. It checks that apply leaves native trust
state unchanged and that project apply leaves user configuration unchanged.

The [final project run](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-final-project.json)
and [final user run](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-final-user.json)
verify shell-field precedence, direct executable arguments, working directory,
literal environment values, matching and excluded tools, and injected context.
Hook payload session IDs and tool arguments correlate with the native tool
completion and marker file. The initial projected
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-projected-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-projected-user.json)
runs remain available. Inline settings pass in
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-inline-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-inline-user.json)
scope.

Explicit denial prevents the tool effect in
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-deny-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-deny-user.json)
scope. User ACP denial occurs before a permission prompt. The timeout probes in
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-timeout-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-timeout-user.json)
scope terminate a five-second hook after one second. The hook does not finish,
but the tool runs through the normal permission flow. The same result holds
with the alias alone in
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-timeout-alias-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-timeout-alias-user.json)
scope. These results exclude timeout as mandatory denial enforcement.

File-level disable tests pass in
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-disabled-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-disabled-user.json)
scope: only the neighbouring file runs. Local HTTP post-tool hooks pass in
[project](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-http-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-http-user.json)
scope. The HTTP handler records the session, tool arguments, and literal header;
its returned context reaches the next model request. These fixtures use the
native localhost opt-in and do not test TLS or header expansion.

### Verification and remaining work

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-hooks.json)
records source hashes and all executed repository checks. Tests cover typed
fields, unknown ordered arrays, import refusal, idempotence, removal, portable
merging, unowned event protection, foreign ownership, global-control refusal,
and inactive-status accuracy. Coverage separates file fields from inline hooks.
The frozen 1,679-entry inventory remains unchanged.

Remaining work includes project ACP discovery, exact shell-rule behavior,
additional hook events, prompt hooks, TLS and header expansion, HTTP permission
decisions, global hook-control authority, Codex native hooks, and the broader
approval and native feature matrices. The full milestone remains incomplete.
No support or release gate is promoted by these results.


## Copilot shell-rule matching and interface boundaries

This follow-up resolves the exact-command expectation from the hook review.
The [allowing-tools reference](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-rules-docs.md),
[fetch metadata](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-rules-docs.source.json),
and [pinned CLI help](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-rules-help.txt)
provide the documented syntax. The help says that matching is usually against
the command name. The tests below use Copilot `1.0.83`, SHA256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`.

### Native matching results

The [runner](../WORKBENCH/conformance/run_native_copilot_shell_rules.py) uses an
isolated home, a trusted fixture repository, and a deterministic local model.
The model requests one generated command. Each result retains the native
command, rule flags, tool-call ID, terminal event, model input, and marker-file
observations. Path grants are separate explicit test inputs. No portable
security profile is projected.

| Tested condition | Observed result | Evidence |
| --- | --- | --- |
| Prompt allowance includes executable and script argument | Denied | [Python](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-absolute-exact.json), [touch](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-touch-exact.json) |
| Prompt allowance uses the absolute executable stem | Allowed | [Stem](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-absolute-stem.json), [prefix](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-absolute-prefix.json) |
| Basename allowance applied to an absolute executable | Denied | [Absolute command](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-basename-stem.json) |
| Basename allowance applied to the same basename command | Allowed | [Basename command](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-command-basename.json) |
| Stem denial with a general shell grant | Denied | [Stem denial](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-deny-stem.json) |
| Full-command denial with a general shell grant | Denied | [Full-command denial](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-deny-full-command.json) |
| Full-command denial specifies a different argument | Allowed | [Other argument](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-deny-other-argument.json) |
| Basename denial applied to an absolute executable | Allowed | [Basename denial](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-deny-basename.json) |
| ACP allowance matches the full command | Allowed without a client permission request | [Full-command ACP](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-full-command-acp.json) |
| ACP allowance specifies only the executable stem | Client request; client denial blocks execution | [Client denial](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-client-denial-acp.json) |
| ACP command uses a different script argument from its allowance | Client request; client denial blocks execution | [Alternate argument](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-alternate-operand-acp.json) |
| ACP matching full-command denial with a general shell grant | Denied without a client permission request | [Denial precedence](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-denial-precedence-acp.json) |

Several early expectations were wrong. Those records remain failed tests,
with their observed effects intact. They are not relabelled as passing runs.
The first [ACP stem probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-stem-acp.json)
also assumed that a rejected tool must return to the model. ACP instead returned
a correlated `rejected` result and ended the turn. The runner now accepts that
specific denial path. It does not treat arbitrary tool failures as denial.

These results require interface-specific interpretation. Prompt allowance and
denial also do not have identical matching behavior in the tested forms.
The results do not prove every executable, shell grammar, or argument pattern.

### Path permissions and hook-fixture correction

The [workspace-only path probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-workspace-paths.json)
denied the absolute interpreter despite its command-stem allowance. Adding
`/usr/bin` in the [executable-path probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-executable-path.json)
allowed it. This establishes the tested executable path requirement without
granting all paths. A [context-only hook probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-context-hook.json)
also passed with the stem allowance, so that hook response did not cause the
original command-rule mismatch.

The first [hook stem probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-stem-project.json)
removed the path grant and failed. Restoring paths in the
[intermediate hook probe](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-stem-paths-project.json)
passed. The hook runner now uses `shell(/usr/bin/python3)` plus `--add-dir /usr/bin`.
The [final scoped-grant hook run](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-scoped-grant-project.json)
and [explicit hook-denial run](../WORKBENCH/evidence/native-draft2-debug/copilot-hooks-scoped-grant-deny-project.json)
both pass. The former broad shell and path grants are no longer required by the
current runner. The interpreter grant is not an argument-level restriction;
the local deterministic provider limits this fixture to its marker script.

### Composed commands

With one executable stem granted, two commands that use that stem both run in
the [same-stem composition probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-compound-same-stem.json).
When only the first of two different executables is granted, neither marker is
written in the [prompt-mode mixed probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-compound-mixed.json).
Granting both through the native general shell rule produces both markers in
the [positive mixed probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-compound-granted.json).
The corresponding [ACP mixed probe](../WORKBENCH/evidence/native-draft2-debug/copilot-shell-compound-mixed-acp.json)
requests permission; client denial prevents both effects. This tests `&&`
composition only. Other operators, substitutions, timeouts, disconnects, and
background descendants still need separate evidence.

### Verification and remaining work

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-shell-rules.json)
records current source hashes, check results, native outcomes, and the retained
failed expectations. The recorded native source and module hashes still match the hook verification.
The current Go suite, changed Python runners, report links, JSON, coverage,
Workbench tests, repository conformance, and compatibility are checked again.

This resolves the earlier unexplained exact-rule failure and narrows the hook
fixture grants. It does not change portable permission refusals or adapter
support. Project ACP hook discovery, additional shell and approval cases,
remaining native configuration families, and the full milestone remain open.


## Codex native hooks and external trust

The official Codex index routes to
[OpenAI's hook reference](https://learn.chatgpt.com/docs/hooks). The
[archived source](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-docs.md)
and [fetch metadata](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-docs.source.json)
retain the document and index hashes. Native tests use Codex `0.154.0`, SHA256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.

### Trust and parser research

Initial [user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-direct-user.json)
and [project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-direct-project.json)
app-server probes ran the model's tool but no hooks. Setting the
[feature flag explicitly](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-enabled-user.json)
did not change that result. The
[listing probe](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-listed-user.json)
then proved discovery: native `hooks/list` returned all four fixture definitions,
with `enabled: true` and `trustStatus: untrusted`. The top-level
`--dangerously-bypass-hook-trust` flag did not make those hooks execute through
this tested app-server path. This does not prove its behavior in every native
interface.

The fixture now trusts only definitions from its exact generated source path,
with commands that call its reviewed script. It reads native keys and hashes,
records those in the isolated native user state, restarts the app-server, and
checks `trusted` status. The direct trusted
[user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-trusted-user.json),
[project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-trusted-project.json),
and [inline-user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-trusted-inline-user.json)
probes pass. No account credentials or external model are used.

An initial [unreviewed probe](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-unreviewed-user.json)
also recorded no hook effects, but lacked the later discovery check. The initial
[disabled-field probe](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-disable-field-user.json)
was similarly inconclusive about that field. The
[listed disabled-field probe](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-trusted-disable-field.json)
provided the decisive warning: `disableAllHooks` is an unknown JSON root field,
so native parsing rejects the whole file.

The [parser matrix](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-parser.json)
distinguishes root and nested handling. Unknown root fields reject the file.
Unknown event, matcher-group, and command-handler fields do not stop the valid
command from being listed. A `prompt` handler is skipped with a warning. An
invalid `command_windows` value rejects parsing, which confirms native alias
deserialization. The mapper validates that alias without changing source spelling.
Windows command execution remains untested.

### Mapping and authority fixes

The draft.2 adapter now has a dedicated Codex hook selector. It checks known
events, matcher groups, and command or MCP-tool handler fields against the
pinned schema. Unknown or skipped members keep the whole event array inactive.
Required profiles refuse that loss. Unknown JSON root fields keep the whole
file inactive, so removing such a field cannot activate a previously rejected
file. Known numeric and structural errors refuse file projection.

Native `prompt` and `agent` handlers remain inactive. The adapter does not use
permissive schema parsing as an activation map. Draft.2 also refuses the portable
`disableAllHooks` mapping for Codex. Replacing it with `[features].hooks = false`
could affect other sources, so the adapter does not make that substitution.
Stable 1.0 and draft.1 projection semantics remain unchanged.

Import excludes the entire `hooks.state` subtree. Both native trust hashes and
per-hook enable state remain external. Credential-bearing event arrays refuse
import before the transaction writes canonical files. No handler is silently
removed from a source array. Arbitrary command and prompt text is not
secret-scanned. Apply reports the separate native review and trust action and
does not grant it.

Unknown-only artifacts now also report `unverified` without inherited positive
evidence. This closes the remaining artifact-level status issue after the
previous field-level correction. Hook event arrays remain units of ownership.
Two repositories can own disjoint Codex events in the same native home;
`--force` cannot replace the other source's event. Removal preserves the other
source and removes an empty file when no settings remain.

### Projected native results

The [runner](../WORKBENCH/conformance/run_native_codex_hooks.py) imports hooks,
applies them to a fresh target, configures native fixture trust separately,
applies again, and checks that native trust state is unchanged. Final runs
import again and compare hook values. User-scope reimport excludes trust state.
Project apply leaves the user configuration unchanged.

The [final project run](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-final-project.json)
and [final user run](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-final-user.json)
pass. Hook payload session and tool-call IDs, exact tool arguments, native
completion events, model context, and marker-file effects agree. The earlier
projected [project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-projected-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-projected-user.json)
runs remain available. Inline projection also passes in
[project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-inline-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-inline-user.json)
scope.

The explicit review-gate tests in
[project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-review-required-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-review-required-user.json)
scope list the definitions but execute none without trust. The changed-definition
runs in [project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-modified-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-modified-user.json)
scope apply a changed command without changing trust state. Native listing marks
one hook `modified` and the other three `trusted`. Only the unchanged matching
hooks run. The changed SessionStart hook contributes no context.

Denial tests pass in
[project](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-deny-project.json)
and [user](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-deny-user.json)
scope: the native denial reaches the next model request and no tool effect is
written. The [user timeout probe](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-timeout-user.json)
returns after the one-second limit, leaves no active timed-out script process,
and allows the tool to run under the fixture's native policy. It does not
establish mandatory permission enforcement or every timeout case.

### Verification and remaining work

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-codex-hooks.json)
records current hashes and executed checks. Go regressions cover numeric bounds,
unknown fields and arrays, skipped handlers, aliases, credential refusal,
external-state exclusion, required native actions, shared ownership, removal,
and the draft.2 disabled-hooks refusal. Coverage links the original hook rows
to the new registry without changing the frozen source inventory.

MCP-hook execution, asynchronous hooks, context-spill limits, Windows commands,
additional events, broader timeout cases, and other native interfaces remain
open. Schema-only declarations remain `unverified`. Full portable permission
mapping and the universal configuration milestone remain incomplete. Adapter
support and release gates are unchanged.

## Codex MCP hook execution and parser limits

### Native evidence and fixes

The current official index links to the
[hook page](https://learn.chatgpt.com/docs/hooks.md). A fresh fetch matches the
[archived source](../WORKBENCH/evidence/native-draft2-debug/codex-hooks-docs.md):
SHA256 `6b4c549c4f1df78e802f289151e54f7677128b8a9929912afb7e8ecd975d2ce9`.
The page documents synchronous MCP handlers, existing-connection prerequisites,
non-blocking error handling, and the unsupported `SessionEnd` combination.

The [first projected test](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-first-project.json)
failed before execution. The native parser rejected `null` inside a nested
input array. This exposed a gap in the pinned schema: it permits arbitrary JSON
input, but native loading converts that input through TOML. The
[parser matrix](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-parser.json)
confirms rejection of direct, object, and array null values. Objects, mixed
arrays, finite floating-point numbers, and tested 64-bit integer bounds parse.
It also confirms that native discovery skips `SessionEnd` MCP handlers with an
explicit warning.

The draft.2 selector now refuses null-valued MCP input and keeps an entire
`SessionEnd` event array inactive if it contains an MCP handler. Optional
profiles can still project other valid events. Required profiles refuse that
loss. Tests cover nested null values, exact large numbers, inactive event
arrays, required profiles, and refusal before target writes. The source format
and stable/draft.1 semantics remain unchanged.

### Execution and authority

The [runner](../WORKBENCH/conformance/run_native_codex_mcp_hooks.py) uses Codex
`0.154.0`, binary SHA256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
It uses a deterministic local Responses provider, a local stdio MCP server,
and separate temporary homes under `/mnt/DATA/tmp`. No real credentials or
external model are used. Native plugin discovery is disabled in final runs.
The MCP connection and native execution policy are explicit fixture setup;
this runner does not claim to project or verify the full MCP configuration.

The runner imports hook configuration, applies it, lists the native definitions,
checks their exact source and values, and configures trust in the isolated
native store. Reapply preserves that store. A second import preserves the hook
values and excludes native trust. Project apply leaves user configuration
unchanged. The native client restarts before the execution check.

| Case | Project evidence | User evidence | Observed result |
| --- | --- | --- | --- |
| File execution | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-execution-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-execution-user.json) | Pre/post hooks, MCP request IDs and arguments, native hook events, model context, and command effects agree. |
| Inline execution | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-inline-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-inline-user.json) | Inline hook values survive import, projection, native trust, reload, and reimport. |
| Denial | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-deny-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-deny-user.json) | Native hook status is blocked; the model receives the reason; no command effect occurs. |
| Missing server | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-missing-server-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-missing-server-user.json) | No pre-hook MCP call; native failure is explicit; command and post-hook complete. |
| Unlisted tool | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-missing-tool-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-missing-tool-user.json) | Codex forwards the unlisted name. The server reports an unavailable tool; command and post-hook complete. |
| Server error | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-error-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-error-user.json) | The native pre-hook fails with the server error; command and post-hook complete. |
| Missing reference | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-template-missing-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-template-missing-user.json) | Missing event field stops the pre-hook before MCP dispatch; command and post-hook complete. |
| Timeout | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-timeout-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-timeout-user.json) | Native wait ends after about one second; command and post-hook complete before the late MCP result. |
| No hook trust | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-unreviewed-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-unreviewed-user.json) | Definitions are listed as untrusted; no hook event or MCP call occurs. |
| Exact integers | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-numbers-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-final-numbers-user.json) | Values at signed and unsigned 64-bit bounds reach the server unchanged. |

Input tests cover a whole-object event reference, embedded string references,
nested arrays and objects, numbers, and booleans. Status messages are verified
as native event metadata; UI display is untested. Non-matching `apply_patch`
hooks do not execute. Native hook events correlate to the same thread, turn,
source path, and command call ID as the MCP payload and command completion.
Only the fixture command is returned by the local model.

The native policy is preconfigured as `approval_policy="never"` and
`sandbox_mode="workspace-write"`. These runs do not prove mandatory portable
approval behavior. No portable security mapping was widened. Errors, missing
references, and timeouts cannot be used as proof of denial.

The timeout server deliberately ignores cancellation and logs its result after
five seconds. Native waiting stops after one second. The next hook completes
before that late result, and late context is absent from the completed model
turn. This proves the timeout does not terminate work in this MCP server.
It does not establish behavior for all servers or remote transports.

### Retained attempts and verification

The [initial unlisted-tool test](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-first-missing-tool-user.json)
expected no dispatch, but the native client sent the name to the server. That
initial fixture returned success for every tool name. The final fixture returns
an explicit missing-tool error. The
[initial timeout test](../WORKBENCH/evidence/native-draft2-debug/codex-mcp-hooks-first-timeout-user.json)
used a serial server, so its five-second wait also delayed the post-hook. The
final server handles calls independently and records the late result. Both
failed expectations remain available with their runner snapshots. Earlier
passing and reviewed probes also remain in the evidence directory.

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-codex-mcp-hooks.json)
contains final source hashes, 20 final native runs, parser checks, and the
stable, draft, Go, Workbench, repository, compatibility, and coverage checks.
MCP field capabilities now use their own evidence. The timeout field identifies
the limit on waiting and does not claim tool termination. Supplemental native
field rows do not increase the original semantic completion count. All 1,679
source entries and their frozen hash remain unchanged.

Remote transports, elicitation, server-timeout precedence, the startup race,
other lifecycle events, asynchronous command hooks, and context-spill limits
remain open. Full portable permission mapping and the universal configuration
milestone remain incomplete. Adapter support and release gates are unchanged.

## Codex background hooks and context-spill limits

### Evidence scope

The [official hook page](https://learn.chatgpt.com/docs/hooks.md) defines
`async` and `additionalContextLimit`. The
[app-server page](https://learn.chatgpt.com/docs/app-server.md) defines the
thread lifecycle. Current documents were fetched through the official index;
the [source record](../WORKBENCH/evidence/native-draft2-debug/codex-background-docs.source.json)
identifies the archived documents and their hashes.

The [runner](../WORKBENCH/conformance/run_native_codex_background_hooks.py) uses
Codex `0.154.0`, binary SHA256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
Each run has an isolated home, workspace, state directory, temporary directory,
and deterministic local Responses provider. Native plugins are disabled. No
real credentials or external model are used. The native policy is explicitly
preconfigured as `approval_policy="never"` with `sandbox_mode="workspace-write"`.
These are configuration tests under that fixture policy, not new portable
approval mappings.

Each run imports and applies the hook configuration, checks exact definitions
and source paths through `hooks/list`, and sets trust only in the isolated
native fixture. Reapply preserves native trust. Reimport preserves the hook
settings and excludes the trust store. Native restart and listing verify the
`async`, timeout, and context-limit values. Project apply leaves the user home
unchanged. Each final record retains the runner snapshot and source hashes.

### Background execution and process lifetime

A release-file gate holds the hook until the triggering command finishes. This
proves that background execution does not wait for hook completion. In the
active-turn case, the local provider releases the hook before it returns a
second tool call. In the next-turn case, the runner releases the hook after the
first turn completes. Later model input includes the correlated hook context.
There is no automatic model request merely because the background hook ends.
A background `permissionDecision: deny` cannot prevent the completed command.

This tested app-server path emits no `hook/started` or `hook/completed`
notifications for background handlers. Evidence instead correlates the trusted
native definition, native command completion ID, hook input thread/turn/tool
IDs, process log, gate state, and later model request. Synchronous context hooks
do emit native start and completion events. UI rendering is not tested.

| Case | Project evidence | User evidence | Observed result |
| --- | --- | --- | --- |
| Next turn | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-next-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-next-user.json) | The command completes before gate release; the next user turn receives hook context. |
| Active turn | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-active-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-active-user.json) | Context reaches a later model step in the same turn. |
| Background denial | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-deny-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-deny-user.json) | The command effect remains; delayed context arrives despite a denial field. |
| Timeout with attached child | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-timeout-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-timeout-user.json) | A one-second timeout stops the hook and its child in the same process group; neither completes its delayed output. |
| Timeout with detached child | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-detached-timeout-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-detached-timeout-user.json) | The hook stops, but a child in a new process group survives and writes its marker after timeout. |
| Unsubscribe | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-unsubscribe-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-unsubscribe-user.json) | The native thread remains loaded; the hook and child remain alive. Releasing the fixture gate lets the hook finish without a model request. |
| Graceful shutdown | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-shutdown-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-shutdown-user.json) | Closing app-server stdin gives exit code zero; the hook and attached child stop before their delayed effects. |
| Archive | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-archive-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-archive-user.json) | The persisted fixture thread leaves the loaded list; its hook and attached child stop. |

The child fixture records its process-group ID. The detached case uses a new
session and process group, so it is not equivalent to an attached child. The
native timeout cannot establish universal process termination. These results
do not prove behavior for other process trees, transports, hook events, or
interfaces. Fixture cleanup is recorded separately from native termination.

Unsubscribe is not shutdown. The official app-server page specifies a
30-minute no-subscriber inactivity grace period. The tests verify that the
thread remains in `thread/loaded/list` immediately after unsubscribe. They do
not wait for or claim to verify expiry of that period.

### Context limits and native runtime files

The fixture returns 82,033 bytes of known context. Tests compare complete spill
file bytes and hashes with the original output. Model input is checked
separately. A small positive limit shortens context and includes the full-file
path. A 32-token limit can leave only the truncation and saved-file notice.
The omitted limit also spills this fixture; zero retains it in model input
without a spill. Two handlers with different limits do not change each other's
output behavior.

| Case | Project evidence | User evidence | Observed result |
| --- | --- | --- | --- |
| Positive limit | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-user.json) | A 128-token limit gives shortened context and the complete saved-file path. |
| Tiny limit | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-tiny-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-tiny-user.json) | A 32-token limit gives the saved-file notice without the context head. |
| Zero limit | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-unlimited-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-unlimited-user.json) | The full output remains in model input and no context spill file is created. |
| Omitted limit | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-default-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-default-user.json) | The native default spills this large output; the exact default threshold is not measured. |
| Independent handlers | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-mixed-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-spill-mixed-user.json) | One handler spills; the zero-limit handler retains the full text. |
| Background spill | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-spill-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-spill-user.json) | Full output is saved and shortened context reaches the next user turn. |
| Background unlimited | [Project](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-unlimited-project.json) | [User](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-background-unlimited-user.json) | The next user turn receives full background context without a spill. |

Inline `config.toml` tests also pass for
[project background execution](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-inline-background-next-project.json),
[user background execution](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-inline-background-next-user.json),
[project context spilling](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-inline-spill-project.json),
and [user context spilling](../WORKBENCH/evidence/native-draft2-debug/codex-background-final-inline-spill-user.json).

The runtime files appear under the fixture's private `TMPDIR/hook_outputs/`
tree. Evidence records file modes as observed; apply does not create, own,
remove, or change these runtime files. The test does not claim automatic native
cleanup or private file modes. The fixture temporary parent is private.
Spill-write failures and alternate temporary-directory configurations remain
untested.

### Retained attempts and capability updates

The [first background test](../WORKBENCH/evidence/native-draft2-debug/codex-background-first-next-project.json)
waited for native hook notifications that this path does not emit. Its log
already showed independent hook completion. The final runner observes process
completion, then checks a later model request. The
[first small-limit test](../WORKBENCH/evidence/native-draft2-debug/codex-background-first-spill-user.json)
expected a visible context head at 32 tokens, but native output contained only
the saved-file notice. The final matrix includes that exact case. The
[first unsubscribe test](../WORKBENCH/evidence/native-draft2-debug/codex-background-reviewed-background-cancel-user.json)
expected immediate shutdown. Live documentation and `thread/loaded/list`
resolved that incorrect expectation; separate archive and graceful-shutdown
tests verify actual session termination. All failed records and runner
snapshots remain available.

Capabilities now identify `async` and `additionalContextLimit` as bounded
execution mappings. Each field has its own scenario evidence. Timeout claims
now distinguish attached and detached descendants. Plan actions identify
background-control limits and the separate ownership of native runtime files.
No portable permission mapping or native trust store was widened.

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-codex-background-hooks.json)
contains 34 final native runs, final source hashes, retained-attempt checks,
and the full stable, draft, Go, Workbench, coverage, and repository checks.
The original 1,679-entry inventory and frozen hash remain unchanged. The full
universal configuration milestone and adapter support gates remain open.

## Plugin selection, shared packages, and discovery interfaces

Draft.2 now has a `plugins` profile for existing native selections. Import reads
recognized configuration and keeps the original packages unchanged. The schema
and CLI preserve stable and draft.1 profile semantics. See the
[package and selection guide](PLUGIN_STANDARD.md) for the format decision.

The review corrected four defects or insufficient checks:

- Credential checks did not traverse arrays inside unknown plugin fields.
  Import now refuses those fields before canonical writes and omits values from
  its error. Regression tests check the complete canonical tree.
- Import checked only the usual native source filename. It now uses declared
  config artifacts in both profiles, retains required status, and merges using
  the adapter's format even when a custom source has no native file suffix.
  Ambiguous multiple source owners cause refusal before writes.
- A native install message or a broad `fixture` string match did not prove
  skill discovery. Tests now check exact identifiers, source, scope, paths,
  version where exposed, file hashes, and the result after disable.
- The handoff described different discovery result fields from those emitted
  by the pinned binaries. The strict first attempts failed and remain in
  `*-selection-final-*.json`. The corrected runner uses the observed fields.
  Copilot's dedicated `skill list --json` discovers the fixture that its
  `plugins list --json --kind plugin,skill` output omits.

| Native CLI | Project evidence | User evidence | Native loading |
| --- | --- | --- | --- |
| Codex 0.154.0 | [Project](../WORKBENCH/evidence/plugin-standard/codex-selection-verified-project.json) | [User](../WORKBENCH/evidence/plugin-standard/codex-selection-verified-user.json) | A versioned package copy; plugin-qualified skill name and exact plugin ID. |
| Copilot 1.0.83 | [Project](../WORKBENCH/evidence/plugin-standard/copilot-selection-verified-project.json) | [User](../WORKBENCH/evidence/plugin-standard/copilot-selection-verified-user.json) | A live reference to the local package; no copied package. |

All four cases use identical Agent Plugins 1.0.0 package bytes. They verify
import, apply without installation, explicit native installation, exact skill
discovery, reimport without package copying, disable, and unchanged package
bytes. Project apply leaves user configuration unchanged. Native project trust
is fixture setup, not adapter output. The earlier untrusted Copilot project
attempt remains available as a separate failed prerequisite check.

The [verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-selections.json)
records deterministic checks, four final native runs, source hashes, and retained
failures. Capabilities and semantic coverage use `native-discovery-only` for
these plugin fields. Unverified overlays retain their separate status. The
original 1,679 source entries and their hash are unchanged.

Remote install, installed-state-only import, resolved package locks, direct
canonical package loading, plugin MCP execution, and native extension loading
remain open. Local installation and discovery do not establish these behaviors
or complete the universal configuration milestone.

## Git marketplace revisions and ignored project updates

The Git follow-up found one adapter status error and one required native action:

- Copilot accepts `extraKnownMarketplaces.<name>.autoUpdate` in project settings
  but ignores it. The adapter had reported that entry as active configuration.
  It now retains the entry unchanged and inactive in project scope; required
  entries refuse activation. Unit tests cover both boolean values, both scopes,
  and required refusal. The [official source](../WORKBENCH/evidence/plugin-standard/copilot-marketplace-settings.source.json)
  distinguishes user opt-in, managed authority, and session types.
- Codex's configured Git marketplace is not available until its native cache
  has been populated. The first native list attempt reports a missing manifest.
  Plan now reports `codex plugin marketplace add SOURCE` with the configured ref
  and sparse paths before package installation. Apply does not perform this
  native operation.

The new fixture serves a generated Git repository over loopback HTTP. Its tag
and default branch contain different package versions and skill bytes. Native
HTTP events correlate with package installation. Codex installs the tag selected
by `ref`; Copilot installs the default-branch package. Neither apply nor disable
fetches Git content. Reimport retains the selection without copying package
assets, and native discovery no longer exposes the skill after disable.

| Native CLI | Project evidence | User evidence |
| --- | --- | --- |
| Codex 0.154.0 | [Project](../WORKBENCH/evidence/plugin-standard/codex-git-verified-project.json) | [User](../WORKBENCH/evidence/plugin-standard/codex-git-verified-user.json) |
| Copilot 1.0.83 | [Project](../WORKBENCH/evidence/plugin-standard/copilot-git-verified-project.json) | [User](../WORKBENCH/evidence/plugin-standard/copilot-git-verified-user.json) |

The local-directory matrix also passed again with the changed CLI. The
[verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-git.json)
checks all eight native runs against current source hashes. It retains the first
two fixture build timeouts and the missing Codex cache attempt. The build
timeouts came from an empty isolated Go cache; fixture builds now use the
existing build cache while native commands keep their isolated homes.

The Codex marketplace `ref` field now has a source-linked discovery mapping.
Public HTTPS/SSH services, authentication, GitHub API sources, automatic updates,
package MCP and native extensions, package locks, and installed-state-only import
remain open. The full milestone and support gates remain unchanged.

## Shared plugin MCP execution and environment differences

The stdio follow-up tests the second portable Agent Plugins component. Both
clients receive identical root `plugin.json`, `mcp.json`, and server bytes. The
local provider requests one exact MCP tool. Approval events, MCP requests and
results, native completion, file effects, and subsequent model input are checked
together. Native fixture policy is explicit and does not pass through apply.

| Case | Codex 0.154.0 | Copilot 1.0.83 |
| --- | --- | --- |
| Native approval already set for the fixture tool | Executes and returns the exact result. | Executes and returns the exact result. |
| Prompt accepted | MCP elicitation acceptance precedes `tools/call`. | ACP permission acceptance precedes `tools/call`. |
| Prompt denied | Native failure; no server call or file effect. | Native failure and turn end; no server call or file effect. |
| `approval_policy=never`, without native tool approval | Refuses the MCP tool. | This Codex setting has no Copilot mapping in the fixture. |
| Disable and restart | Tool is absent; server does not start; package/data remain. | Tool is absent; server does not start; package/data remain. |

Each applicable row passes in project and user scope. The
[verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-mcp.json)
covers 14 MCP runs and eight repeated local/Git marketplace runs with current
source hashes. Native MCP execution is a separate `bounded-fixture-execution`
capability. The original semantic inventory count does not increase.

The test also finds two Copilot standard violations. `${ODA_AMBIENT}` in a
package environment value expands to the ambient fixture value, although the
standard requires it to stay literal. `${PLUGIN_DATA}` in another environment
value expands during install inspection but stays literal during the ACP
session. Both launches still receive correct reserved `PLUGIN_ROOT` and
`PLUGIN_DATA`, arguments, and working directory. Codex passes these tested
environment checks. See the
[requirements excerpt](../WORKBENCH/evidence/plugin-standard/agent-plugins-environment-requirements.md).
Capabilities now retain these limits instead of implying full package
conformance. The adapter leaves the original package unchanged.

Failed probes remain available. The initial Codex model fixture omitted its MCP
namespace. The next probe correctly reached native tool dispatch but met the
unapproved-tool refusal under `never`. The initial Copilot log search matched
both fixture and native session logs; the runner now uses a unique fixture
filename. Protocol `_meta` fields made an overly strict whole-params comparison
fail; the test now checks tool name and arguments while retaining metadata.
The first denial assertion incorrectly required a model follow-up from Copilot;
the native turn had ended after rejection. These fixture corrections do not
change native approval or package semantics.

Approval timeout/disconnect cases, other MCP transports, path containment,
component isolation, package native extensions, installed-state-only import,
public services, and remaining configuration families still need work. Portable
`ask` and combined native/security gates remain unchanged.

## Legacy preferences, object containers, and array selection

The configuration audit found four implementation defects:

- Copilot user import read `settings.json` only. It now also reads recognized
  legacy preferences from `config.json` without copying application state or
  modifying the source file. Documented unmapped preferences remain inactive.
- Pending native migration could silently overwrite a newly projected setting.
  Native `1.0.83` replaces matching modern roots with legacy roots, including
  whole objects. Apply now refuses conflicting writes and removals, including
  with force/adopt, and repeats the check when it rebuilds under the target lock.
  Migration remains a native operation.
- The field selector required explicit object-container declarations even when
  known leaves existed. It now derives containers such as `ide`, `subagents`,
  and `tabs` from their registered descendants. Unknown descendants remain
  inactive. The type and capability declarations use the same inferred parents.
- The selector accepted `array<string>` members but rejected the documented
  `string[]` form. Both indexed forms now work; named properties do not become
  array members. Invalid/unknown array members still keep the array inactive.

The type inventory also conflated the boolean `memory` preference with the
saved-permission operation `memory`. It now treats the preference as a user
setting and keeps the operation in its separate external context. `bannerStyle`
has its documented enum mapping. `trustedFolders` is classified as external
authority, with its existing semantic ID preserved.

The [native runner](../WORKBENCH/conformance/run_native_copilot_legacy.py)
compares three source layouts against migration in independent control homes:
legacy-only preferences, conflicting scalar/array values, and conflicting
objects. Import, apply, native startup, and reimport preserve the expected
configuration. Source state is unchanged by the adapter. New preferences are
private. A pending conflicting `memory` value blocks apply before writes.

Go regressions cover malformed legacy JSONC, external credentials and trust,
whole-root precedence, removal, equal legacy values, source permissions,
project isolation, known containers, arrays, and unknown descendants. The first
ad hoc native probe used a plain JSON parser on Copilot's generated comment
header; the [failed parser record](../WORKBENCH/evidence/native-draft2-debug/copilot-legacy-jsonc-probe-failure.json)
retains the observed files and limits of that attempt. The native runner handles the
generated whole-line comments and does not claim to be a general JSONC parser.

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-legacy.json)
checks three migration cases and eight repeated marketplace cases with current
source hashes. Earlier MCP execution records retain their historical source
hashes. These results do not establish native behavior for every newly reachable
preference. The remaining field and security evidence work stays open.

## Copilot preference values and terminal behavior

The next audit found documented user preferences without activation mappings:
`defaultMode`, `inlineImages`, `inlineImageLiveWindow`, `notifications`, and
the `statusLine` leaves. They now have explicit type and scope declarations.
The history limit also needed its documented integer bound. Invalid integer
values, unknown tab names, and requests to hide the Session tab remain inactive
instead of being reported as successfully activated. Arrays remain atomic, and
required invalid values refuse before writes. `defaultPermissionMode` has its
documented enum declaration but remains blocked by the security gate.

The [native test](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-final.json)
uses canonical apply before each of three fresh terminal sessions:

| Phase | Native observations |
| --- | --- |
| Plan | Plan indicator; Sessions before Current; Gists hidden; status output in column four; nine command calls with one-second timer gaps after startup. |
| Interactive | Plan indicator absent; Current before Sessions; Gists restored; status output in column one; five command calls with two-second timer gaps after startup. |
| Removed | Tab bar absent; removed status command produces no new file events or terminal output. |

Each status event contains native session JSON. The tests correlate its session
ID, workspace, version, and model with command output. The sessions use distinct
IDs. The native process remains live for each observation window and is then
closed by the fixture. No model turn is submitted. Apply preserves native
application state; new settings have mode `0600`. Native startup leaves the
projected settings unchanged, and reimport preserves each phase's values.

The [first attempt](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-first.json)
retains a fixture assertion failure: the terminal reached column one with CRLF,
while the assertion expected an explicit cursor-position sequence. The corrected
check accepts both representations. This is not a native padding failure.

Capabilities and semantic coverage link the measured preference subset to this
evidence. Notification delivery, image rendering, autopilot execution, command
history behavior, command failure handling, and event-only status refresh remain
unverified. The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-preferences.json)
checks the current CLI, repository, and repeated plugin/migration fixtures.
Earlier verification records retain the source hashes from their own revisions.

## Copilot subagent dispatch and native account prerequisites

The next coverage audit found a declared `subagents.agents` object whose child
settings could not activate. The adapter now maps each named agent's `model`,
`effortLevel`, and `contextTier`, including `inherit`. It treats the agent name as
a selector and continues to reject unknown nested settings. Import, apply, and
reimport preserve these fields. Project scope keeps them inactive.

The adapter also checks the documented upper bounds for concurrency and depth,
keeps ignored `rubber-duck` disable requests inactive, and preserves disable-array
atomicity. The implemented numeric subset uses positive integers; exact lower
bounds and limit enforcement need a native usage-based billing fixture.

Four [pinned native cases](../WORKBENCH/conformance/run_native_copilot_subagents.py)
test inherit, overrides, disable, and limits. Native provider requests prove that
the per-agent model and effort override applies to the child and leaves the
parent choice intact. The native event records the explicit long-context tier;
it omits the default tier. The test preserves that absence. It does not measure
actual context capacity.

The disabled custom agent is absent from discovery and rejects an explicit task
dispatch before a child starts. In the separate limit case, two nested agents
complete despite `maxDepth: 1` and `maxConcurrency: 1`. The official reference
states that those settings are ignored outside usage-based billing. Plan and
capability output now report this prerequisite instead of implying that apply
establishes runtime resource limits. Account state remains external.

Two fixture failures remain available with their original runner snapshots:

- [Tool parser](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-first-override.json):
  the request included a `tool_search` entry without a `function` member.
- [Default tier parser](../WORKBENCH/evidence/native-draft2-debug/copilot-subagents-final-inherit.json):
  the native configured event omitted the default `contextTier` member.

Both attempts executed the child effect before the assertion failed. Neither
is evidence of a native dispatch failure. The corrected test correlates model
requests, native events, approval, file effects, and import preservation.
The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-copilot-subagents.json)
checks the current CLI and repeated plugin, migration, preference, and dispatch
fixtures. Remaining native configuration and security work stays open.

The repeated preference fixture also retained a
[terminal cleanup failure](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-subagents.json).
Pexpect reported that it could not terminate the child within its short wait.
A subsequent process inspection found no remaining process in that fixture's
native home. The runner now saves terminal output before cleanup and, if close
fails, polls the same child for up to five seconds. It does not start a replacement
session while that child remains live. The
[repeated preference result](../WORKBENCH/evidence/native-draft2-debug/copilot-preferences-subagents-verified.json)
records process state after teardown as well as the original observations.

## Shared skill discovery and Codex selector references

Both Codex and Copilot already discover `.agents/skills` at project scope.
The [skills.sh source review](PLUGIN_STANDARD.md#skill-distribution-with-skillssh)
confirms that its CLI documents this destination for both clients. The adapter
does not need to install a second project copy or provide a new skill loader.

This audit found two separate Codex selector defects. Native import copied user
skills while retaining absolute selectors to the old native home. It now makes
references to copied packages relative to the target native home. External
relative references keep their original absolute targets, and `~` expressions
retain native HOME expansion. Source configuration, assets, and permissions stay
unchanged during import.

The [path probe](../WORKBENCH/evidence/native-draft2-debug/codex-skill-path-probe.json)
found that file selectors disable a skill, while directory selectors leave it
enabled. The [HOME probe](../WORKBENCH/evidence/native-draft2-debug/codex-skill-tilde-probe.json)
confirmed native tilde expansion. Directory selectors now remain inactive, with
atomic selector arrays and required-value refusal. The coverage generator also
matches schema `[]` indexes to documented `<index>` fields without changing
semantic IDs or the frozen source entries.

The [project discovery attempt](../WORKBENCH/evidence/native-draft2-debug/codex-skills-first-project.json),
[absolute-path control](../WORKBENCH/evidence/native-draft2-debug/codex-skills-absolute-project.json),
and [model-context control](../WORKBENCH/evidence/native-draft2-debug/codex-skills-catalog-project.json)
remain available. They show that loaded project selectors do not affect discovery
or the initial model catalog in Codex `0.154.0`. Required project selectors now
refuse before writes. The final project fixture performs direct native setup
after that refusal to verify the limitation; it does not claim successful
adapter projection of the ignored setting.

The [user result](../WORKBENCH/evidence/native-draft2-debug/codex-skills-verified-user.json)
correlates four discovery responses with four completed local-provider turns:
source disabled, relocated disabled, enabled, and disabled again. Reimport keeps
the same relative selector. The
[project result](../WORKBENCH/evidence/native-draft2-debug/codex-skills-verified-project.json)
records the loaded value, unchanged user configuration, and three completed
model turns that expose the ignored selector. No skill script executes.

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-codex-skills.json)
covers repository checks and these affected native cases. Previous plugin and
Copilot native records keep their original source hashes; they were not rerun
for this Codex skill-reference change. Broader native coverage stays incomplete.

## Coverage syntax, refusal gates, and milestone boundaries

The coverage audit found filesystem access values and path tokens counted as
settings, a duplicate field hidden by TOML table notation, and the empty keymap
example counted separately from its setting. These records now link to their
settings and do not increase feature counts. All 1,397 IDs and all 1,679 original
source-row links remain unchanged; the tests compare them with the retained
[baseline](../WORKBENCH/evidence/native-draft2-debug/coverage-classification-baseline.json).

The map also applies the source's user-only scope for desktop file handlers and
marks Windows and desktop-app records outside the Linux CLI milestone. This
adds no native support claim. Windows onboarding acknowledgement is runtime
state. Managed-policy records keep their external disposition.

Some permission children lacked a field declaration even though the compiled
adapter already refuses their parent setting. The map now reads that refusal
from capabilities and records `gate_scopes` separately from mapping scopes.
It does not infer a field validator or native enforcement from the refusal.

Coverage format version 3 distinguishes 1,397 records, 1,385 counted features,
and 1,361 Linux CLI milestone features. There are still 51 mapping-pending
features, 83 security-evidence-required features, and 232 declared validators
whose native behavior is not established by the map. These classifications
are not completion totals. The
[verification record](../WORKBENCH/evidence/native-draft2-debug/verification-coverage-classification.json)
checks source preservation, coverage tests, the repository, and the unchanged
CLI against its existing pinned skill records. No native test is repeated for
this coverage-only change.

## Codex telemetry scope and credential values

The [direct native probe](../WORKBENCH/evidence/native-draft2-debug/codex-otel-direct-first.json)
confirmed that Codex `0.154.0` ignores project `otel`, even in a trusted workspace.
The adapter had accepted it. Required project telemetry now refuses before
writes, and optional project telemetry remains inactive. User telemetry keeps
its native object structure and exporter variants.

Regression tests also reproduced four credential leaks through activation:
collector URL user information, an access-token query parameter, a Cookie
header, and an X-Api-Key header. These values now remain external. Import
records excluded field paths without values and leaves source configuration
unchanged. TLS certificate and private-key paths remain external references;
apply does not read or copy the referenced files. Unit tests preserve both
HTTP and gRPC configuration objects with absolute TLS references. This is
configuration preservation evidence, not native TLS or gRPC execution evidence.

The [adapter result](../WORKBENCH/evidence/native-draft2-debug/codex-otel-adapter-final.json)
verifies import, apply, reimport, and native execution. Three completed turns
use local model and collector services. Native prompt events contain the
synthetic prompt when enabled and `[REDACTED]` when disabled. Conversation IDs
match native thread events; trace IDs correlate exported logs and spans.
Resource environment tags and fixture headers match user settings. Project
collector and prompt-export overrides have no effect. The project control is
written directly by the fixture after the adapter refuses it. New user config
files have mode `0600`, and native execution leaves them unchanged.

The [earlier adapter run](../WORKBENCH/evidence/native-draft2-debug/codex-otel-adapter-first.json)
also passed; the final run adds explicit prompt-event, environment, and trace
correlation assertions. Both runner snapshots remain available. The
[verification record](../WORKBENCH/evidence/native-draft2-debug/verification-codex-otel-final.json)
checks the affected native case, source hashes, and repository gates. Existing
skill and plugin records retain their original implementation hashes.

The coverage map now resolves documented exporter `<id>` paths to the finite
schema variants. `validation_paths` shows both transports for shared fields
and HTTP only for `protocol`. All 1,679 source entries and all 1,397 record IDs
remain intact. TLS fields are validator declarations; the measured HTTP JSON
subset has separate native evidence. Metrics delivery, TLS, gRPC, binary
encoding, live reload, and broader native coverage remain open.

The [first full verification](../WORKBENCH/evidence/native-draft2-debug/verification-codex-otel.json)
passed 26 of 27 checks. The compatibility check found that the new limitation
was absent from the CLI capabilities text. The text is now synchronized with
`CLI/compatibility.json`. The final native run uses the resulting source hashes;
the earlier native results and failed verification remain available.

## Codex telemetry TLS references and client identities

The [first CA-only import test](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-ca-relative-first.json)
reproduced a reference relocation defect. Native delivery worked from the source
home, then failed after apply because the relative CA path pointed into the
new home. User import now resolves known relative TLS references against the
source configuration directory. The adapter does not read, copy, or own the
referenced files. Absolute paths and native HOME expressions stay unchanged.

The [pinned source review](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-sources.json)
records Codex commit `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`, from tag
`rust-v0.154.0`. Its TLS fields use `AbsolutePathBuf`; the saved implementation
shows base-path resolution and native HOME expansion. The saved exporter source
also shows the client certificate/key pairing and exporter builder calls.
The measured native binary remains pinned independently by SHA256.

The final [relative](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-ca-relative-final.json),
[absolute](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-ca-absolute-final.json),
and [HOME](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-ca-home-final.json)
CA tests each complete source and relocated model turns. A local TLS collector
receives logs and traces; native thread events and exported trace IDs correlate.
Import/apply/reimport preserves the reference targets. Source file hashes and
permissions stay unchanged, and no certificate or private-key files appear in
the canonical tree or target home. New user config remains mode `0600`.

The HTTP identity tests produced a separate native limitation. Codex `0.154.0`
reports `Could not create otel exporter: builder error` with valid EC and RSA
client identities. Metrics reports the same failure with the additional
`invalid OTLP metrics configuration` context. Independent Python TLS requests
succeed against the same collector with the same CA and client identity. The
collector verifies the client certificate serial. Each native test completes
a deterministic local model turn but sends no telemetry to that collector.

| Exporter | EC identity | RSA identity |
| --- | --- | --- |
| Logs | [Failure verified](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-identity-exporter-ec-final.json) | [Failure verified](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-identity-exporter-rsa-final.json) |
| Traces | [Failure verified](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-identity-trace_exporter-ec-final.json) | [Failure verified](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-identity-trace_exporter-rsa-final.json) |
| Metrics | [Failure verified](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-identity-metrics_exporter-ec-final.json) | [Failure verified](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-identity-metrics_exporter-rsa-final.json) |

The adapter now refuses required HTTP identity settings and keeps each optional
exporter inactive as a whole. It does not remove identity fields and activate
an unauthenticated exporter. Import still preserves these native references.
gRPC identity behavior has not been measured and is not covered by this refusal.
Coverage records the HTTP-specific limitation without claiming a gRPC result.

All early attempts remain available. The first independent TLS controls failed
because the fixture certificates lacked Authority Key Identifier extensions.
The corrected certificates include subject and authority identifiers and key
usage extensions; the control then passes with strict certificate checks.
The first metrics assertion expected the shorter logs/traces error text; the
retained metrics record contains the actual prefixed builder error. These are
fixture failures and do not establish native behavior by themselves.

The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-codex-otel-tls.json)
checks ten current native records, including the
[repeated HTTP test](../WORKBENCH/evidence/native-draft2-debug/codex-otel-tls-http-final.json).
TLS records include CLI Go source, embedded schema, go.mod, go.sum, helper,
and runner hashes. Stable and draft conformance, Go race/vet, Workbench,
compatibility, and coverage checks remain required. Binary OTLP encoding,
gRPC, metrics delivery, certificate rotation, live reload, and broader native
security and configuration work remain open.

## Scoped instruction link validation

Validation of the SPEC starter exposed a CLI discrepancy. The Python
conformance checks accepted `SPEC/examples/basic/AGENTS.md`, a tracked link to
its own `.agents/AGENTS.md`. CLI validation accepted that form only at the
selected repository root and refused the nested example link as non-regular.

The CLI now accepts the same canonical compatibility link within a nested
scope. The source directory and canonical file must be real filesystem
entries. The resolved link must equal the scope's own canonical file. External
files, ancestor instructions, missing files, cycles, canonical file links,
and canonical directory links remain refused. No arbitrary scoped symlink
support follows from this correction. The Claude file projection uses the
same validation rule for its scoped bridges. No Claude native test was run.

The regression fixtures apply Codex, Copilot, and Claude file projections with
root and nested compatibility links plus ordinary deeper instructions. They
verify that scoped links remain unchanged and that Claude bridges stay outside
canonical directories. Unsafe cases test refusal before writes with `--force`.
The fixture uses a link-aware snapshot so that broken and cyclic links can be
checked without following them. An early fixture incorrectly used the generic
content snapshot and omitted the existing root compatibility link; both test
setup errors were corrected before verification.

The SPEC starter now includes the requested empty `.agents/plugins/.gitkeep`.
Its manifest remains stable `1.0.0`; an empty unselected directory activates no
plugin profile. The [verification record](../WORKBENCH/evidence/native-draft2-debug/verification-scoped-instruction-links.json)
checks the starter with the CLI as well as the specification conformance
suite. Repository-only verification states that native tests were not run.
Existing native evidence remains historical at its recorded source hashes.

This change does not resolve initial instruction projection when a project
has no root compatibility file. The new fixture exposed that separate gap:
apply did not create a root Claude bridge from `.agents/AGENTS.md` alone.
Root discovery and projection need a separate cross-adapter lifecycle review.
Telemetry transports, remaining native settings, and security evidence also
remain open.

The later [OTLP transport review](CODEX_OTEL_TRANSPORTS.md) adds user-scope gRPC
and HTTP binary delivery for logs, traces, and metrics, including CA trust and
gRPC client certificates. It also fixes a native-import defect that removed
authentication while keeping an exporter active. Exporters with excluded
credentials and unsupported optional HTTP identities now use explicit `none`;
required content still blocks apply. That report supersedes the open transport
and metrics items above. Remote collectors, certificate rotation, live reload,
and the wider security matrix remain outside this evidence.

## Stable import credentials, policy, and rollback

The continuation of the project extension audit found five stable import
defects. These are adapter and file-transaction defects; they do not require a
native session to reproduce.

1. Codex literal environment and header values could reach canonical files
   before versioned validation returned an error.
2. A mixture of literal values and native environment references could discard
   the literals and report success. A literal that looked like the internal
   portable reference URI could also be interpreted as a reference.
3. Unknown MCP fields, including activation, authentication, and tool-filter
   controls, could disappear during decoding into a Go struct.
4. Forced import could remove required capabilities and selected profiles from
   an existing stable manifest.
5. Final validation or a later write error could leave imported files and
   backups in the repository. An external instruction link could fail only
   after its contents had been copied into the canonical tree.

The importer now checks native MCP fields before typed decoding and validates
the canonical representation before writes. It checks instruction discovery
before reading the instruction source. It preserves manifest requirements,
metadata, and selected profiles. A private staging tree validates imported and
retained selected content. Unselected native packages and runtime state are
not copied into staging.

The import transaction includes each output and backup. It retains existing
file modes and uses `0600` for new backups. New skill scripts retain their
executable mode. Target snapshots detect changes during validation and before
writes. On failure, rollback covers attempted outputs and created directories;
a concurrent edit to a target that was not written remains intact. This is not
a claim of the native user-scope locking and authority model for stable import.

Evidence:

- [Regression tests](../CLI/internal/config/import_safety_test.go) cover all
  three stable clients' unknown-field refusal, Codex literals, retained policy,
  instruction links, write failure, backups, permissions, and target changes.
  The Claude cases use configuration fixtures only; native Claude testing
  remains skipped.
- [Committed-baseline failures](../WORKBENCH/evidence/project-tools/stable-import-safety-baseline.json)
  come from a separate checkout of the recorded CLI commit, with a frozen copy
  of the new public-API tests. Each named defect fails there.
- [First regression attempt](../WORKBENCH/evidence/project-tools/stable-import-safety-before.json)
  is retained. Its policy test used an incorrect object shape for `requires`;
  the baseline record uses the correct array and reproduces policy removal.
- [Public CLI checks](../WORKBENCH/evidence/project-tools/stable-import-cli.json)
  record four refusals with unchanged files, successful policy-preserving
  import, private backups, and the resulting mandatory-capability plan refusal.

Run the current regression tests from `CLI`:

```sh
go test ./internal/config -run 'TestStable.*Import|TestStableImport' -count=1
```

The project extension verifier checks the public CLI record against current
source hashes. These corrections do not clear native environment-reference,
plugin authentication, telemetry, or wider security-evidence gates.

## Initial stable instruction projection

Stable plan/apply/sync now creates a missing root compatibility link to
`.agents/AGENTS.md`. The transaction records link ownership, deduplicates the
same link across adapters, and creates the initial Claude file bridge. It does
not replace an existing user instruction file. Canonical edits appear through
the link without a copied instruction tree. A managed link replaced by a
regular file must be restored before apply; force does not remove that file.

Regression tests cover read-only plans, single-adapter and all-adapter apply,
repeated sync, canonical edits, file-write failure rollback, a root file that
appears after planning, and a concurrent replacement that rollback must retain.
The placeholder check also permits an empty regular `.gitkeep` in unselected
skills; nonempty content and indirect entries remain refused.

The [pre-commit verification](../WORKBENCH/evidence/native-draft2-debug/verification-cumulative-precommit.json)
covers the final cumulative source state. This is CLI projection evidence.
Native discovery was not rerun and Claude native testing remains skipped.
Draft.2 project interoperability with an existing root compatibility link,
remaining telemetry transports, and the broader security matrix remain open.

## Draft.2 instruction links and re-import

The stable adapter creates `AGENTS.md -> .agents/AGENTS.md`. Draft.2 Codex
apply previously refused that link. Copilot apply created a second instruction
copy. The project projection now verifies and reuses the canonical link.
Codex releases obsolete ownership of the copied file without changing the
link. Copilot removes only an unchanged copy owned by this source repository.
Modified copies and foreign ownership remain protected. Unit tests cover
migration, repeated apply, backups, rollback, and forced-conflict refusal.

Further round-trip tests found two import defects. Codex refused its verified
canonical link as a native symlink. Copilot reported no recognized artifact
when the canonical link was the only native instruction source. Import now
reads the verified canonical file directly for both clients. It preserves the
link and canonical bytes. External, broken, cyclic, and indirect canonical
links remain refused. A conflicting Copilot instruction copy cannot replace
canonical content, including with `--force`. Existing stable and draft.1
manifests still require a separate migration decision.

The failing import test is retained in
[instruction-link-import-regression-first.json](../WORKBENCH/evidence/native-draft2-debug/instruction-link-import-regression-first.json).
Native Linux runs use Codex `0.154.0` and Copilot `1.0.83`, isolated native
homes, and a local model endpoint. Four cases cover a stable-created link and
a link that replaces an old draft.2 instruction copy for each client:

- [Codex, stable link](../WORKBENCH/evidence/native-draft2-debug/codex-instruction-link-stable-final.json)
- [Codex, migrated copy](../WORKBENCH/evidence/native-draft2-debug/codex-instruction-link-native-final.json)
- [Copilot, stable link](../WORKBENCH/evidence/native-draft2-debug/copilot-instruction-link-stable-final.json)
- [Copilot, migrated copy](../WORKBENCH/evidence/native-draft2-debug/copilot-instruction-link-native-final.json)

Each case performs apply, import, and a fresh native turn for the original
and updated instructions. The captured model request contains the selected
instruction marker exactly once. The later request contains no stale marker.
Native events confirm completion. The link inode remains unchanged, no
Copilot duplicate remains, repeat plans have no writes, and project apply
leaves native user files unchanged. This is evidence for initial context in
fresh sessions, not live reload, tool execution, or every precedence case.

The first Copilot attempts are retained. Their instruction checks passed,
but the fixture incorrectly required native `config.json` bytes to remain
unchanged after a session. Copilot adds first-launch metadata to this state
file. The corrected fixture checks unchanged user files across adapter apply
and unchanged `trustedFolders` across native execution separately. It records
the native state-file change instead of treating it as an adapter write.

Official instruction sources and hashes are retained in
[instruction-link-docs.sources.json](../WORKBENCH/evidence/native-draft2-debug/instruction-link-docs.sources.json).
Run the affected repository and native-evidence checks with:

```sh
python3 WORKBENCH/conformance/verify_plugin_selections.py \
  --with-instruction-links --output /absolute/path/to/new-verification.json
```

This correction does not complete the universal milestone or change release
support gates. Telemetry transports and the wider native security matrix
remain open.

## Runtime authentication units

The [authentication report](NATIVE_AUTHENTICATION.md) extends the telemetry
audit to MCP, model-provider, and LSP definitions. The earlier importer could
exclude a credential while keeping its runtime definition active. It also
misclassified Codex `requires_openai_auth` as account material. Both paths are
corrected: unsafe filtering refuses the operation, while the Boolean remains
configuration. Required portable policy and external account files stay intact.

A retained native baseline shows the model request lose authentication after
relocation. Fixed on/off controls complete four native sessions with correlated
local requests and unchanged account files. The eight telemetry cases were
also rerun against the current source as `*-auth-current.json`; the earlier
`*-final.json` files remain historical evidence. The wider milestone stays open.

```sh
python3 WORKBENCH/conformance/verify_plugin_selections.py \
  --repository-only --with-project-extensions --with-otel-transports \
  --with-provider-auth --output /absolute/path/to/new-verification.json
```

## Provider token-command mapping and native fallback

The [token-command audit](CODEX_COMMAND_AUTHENTICATION.md) corrected a valid
configuration refusal for Codex `model_providers.<id>.auth`. It also found
that the native client sends unauthenticated model requests after helper
failures. Nine paired cases now cover configuration preservation, literal
arguments and working directory, timeout, cached tokens, timed refresh, 401
retry, and five native failure modes. The failed enforcement tests remain
evidence; later observation tests do not claim authentication enforcement.

Current telemetry, account-selection, and stable-import receipts use
`*-command-current.json`. They were rerun after the Go changes. Use
`--with-command-auth` with the repository verifier to check the new cases.
The universal milestone and release support remain incomplete.

## Codex project scope and ownership cleanup

The [project-scope report](CODEX_PROJECT_SCOPE.md) fixes a false activation
claim: the CLI projected project provider settings that the native client
ignored. The model control proves that the trusted project was loaded, while
the request still used the user provider. Required profiles now refuse before
writes; optional source stays intact and inactive. Unchanged older ownership
can be removed without changing unowned fields. Modified owned values remain
conflicts even with force. Thirteen unconditional native restrictions have
effective-configuration evidence. Broker-dependent restrictions remain separate.

Current telemetry, account-selection, token-command, and stable-import receipts
use `*-scope-current.json`. The earlier records remain unchanged. Add
`--with-project-scope` to the verification command to check the new evidence.

## Bounded agent-role overrides

The [role audit](CODEX_ROLE_OVERRIDES.md) confirms a separate child override
contract. The earlier adapter accepted a role provider that Codex ignored.
Required agent files now refuse ignored fields before writes. Optional files
stay intact and inactive as a whole. Project and user role discovery share
this contract; general project-config stripping does not apply to role files.

Eight native cases cover inherited provider selection, child model and effort,
shell-tool reduction, skill-instruction reduction, and skill-selector reduction.
Each case retains completed child and parent events with local model requests.
The official broad configuration claim conflicts with the pinned result and
is retained in the report. Other role controls still require separate evidence.

Current telemetry, account-selection, token-command, project-scope, and
stable-import receipts use `*-role-checked.json`. Add `--with-role-scope` to
the repository verifier to check this audit. Historical evidence remains
unchanged. The universal milestone remains incomplete.

## Role-file and skill-selector relocation

The [reference audit](CODEX_ROLE_REFERENCES.md) found two more import defects.
A relative role declaration stopped loading after native-home relocation.
A relative selector inside a copied role stopped disabling its skill. Import
now preserves external references as absolute source paths and keeps references
to mapped assets relative to their native destination layout. Malformed skill
selectors are not repaired into active selectors.

Ten project/user cases verify original and relocated child runs, reference
values after reimport, unchanged source files, and preserved skill disables.
Current regression receipts use `*-reference-current.json`. Add
`--with-role-references` to the combined verifier. The frozen source inventory
still contains exactly 1,679 rows; the semantic map records reference evidence
without adding duplicate source entries.

The role-scope regression receipts use `*-reference-corrected.json`. Their
import check now compares resolved skill paths because a mapped user-skill
reference can become relative to the agent file. The earlier exact-text
assertion failure is retained. Apply must still leave the imported optional
artifact unchanged.

## Copilot skill metadata and test corrections

The [metadata audit](COPILOT_SKILL_METADATA.md) adds 28 native cases in project
and user scope. It separates skill discovery, model visibility, slash-command
visibility, body loading, tool approval, and file effects. The tested wildcard
tool allowance and custom argument hint do not have the documented effects.
The visibility controls do affect the tested invocation methods.

Two fixture assumptions were corrected. An absolute Python command triggered
directory access, which does not prove shell-tool approval behavior. A skill
catalog check read only messages, although Copilot can put the catalog in its
tool description. The corrected runner uses a shell built-in and checks the
complete model input. Failed records remain in place. The verifier rejects
missing native events, unrelated approval identifiers, and contradictory
file effects.

Use `--with-copilot-skill-metadata` with the combined verifier. Draft.2 now
reports each selected skill, its invocation controls, and metadata losses.
It refuses malformed discovery metadata before writes while stable Markdown
rules remain unchanged. User-scope cases check byte preservation through import,
apply, and reimport. The six metadata records link to these bounded mappings;
their native limitations remain explicit. The source inventory is unchanged.

The metadata cycle used `-reported-checked.json` receipts. Its earlier native
regression families and stable import checks used `*-skill-rechecked.json`
receipts to match the YAML value-validation change.
Prior evidence stays in place, including three command-authentication cases
and two provider-authentication cases that timed out during concurrent
refreshes. Serial reruns pass. Provider-authentication receipts use
`*-skill-serial.json`; the other refreshed families keep the names above.
These retries do not establish the cause of the startup timeouts.

The coverage audit also links discovery-table references for
`COPILOT_SKILLS_DIRS`, `COPILOT_CUSTOM_INSTRUCTIONS_DIRS`, and `--add-dir` to
their existing invocation records. These are not pending configuration
adapters or three additional semantic features. Apply does not set these
options or make their directory-trust decisions. All source rows and record
identifiers remain intact.

## Copilot project skill import and marker backups

The [project skill import report](COPILOT_SKILL_IMPORT.md) records a native
loss from omitted `.github/skills/` and `.claude/skills/` packages. Draft.2 now
imports complete packages, selects `skills`, and records source paths.
Conflicting packages refuse import, plan, and apply even with `--force`.
The two fixed origin cases each execute a contained script before and after
relocation, with matching native approval, tool events, and file effects.

A public CLI check also found that an adjacent backup of an empty canonical
skill marker made the imported tree invalid. The marker backup now goes under
`.agents/state/import-backups/`. Removal remains transactional and checks for
concurrent changes. The original native marker remains unchanged.

Current project-import receipts end in `-final.json`. The Codex regression
families, Copilot metadata matrix, and stable CLI import check use
`*-project-skills-final.json`, with source hashes from the marker-backup fix.
Earlier receipts remain historical, including the command-authentication
`empty` case that timed out in the preceding `project-skills` refresh. The
current case passes; this does not establish the cause of that timeout.

Use `--with-copilot-skill-import` in the combined verifier. The coverage map
now links the two project discovery locations to implemented import mappings.
The frozen inventory still has 1,679 source rows. Full coverage and adapter
support remain incomplete.

## Recursive Copilot instruction loss

The [recursive instruction audit](COPILOT_RECURSIVE_INSTRUCTIONS.md) identifies
another import loss in both scopes. The previous flat-directory scan omitted
native nested instruction files. Draft.2 now preserves their relative names
under the registered directory and applies the existing ownership, removal,
backup, and rollback checks to them. Go and public CLI failures are retained.

Native source/relocated pairs confirm that unconditional nested bodies now
reach the model. The separate `applyTo: "**/*.go"` read probe did not load
either matching body, including its flat control. Later review found the native
catalog in those same requests: Copilot tells the model to read the listed
instruction files. The probe had skipped those reads. The automatic-injection
expectation was incorrect; the failed evidence stays intact.

The coverage map links the two recursive discovery locations to this bounded
mapping and identifies the user directory-overview row as an alias. Original
source rows and identifiers remain unchanged. The first import fix used
`copilot-recursive-instructions-*-final.json`; earlier families retain their
historical source hashes. Use `--with-copilot-recursive-instructions` to check
this cycle with the combined repository verifier.

## Copilot instruction catalogs and user read permission

Nine native trigger experiments separate catalog discovery from automatic body
injection. Mentions, ACP resource links, tracked files, later turns, and edits
do not replace the model's instruction reads for the tested non-global patterns.
The native prompt lists their paths and asks the model to read them. An approved
edit can succeed when the fixture model skips that guidance. This is not an
adapter conversion loss or an enforced instruction policy.

The revised source/relocated cases read instruction paths from the native table.
Project files load directly through `view`. User files outside trusted
directories prompt for exact-path read permission. Native allow loads both
bodies; denial stops the tested read and leaves the bodies absent. Plan now
reports this native prerequisite in user scope. Apply leaves trust unchanged.

Current recursive-instruction receipts end in `-catalog-checked.json`. The
verifier checks native table rows, explicit reads, allow/deny events, source
preservation, and completed native interactions. The retained trigger cases
and the wrong early test expectations have separate checks. See the
[updated report](COPILOT_RECURSIVE_INSTRUCTIONS.md).

## Coverage links for existing native declarations

The [coverage audit](FEATURE_INVENTORY.md#native-path-and-declaration-links)
found missing links between source records and compiled capabilities. Four
Copilot user artifact paths now link to their existing mappings. Two concrete
Bedrock fields now resolve to generic provider validators, while their native
status remains unverified. Two repeated user-path records are aliases. The
audit also corrects project/user discovery scopes and removes stale prose
counts from the inventory report.

The retained failed tests distinguish an absent artifact link, an absent
named-object validator link, and a duplicate counted feature. Tests also check
ambiguous wildcard matches, component boundaries, missing declarations, finite
telemetry aliases, and preservation of source IDs. This changes the accuracy
of the coverage map; it does not change adapter behavior or promote support.

## Copilot regular root instructions

The [root instruction audit](COPILOT_ROOT_INSTRUCTIONS.md) reproduced two
import defects: root-only sources failed, and sources with a distinct native
instruction file silently lost the root body. Draft.2 now imports regular root
instructions without potential file references and refuses distinct bodies.
Plan and apply also refuse stale regular root instructions after a canonical
policy edit. Force does not bypass these checks.

Pinned native sessions confirm source and relocated instruction loading,
combined distinct bodies, and a changed `@policy.md` reference base after a raw
copy into `.github`. Reference conversion remains refused and incomplete.
Current root receipts use `-reviewed.json`; recursive instruction regressions
use `-root-reviewed.json`. Previous receipts remain historical. The combined
verification record is `verification-copilot-root-instructions-verified.json`.
This fixes a loss path; it does not complete the wider milestone.

The first combined root-instruction check failed because the stable-import
receipt recorded earlier Go hashes. The public CLI cases were rerun against
the final source. The rejected combined receipt remains unchanged.

## Native agent instruction mapping

The [native agent instruction audit](COPILOT_AGENT_INSTRUCTIONS.md) removes the
remaining root-reference and distinct-body import refusals through a fixed
path mapping. It also imports `CLAUDE.md`, `.claude/CLAUDE.md`, and `GEMINI.md`.
Source bytes, file locations, native reference bases, canonical portable policy,
and custom native source paths remain intact. Native files use individual
ownership, removal, private backups, and transactional rollback.

Eight native sessions cover combined and isolated files from repository-root
and `.claude` working directories. They show that `.claude/CLAUDE.md` is not
loaded by the tested root sessions, even without a root Claude file. It loads
from `.claude`, including beside root `CLAUDE.md`. Three failed expectations
remain recorded. The observed behavior is a scope difference, not an inferred
root-file precedence rule.

Current receipts and remaining limits are listed in the report. The combined
check uses `verification-copilot-agent-instructions-final.json`. Earlier
root-reference refusals and evidence receipts remain historical. The broader
milestone remains incomplete.

## Canonical Copilot instruction binding

The [canonical instruction audit](COPILOT_CANONICAL_INSTRUCTIONS.md) closes the
distinct-body import refusal beside a verified canonical root link. Native
evidence first established that root references through the link use the
project root. The adapter now records a fixed `canonical-instructions`
binding to portable `.agents/AGENTS.md`. Relocation creates a managed root
file and retains a separate native Copilot body. Later core edits update only
the root output. References remain external dependencies.

Six native sessions cover source links, relocation, and updated core loading.
Go regressions cover ownership release, foreign ownership refusal, identical
source basenames in separate namespaces, duplicate targets, malformed names,
unsafe removal, and transaction rollback. Earlier failures remain recorded.
Root, agent-file, and recursive instruction regressions use new `-canonical`
receipts. The combined check uses
`verification-copilot-canonical-instructions-final.json`. This result does
not promote adapter support or complete the wider milestone.

## Shared project skill source import

The [shared skill import audit](COPILOT_SKILL_IMPORT.md) found two remaining
failures at `.agents/skills`. A bare native skill tree failed the manifest
guard, although Copilot loaded and executed its skill. In an existing draft.2
tree, the source list omitted the shared packages and left `skills` unselected.
Both Go failures and the successful native source session followed by import
failure remain recorded.

Draft.2 Copilot project import now selects shared packages in place. A bare
tree can establish metadata only for recognized skill packages and an optional
regular canonical instruction file. Existing manifests, policy, source modes,
and inodes remain protected. Malformed manifests, other unversioned content,
symlinks, identity conflicts, and user-scope attempts refuse before writes.

Six current skill sessions cover `.agents`, `.github`, and `.claude` origins
before import and after relocation. Native events, model requests, a scoped
approval, and an asset-derived file effect establish execution. The coverage
map now records the existing `.agents/skills/` feature as mapped without adding
a new semantic feature. Thirteen mappings remain pending. Instruction native
regressions use refreshed `-shared` receipts. The combined record is
`verification-copilot-shared-skills-final.json`. The milestone remains open.

## User instruction source routing

The [user instruction audit](COPILOT_USER_INSTRUCTIONS.md) found that reimport
could add a second assignment to `copilot-instructions.md`. It reset a custom
namespace source to the native filename. A retained native run loaded the
source and relocated instructions correctly before reimport changed the
declaration. The importer now reuses the existing user instruction artifact.
Go tests also cover Codex source routing, duplicate targets, policy conflicts,
foreign ownership, scope, file modes, removal, and final-write rollback.

Twelve Copilot native sessions cover default and explicit user homes, a custom
source, relative and child references, updated loading, and removal. Adapter
calls preserve external configuration bytes and trust. An earlier probe's
incorrect first-launch metadata check and the Go test's import-lock snapshot
failure remain recorded. The coverage map now records the existing user
instruction discovery feature as mapped. Twelve mappings remain pending;
semantic feature counts and release gates are unchanged.

Current skill and instruction regressions use `-user-instructions` receipts.
The combined record is `verification-copilot-user-instructions-final.json`.
The wider milestone remains incomplete.
