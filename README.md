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
