# Rules

Scoped or conditional behavioral rules — apply only in certain contexts (a
given path, language, file type, or trigger condition), as opposed to the
unconditional, always-on guidance in `../instructions/`.

**Scope note vs. `../instructions/`:** see `../instructions/README.md` for
the full distinction. In short: `instructions` = always loaded; `rules` =
conditionally applied. Cursor's `.cursor/rules/*.mdc` files confirm this is a
first-class, widely-used concept beyond OpenAI: each rule file carries
activation metadata (always-on, glob/path-matched, manually triggered, or
"model decides"). When mapping from a harness that has only one mechanism,
prefer `instructions` and use `rules` only for genuinely conditional/scoped
guidance.

Maps to: OpenAI's agent-configuration rules concept, Cursor's
`.cursor/rules/*.mdc` files, Windsurf Cascade's `.windsurf/rules/`;
conditionally-scoped instruction mechanisms in other harnesses (e.g.
path-scoped instruction files) fall here too.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`, including
the `rules/*.md` filename pattern and required `applyTo` front matter key.
