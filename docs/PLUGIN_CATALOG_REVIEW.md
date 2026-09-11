# Plugin catalog review

Reviewed on 2026-09-11. This review covers catalog structure and selected
package sources. It does not audit every listed plugin or establish native
compatibility. [Source commits and hashes](../WORKBENCH/evidence/plugin-catalog-review/20260911.json)
are recorded in Workbench. No catalog or additional native plugin was installed
as part of this review.

| Source | Role | Use for Open-Dot-Agents |
| --- | --- | --- |
| [Awesome Codex Plugins](https://github.com/hashgraph-online/awesome-codex-plugins/tree/5cb6096acff379bdeb7f48ed3a41c347c1e31851) | Community catalog and Codex marketplace; 220 entries at this commit | Find candidates and inspect real plugin manifests. Check original sources as well as mirrored bundles. |
| [GitHub Copilot Plugins](https://github.com/github/copilot-plugins/tree/fbf7c536a5c7af0c94ff5f528a39004c55129e6d) | Official GitHub collection with local and external plugin sources | Use as a Copilot packaging reference. Advanced Security is the most relevant listed package for this project. |
| [wshobson/agents](https://github.com/wshobson/agents/tree/a30778f8c4e6b0a87567941b7cca4f534bf642b6) | Shared source with native registries and generated client artifacts | Review selected skills from `security-scanning`, `developer-essentials`, and `python-development`. Use its adapter design as comparison evidence. |
| [Awesome DSH Plugin](https://github.com/awesome-dsh-plugin/awesome-dsh-plugin/tree/5b7be0b95ecaf31e8782d38c41811a863b04432b) | Community catalog for DeepSeek Harness | Reference for a possible future DSH adapter. Its `dsh.bundle` packages require the DSH runtime. |

## Candidates and current decisions

- **Trail of Bits property-based testing:** already installed as a shared skill
  under `.agents/skills/property-based-testing`. Both pinned clients discovered
  it. It fits round-trip, normalization, and ownership invariants.
- **wshobson security-scanning:** candidate skills include security requirement
  extraction, threat analysis, and mitigation mapping. These can help connect
  portable requirements to native security tests. Individual skill content can
  be shared; its agents and command workflows need native mapping checks.
- **wshobson python-development:** resource management and testing skills are
  relevant to Workbench cleanup, exception paths, and test quality. The package
  contains 16 skills; selecting the whole package would also add unrelated
  Python guidance.
- **Copilot Advanced Security:** potentially useful for dependency and secret
  scanning. The inspected plugin uses the GitHub MCP endpoint and selects
  `code_security`, `secret_protection`, `security_advisories`, and `dependabot`.
  Native authentication and the required tools must be available. The secret
  scan sends selected content to the remote tool. This review did not configure
  authentication, invoke scans, or establish access to these tools.
- **falsegreen-skill from the Codex catalog:** a possible additional test-review
  skill for Workbench Python. Its declared language coverage does not include
  Go. Its package and findings need further review before use as a check.
- **DSH packages:** no direct addition for the current Linux Codex/Copilot
  milestone. DSH extensions can replace runtime components, including tools,
  sandboxes, and the agent loop. A file adapter cannot supply those runtime
  capabilities to another client.

## Concrete packaging findings

The Hashgraph catalog publishes `.agents/plugins/marketplace.json` and mirrors
packages under `plugins/`. A listing and its scanner score are discovery
signals. They do not prove Open-Dot-Agents conformance or matching behavior in
Codex and Copilot. The catalog's scanner/CI requirements apply to submissions
to that catalog; they are not requirements of this repository.

GitHub's catalog uses `.github/plugin/marketplace.json`. Its
`.claude-plugin/marketplace.json` is a Git symlink to that file. A raw download
of the symlink returns its target text, not JSON. A catalog reader must detect
that form and enforce safe resolution. GitHub's Advanced Security package uses
`.github/plugin/plugin.json` and `.mcp.json`; the shared skills and MCP protocol
do not remove authentication or native metadata requirements.

The inspected `wshobson/agents` Codex manifests for `debugging-toolkit` and
`unit-testing` both declare `"skills": "./skills/"`. The Git tree at the pinned
commit contains neither directory. Both packages have agent and command
sources, but the inspected Codex manifests expose no complete skill tree.
This is a source-layout finding; no native installation test was run. The
`security-scanning`, `developer-essentials`, and `python-development` packages
have 5, 11, and 16 source skills respectively.

That repository's Copilot documentation describes generated `.copilot/`
artifacts and a global symlink installer. Those are separate integration
choices. This project uses the shared `.agents/skills` discovery path already
verified for both pinned clients. An external installer must not replace
existing native homes or become part of reference CLI `apply`.

## Reuse boundary

Pin each selected source revision, retain its license, inspect all declared
assets, and test the resulting native artifact. Shared `SKILL.md` files can
stay in `.agents/skills`. Agent definitions, hooks, plugin manifests, and
installation operations keep their native formats and authority boundaries.

The catalog licenses are Apache-2.0 (Hashgraph), MIT (GitHub and wshobson), and
CC0-1.0 (Awesome DSH). External plugins retain their own licenses. These catalog
licenses do not relicense every linked or mirrored package.

## AI Hero and Matt Pocock skills

The [AI Hero skills page](https://www.aihero.dev/skills) points to
[`mattpocock/skills`](https://github.com/mattpocock/skills). At revision
`3cca18b368ae95cdbdebbff572ccafa662551015`, the repository contains 37 `SKILL.md`
files and an MIT license. The current README offers a Claude Code plugin and
individual skill installation. It describes a native Codex plugin as planned.
The [review record](../WORKBENCH/evidence/plugin-catalog-review/mattpocock-20260911.json)
contains the source inventory and hashes.

`diagnosing-bugs` is now installed under the project's shared `.agents/skills`
directory. It adds a useful sequence: reproduce the failure, reduce the case,
test hypotheses, fix the cause, and run a regression check. Its supporting
human reproduction template and native display metadata remain with the skill.
Both Codex and Copilot discover it directly. No skills.sh installer or remote
setup script was run.

`codebase-design` is another relevant reference, but it overlaps with the
project's existing architecture and specification guidance. The wider workflow
bundle is not installed. `code-review` calls for parallel agents, while
`setup-matt-pocock-skills` changes instruction and issue-tracking conventions.
Those workflows are not needed for the selected standalone diagnostic skill.
The upstream merge skill's “never abort” rule must not become a project Git
history-safety rule.

## Requested OpenAI GitHub package

The separate [project setup](../.agents/plugins/com.openai.codex/README.md) adds
the exact requested package from `openai/plugins`. This supersedes catalog-only
review for that package. Its six files are unchanged at revision
`d416fd5a43426019986b1e489506db3db66dee3d`, version `0.1.11`.
Codex project discovery is verified. Copilot installation can succeed while
ignoring the package's bearer-token field; that package remains inactive in
Copilot. The [readiness report](PROJECT_EXTENSIONS_READINESS.md) records the
native tests and remaining limits.
