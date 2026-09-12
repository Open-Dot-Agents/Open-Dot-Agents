# Global configuration

Draft.2 supports a global canonical tree at `~/.agents`, resolved from the
user's absolute home directory. It can hold the same portable profiles,
native namespaces, skills, tools, and plugin selections as a project tree.
The feature requires `--experimental` and remains subject to adapter limits.

For an empty location:

```sh
agents init --global --experimental
agents validate --global --experimental
agents plan --global --experimental --vendor codex --native-home "$HOME/.codex" --format json
agents apply --global --experimental --vendor codex --native-home "$HOME/.codex"
agents plan --global --experimental --vendor copilot --native-home "$HOME/.copilot" --format json
agents apply --global --experimental --vendor copilot --native-home "$HOME/.copilot"
```

`--global` selects the canonical source and implies user scope for import,
plan, apply, and sync. Each write still requires an explicit native home.
Do not combine `--global` with `--root` or project scope. The current working
directory does not select the global source. Initialization creates private
files in one transaction and does not create `~/AGENTS.md`.

The starter binds one `~/.agents/AGENTS.md` to Codex user `AGENTS.md` and
Copilot user `copilot-instructions.md`. The namespace profiles refer to the
same core file. Content with `@` requires an explicit native instruction
artifact because reference relocation is not verified. Native user references
retain their existing external-file requirements.

Global values supply user defaults after apply. Project values can override
those defaults where the native field supports project scope. Instructions
can accumulate instead of replacing complete files. Native restrictions,
such as user-only provider configuration, remain in force. Project operations
report the detected global source and leave user destinations unchanged.

Portable requirements are cumulative. Draft.2 project preflight also checks
the global manifest and portable requirements. A selected global permissions
or sandbox profile currently refuses project projection because combined
enforcement is not verified. `--force` cannot discard that requirement.
Malformed global manifests also refuse projection. A directory with no
manifest is not treated as an Open-Dot-Agents global policy.

When the source is `~/.agents` and the destination is the matching default
native home, selected shared skills remain at `~/.agents/skills`. The adapter
does not create a second native package. The native client must use the same
`HOME`. Unselected nonempty shared skills cause refusal because native
discovery can still load them. Projection to another native home retains
the explicit user package-copy rules.

Global settings and assets keep ownership under their canonical source path.
A project cannot replace or remove that ownership through `--force`. Private
backups and rollback use the existing user-scope transaction rules. Import
reads recognized native paths; it does not copy the complete native home.
Trust, credentials, managed policy, account state, and live sessions stay
external. Plugin installation and authentication remain native operations.

If `$HOME/.agents` uses another format, such as a `dot-agents` tree with
`config.json`, global initialization leaves it unchanged and refuses the
nonempty location, including with `--force`. Migration needs explicit source
mappings and a complete package check before a manifest can select existing
content. No automatic format conversion is claimed.

The [native evidence](../WORKBENCH/conformance/verify_global_config.py) covers
six fresh sessions with Codex `0.154.0` and Copilot `1.0.83`. Both clients load
the global core and discover the shared user skill without a second package.
They also load the project instruction body. Codex uses the project model
override and returns to the global model after removal. These checks do not
establish model compliance with instructions or universal native field
precedence. The saved failed Copilot attempts identify test provider and
JSONC handling defects; the final streaming probe verifies native output.
