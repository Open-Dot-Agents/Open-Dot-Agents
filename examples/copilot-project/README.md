# Real Copilot CLI acceptance project

This fixture exercises a real `.agents` projection into GitHub Copilot CLI
instructions, scoped rules, custom agents, hooks, skills, and MCP configuration.
The deterministic round-trip gate also covers prompt files, plugin metadata,
nested repository instruction paths, and both import/export directions.
It also loads the repository-level `.github/copilot/settings.json` projection
used for declarative plugin and marketplace activation.
The native CLI also mounts the generated `.github/plugin/plugin.json` through
`--plugin-dir` and confirms that the projected plugin is listed.

From the repository root, run:

```sh
task cli:test:copilot:real
```

The task requires an installed and authenticated `copilot` CLI. It runs in an
isolated copy, invokes live model and MCP services, and may consume Copilot
premium requests. Failed probes and their logs are preserved at the path shown
by the task.
