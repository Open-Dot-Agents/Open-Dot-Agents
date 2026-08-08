# Schema contracts (v1)

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

- `format_version: 1`
- `canonical_root: ".agents"`

across compatible implementations.

## Contract and compatibility

These schemas are the source of truth for structural validation and are intended to
be stable per major version while allowing controlled extension.

See neighboring category READMEs and `.agents/mappings.yaml` for how each schema field
is used in practice.
