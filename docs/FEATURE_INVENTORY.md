# Codex and Copilot CLI feature inventory

This inventory records the CLI configuration surface as observed on
2026-09-10. Native help is pinned to Codex 0.154.0 and Copilot 1.0.83. The live
web references can describe newer behavior. A documentation row does not prove
that the pinned binary implements it.

The [machine-readable inventory](../.agents/features/codex-copilot.json) has
1,679 source entries. Entries retain aliases, repeated context, original
reference text, source line or rendered table row, classification, and feature
area. They are not 1,679 distinct portable features. Common scope, precedence,
lifecycle, status, version, and limit fields are in `record_defaults`. Each
entry's quotation and source section give field-specific details. Unresolved
native semantics stay unresolved; they are not inferred as portable behavior.

The source set contains the complete structured settings, options, commands,
shortcuts, and field tables in the listed references, plus installed CLI help.
It includes Codex's rendered command tables because the Markdown page omits
some imported table data. It also includes Copilot's separate configuration
directory reference. Native flags absent from a live page remain visible in
native-help entries. The hashes identify the exact downloaded sources. This
is a complete extraction of that bounded corpus, not a claim about undocumented
internals, every prose example, or future vendor releases.

## Classification

| Class | Meaning | Current implementation |
| --- | --- | --- |
| `existing-portable` | A field has a 1.0 portable counterpart. | Existing adapter rules and refusals apply. It is not a full vendor feature guarantee. |
| `proposed-portable` | A native security control is a candidate for shared semantics. | Draft validation plus a narrow Codex direct-shell mapping; other policies refuse. |
| `vendor-extension` | A native configuration detail needs a named extension or later proposal. | Inventory only. An extension does not override portable restrictions. |
| `runtime-operation` | A command, flag, environment input, or UI action operates the harness. | Outside persistent portable configuration. Related policy can still have a portable counterpart. |

Per-entry classification uses documented names and context. A native syntax
variant is not automatically equivalent to a portable field. Use this feature
matrix to assess the broader capability, including prose-only behavior.

## Feature map

| Area | Codex CLI surface | Copilot CLI surface | Open-Dot-Agents disposition |
| --- | --- | --- | --- |
| Instructions and scope | AGENTS files, overrides, discovery limits, trust | AGENTS and Copilot instruction files, path rules, imports | 1.0 core for instructions and nearest-file scope; native overrides and discovery controls remain extensions. |
| Skills | Skill discovery, enable/disable, skill metadata | Skill frontmatter, locations, install/list/remove | 1.0 skill content; native catalogue operations and extra metadata remain extensions or runtime operations. |
| Hooks | Lifecycle commands, trust, disable state | Event handlers, approval responses, input/output | 1.0 command hooks with explicit event/matcher refusals; native payload semantics remain native. |
| MCP | Stdio, remote, environment references, timeouts, approvals | Server catalogues, tools, transport, caches, enterprise allowlist | 1.0 MCP subset; unsupported references already refuse. Extra tools, caches, auth, and policy remain native. |
| Tool permissions | Approval policy, rules, permission profiles | Allow/deny tool patterns, path and URL approvals | Experimental permissions profile. Exact commands and `deny > ask > allow`; only Codex shell default-allow permissions map in the first subset. |
| Filesystem isolation | Legacy sandbox and beta filesystem profiles | Built-in path checks and local OS sandbox | Experimental sandbox profile. Read/write/deny and overlap rules; presets are not equivalent. |
| Network isolation | Native proxy, host rules, local/private controls | Sandbox network modes, cooperative proxy, URL checks | Experimental sandbox profile. Separate subprocess, web, remote MCP, local/private scopes. |
| Trust and authority | Trusted projects, managed requirements, flags and config layers | Trusted folders, enterprise floors, runtime grants, settings layers | Repository requirements cannot self-grant trust. Authority must be verified before native projection. |
| Credentials | Auth stores, environment, provider credentials | Login, token sources, keychains, MCP secret/OAuth stores | No credential values in portable data. Draft exposure requirements; account setup stays native. |
| Models and reasoning | Models, providers, reasoning effort, profiles | Model selection, providers, reasoning, model access | Later agents/models proposal; provider-specific settings stay named extensions. |
| Custom agents and delegation | Agent roles, limits, thread controls | Agent frontmatter, built-in agents, sidekicks, scoped communication | Later agents proposal. Security coverage must include descendants explicitly. |
| Planning | Plan mode, collaboration, plan updates | Plan mode, autopilot, plan-then-autopilot | Later planning proposal. Runtime mode commands are not file semantics. |
| Context and compaction | Context limits, compaction prompts, search, images | Context tools, compaction, attachments, session context | Later context proposal; native token accounting remains an extension. |
| Sessions and memory | Resume/fork, history, memory, checkpoints | Session stores, sidebar, resume, checkpoints, memory | Later sessions proposal. Do not copy private runtime stores into a repository. |
| Plugins | Marketplaces, installation, discovery and policy | Plugins, marketplaces, skills/MCP/LSP bundles | Named extensions now; later interoperability work. Plugin lifecycle operations stay native. |
| LSP and automation | Code tools, background work, notification hooks | Language servers, schedules, sidekick triggers | Later proposal. Sandbox coverage must include these processes before any enforcement claim. |
| Remote and protocol interfaces | App server, remote control, MCP server, transports | ACP, remote delegation, protocol and monitoring options | Protocol/runtime operations and named extensions. No launcher in the reference CLI. |
| Logs, telemetry, UI, updates | OTel, terminal options, diagnostics, completions | OTel, tracing, key bindings, settings UI, updates | Native runtime and user settings. No portable UI or telemetry endpoint contract. |

