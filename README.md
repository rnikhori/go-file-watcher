# go-file-watcher

A CLI that recursively watches a directory for file system changes and prints each event with a timestamp.

## Contents

- [Build & run](#build--run)
- [Build a binary](#build-a-binary)
- [Ignoring paths](#ignoring-paths)
- [Filtering by extension](#filtering-by-extension)
- [JSON output](#json-output)
- [Debouncing rapid events](#debouncing-rapid-events)
- [Exiting after the first event](#exiting-after-the-first-event)
- [Summary on exit](#summary-on-exit)
- [Flags reference](#flags-reference)

## Build & run

```sh
go run . /path/to/dir
```

If no path is given, it watches the current directory (`.`).

With the watcher running, open another terminal and create, edit, or remove a file inside the watched directory to see events printed. Press `Ctrl+C` to stop.

## Build a binary

```sh
go build -o file-watcher .
./file-watcher /path/to/dir
```

## Ignoring paths

Exclude paths from being watched and reported with gitignore-style glob
patterns, either inline:

```sh
go run . --ignore='*.log,node_modules,build' /path/to/dir
```

or via a `.watchignore` file in the watched directory (one pattern per
line, blank lines and `#` comments are skipped):

```
# dependency directories
node_modules
vendor

# build output
build
*.log
```

Patterns are matched against the relative path, its basename, and each
path segment, so `node_modules` excludes that directory anywhere in the
tree, not just at the root. Use `--ignore-file` to point at a different
file (relative to the watched root unless given as an absolute path).

## Filtering by extension

Only report events for files with specific extensions:

```sh
go run . --ext=.go,.txt /path/to/dir
```

The leading dot is optional (`--ext=go,txt` also works) and matching is
case-insensitive. Directory create/remove events are always processed
(so new subdirectories keep getting watched) regardless of `--ext`.

## JSON output

For machine consumption, emit one JSON object per line instead of plain
text:

```sh
go run . --json /path/to/dir
```

```json
{"time":"2024-01-02T03:04:05Z","op":"CREATE","path":"/path/to/dir/foo.go"}
```

The startup banner and shutdown message are written to stderr, so stdout
carries only event output and can be piped directly into `jq` or similar
tools.

## Debouncing rapid events

Editors and build tools often produce several events for a single logical
save. Coalesce events for the same path within a window into one:

```sh
go run . --debounce=300ms /path/to/dir
```

The first event for a path is always reported immediately; subsequent
events for that same path within the window are suppressed. The window
resets from the last *reported* event, so a sustained burst still
produces periodic output rather than being silenced indefinitely.

## Exiting after the first event

For scripting or CI, watch until the first matching event (after any
ignore/extension filtering) and then exit automatically:

```sh
go run . --once --ext=.go /path/to/dir
```

Combine with `--json` for a single machine-readable line describing
whatever changed, then let the calling script decide what to do next.

## Summary on exit

Print a count of events by operation type when the watcher stops:

```sh
go run . --summary /path/to/dir
# ^C
# stopping watcher
# summary: total=7 CREATE=3 REMOVE=1 WRITE=3
```

The summary is written to stderr, same as the startup banner and
shutdown message.

## Flags reference

| Flag | Default | Description |
| --- | --- | --- |
| `--ignore` | *(none)* | Comma-separated glob patterns to exclude from watching and reporting. |
| `--ignore-file` | `.watchignore` | Gitignore-style file (relative to the watched root) with additional ignore patterns. |
| `--ext` | *(all)* | Comma-separated file extensions to report, e.g. `.go,.txt`. |
| `--json` | `false` | Emit one JSON object per line instead of plain text. |
| `--debounce` | `0` (disabled) | Coalesce repeated events for the same path within this window into one, e.g. `300ms`. |
| `--once` | `false` | Exit after the first matching event is reported. |
| `--summary` | `false` | Print a count of events by operation type on exit. |

Run `file-watcher --help` (or `go run . --help`) to see this list along
with a few example invocations.
