# AGENTS

Authoritative engineering rules and project memory for `vs-modlist-generator`.

## Project Overview

**Vintage Story Mod List Generator** — a portable, self-contained Go utility that scans a directory of `.zip` archives (game mods), extracts their `modinfo.json` metadata, and outputs a sorted mod manifest in plain-text or Markdown format.

- **Language**: Go 1.27.4 (stdlib only, zero dependencies)
- **Original implementation**: JavaScript (`modlist-generator.js`) — the initial source of truth for specification and behavior; the Go rewrite is idiomatic but functionally derived from it.
- **Module path**: `github.com/jena-th/vs-modlist-generator`
- **Repository**: https://github.com/jena-th/vs-modlist-generator
- **Latest release**: v1.4.1 (2026-11-23)

## Architecture

### File layout

```
modlist-generator.go   — single-file Go application (current primary implementation)
modlist-generator.js   — initial reference implementation; **source of truth** for original spec and behavior
build.sh               — cross-platform build script (Win, Linux, macOS Intel/ARM)
build.sh              — cross-platform build script (Win, Linux, macOS Intel/ARM)
release.sh            — create GitHub release with assets via `gh` CLI
CHANGELOG.md          — versioned changelog following semver-style entries
go.mod                — Go module file, no external dependencies
```

### Design principles

- **Single-file binary** — everything lives in `modlist-generator.go`; no package splitting needed.
- **Result monad pattern** — custom `result[T]` type for error collection without panics; errors are accumulated and reported at the end rather than failing fast.
- **No external dependencies** — pure stdlib (`archive/zip`, `regexp`, `slices`, etc.). Zero-runtime deployment on any platform with Go available (or pre-built binary).
- **Portable output** — plain text by default, Markdown table via `--markdown` / `-m`.

### Core data flow

1. Read all `.zip` files from current working directory (sorted alphabetically)
2. Open each archive and extract `modinfo.json` using regex field extraction
3. Classify extracted fields into `modEntry` or `failedEntry`
4. Sort results by name, failures by zip filename
5. Render via `rawRender()` (plain text) or `mdRender()` (Markdown table + error block)

### Key types

- `result[T]` — generic result monad with `fold`, `match`, and `unwrap` helpers
- `modEntry` — successful parse: `{name, version, description, zipName}`
- `failedEntry` — failed parse: `{zipName, fields map[string]error}`

### CLI interface (docopt-style)

```
Usage:  modlist-generator [options]
Options:
  -m, --markdown    Output as a Markdown table with summary section.
                    (default: plain text one-per-line)
  -h, --help        Show this help message and exit.
```

The `helpMsg` constant lives at package level near the top of the file for code locality alongside usage comments.

## Build & Release Process

### Building locally

```bash
bash build.sh
# Outputs to build/: Windows .exe, Linux amd64, macOS Intel + ARM
```

### Releasing

1. Bump version in `go.mod` (module path) and add entry to `CHANGELOG.md`
2. Commit with conventional-commits prefix: `chore(release): v1.x.y — description`
3. Tag: `git tag v1.x.y`
4. Push branch + tags: `git push origin main --tags`
5. Create release: `gh release create v1.x.y --title "Release x.y.z" --generate-notes build/*`

Or run `bash release.sh vX.Y.Z` (requires `GITHUB_USER=jena-th` and `GITHUB_REPO=vs-modlist-generator` env vars).

### CI/CD

- **No GitHub Actions runners** — all builds are local.
- Conventional Commits enforcement exists in `.github/workflows/conventional-commits.yml`.
- Release workflow lives in `.github/workflows/release.yml`.

## Development conventions

- **Every change must be documented in `CHANGELOG.md`** — add an entry under the appropriate section (`Added`, `Changed`, `Fixed`, etc.) before committing. No undocumented changes, even trivial ones.
- Use conventional commits for all messages (`feat:`, `fix:`, `docs:`, `chore:`, etc.)
- Track work in `tasks.md` — plan features, estimate cost, and log modifications there first.
- Go formatting via `gofmt` / editor integration.
- No external test framework; the stdlib `testing` package is available if needed but currently unused — tests can be added as new `*_test.go` files alongside the main file.
- Pre-built binaries are uploaded as GitHub release assets for all four target platforms.

## Known constraints

- Only reads `.zip` archives in the current working directory (does not recurse into subdirectories).
- Regex-based field extraction from `modinfo.json` — tolerant of formatting variations but not a full JSON parser.
- Failed entries are reported, never silently dropped — every archive gets at least one outcome (success or failure detail).
