# Codex acceptance project

This project is a complete canonical `.agents` example for exercising the
Open-Dot-Agents Codex adapter. It intentionally contains only categories that
the v0.0.1 Codex mapping supports: instructions, agents, skills, and tools.

Run the end-to-end acceptance flow from the repository root:

```bash
task cli:test:codex:real
```

The task copies this project into an ignored temporary directory inside the
trusted checkout, generates Codex output there, and invokes Codex against the
generated configuration. It requires an authenticated Codex CLI, network
access, `jq`, and consumes model tokens. Before live discovery, it validates
the generated TOML against OpenAI's official Codex config JSON Schema and
loads it through the installed CLI with `--strict-config`.

Do not force generation over an existing hand-authored `.codex/config.toml`
without reviewing `generate --dry-run --diff`: v0.0.1 owns and replaces the
complete generated file.
