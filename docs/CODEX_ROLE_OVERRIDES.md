# Codex child-role overrides

Draft.2 now checks Codex agent files against the bounded child-role contract.
It no longer treats an agent file as a general configuration layer. Required
files with ignored fields are refused before writes. Optional files stay
intact in canonical source and inactive as a whole. The same rule applies to
project and user discovery.

## Defect and native result

The earlier validator accepted `model_provider='child'` in a user agent file.
The CLI wrote the file and reported success. Codex applied the child model,
reasoning effort, and instructions, but sent the child request to the parent
provider endpoint. Both the direct native run and the projected run completed.
The [failed adapter check](../WORKBENCH/evidence/native-draft2-debug/codex-role-user-before.json)
retains the projected file, model requests, and correlated native events.

The current [official agent reference](https://learn.chatgpt.com/docs/agent-configuration/subagents.md)
says that agent files can override normal session settings. It also lists
`sandbox_mode` and `mcp_servers` as agent-file examples. The pinned source and
the provider probe show a narrower contract. The
[documentation capture](../WORKBENCH/evidence/native-draft2-debug/codex-role-audit.sources.json)
and [pinned role source](../WORKBENCH/evidence/native-draft2-debug/codex-agent-role-scope.sources.json)
remain available to inspect this conflict. The adapter uses the bounded native
contract and does not treat parsing as activation.

## Mapping

The pinned `AgentRoleOverrides` structure copies seven root settings:

- `developer_instructions`
- `model`
- `model_reasoning_effort`
- `model_reasoning_summary`
- `model_verbosity`
- `personality`
- `service_tier`

These fields still require pinned schema validation. The source also permits
`false` for six feature flags: `shell_tool`, `apps`, `personality`, `plugins`,
`memories`, and `request_permissions_tool`. Skill controls can disable entries
in `skills.config`, disable `skills.bundled`, or set
`skills.include_instructions=false`. The native role loader discards skill
enables and `skills.max_context_tokens`.

Other configuration roots, unknown fields, malformed values, and capability
increases keep the whole file inactive. This includes provider definitions,
MCP definitions, context-window limits, compaction controls, notifications,
and nested delegation configuration. The adapter does not filter an agent
file into a different active file.

Role overrides use a native session layer in both discovery scopes. General
project-config restrictions do not apply to that layer. The native tests
therefore check project role skill selectors separately from the ignored
selectors in project `config.toml`.

## Native evidence

The matrix uses Linux Codex `0.154.0`, with SHA256
`3188814c35471432d4123203e0eb38e5bddc60226e3d7ddf0e59e649ea140022`.
The exact binary hash in each receipt is checked by the verifier. Source
revision: `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`.

Each case runs in project and user scope with an isolated native home,
preconfigured project trust, and local model transport. No account file or
credential is copied. Each native phase correlates a completed spawn, child
turn, parent turn, unique prompt marker, and model request.

| Case | Direct native observation | Adapter and native check |
| --- | --- | --- |
| Provider | Child model and effort apply; parent provider stays active | Ignored provider field is refused; valid model and instruction file works |
| Shell tool | Role `true` cannot enable the tool over parent `false` | Role `false` removes `exec_command` from child model tools when parent has it |
| Skill instructions | Role `true` cannot enable instructions over parent `false` | Role `false` removes the skill catalog marker from child input |
| Skill selector | Role cannot enable a skill disabled by the parent | Role can disable the selected skill while the parent still lists it |

Eight passing cases contain 16 native processes and 32 completed parent or
child turns. The
[verifier](../WORKBENCH/conformance/verify_codex_role_scope.py) checks the
`codex-role-<case>-<scope>-checked.json` receipts, frozen runner hashes,
current Go source hashes, source captures, required refusal, optional
preservation, and unchanged user configuration. Failed fixture attempts are
retained, including invalid TOML, an incomplete canonical fixture, and a
skill ownership conflict. One user-scope probe found the imported fixture
skill through direct `.agents/skills` discovery as a second copy. The corrected
probe puts the user canonical source outside the native workspace and selects
only the native profile. Fixture skills remain external to role projection.

```sh
python3 WORKBENCH/conformance/verify_codex_role_scope.py
```

The tests establish child request configuration. They do not establish shell
execution isolation, all six feature reductions, bundled-skill behavior,
every scalar value, public-provider model availability, or approval
enforcement. Parent security requirements remain in force. Full adapter
support, release gates, and milestone status remain unchanged.
