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

	w, err := watcher.New(watcher.Options{Root: cfg.Root, Ignore: ignoreMatcher})
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer w.Close()

	fmt.Printf("watching %s (recursive) — press Ctrl+C to stop\n", w.Root())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case event, ok := <-w.Events():
			if !ok {
				return
			}
			handleEvent(w, event, extMatcher)
		case err, ok := <-w.Errors():
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		case <-sigCh:
			fmt.Println("\nstopping watcher")
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

func handleEvent(w *watcher.Watcher, event fsnotify.Event, ext *extfilter.Filter) {
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

	fmt.Printf("[%s] %-7s %s\n", time.Now().Format("2006-01-02 15:04:05"), event.Op, event.Name)
}
