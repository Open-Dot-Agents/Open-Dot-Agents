# Security Policy

Report vulnerabilities privately through GitHub Security Advisories. Do not
open public issues for credential exposure, path traversal, adapter integrity,
or arbitrary-execution vulnerabilities.

Adapters are third-party executables and must be explicitly locked and verified.
Checksums provide integrity but do not establish publisher trust. Never commit
credentials to `.agents`; use environment references.
