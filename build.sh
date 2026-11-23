#!/usr/bin/env bash
set -euo pipefail

SRC="modlist-generator.go"
OUT="builds"
mkdir -p "$OUT"

targets=(
  "windows amd64 modlist-generator.exe"
  "linux   amd64 modlist-generator-linux-amd64"
  "darwin  amd64 modlist-generator-mac-intel"
  "darwin  arm64 modlist-generator-mac-arm"
)

for target in "${targets[@]}"; do
  read -r goos arch name <<< "$target"
  echo "→ $goos/$arch"
  GOOS=$goos GOARCH=$arch go build -o "$OUT/$name" "$SRC"
done

echo "Done. Binaries in $OUT/"
ls -lh "$OUT/"
