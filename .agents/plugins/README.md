# Plugins

Installable, distributable bundles that combine any of the other
categories — e.g. a single package that ships an agent, a couple of skills,
a hook, and an MCP server together — so they can be shared and versioned as
one unit instead of copy-pasted piecemeal.

Maps to: GitHub Copilot CLI plugins (installable via a plugin marketplace).
Other harnesses do not currently have a first-class plugin/marketplace
mechanism; sharing a bundle of configuration there typically means copying
the relevant subfolders/files directly (e.g. cloning another repo's
`.claude/` contents).

Expected content: a manifest describing the plugin (name, version,
description, included categories/paths) plus the bundled category
subfolders themselves.

See `../mappings.yaml` for links to each vendor's canonical documentation.

Machine contract metadata is in `../schema/v1/agents.schema.json`.
