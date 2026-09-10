# Plugin reuse and package format

Status: experimental draft.2 selection import and projection. Direct loading of
local canonical packages and public marketplace installation tests remain open.

Use existing marketplace and installed plugins as the normal workflow.
`.agents/plugins/` records native selections, sources, activation choices, and
settings. It keeps native version or revision fields where the client defines
them. Resolved package locks and direct local package loading remain open.

Use [Agent Plugins 1.0.0](https://agent-plugins.org/specification) for portable
packages. A future local package loader can use `.agents/plugins/<name>/`. Keep the root manifest and
component formats intact. Import references to existing native packages without
requiring users to rewrite or vendor them. Open-Dot-Agents does not need a
competing package manifest.

The [upstream repository](https://github.com/agentplugins/agent-plugins-spec)
at commit `ff8ab5e392cc87bd88d87c060815a87490e51003` identifies 1.0.0 as the
published release and 1.1.0 as a working draft. The
[review record](../WORKBENCH/evidence/plugin-standard/agent-plugins-1.0.0-review.json)
contains immutable source links and hashes. The technical charter provides
community governance and prevents a single vendor from controlling a majority
of core maintainer seats. Schemas and software use Apache-2.0; specification
text and documentation use CC-BY-4.0. Preserve the applicable notices when
vendoring material.

## Existing marketplace and installed plugins

A marketplace is a distribution source. Agent Plugins specifies a package
format; it does not standardize marketplace catalogs, install commands, or
native plugin state. Support both standards-based packages and recognized
native Codex/Copilot package formats through their respective adapters.

The integration must:

- Import an existing plugin selection and its source identity from recognized
  native configuration. Preserve the upstream package and native components.
- Store the user's source, version constraint, activation choice, and scoped
  settings under `.agents/plugins/`. Record the resolved revision or digest
  separately so later checks can identify exactly what was tested.
- Allow a project to use a package already installed in a native home, or a
  package available from a configured marketplace. Do not require copying the
  package into the repository or authoring a replacement plugin.
- Report whether each target can load the package and its selected components.
  Marketplace availability is not proof of cross-client compatibility. A
  native extension cannot become portable merely by moving its files.
- Keep package selection separate from installation, updates, login, trust,
  credentials, and runtime data. Apply writes configuration and reports any
  required native installation or activation action. It does not silently
  install or execute a marketplace package.
- Offer vendored or locally authored Agent Plugins packages when the user wants
  a self-contained repository. Use the same ownership and compatibility checks
  for these sources.

Native fixtures use a generated local marketplace and package. No published
package has been installed as part of this verification. Selection import and
projection are implemented. Git fetching is tested with an isolated HTTP server;
public marketplace services and authentication remain unverified.

## Shared packages and native selections

One package can serve several harnesses when each client supports its format
and components. Agent Plugins 1.0.0 supplies the shared skills and MCP format.
Hooks, agents, LSP, and other native features need client-specific extensions.
A native legacy plugin does not become portable through its marketplace.

The package and marketplace catalog are separate. Different native catalogs
can refer to the same unchanged package. The CLI keeps native references:

```text
.agents/
  manifest.json                    # draft.2, profiles: ["plugins"]
  plugins/
    com.openai.codex/
      profile.json
      config.toml                  # plugins, marketplaces
    com.github.copilot/
      profile.json
      settings.json                # enabledPlugins, extraKnownMarketplaces
```

Each `profile.json` records namespace, pinned harness version, project or user
scope, required status, and config artifact sources. The adapter determines the
destination. These files select packages; they are not package manifests.
See the [example](../SPEC/examples/plugins-draft/.agents/manifest.json).

Import reads selection configuration from recognized native files. It does not
copy installed packages or infer selections from installation databases alone.
An existing selection stays in its declared artifact, including a custom source
filename, and retains its required status. An import with conflicting values,
ambiguous source ownership, or embedded credentials fails before canonical
writes. Unknown optional entries remain unchanged and inactive. Unknown
required entries block activation. Native approval overrides cannot bypass
portable security requirements.

The pinned selector accepts Codex plugin enablement, per-server enablement,
and marketplace source fields that pass its schema and authority checks.
Copilot accepts boolean plugin selections and directory, Git, or GitHub
marketplace source declarations. Boolean `autoUpdate` is mapped only in user
scope. Copilot ignores that field in project settings, so an entry with that
field remains inactive there. Required entries refuse activation. Native parser
acceptance of a declaration does not prove automatic updates or a public fetch.
Codex marketplace refresh timestamps and resolved runtime revisions remain
external. Source URLs cannot contain passwords, HTTP user information, or query
data. Credentials and trust must be configured through the native client.

## Native verification

The same package bytes, with root `plugin.json` and one `SKILL.md`, are tested
with Codex `0.154.0` and Copilot `1.0.83`. The runner uses isolated project and
user fixtures. Native trust is explicit fixture setup. Apply only writes
selection configuration. Installation is a separate native command.

| Check | Codex | Copilot |
| --- | --- | --- |
| Native install | `codex plugin add fixture@oda-fixture` caches a versioned copy. | `copilot plugin install fixture@oda-fixture` loads the local package directly. |
| Skill discovery | App-server `skills/list` reports `fixture:fixture` and the exact plugin ID and cache path. | `copilot skill list --json` reports `fixture`, source `plugin`, and its original source path. |
| Disable and reload | The fixture skill disappears; the cached package remains. | The fixture skill disappears; the plugin record is disabled and source files remain. |
| Project apply | User files are unchanged. | User files are unchanged. |

Copilot's `plugins list --json --kind plugin,skill` lists the live plugin but
omits its skill in this fixture. The runner checks the separate skill command.
The initial broad string checks were insufficient; all failed strict checks and
the earlier untrusted-project attempt are retained.

These tests establish installation and discovery for the local fixture. They
do not establish MCP execution, all Agent Plugins loading rules,
or cross-client support for arbitrary marketplace packages. A separate stdio
MCP execution test is described below. The
[verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-selections.json)
links each result and records source hashes.

## Git marketplace sources

The follow-up uses a loopback HTTP Git server and a generated marketplace with
two revisions. Native HTTP requests correlate with the installed package bytes.
Apply makes no Git request. Both clients discover the installed skill, preserve
the selection during reimport, and stop exposing the skill after disable.

Codex requires `codex plugin marketplace add SOURCE --ref REF` before
`codex plugin add NAME@MARKETPLACE` when its Git marketplace cache is absent.
Plan reports this prerequisite. A configuration entry alone does not create
that cache. The test requests a tag that differs from the default branch and
checks the exact installed bytes.

Copilot's `copilot plugin install NAME@MARKETPLACE` fetches the configured Git
source without a separate marketplace-add step. The test verifies the package
from the default branch. Unlike a local directory source, this Git source
produces a copy under the native installed-plugin directory.

Both project and user cases pass. This proves the tested Git HTTP flow; it
does not prove public HTTPS/SSH access, credentials, GitHub API source handling,
or automatic update behavior. See the
[Git verification record](../WORKBENCH/evidence/plugin-standard/verification-plugin-git.json).

## Stdio MCP execution and native limits

The same generated package with root `plugin.json`, root `mcp.json`, and a Python
stdio server executes through both clients. A deterministic local model requests
one exact tool. The test correlates the native tool event, MCP request and result,
file effect, and the result sent back to the model. Apply still writes only the
selection. The package and its persistent data remain outside adapter ownership.

Fourteen project/user cases cover native tool approval set before execution,
explicit approval acceptance and denial, and Codex's `approval_policy=never`
refusal when the tool has no prior native approval. The latter is not an automatic
grant. Native fixture policy approves only the generated tool. The adapter does
not project that approval or weaken a portable permission requirement.

Approval responses occur before successful MCP calls. Denied calls produce no
MCP `tools/call` request and no file effect. Copilot ends the denied turn without
another model request; Codex returns the native failure to the model. After
disable and restart, neither client exposes the tool or starts its server. The
installed package and existing plugin data remain unchanged.

The [standard](../WORKBENCH/evidence/plugin-standard/agent-plugins-environment-requirements.md)
requires unknown placeholders to remain literal and `${PLUGIN_DATA}` to expand
in environment values. The pinned Copilot runtime does not meet these two checks:

| Fixture value | Codex 0.154.0 | Copilot 1.0.83 |
| --- | --- | --- |
| `ODA_UNKNOWN=${ODA_AMBIENT}` with an ambient fixture value | Remains literal. | Expands to the ambient value. |
| `ODA_DATA=${PLUGIN_DATA}` | Expands at both tested starts. | Expands during install inspection, but remains literal during the ACP session. |
| Reserved `PLUGIN_ROOT` and `PLUGIN_DATA`, argument paths, and working directory | Correct in the tested fixture. | Correct in the tested fixture. |

These native limits are reported in plan and capability output. Packages are not
rewritten to hide them. The evidence field `standard_environment_conformance`
covers only the tested environment values; it is not a full conformance result.
Other transports, path rejection, component isolation, native extensions, and
approval timeout/disconnect behavior still require tests. The
[MCP verification](../WORKBENCH/evidence/plugin-standard/verification-plugin-mcp.json)
checks 22 current native runs, including the repeated skill and Git matrices.

## Proposed local package layout

```text
.agents/
  plugins/
    example-plugin/
      plugin.json
      skills/
        example-skill/
          SKILL.md
      mcp.json
```

Only `plugin.json` is required by the package format. Skills and MCP are
optional component types. Each plugin root is self-contained and can be shared
without the surrounding Open-Dot-Agents repository configuration.

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
  "name": "example-plugin",
  "version": "1.0.0"
}
```

Agent Plugins does not standardize the parent installation directory. Choosing
`.agents/plugins/` is an Open-Dot-Agents convention. The specification itself
uses a home `.agents/plugins/` location in a non-normative example.

## Reuse and remaining repository responsibilities

| Area | Agent Plugins 1.0.0 | Open-Dot-Agents integration |
| --- | --- | --- |
| Package manifest | Root `plugin.json` and pinned canonical schema identifier. | Load the original format with locally pinned schemas. |
| Skills | Immediate `skills/*/SKILL.md` children; Agent Skills defines their content. | Preserve package assets, discovery, identity, and native activation evidence. |
| MCP | Root `mcp.json`, typed transports, per-server failure boundaries. | Map each supported transport and value without changing its semantics. |
| Native extensions | Reverse-domain manifest namespaces and matching root directories. | Use a vendor namespace only with defined vendor semantics. Unknown namespaces remain inactive and unchanged. |
| Package path containment | Resolved package paths must stay inside the plugin root. | Reuse descriptor-relative filesystem checks and test allowed internal links and denied escapes. |
| Mutable plugin data | Native client provides persistent per-instance `PLUGIN_DATA`. | Keep runtime data out of tracked packages and ownership records for static configuration. |
| Project selection and scope | Outside the package format. | Define selection, project/user scope, required status, ownership, collisions, and removal separately from `plugin.json`. |
| Marketplaces and installation | No portable marketplace or installation command contract. | Keep installation, login, trust, and runtime actions explicit; apply writes configuration only. |

Hooks, agents, LSP, automation, and permissions are not core component types in
Agent Plugins 1.0.0. Use defined native extensions for those features. Do not
put new portable fields in root `plugin.json` or rename native behavior as a
shared standard. Existing native profiles remain available for host settings.

## Required semantic checks

The normative specification governs loading behavior, not JSON Schema alone.
A closed-schema validator cannot be used as a universal reject-all policy:

- Unknown root manifest fields are reported and ignored. A non-object
  `extensions` field is also non-fatal. Other invalid manifest fields reject
  the package.
- Unknown extension namespaces are ignored without validating their values.
- An invalid component does not disable independently valid components.
  Invalid MCP entries are skipped separately. Unsupported transports do not
  become different transports.
- Only `${PLUGIN_ROOT}` and `${PLUGIN_DATA}` expand, once, in MCP `args`, `env`
  values, and `cwd`. Unknown placeholders stay literal. There is no expansion
  in `command`, URLs, headers, or environment keys.
- The default MCP working directory is the plugin root. Plugin data must be
  separate, writable, and preserved across updates. Reserved environment
  variables come from the client, not the package.
- Remote URL rules, header casing, origin boundaries, and credential exclusions
  must be preserved. The standard supplies no OAuth configuration or portable
  credential-reference format.

Open-Dot-Agents portable security requirements remain in force. A required
selection cannot activate with a documented loss. This repository-level gate
must stay separate from the package parser's specified partial-load behavior.

[OpenAI's builder documentation](https://learn.chatgpt.com/docs/build-plugins.md)
already uses Agent Plugins 1.0.0 in its manual portable example. It separately
describes the `.codex-plugin/plugin.json` compatibility scaffold. This is
documentation evidence, not proof that the pinned Codex or Copilot executables
implement every loading rule. Version-pinned discovery and execution tests are
required before making native support claims. Claude Code execution remains
outside the current test scope.

## Skill distribution with skills.sh

[skills.sh](https://skills.sh) is a directory for discovering reusable skills.
Its [open-source CLI](https://github.com/vercel-labs/skills/tree/80feb48868972d518436f26711509bc78595b5cb)
installs existing `SKILL.md` packages from repositories and other supported
sources. The format comes from [Agent Skills](https://agentskills.io).
Agent Plugins serves a separate purpose: it bundles skills, MCP servers, and
defined native extensions in one plugin package.

The pinned [skills CLI documentation](../WORKBENCH/evidence/native-draft2-debug/skills-sh-source.md)
lists `.agents/skills/` as the project destination for both `codex` and
`github-copilot`. This matches the canonical Open-Dot-Agents skill tree.
Compatible packages can remain unchanged and serve both clients. Features
specific to one client still need that client's support.

Use skills.sh for discovery and its CLI for an explicit installation or update
operation. Keep the selected skill files in `.agents/skills` and select the
portable `skills` profile. Open-Dot-Agents validates and projects the resulting
tree; `apply` does not invoke `npx`, install packages, or run `skills use`.
The skills CLI's `use --agent` mode starts a native agent and is a runtime
operation. User installations have separate native destinations, including
`~/.codex/skills` and `~/.copilot/skills`.

The upstream CLI documents copy and symlink installation modes. The canonical
tree must satisfy Open-Dot-Agents file and path checks after installation.
This documentation review does not establish version-pinned installer behavior
or compatibility for every package listed on skills.sh.
