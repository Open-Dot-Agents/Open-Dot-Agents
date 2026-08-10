#!/usr/bin/env bash
set -euo pipefail

dist_dir="${DOTA_RELEASE_DIST:-dist}"
artifacts_file="$dist_dir/artifacts.json"
checksums_file="$dist_dir/checksums.txt"
publisher_dir="$dist_dir/publisher-manifests"

test -s "$artifacts_file"
test -s "$checksums_file"

assert_count() {
  local expression="$1"
  local expected="$2"
  local description="$3"
  local actual
  actual=$(jq "[$expression] | length" "$artifacts_file")
  if [ "$actual" -ne "$expected" ]; then
    echo "$description: got $actual, want $expected" >&2
    exit 1
  fi
}

assert_count '.[] | select(.type == "Archive" and .extra.ID == "bundle")' 6 "platform bundles"
assert_count '.[] | select(.type == "SBOM")' 6 "archive SBOMs"
for target in codex copilot claude; do
  assert_count ".[] | select(.type == \"Binary\" and .internal_type == 2 and .extra.ID == \"adapter-$target\")" 6 "$target installable adapters"
  while IFS=$'\t' read -r installable bundled; do
    cmp "$installable" "$bundled"
  done < <(jq -r \
    --arg release_id "adapter-$target" \
    --arg bundle_id "dota-adapter-$target-bundle" \
    '. as $artifacts
      | $artifacts[]
      | select(.type == "Binary" and .internal_type == 2 and .extra.ID == $release_id) as $release
      | [
          $release.path,
          ($artifacts[]
            | select(
                .type == "Binary"
                and .internal_type == 4
                and .extra.ID == $bundle_id
                and .goos == $release.goos
                and .goarch == $release.goarch)
            | .path)
        ]
      | @tsv' "$artifacts_file")
done

while IFS=$'\t' read -r name path; do
  expected=$(awk -v artifact="$name" '$2 == artifact { print $1 }' "$checksums_file")
  if [ -z "$expected" ]; then
    echo "checksum entry missing for $name" >&2
    exit 1
  fi
  actual=$(sha256sum "$path" | awk '{ print $1 }')
  if [ "$actual" != "$expected" ]; then
    echo "checksum mismatch for $name" >&2
    exit 1
  fi
done < <(jq -r '.[] | select(.type == "Archive" or .type == "SBOM" or (.type == "Binary" and .internal_type == 2)) | [.name, .path] | @tsv' "$artifacts_file")

while IFS= read -r path; do
  jq -e '.spdxVersion | startswith("SPDX-")' "$path" >/dev/null
done < <(jq -r '.[] | select(.type == "SBOM") | .path' "$artifacts_file")

while IFS= read -r path; do
  case "$path" in
    *.tar.gz) contents=$(tar -tzf "$path") ;;
    *.zip) contents=$(unzip -Z1 "$path") ;;
    *) echo "unexpected bundle format: $path" >&2; exit 1 ;;
  esac
  for binary in dota dota-adapter-codex dota-adapter-copilot dota-adapter-claude; do
    if ! printf '%s\n' "$contents" | sed 's/\.exe$//' | grep -Fx "$binary" >/dev/null; then
      echo "$path does not contain $binary" >&2
      exit 1
    fi
  done
done < <(jq -r '.[] | select(.type == "Archive" and .extra.ID == "bundle") | .path' "$artifacts_file")

for target in codex copilot claude; do
  manifest="$publisher_dir/org.open-dot-agents.$target.json"
  jq -e \
    --arg id "org.open-dot-agents.$target" \
    '.id == $id
      and .protocolVersion == "1.0"
      and .capabilities == ["validate", "import", "export"]
      and (.artifacts | length) == 6
      and ([.artifacts[] | (.os + "/" + .arch)] | unique | length) == 6
      and all(.artifacts[]; (.sha256 | test("^[a-f0-9]{64}$")) and (.url | startswith("https://")))' \
    "$manifest" >/dev/null
done

(
  cd "$publisher_dir"
  sha256sum -c checksums.txt >/dev/null
)

manifest_root=$(mktemp -d /tmp/dota-release-manifests.XXXXXX)
binary_root=$(mktemp -d /tmp/dota-release-binaries.XXXXXX)
cleanup() {
  case "$manifest_root" in
    /tmp/dota-release-manifests.*) rm -rf -- "$manifest_root" ;;
    *) echo "refusing to clean unexpected manifest probe path: $manifest_root" >&2 ;;
  esac
  case "$binary_root" in
    /tmp/dota-release-binaries.*) rm -rf -- "$binary_root" ;;
    *) echo "refusing to clean unexpected binary probe path: $binary_root" >&2 ;;
  esac
}
trap cleanup EXIT

cli/bin/dota init --root "$manifest_root" >/dev/null
for target in codex copilot claude; do
  cli/bin/dota adapter add \
    --root "$manifest_root" \
    --manifest "$publisher_dir/org.open-dot-agents.$target.json"
done
cli/bin/dota validate --root "$manifest_root" >/dev/null

cp -a conformance/v1/valid/minimal/. "$binary_root/"
for target in codex copilot claude; do
  adapter_path=$(jq -r ".[] | select(.type == \"Binary\" and .internal_type == 2 and .extra.ID == \"adapter-$target\" and .goos == \"linux\" and .goarch == \"amd64\") | .path" "$artifacts_file")
  test -x "$adapter_path"
  cli/bin/dota adapter add \
    --root "$binary_root" \
    --id "org.open-dot-agents.$target" \
    --version "$(jq -r '.version' "$publisher_dir/org.open-dot-agents.$target.json")" \
    --path "$adapter_path"
  cli/bin/dota conformance adapter \
    --root "$binary_root" \
    --adapter "org.open-dot-agents.$target" >/dev/null
done

echo "Release artifacts and publisher manifests passed inspection."