## Scope, precedence, and lifecycle

Codex loads project config only for trusted projects. Some host-owned keys
cannot be set in project config. Command-line overrides and managed
requirements affect the result. New permission profiles and legacy sandbox
keys can conflict across layers; legacy keys can cause the newer profiles to
be ignored. Domain policy needs an active native proxy. These are activation
constraints, not details that an exporter can discard.

Copilot separates user settings, repository settings, local settings, and
managed policy. Live docs describe settings migration and managed sandbox
floors. The pinned help remains a separate record. Persisted tool approvals,
trust stores, credentials, sessions, and plugin state have separate lifecycles.
Startup discovery, explicit reload, session approval, and settings commands
must not be treated as one reload contract.

IDE extensions and cloud execution use different hosts, trust, network,
credentials, and lifecycle controls. This work inventories relevant differences
but implements no IDE or cloud security adapter. Future adapters need their
own versions and evidence.

## Sources and reproduction

Start with the official [Codex index](https://learn.chatgpt.com/docs/llms.txt)
and [GitHub index](https://docs.github.com/llms.txt). Principal references:

- [Codex configuration](https://learn.chatgpt.com/docs/config-file/config-reference.md),
  [commands](https://learn.chatgpt.com/docs/developer-commands?surface=cli),
  [permissions](https://learn.chatgpt.com/docs/permissions.md), and
  [security](https://learn.chatgpt.com/docs/agent-approvals-security.md).
- [Copilot commands](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference),
  [configuration](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference),
  [sandbox settings](https://docs.github.com/en/copilot/how-tos/cloud-and-local-sandboxes/configuring-local-sandbox-settings),
  and [sandbox limits](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/understanding-local-sandboxing).

The inventory tool reads saved public sources only:

```sh
python3 .agents/tools/build-feature-inventory.py \
  --source-dir WORKBENCH/evidence/results/security-draft-final/sources \
  --output .agents/features/codex-copilot.json
```

Source files and their hashes are retained in the local evidence bundle. They
are not credentials or native user settings. Refreshing the corpus requires
new downloads, new hashes, and a review of changed entries. The tool does not
fetch or execute vendor commands itself.

Current mapping limits and observed results are in [security profiles](SECURITY_PROFILES.md).
The source inventory retains the original bounded corpus.
