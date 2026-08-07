# Memories

Persistent knowledge that survives across sessions and context resets —
learned facts, conventions, or user preferences an agent should recall in
future sessions without being re-taught.

**Scope note vs. `../rules/` and `../instructions/`:** memories are
*implicitly/automatically captured* by the agent during a session (or saved
on explicit request, e.g. "remember that..."), not hand-authored ahead of
time like rules/instructions — and they are typically workspace- or
user-scoped rather than meant for version control alongside the rest of the
project's `.agents/` config. Windsurf's Cascade makes this split explicit:
Rules are user-authored and checked into git; Memories are auto-generated
and *not* shared across workspaces. Follow that model here: don't put
hand-written standards in `memories/` — that belongs in `../rules/` or
`../instructions/`.

**Cross-harness status:** GitHub Copilot CLI has an experimental/evolving
memories feature; Windsurf's Cascade has first-class memories; Claude Code
and OpenAI Codex CLI do not (as of this writing) expose an equivalent
first-class, file-based memory store — persistent context there is usually
folded back into `../instructions/` (e.g. appending learned facts to
`CLAUDE.md`). Treat this category as forward-looking: define a vendor-neutral
format now so tooling has somewhere to converge as more harnesses add native
memory support.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json` for
category shape and future validation integration.
