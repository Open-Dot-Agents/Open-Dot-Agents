# Security Policy

## Supported components

Security fixes are accepted for the current releases of the specification,
reference CLI, and published conformance tooling. Experimental workbench
mappings are best-effort until they graduate to a supported adapter.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Report it privately
to the maintainers through the repository's GitHub private vulnerability
reporting flow. Include affected component and version, reproduction steps,
impact, and any proposed mitigation.

Maintainers will acknowledge reports within seven days, assess severity,
coordinate a fix, and publish an advisory once users can update safely.

## Configuration safety

Never place credentials, tokens, private URLs, or customer configuration in
issues, examples, fixtures, or commits. Canonical files may reference
environment-variable names, but adapters must not expose secret values in
diagnostics, diffs, backups, or generated examples.
