# Vintage Story Mod List Generator

Portable, zero-dependency mod list generator for [Vintage Story](https://vintagestory.at/) — reads all `.zip` archives in the current working directory and outputs a clean mod manifest.

[![License: GPL-3.0](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/jena-th/vs-modlist-generator?label=release&color=purple)](https://github.com/jena-th/vs-modlist-generator/releases/latest)
[![Go Stable](https://img.shields.io/badge/Go-stable-green)](https://go.dev/dl/)

## Why

Vintage Story stores each mod inside a `.zip` archive with a `modinfo.json` metadata file. There is no built-in tool that reads those archives and produces a sorted, human-readable manifest — especially not one that runs natively on Windows, Linux and macOS without installing anything extra.

This project fills that gap: drop it into your `Mods/` folder, run it once, get the list.

## Install

Download the latest pre-built binary for your platform from the [Releases](https://github.com/jena-th/vs-modlist-generator/releases/latest) page — no compiler or runtime required.

| Platform | Binary |
| -------- | ------ |
| Windows  | `modlist-generator.exe` |
| Linux    | `modlist-generator-linux-amd64` |
| macOS (Intel)     | `modlist-generator-mac-intel` |
| macOS (Apple Silicon) | `modlist-generator-mac-arm` |

**Windows** — download `modlist-generator.exe` and place it somewhere on your `PATH` (e.g. `%USERPROFILE%\AppData\Local\Microsoft\WindowsApps`) or run it directly from inside the `Mods/` folder.

**macOS** — download the binary matching your chip (`modlist-generator-mac-arm` for
Apple Silicon, `modlist-generator-mac-intel` for Intel):

```bash
curl -LO https://github.com/jena-th/vs-modlist-generator/releases/latest/download/modlist-generator-mac-arm
chmod +x modlist-generator-mac-arm
mv modlist-generator-mac-arm /usr/local/bin/modlist-generator  # Apple Silicon
```

**Linux** — download `modlist-generator-linux-amd64`:

```bash
curl -LO https://github.com/jena-th/vs-modlist-generator/releases/latest/download/modlist-generator-linux-amd64
chmod +x modlist-generator-linux-amd64
mv modlist-generator-linux-amd64 ~/bin/modlist-generator
```

> **Note:** The user-local `bin/` folder varies by distro. On Arch Linux derivatives like EndeavourOS it's typically `~/.local/bin`, so the move command would be:
>
> ```bash
> mv modlist-generator-linux-amd64 ~/.local/bin/modlist-generator
> ```

Or build from source with Go ≥ 1.20:

```bash
go build -o modlist-generator modlist-generator.go
```

## Usage

Run the executable inside your `Mods/` directory (where all `.zip` files live):

```bash
cd /path/to/VintageStory/Mods
modlist-generator              # plain text output (default)
modlist-generator --markdown   # Markdown table with summary section
```

Plain-text mode prints one mod per line (`Name Version`). Markdown mode produces a sortable table plus an error-reporting block.

### Output example (plain text)

```txt
Ancient Foundations 1.20.5
Better Furnaces 3.1.0
# ERROR `broken-mod.zip`: unable to extract field "name"
```

### Options

| Flag | Description |
| ---- | ----------- |
| `-m`, `--markdown` | Render as a Markdown table with summary section |
| `-h`, `--help`     | Show help and exit |

The command-line interface follows the [docopt](http://docopt.org/) philosophy: a human-readable usage block drives the entire parser, so every flag you see is declared in plain text near the top of the program. That means the help output is never out of sync with the source code — it *is* the documentation, embedded directly into `modlist-generator.go`.

## Platform-Specific Usage

The generator writes the mod list to **standard output** — nothing is saved to disk automatically. Redirect it into a file or pipe it to your clipboard.

### Windows

Open **Command Prompt** or **PowerShell**, navigate to your `Mods/` folder, then run:

```cmd
cd C:\Games\VintageStory\Mods
modlist-generator > modlist.txt
```

Or with Markdown output:

```cmd
modlist-generator --markdown > mods.md
```

To copy directly to the clipboard (PowerShell only):

```powershell
modlist-generator | clip
```

### macOS

Open **Terminal**, navigate to your `Mods/` folder, then run:

```bash
cd /path/to/VintageStory/Mods
modlist-generator > modlist.txt
```

Copy to the macOS clipboard instead of a file:

```bash
modlist-generator | pbcopy
```

### Linux

Open your terminal, navigate to the `Mods/` folder, then run:

```bash
cd /path/to/VintageStory/Mods
modlist-generator > modlist.txt
# or with Markdown output
modlist-generator --markdown > mods.md
```

Copy to the clipboard (requires `xclip` or `wl-copy`):

```bash
modlist-generator | xclip -selection clipboard  # X11
# or
modlist-generator | wl-copy                    # Wayland
```

## How It Works

1. Scans the working directory for `.zip` files (alphabetical order).
2. Opens each archive, locates `modinfo.json`.
3. Parses the `name` and `version` fields via regex — not because a JSON library is unavailable, but because many mod authors ship malformed or non-standard metadata: inconsistent key casing (the official docs say lowercase, some mods use TitleCase), legacy fields that confuse strict parsers, stray BOM bytes, CRLF line endings, trailing commas. The regex approach tolerates all of these without silently dropping the mod.
4. Sorts successful entries by name; failed entries are listed separately with the exact error reason.

## Documentation

Nothing deeper than what this README covers — it's a single-file Go program under 1 KB of dependencies (none at runtime).

If you want to understand the internals, read `modlist-generator.go` top-to-bottom: result monad → helpers → modinfo pipeline → zip extraction → rendering → main.

## Contributing

Contributions are welcome — open an issue or submit a PR. Please follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages and PR titles; the CI enforces this automatically.

To add new release targets, edit `build.sh` (the Go cross-compilation table).

## LLM Usage Disclosure

The Go rewrite of this tool was produced by the author with assistance from an AI coding assistant:

* **`modlist-generator.go`** — generated by **[Qwen3.6-27B-Claude-Opus-Sonnet-DistilledV2-MTP](https://huggingface.co/Brian6145/Qwen3.6-27B-Claude-Opus-Sonnet-DistilledV2-MTP-GGUF)** running through **[LM Studio](https://lmstudio.ai/)**, under the **[pi agent harness](https://pi.dev/)**. Based on the author's JavaScript reference implementation (`modlist-generator.js`). The model was instructed to port the existing logic into idiomatic Go, preserve all functionality (including BOM/CRLF handling and result monads), and keep the code self-contained.
* **`README.md`** — drafted by the same agent using the blog post ["README Best Practices"](https://markdownpreviewlive.com/blog/readme-best-practices/) as a structural template. The author supplied project-specific details (platform binaries, Vintage Story context, regex rationale) and reviewed every section before committing.

All output was read back, verified against diagnostics, and edited by the author where inconsistencies were found. The agent served as a fast draft engine; design decisions, content accuracy, and final wording remain the author's responsibility.

**Code responsibility:** all code in this repository is under the full authority of **Jena-Th**. If you find bugs, have suggestions, or want to report issues — file them against this project. No AI vendor or model maintainer is involved in any way.

## License

[GPL v3](LICENSE).
