package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/rnikhori/go-file-watcher/internal/config"
	"github.com/rnikhori/go-file-watcher/internal/extfilter"
	"github.com/rnikhori/go-file-watcher/internal/ignore"
	"github.com/rnikhori/go-file-watcher/internal/output"
	"github.com/rnikhori/go-file-watcher/internal/watcher"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("%v", err)
	}

	ignoreMatcher, err := buildIgnoreMatcher(cfg)
	if err != nil {
		log.Fatalf("loading ignore file: %v", err)
	}
	extMatcher := extfilter.New(cfg.Ext)

	var formatter output.Formatter
	if cfg.JSON {
		formatter = output.NewJSON(os.Stdout)
	} else {
		formatter = output.NewText(os.Stdout)
	}

	w, err := watcher.New(watcher.Options{Root: cfg.Root, Ignore: ignoreMatcher})
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer w.Close()

	fmt.Fprintf(os.Stderr, "watching %s (recursive) — press Ctrl+C to stop\n", w.Root())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case event, ok := <-w.Events():
			if !ok {
				return
			}
			handleEvent(w, event, extMatcher, formatter)
		case err, ok := <-w.Errors():
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		case <-sigCh:
			fmt.Fprintln(os.Stderr, "\nstopping watcher")
			return
		}
	}
}

func buildIgnoreMatcher(cfg *config.Config) (*ignore.Matcher, error) {
	patterns := append([]string(nil), cfg.IgnorePatterns...)

	ignoreFilePath := cfg.IgnoreFile
	if !filepath.IsAbs(ignoreFilePath) {
		ignoreFilePath = filepath.Join(cfg.Root, ignoreFilePath)
	}
	filePatterns, err := ignore.LoadFile(ignoreFilePath)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, filePatterns...)

	return ignore.New(patterns), nil
}

func handleEvent(w *watcher.Watcher, event fsnotify.Event, ext *extfilter.Filter, formatter output.Formatter) {
	if w.IsIgnored(event.Name) {
		return
	}

	isDir := false
	if event.Op&fsnotify.Create != 0 {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			isDir = true
			if err := w.AddRecursive(event.Name); err != nil {
				log.Printf("watching new dir %s: %v", event.Name, err)
			}
		}
	}

	if !isDir && !ext.Match(event.Name) {
		return
	}

	if err := formatter.Write(event, time.Now()); err != nil {
		log.Printf("writing event: %v", err)
	}
}
