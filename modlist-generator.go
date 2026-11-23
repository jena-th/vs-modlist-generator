package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// helpMsg — CLI usage string (docopt-style).  Lives as high as possible in the
// file; Go requires imports to appear before other declarations.
const helpMsg = `
Usage:
  modlist-generator [options]

Options:
  -m, --markdown       Output as a Markdown table with summary section.
                       (default: plain text one-per-line)
  -h, --help           Show this help message and exit.

Notes:
  Reads all .zip files in the current working directory. Each archive is
  opened for its modinfo.json entry; if the file is missing or malformed
  the mod lands in a "Failed Entries" block instead of being silently dropped.
`

// ---------------------------------------------------------------------------
// Result monad — Go-friendly, no panics.  Errors are collected and reported.
// ---------------------------------------------------------------------------

type result[T any] struct {
	value   T
	err     error
	isError bool
}

func ok[T any](v T) result[T]         { return result[T]{value: v, isError: false} }
func fail[T any](err error) result[T] { return result[T]{err: err, isError: true} }

// fold converts a result into the same output type with separate handlers.
func (r result[T]) fold(failFn func(error), okFn func(T)) {
	if r.isError {
		failFn(r.err)
	} else {
		okFn(r.value)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var lineJoin = func(lines []string) string { return strings.Join(lines, "\n") }

func escapeRegex(s string) string {
	return regexp.QuoteMeta(s)
}

func extractValue(raw, key string) result[string] {
	re := fmt.Sprintf(`"%s"\s*:\s*"((?:\\.|[^"\\])*)",`, escapeRegex(key))
	matched, err := regexp.MatchString(re, raw)
	if err != nil {
		return fail[string](fmt.Errorf("regex error: %w", err))
	}
	if !matched {
		re = fmt.Sprintf(`"%s"\s*:\s*"((?:\\.|[^"\\])*)"`, escapeRegex(key))
		matched, err := regexp.MatchString(re, raw)
		if err != nil || !matched {
			return fail[string](fmt.Errorf("unable to extract field %q", key))
		}
		rx := regexp.MustCompile(re)
		submatch := rx.FindStringSubmatch(raw)
		return ok(submatch[1])
	}
	rex := regexp.MustCompile(re)
	submatch := rex.FindStringSubmatch(raw)
	if submatch == nil {
		return fail[string](fmt.Errorf("unable to extract field %q", key))
	}
	return ok(submatch[1])
}

func extractValuesObject(keys ...string) func(string) map[string]result[string] {
	return func(raw string) map[string]result[string] {
		out := make(map[string]result[string], len(keys))
		for _, k := range keys {
			out[k] = extractValue(raw, k)
		}
		return out
	}
}

func partition[K comparable, V any](fields map[K]result[V]) (success map[K]V, failed map[K]error) {
	success = make(map[K]V)
	failed = make(map[K]error)
	for key, r := range fields {
		if r.isError {
			failed[key] = r.err
		} else {
			success[key] = r.value
		}
	}
	return
}

// ---------------------------------------------------------------------------
// modinfo pipeline
// ---------------------------------------------------------------------------

func processModInfo(data []byte) map[string]result[string] {
	text := string(bytes.ReplaceAll(data, []byte{0xEF, 0xBB, 0xBF}, nil)) // strip BOM
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = normalizeLineEndingsAndTrailingCommas(text)

	// Lower-case the first character of every JSON key to match original behaviour.
	text = lowerFirstKeyChar(text)

	return extractValuesObject("name", "version")(text)
}

func normalizeLineEndingsAndTrailingCommas(s string) string {
	// Remove trailing commas before closing braces: ,\s+\}  →  \}
	re := regexp.MustCompile(`,\s+\}`)
	return re.ReplaceAllString(s, "}")
}

// lowerFirstKeyChar transforms JSON keys whose first character is uppercase,
// so that field lookups work regardless of how the mod author capitalised the key.
// It targets only bare-key tokens (no preceding \",) to avoid touching values.
var keyStart = regexp.MustCompile(`(?m)^\s*"([A-Za-z])`)

func lowerFirstKeyChar(in string) string {
	return keyStart.ReplaceAllStringFunc(in, func(m string) string {
		trimmed := strings.TrimLeft(m, " \t")
		prefix := m[:len(m)-len(trimmed)]
		if len(trimmed) >= 2 && trimmed[1] >= 'A' && trimmed[1] <= 'Z' {
			return fmt.Sprintf("%s\"%c", prefix, trimmed[1]+('a'-'A'))
		}
		return m
	})
}

// ---------------------------------------------------------------------------
// Zip extraction (pure Go — no OS dependency)
// ---------------------------------------------------------------------------

func extractModInfoFromZip(zipPath string) result[map[string]result[string]] {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fail[map[string]result[string]](fmt.Errorf("cannot open %s: %w", zipPath, err))
	}
	defer zr.Close()

	for _, f := range zr.File {
		if !f.FileInfo().IsDir() && strings.EqualFold(f.Name, "modinfo.json") {
			rc, err := f.Open()
			if err != nil {
				return fail[map[string]result[string]](fmt.Errorf("cannot open entry %s: %w", f.Name, err))
			}

			buf := bytes.NewBuffer(nil)
			_, copyErr := buf.ReadFrom(rc)
			rc.Close()
			if copyErr != nil {
				return fail[map[string]result[string]](copyErr)
			}
			return ok(processModInfo(buf.Bytes()))
		}
	}

	// Try with path separator fallback (Windows zip may contain backslashes).
	for _, f := range zr.File {
		nameLower := strings.ToLower(f.Name)
		if !f.FileInfo().IsDir() && (strings.EqualFold(filepath.ToSlash(f.Name), "modinfo.json") ||
			strings.Contains(nameLower, "modinfo.json")) {
			rc, err := f.Open()
			if err != nil {
				return fail[map[string]result[string]](fmt.Errorf("cannot open entry %s: %w", f.Name, err))
			}

			buf := bytes.NewBuffer(nil)
			_, copyErr := buf.ReadFrom(rc)
			rc.Close()
			if copyErr != nil {
				return fail[map[string]result[string]](copyErr)
			}
			return ok(processModInfo(buf.Bytes()))
		}
	}

	return fail[map[string]result[string]](fmt.Errorf("modinfo.json not found in archive"))
}

// ---------------------------------------------------------------------------
// Mod record & failure tracking
// ---------------------------------------------------------------------------

type modEntry struct {
	name    string
	version string
	zipName string
}

type failedEntry struct {
	zipName string
	fields  map[string]error
}

func classify(fields map[string]result[string], zipFile string) (modEntry, *failedEntry) {
	success, fieldErrors := partition(fields)

	if len(fieldErrors) > 0 {
		return modEntry{}, &failedEntry{zipName: zipFile, fields: fieldErrors}
	}

	var name, version string
	for k, v := range success {
		switch k {
		case "name":
			name = v
		case "version":
			version = v
		}
	}
	return modEntry{zipName: zipFile, name: name, version: version}, nil
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

func rawRender(results []modEntry, failed []*failedEntry) string {
	var lines []string
	for _, m := range results {
		lines = append(lines, fmt.Sprintf("%s %s", m.name, m.version))
	}
	for _, f := range failed {
		errParts := make([]string, 0, len(f.fields))
		for _, err := range f.fields {
			errParts = append(errParts, err.Error())
		}
		slices.Sort(errParts) // deterministic output
		lines = append(lines, fmt.Sprintf("# ERROR `%s`: %s", f.zipName, strings.Join(errParts, ", ")))
	}
	return strings.Join(lines, "\n") + "\n"
}

func mdRender(results []modEntry, failed []*failedEntry) string {
	var lines []string
	lines = append(lines, "# Mod List\n")
	lines = append(lines, "| Name | Version |")
	lines = append(lines, "| ------ | --------- |")
	for _, m := range results {
		lines = append(lines, fmt.Sprintf("| `%s` | `%s` |", m.name, m.version))
	}
	lines = append(lines, "")
	lines = append(lines, "## Summary")
	lines = append(lines, "")
	totalStr := fmt.Sprintf("Total entries: **%d**", len(results))
	if len(failed) > 0 {
		totalStr += fmt.Sprintf(" (%d missing metadata)", len(failed))
	}
	lines = append(lines, totalStr)

	if len(failed) > 0 {
		lines = append(lines, "")
		lines = append(lines, "## Failed Entries")
		for _, f := range failed {
			errParts := make([]string, 0, len(f.fields))
			for _, err := range f.fields {
				errParts = append(errParts, err.Error())
			}
			slices.Sort(errParts)
			lines = append(lines, fmt.Sprintf("* `%s` - %s", f.zipName, strings.Join(errParts, ", ")))
		}
	}

	return lineJoin(lines) + "\n"
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			fmt.Print(helpMsg)
			return
		}
	}

	modDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
		return // unreachable but satisfies compiler
	}

	useMarkdown := false
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, "--markdown") || strings.EqualFold(arg, "-m") {
			useMarkdown = true
			break
		}
	}

	// Collect .zip files.
	entries, err := os.ReadDir(modDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read directory %s: %v\n", modDir, err)
		os.Exit(1)
		return // unreachable but satisfies compiler
	}

	var zipFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".zip") {
			zipFiles = append(zipFiles, e.Name())
		}
	}
	slices.Sort(zipFiles)

	if len(zipFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No .zip files found in Mods directory.")
		os.Exit(0)
		return // unreachable but satisfies compiler
	}

	var results []modEntry
	var failures []*failedEntry

	for _, zf := range zipFiles {
		r := extractModInfoFromZip(filepath.Join(modDir, zf))

		r.fold(
			func(err error) {
				failures = append(failures, &failedEntry{zipName: zf, fields: map[string]error{"open": err}})
			},
			func(fields map[string]result[string]) {
				entry, fail := classify(fields, zf)
				if fail != nil {
					failures = append(failures, fail)
				} else if entry.name != "" || entry.version != "" {
					results = append(results, entry)
				}
			},
		)
	}

	// Sort results alphabetically by name.
	slices.SortFunc(results, func(a, b modEntry) int {
		if a.name == "" && b.name == "" {
			return 0
		}
		return strings.Compare(a.name, b.name)
	})

	// Sort failures alphabetically by zip file name.
	slices.SortFunc(failures, func(a, b *failedEntry) int {
		return strings.Compare(a.zipName, b.zipName)
	})

	output := rawRender(results, failures)
	if useMarkdown {
		output = mdRender(results, failures)
	}

	fmt.Print(output)
}
