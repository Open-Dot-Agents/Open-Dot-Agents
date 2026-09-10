# Experimental permissions and sandbox profiles

The reference CLI validates `1.1.0-draft.1` and projects a narrow Codex Linux
shell subset. Other security policies are refused. This is an unratified
proposal. Stable 1.0 support and release gates are unchanged.

See the [draft specification](../SPEC/spec/1.1-draft/SPECIFICATION.md),
[feature inventory](FEATURE_INVENTORY.md), and
[native evidence](../WORKBENCH/evidence/CODEX_SECURITY_SUBSET.md).

## Codex shell subset

This mapping covers direct `codex sandbox` commands and their descendants.
It requires the pinned Codex 0.154.0 Linux amd64 binary and an unprivileged
user. It does not establish coverage for model tools, approval flows, hooks,
MCP, LSP, or delegated agents.

The [example](../SPEC/examples/codex-linux-security/.agents/sandbox/sandbox.json)
has these requirements:

- Coverage is exactly `shell`. The sandbox profile is required.
- Filesystem default is `read`. Workspace rules can grant `write`, `read`, or
  `deny`. A child write rule cannot weaken a parent read or deny rule.
- `filesystem.runtime: ["process"]` explicitly permits native private device
  and process-information mounts and caller-supplied standard streams. It
  does not grant shared temporary, cache, or ordinary host-file writes.
- `.agents`, `.codex`, and `.git` have read access. A real local `.git`
  directory is required. Declared paths must exist, except prospective `.codex`.
- Existing workspace hard links and alternate or nested mount views are
  refused. Restricted trees cannot contain symbolic links.
- Subprocess network default, local, and private access are `deny`, with no
  host rules. `web` and `remoteMCP` are `allow` and outside coverage.
- Credential environment and file access explicitly use `inherit`; `allow`
  is empty. This subset does not provide credential isolation.
- Optional permissions have shell coverage, default `allow`, and no rules.
  `ask`, `deny`, and exact-command rules are refused.
- Selected tools, hooks, skills, and other existing owned projections cannot
  be combined with this first subset.

## Native authority and use

Use a configuration-only native home outside the workspace. Its `config.toml`
must already trust the exact absolute workspace path. Only that file and
native-created `tmp` state can exist in the home. Account and cached policy
state are outside this subset. The CLI does not grant trust.

Managed/system configuration, ancestor project configuration, legacy sandbox
keys, and unknown native settings cause refusal. The CLI checks the native
binary hash and runs a read-only native startup probe before projection. Plan
can create native temporary state; it does not change repository or trust
configuration.

Set `workspace` and `native_home` to your existing absolute paths:

```sh
agents validate --experimental --root "$workspace" --format json
agents plan --experimental --vendor codex --root "$workspace" \
  --codex-home "$native_home" --check --format json
agents apply --experimental --vendor codex --root "$workspace" \
  --codex-home "$native_home" --format json
```

Use a fresh successful plan before native execution. Use its exact
`security.native_invocation`, replacing only the executable and argument
placeholders. Replace the process environment with `security.native_environment`;
do not merge caller environment variables into it. The plan sets a fixed
system `PATH` and the selected `CODEX_HOME`. No launcher is installed.
Changed authority, paths, mounts, or native binaries invalidate the preflight.
An ordinary interactive Codex session is outside this evidence scope.

Plan JSON also reports declared and normalized policy, projected settings,
automatic grants, unresolved controls, authority hashes, coverage, and evidence
scope. `native-subset` does not mean full adapter support.

## Ownership, refusal, and removal

The CLI owns two marked segments in `.codex/config.toml`: root selectors and
the named `open-dot-agents` permissions profile. It records their hash in
`.agents/.state/reference-cli/codex.json`. Unrelated bytes stay intact.
Equivalent unowned segments require `--adopt`. Modified owned segments require
an explicit `--force --backup`. Force cannot replace conflicting unmarked
security settings or make an unsupported policy applicable.

To remove the owned policy, remove its profile selections and requirements
from the draft manifest. Review and apply with `--experimental` and the same
`--codex-home`. The recorded explicit-profile invocation then fails because
the profile is absent. Other native invocations use their own settings and
remain outside this contract. A failed apply leaves previous settings active.

Native security import remains unavailable. Import cannot overwrite a draft
manifest with a stable manifest, even with force. The legacy Go export API
refuses selected security profiles. Use plan/apply/sync for the Codex subset.

| Diagnostic | Meaning |
| --- | --- |
| `ODA-SECURITY-0001` | Draft input needs experimental opt-in. |
| `ODA-SECURITY-0002` | The requested policy has no verified mapping in this context. |
| `ODA-SECURITY-0003` | An unknown required extension blocks activation. |
| `ODA-SECURITY-0004` | Security import would be unavailable or lose draft policy. |
| `ODA-SECURITY-0005` | An optional extension is preserved and inactive. |
| `ODA-SECURITY-0006` | A Codex subset, authority, ownership, or prerequisite check failed. |

Plan is an inspection command; `--check` makes refusal return a nonzero status.
Apply and sync fail on refused policies. Empty projected settings after refusal
do not mean enforced denial. Existing native settings can still grant access.

## Copilot and remaining work

Copilot 1.0.83 native sandbox tests reached a host Unix socket with
`allowLocalNetwork=false`. That conflicts with the draft local-network denial.
The CLI continues to refuse Copilot security projection. See the
[assessment](../WORKBENCH/evidence/COPILOT_SECURITY_ASSESSMENT.md) for the four
attempts, prerequisite setup, and temporary-directory limits.

The [60-scenario record](../WORKBENCH/evidence/SECURITY_SCENARIOS.json) separates
enforcement, refusal, lifecycle checks, deterministic-only tests, and pending
work. Model tools, approval flows, additional execution scopes, native import,
and other platforms remain later work. Claude native verification is skipped
at the user's request.
