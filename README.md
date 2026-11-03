# go-file-watcher

A CLI that recursively watches a directory for file system changes and prints each event with a timestamp.

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
