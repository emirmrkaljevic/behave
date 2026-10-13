#!/usr/bin/env bash
# Build MCPB bundles from goreleaser archives and write server.json for the
# MCP Registry. Usage: scripts/mcpb.sh <version> <dist-dir> <out-dir>
set -euo pipefail
version=$1 dist=$2 out=$3
repo=https://github.com/emirmrkaljevic/behave
mkdir -p "$out"
packages=()
for os in linux darwin; do
  for arch in amd64 arm64; do
    tgz="$dist/behave_${version}_${os}_${arch}.tar.gz"
    [ -f "$tgz" ] || { echo "missing $tgz" >&2; exit 1; }
    work=$(mktemp -d)
    mkdir -p "$work/server"
    tar -xzf "$tgz" -C "$work/server" behave
    cat > "$work/manifest.json" <<JSON
{
  "manifest_version": "0.2",
  "name": "behave",
  "display_name": "behave",
  "version": "$version",
  "description": "Record how a command-line tool behaves and check that another implementation behaves the same.",
  "author": { "name": "Emir Mrkaljević", "url": "https://github.com/emirmrkaljevic" },
  "repository": { "type": "git", "url": "$repo" },
  "homepage": "$repo",
  "license": "MIT",
  "server": {
    "type": "binary",
    "entry_point": "server/behave",
    "mcp_config": { "command": "\${__dirname}/server/behave", "args": ["mcp"] }
  },
  "compatibility": { "platforms": ["$os"] }
}
JSON
    file="behave_${version}_${os}_${arch}.mcpb"
    python3 -c "import shutil,sys; shutil.make_archive(sys.argv[1], 'zip', sys.argv[2])" "$out/$file.tmp" "$work" && mv "$out/$file.tmp.zip" "$out/$file"
    rm -rf "$work"
    sha=$(sha256sum "$out/$file" | cut -d' ' -f1)
    packages+=("{\"registryType\":\"mcpb\",\"identifier\":\"$repo/releases/download/v$version/$file\",\"fileSha256\":\"$sha\",\"transport\":{\"type\":\"stdio\"}}")
  done
done
pk=$(IFS=,; echo "${packages[*]}")
cat > "$out/server.json" <<JSON
{
  "\$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.emirmrkaljevic/behave",
  "title": "behave",
  "description": "Record how any CLI behaves and check that a rewrite behaves the same (golden tests).",
  "repository": { "url": "$repo", "source": "github" },
  "version": "$version",
  "packages": [$pk]
}
JSON
jq . "$out/server.json" >/dev/null
echo "wrote $(ls "$out" | wc -l) files to $out"
