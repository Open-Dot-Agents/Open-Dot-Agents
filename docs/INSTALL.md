# Installing the Reference CLI

## Build from source

The reference CLI is written in Go. From a checked-out release:

```sh
cd CLI
go install ./cmd/agents
agents help
```

To build without installing:

```sh
cd CLI
go build -o bin/agents ./cmd/agents
./bin/agents help
```

## GitHub release assets

Published GitHub releases provide SHA-256-checksummed archives, per-platform
SPDX JSON SBOMs, and GitHub artifact attestations for the reference CLI. Draft
releases are for maintainer review and are not final
release assets until published. Release assets do not imply a supported native
adapter or harness claim. Download the artifact for your platform from the
published release. It contains one of:

| Platform | Archive |
| --- | --- |
| Linux amd64 | `agents-linux-amd64.tar.gz` |
| Linux arm64 | `agents-linux-arm64.tar.gz` |
| macOS amd64 | `agents-darwin-amd64.tar.gz` |
| macOS arm64 | `agents-darwin-arm64.tar.gz` |
| Windows amd64 | `agents-windows-amd64.zip` |
| Windows arm64 | `agents-windows-arm64.zip` |

In the directory containing the archive and its `.sha256` file, verify and
extract it:

```sh
# Linux
sha256sum -c agents-linux-amd64.tar.gz.sha256
tar -xzf agents-linux-amd64.tar.gz

# macOS
shasum -a 256 -c agents-darwin-arm64.tar.gz.sha256
tar -xzf agents-darwin-arm64.tar.gz
```

```powershell
# Windows PowerShell
$archive = 'agents-windows-amd64.zip'
$expected = (Get-Content "$archive.sha256").Split()[0]
$actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash
if ($actual -ne $expected) { throw 'SHA-256 checksum mismatch' }
Expand-Archive -LiteralPath $archive -DestinationPath .
```

Run `./agents version` and `./agents help` on Linux or macOS, or
`.\agents.exe version` and `.\agents.exe help` on Windows, then move the
binary to a directory on your `PATH` if desired. Verify provenance with GitHub
CLI after authenticating:

```sh
gh attestation verify agents-linux-amd64.tar.gz \
  --repo Open-Dot-Agents/Open-Dot-Agents
```

The project does not publish package-manager installations yet. Inspect the
matching `.spdx.json` file before deployment; see the [release process](RELEASING.md).
