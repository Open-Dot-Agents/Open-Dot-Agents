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
| Copilot 1.0.83 | Existing native approval, filesystem, and network settings are unchanged. Automatic local execution is not guaranteed. | Development decisions are included in project instructions where a native copy is used; canonical instruction references are preserved. |

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
