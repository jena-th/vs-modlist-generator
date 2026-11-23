# Changelog

All notable changes to this project will be documented in this file.

## [1.0.1] — 2026-11-23

### Added
- Pre-built binaries for Windows, Linux (x86_64), macOS Intel and Apple Silicon uploaded as release assets.
- `release.sh` script to automate building and publishing tagged releases via the GitHub CLI.

### Changed
- Build output folder renamed from `builds/` to `build/` and added to `.gitignore`.

## [1.0.0] — 2026-11-23 *(superseded)*

Initial release of the Vintage Story Mod List Generator.

### Added
- Single-file Go binary that scans a directory for `.zip` archives, extracts `modinfo.json`, and prints a sorted mod manifest.
- Markdown output mode (`--markdown`) producing a sortable table plus error-reporting block.
- Cross-platform build matrix (Windows, Linux, macOS).
- Conventional Commits enforcement via CI checks.

### Changed
- Replaced the original JavaScript reference implementation with idiomatic Go for zero-runtime deployment.

## [0.1.0] — 2026-11-15 *(pre-release)*

Prototype release: initial commit, project scaffolding and first working version of the mod list generator.
