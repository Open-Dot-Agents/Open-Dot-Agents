# Codex keymap validation and native limits

The adapter now validates the keybinding grammar used by Codex `0.154.0`.
Its JSON schema describes strings and arrays, but the native loader also uses
a custom parser. Before this fix, `ctrl+y` passed the adapter and failed native
configuration loading. It now refuses activation before writes. The accepted
spelling is `ctrl-y`.

Bindings can contain one or two space-separated strokes. An array gives
alternative bindings; an empty array explicitly unbinds the action. Each
stroke can use the documented modifier aliases, printable ASCII keys, named
keys, or function keys through F24. Duplicate or misplaced modifiers, unknown
keys, invalid array members, and longer chords remain inactive. Required
invalid content blocks apply. Validation preserves spelling, alternative order,
and empty arrays. Native loading can normalize spelling.

Only context and action names in the pinned schema activate. The wildcard
`tui.keymap.<name>.<name>` represents that closed field family. For example,
`open_external_editor` belongs to the `global` context, not the composer
context. Approval action labels do not change portable security gates.

The [runner](../WORKBENCH/conformance/run_native_codex_keymap.py) and
[verifier](../WORKBENCH/conformance/verify_codex_keymap.py) cover four fresh
sessions in each of project and user scope. They observe the default editor
shortcut, replace it with F7, update it to `ctrl-x ctrl-e`, then unbind it with
`[]`. They correlate key input with native process identity, editor child
identity, the written effect, and reloaded composer text. Apply does not execute
the editor. Import preserves the selected binding. Other actions have typed
validation but do not have terminal execution evidence.

Three documented fields have version-bound native limitations:

| Field | Pinned observation |
| --- | --- |
| `model_supports_reasoning_summaries` | Omitted, `false`, and invalid-string values produce the same reasoning metadata. The field is absent from the pinned schema. |
| `mcp_servers.<name>.experimental_environment` | `remote` still starts the stdio fixture as a local child. Its ready event and local effect agree. The field is absent from the pinned server schema. |
| `features.rollout_budget.reminder_interval_tokens` | Configuration loading fails. A control with the current `reminder_at_remaining_tokens` setting succeeds. |

The [gap verifier](../WORKBENCH/conformance/verify_codex_setting_gaps.py)
checks six completed native turns and two configuration refusals. Required
content using these three fields remains refused; optional content stays
inactive. The evidence covers the pinned app-server and isolated fixtures.
Other versions and interfaces remain unverified. Raw `config/read` content
alone does not establish that a setting takes effect.

Failed attempts remain stored. One terminal attempt correctly refused an
action in the wrong context. An early rollout control omitted the required
current reminder list; corrected controls use separate receipts. The frozen
source inventory is unchanged. No adapter support claim changes.
