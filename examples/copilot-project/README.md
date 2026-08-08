# Real Copilot CLI acceptance project

This fixture exercises a real `.agents` projection into GitHub Copilot CLI
instructions, scoped rules, custom agents, hooks, skills, and MCP configuration.

From the repository root, run:

```sh
task cli:test:copilot:real
```

The task requires an installed and authenticated `copilot` CLI. It runs in an
isolated copy, invokes live model and MCP services, and may consume Copilot
premium requests. Failed probes and their logs are preserved at the path shown
by the task.
