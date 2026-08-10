#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: $0 v1.x.y[-prerelease]" >&2
  exit 2
fi

release_tag="$1"
if [[ ! "$release_tag" =~ ^v1\.[0-9]+\.[0-9]+(-[0-9A-Za-z][0-9A-Za-z.-]*)?$ ]]; then
  echo "release tag must be a v1 semantic version: $release_tag" >&2
  exit 2
fi

release_version="${release_tag#v}"
repository="${DOTA_RELEASE_REPOSITORY:-Open-Dot-Agents/Open-Dot-Agents}"
dist_dir="${DOTA_RELEASE_DIST:-dist}"
artifacts_file="$dist_dir/artifacts.json"
checksums_file="$dist_dir/checksums.txt"
output_dir="$dist_dir/publisher-manifests"

test -s "$artifacts_file"
test -s "$checksums_file"
mkdir -p "$output_dir"

for target in codex copilot claude; do
  archive_id="adapter-$target"
  adapter_id="org.open-dot-agents.$target"
  case "$target" in
    codex) adapter_name="Open-Dot-Agents Codex adapter" ;;
    copilot) adapter_name="Open-Dot-Agents Copilot adapter" ;;
    claude) adapter_name="Open-Dot-Agents Claude adapter" ;;
  esac

  jq \
    --arg archive_id "$archive_id" \
    --arg adapter_id "$adapter_id" \
    --arg adapter_name "$adapter_name" \
    --arg release_version "$release_version" \
    --arg release_base "https://github.com/$repository/releases/download/$release_tag" \
    --rawfile checksums "$checksums_file" \
    '
      def digest($name):
        ($checksums
          | split("\n")
          | map(select(endswith("  " + $name)))) as $matches
        | if ($matches | length) != 1 then
            error("missing or duplicate checksum for " + $name)
          else
            ($matches[0] | split("  ")[0])
          end;
      [
        .[]
        | select(
            .type == "Binary"
            and .internal_type == 2
            and .extra.ID == $archive_id)
      ] as $artifacts
      | if ($artifacts | length) != 6 then
          error($archive_id + " must have exactly six platform artifacts")
        else
          {
            "$schema": "https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/adapter-publisher.schema.json",
            id: $adapter_id,
            name: $adapter_name,
            version: $release_version,
            protocolVersion: "1.0",
            capabilities: ["validate", "import", "export"],
            artifacts: (
              $artifacts
              | sort_by(.goos, .goarch)
              | map(. as $artifact | {
                  os: $artifact.goos,
                  arch: $artifact.goarch,
                  url: ($release_base + "/" + $artifact.name),
                  sha256: digest($artifact.name)
                }))
          }
        end
    ' "$artifacts_file" >"$output_dir/$adapter_id.json"
done

(
  cd "$output_dir"
  sha256sum ./*.json >checksums.txt
)

echo "Generated publisher manifests in $output_dir"
