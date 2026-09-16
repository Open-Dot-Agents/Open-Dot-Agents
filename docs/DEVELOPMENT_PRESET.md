# Development preset

The practical preset permits normal local development and local commits. It
requires the agent to obtain approval for push, publish, deployment, messages,
and destructive work. Those operation decisions are agent instructions.
Native permissions enforce only the boundaries listed below.

## Configure

For a new project:

    agents init --preset development --experimental

For an existing canonical tree without a permissions or sandbox profile:

    agents init --preset development --experimental --adopt

Setup creates two files:

- .agents/permissions/development.json: edit the development decisions.
- .agents/guardrails/development.md: edit the instructions for how to apply them.

The canonical .agents/AGENTS.md refers to both files. Setup preserves existing
instructions and backs up an existing manifest and changed canonical instructions.
Practical setup uses draft.2. Existing security selections cause refusal; they
are never silently converted from strict enforcement to guidance.

The default configuration is:

    {
      "enforcement": "practical",
      "version": "1",
      "preset": "development",
      "project_work": "allow",
      "local_commits": "allow",
      "external_changes": "ask",
      "destructive_work": "ask",
      "protected_paths": [".env", "secrets"]
    }

Allow means proceed within the user's task. Ask means obtain approval for the
specific action. Deny means do not perform it. Scripts and hooks must follow
the decisions for their effects. An allowed test does not authorize a hidden
push. Existing approval for the same action remains valid in the session.
Protected paths are literal project-relative paths, including descendants;
.env does not also mean .env.local or packages/api/.env.

## Apply and update

    agents plan --preset development --experimental --vendor codex
    agents apply --preset development --experimental --vendor codex

The explicit preset option updates only Codex development settings and
preserves other profiles and their requirements. Full repository apply still
checks every selected profile. For Copilot instruction projection, use the
ordinary plan/apply commands without the preset option. Plan reports
native settings and guidance limits. Apply writes configuration; start a new
trusted native session to load it. It does not grant trust or modify the
current session's permissions. After editing either configuration file, run
plan and apply again. Repeat apply makes no changes.

| Target | Native configuration | Agent guidance |
| --- | --- | --- |
| Codex 0.154.0 | Extends the workspace profile; project writes and .git writes for allow; protected-path deny entries; command network disabled; on-request approval reviewed by the user. | Development decisions are included in developer instructions. |
| Copilot 1.0.84-9 | Existing native approval, filesystem, and network settings are unchanged. Automatic local execution is not guaranteed. | Development decisions are included in project instructions where a native copy is used; canonical instruction references are preserved. |

Codex workspace write access includes deletion. Git write access includes
history changes. Native profiles cannot distinguish every operation inside
scripts. Protected-path behavior depends on the native platform and path
state. Readable host paths and inherited credentials are not isolated by
this preset. Hooks, MCP, apps, browsers, and delegated agents use separate
native controls; the guidance still tells the agent to request approval for
their external effects. Administrator and session policy can impose additional
restrictions. These limits are visible in text and JSON plans.

The preset does not add blanket command allowances, command wrappers, or an
external runtime. Command network access remains disabled even when external
changes are allowed by guidance; native escalation is a separate authority.

## Diagnose setup

    agents doctor --experimental --vendor codex
    agents doctor --experimental --vendor copilot --format json

Select a vendor explicitly. The root defaults to the current directory; use
`--root /path/to/project` to inspect another project. Text is the default.
Codex inspection uses the development-only plan. Copilot inspection uses the
full project plan, including unrelated selected profiles and requirements.
Doctor reports that scope and does not write project or user configuration.

| Configuration state | Meaning and next step |
| --- | --- |
| `missing` | Initialize a new tree, or use the shown adoption command for an existing valid tree without security profiles. |
| `invalid` | Correct the reported manifest, policy, or guardrail path, then run doctor again. Configuration values are omitted. |
| `needs-apply` | Review the shown plan, then apply the reviewed changes. |
| `conflict` | Review ownership and local changes with the shown plan. Doctor does not force an overwrite. |
| `legacy-settings` | Review the scoped migration plan with force and backup before applying it. |
| `removal-pending` | Permissions are deselected; review and apply the remaining managed removal. |
| `refused` | Strict enforcement or another selected security contract cannot use this development mapping. The refusal remains in place. |
| `blocked` | Planning refused configuration, ownership, or required capabilities. Inspect the plan. |
| `unknown` | Required configuration overlaps a protected path or the root cannot be inspected. Correct the reported cause before retrying. |
| `current` | The managed files match the plan in the reported scope. Start a new native session after apply. |
| `not-selected` | Development is deselected and no managed change remains in this scope. |

