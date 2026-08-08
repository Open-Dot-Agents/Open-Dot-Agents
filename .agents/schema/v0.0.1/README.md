# Schema contracts (v0.0.1)

This folder contains the canonical machine-readable contracts used by the
Open-Dot-Agents adapter and validation tooling.

## Files

- `agents.schema.json`  
  Canonical contracts for `.agents/` category structure, including filenames,
  required fields, parser expectations, and supported metadata for each category.
- `mappings.schema.json`  
  Contract for `.agents/mappings.yaml` status/compatibility metadata across targets.

## Validation baseline

Tooling should use:

- `format_version: "v0.0.1"`
- `canonical_root: ".agents"`

across compatible implementations.

## Contract and compatibility

These schemas are the source of truth for structural validation and are intended to
be stable per major version while allowing controlled extension.

Target status values describe the current `oda` projection:

- `supported`: emitted directly without a structural transform
- `mapped`: emitted through a defined vendor-specific transform
- `partial`: only a documented subset is emitted
- `unsupported`: the adapter emits no output for the category

These values do not claim that a harness lacks an equivalent native feature.

See neighboring category READMEs and `.agents/mappings.yaml` for how each schema field
is used in practice.
