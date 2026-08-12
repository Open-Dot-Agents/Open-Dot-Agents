# Versioning and Compatibility

The specification and reference CLI use Semantic Versioning.

| Release | Meaning |
| --- | --- |
| Major | A breaking change to canonical semantics, schema, profile behavior, or CLI contract. |
| Minor | A backward-compatible profile, field, adapter capability, or command addition. |
| Patch | A compatible clarification, defect fix, security fix, or documentation correction. |

Every canonical `.agents/` manifest declares the specification version it
targets. Adapters report the specification versions and profiles they support.
A projection is conformant only when the compatibility matrix records a
passing result for the exact adapter and harness versions.

Extension fields must use a namespaced key. An adapter that does not support an
extension must preserve it when possible or emit an explicit diagnostic. It
must never silently remove portable data.
