# Copilot parent skills

Copilot `1.0.83` inherits packages from parent `.github/skills`, `.agents/skills`,
and `.claude/skills` directories. The nearest Git repository boundary limits
discovery. A local definition can override the same inherited skill name.

Configure the owning parent explicitly:

```sh
agents import --vendor copilot --root /workspace/monorepo --experimental
agents plan --vendor copilot --root /workspace/monorepo --experimental
agents apply --vendor copilot --root /workspace/monorepo --experimental
cd /workspace/monorepo/packages/child
copilot skill list --json
```

Import reads recognized paths under the selected root and preserves complete
packages in that root's `.agents/skills`. Copilot can discover these canonical
packages from a descendant working directory. Apply does not create copies in
each child. Import at a child root does not capture parent packages.

Plans expose the parent mapping and describe inherited discovery as an external
source for child operations. This is a boundary declaration, not an enumeration
of every ancestor package. The adapter does not read or own those packages
through the child root. Use the native catalogue from the intended working
directory to inspect effective and shadowed sources. Existing native trust and
repository boundaries remain in force; the adapter does not grant trust.

The [lifecycle verifier](../WORKBENCH/conformance/verify_copilot_parent_skills.py)
checks six native sessions: source inheritance, relocation of the canonical
parent, a parent asset update, a local child override, fallback after removal
of that override, and a nested Git repository that stops inheritance. Each
active case has a correlated skill load, approved command, and observable asset
effect. The absent case has no command approval or effect. Import preserves
the source bytes, permissions, inodes, and modification times. Child operations
leave parent assets unchanged.

Six additional catalogue probes cover all three native parent directories,
with no Git repository, one parent repository, and a nested repository. They
confirm source provenance and local precedence. These tests use Copilot only;
they are not Claude Code native tests.

The first lifecycle attempt used the wrong evidence field name and remains
stored. The corrected runner retains native session events and rejects missing
or uncorrelated effects. No adapter support promotion follows from this bounded
mapping.
