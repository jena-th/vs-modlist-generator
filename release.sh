#!/usr/bin/env bash
# release.sh — build all targets, create a git tag, and push binaries as a GitHub Release.
# Usage:  ./release.sh v1.0.0   (tag name required)
set -euo pipefail

TAG="$1"

echo "=== Building all platforms ==="
bash build.sh

echo ""
echo "=== Creating release $TAG ==="
gh release create "$TAG" \
  --title "Release ${TAG#v}" \
  --generate-notes \

  build/*

echo ""
echo "Done — tag: https://github.com/${GITHUB_USER:-jena-th}/${GITHUB_REPO:-vs-modlist-generator}/releases/tag/$TAG"