The report has separate layers:

- **Requested policy:** selected development decisions, not permission grants.
- **Configuration:** managed files and ownership on disk. Executable lookup uses
  `CODEX_BIN` or `COPILOT_BIN`, then `PATH` when the variable is unset. A found
  executable is not run, and its version is not verified.
- **Process filesystem:** on Linux, mount metadata for the workspace and Git
  directories, including nested repositories, submodules, and the common
  directory of linked worktrees. A detected read-only mount restricts this
  process. No detected read-only mount does not prove write or commit access.
- **Active session:** unknown. Matching files do not verify native runtime
  permissions, administrator policy, saved approvals, or separate tool controls.

Inspect the app or session permission profile if a requested operation remains
blocked. Project policy cannot remove a host restriction. Doctor does not
recommend weakening permissions or changing the real user home.

Exit `0` means the inspected configuration needs no action and no blocking
mount restriction was detected. Exit `1` means an actionable finding or error.
Unknown runtime authority alone does not fail the command. JSON has its own
`schema_version` (`1.0.0`), a `ready` result, `configuration_state`, `scope`,
and checks with identifiers, status, explanation, paths, and optional command
argument arrays or a `next_step`. Check status is `ok`, `action-required`, or
`unknown`. Text and JSON report the same findings. Read-only Git mounts can
make `ready` false even when `configuration_state` is `current`.

Doctor does not open protected data or account credential stores. It omits
configuration bodies and sensitive values. It does not start native sessions,
even to request a version: [Copilot startup can migrate configuration](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference).
All suggested repairs are separate commands for review.

## Conflicts and removal

A project config that uses legacy sandbox_mode or sandbox_workspace_write
can be migrated with an explicit scoped plan and apply:

    agents plan --preset development --experimental --vendor codex --force --backup
    agents apply --preset development --experimental --vendor codex --force --backup

This replaces only development settings and removes legacy sandbox keys. It
backs up the original config file. Use force and backup for this initial
migration; subsequent managed updates do not need them. An existing backup
is not overwritten. Other native values, including models, MCP,
plugins, credentials, and trust, remain subject to normal ownership checks.
The preset does not erase them. Existing identical values can be adopted with
apply --adopt. Review conflicts explicitly; force is not a policy override.

To remove the projection, deselect permissions from manifest profiles and
requires, then apply each affected vendor again. Owned development settings
and generated decision text are removed. Editable canonical files remain.
Native user/default settings then take effect. Remove the conditional canonical
reference if the development files are no longer needed. Native transactions
provide conflict detection, update checks, and rollback.

## Strict mode

    agents init --preset development --enforcement strict --experimental

A missing enforcement field also means strict. Strict mode retains the
original requirement for equivalent native enforcement of every operation.
It currently refuses activation. Changing a strict file to practical is an
explicit policy change; practical mode also needs a draft.2 manifest and the
guardrails file. Validation alone never establishes native enforcement.

See the [extension contract](../SPEC/spec/1.1-draft/DEVELOPMENT_PRESET.md) and
[acceptance record](../WORKBENCH/evidence/DEVELOPMENT_PRESET.md). No adapter
support claim follows from installing this preset.

## Workflow evidence

The completed 2026-09-16 baseline passed 28 required native cases: 20 for
Codex 0.154.0 and eight for Copilot 1.0.84-9. Five additional Codex cases
record guidance-only effects; they cannot prove enforced operation approval.
See the [workflow evidence record](../WORKBENCH/evidence/DEVELOPMENT_WORKFLOWS.md)
for exact executable and receipt hashes. New doctor checks run within that
campaign. Changes to the CLI, tests, or runner require a fresh campaign before
current-source acceptance. Full adapter support remains false.
