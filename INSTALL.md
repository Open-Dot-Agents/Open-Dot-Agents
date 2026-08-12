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

## Release candidates

The release-candidate workflow publishes unsigned, SHA-256-checksummed
archives for evaluation. They are workflow artifacts, not final release
assets, and do not imply a supported adapter or harness claim. Download the
artifact for your platform from the matching workflow run and unpack GitHub's
artifact-download wrapper. It contains one of:

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

Run `./agents help` on Linux or macOS, or `.\agents.exe help` on Windows,
then move the binary to a directory on your `PATH` if desired. The project
does not publish package-manager installations. For non-candidate use, select
an applicable published release asset or build a reviewed tagged source
checkout. Do not assume a release asset has a signature, SBOM, or provenance
unless its release documentation explicitly provides it; see the
[release process](RELEASING.md).
