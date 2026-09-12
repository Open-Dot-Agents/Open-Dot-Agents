# Copilot user skill import and projection

Draft.2 can import personal skills from an explicitly selected `.copilot` home
or shared skills from an explicitly selected `.agents` home. Both sources use
the portable `skills` profile. Apply writes user skills to the explicitly
selected Copilot home. It does not scan a sibling home or a parent repository.

For a dedicated draft.2 configuration repository, shared skill import can use:

```sh
agents import --vendor copilot --experimental --scope user \
  --root /path/to/config-repository --native-home "$HOME/.agents"
agents plan --vendor copilot --experimental --scope user \
  --root /path/to/config-repository --native-home "$HOME/.copilot"
agents apply --vendor copilot --experimental --scope user \
  --root /path/to/config-repository --native-home "$HOME/.copilot"
```

Use `.copilot` as the import home for personal Copilot skills. The selected
home is a source directory during import and a destination during apply.
Native trust remains unchanged. Project skills continue to use `.agents/skills`
directly.

## Package and ownership safety

The audit reproduced two cases in both adapters where individual file merges
created a package that existed in neither source. User import kept an old
canonical asset and added a new native asset. User adoption likewise combined
unowned target assets with canonical assets. Both operations now compare whole
packages. Identical packages can be retained or adopted; different packages
under the same directory name refuse. Copilot native-name collisions also
refuse before writes. Force does not override these checks or foreign ownership.

Owned package updates and removals remain supported. An empty `.gitkeep` is
removed when user import selects actual packages. Private backups for projected
user skills now use the XDG state tree, indexed by vendor and native home,
under `native/<vendor>/backups/<home-hash>/skills/`. A sidecar backup inside a
skill would become an active package asset. Existing conflicting backups
remain protected. Older sidecar backups need separate review before adoption.

Go regressions cover both vendors, repeated import, package conflicts,
adoption, native identity conflicts, marker removal, owned updates, private
backups, and rollback after the final removal write. Source packages remain
unchanged. This cycle does not establish native Codex skill execution.

## Native evidence

The source receipt (`WORKBENCH/evidence/native-draft2-debug/copilot-user-skills.sources.json`)
retains GitHub's documentation index and skill-location table. The table lists
personal Copilot and shared agent skills after project and inherited skills.
The adapter does not infer that this priority authorizes copying other homes.

Native tests use Copilot `1.0.83`, binary SHA-256
`a3262c4513ef1fc2ca21485261ca73196977ad76bd5e7990fb572f6134aaeedd`,
isolated homes, and a local model provider. Runtime workspaces are separate
from the canonical repository. The native catalogue must report
`personal-copilot` or `personal-agents` at the expected path. Eight sessions
cover source discovery, relocated execution, updated asset execution, and
removal for both origins. Model requests, skill invocation events, exact
command approvals, completed commands, and asset-derived file effects are
correlated. Removing the user skill stops its discovery and execution.

Receipts are `copilot-user-skills-{copilot,agents}-verified.json`. The
[verifier](../WORKBENCH/conformance/verify_copilot_user_skills.py) checks native
provenance and execution separately from copied-file hashes. Its tests reject
project fallback, stale effects, wrong paths, mismatched approvals, and reused
loaded phases after removal. Every adapter call records unchanged external
configuration hashes.

Earlier failures remain recorded: package merge, partial adoption, marker
retention, sidecar backup pollution, a fixture command denied because the test
client received a directory instead of the exact executable path, and project
discovery masking user removal. A cleanup receipt records removal of four
ownership records from expired test directories after the initial regression
omitted its isolated XDG state setting. Tests now set that directory explicitly.

These results do not prove arbitrary model compliance, project precedence,
inherited discovery, live reload, or full adapter support. Stable 1.0 and
draft.1 semantics remain unchanged. The wider milestone remains incomplete.
