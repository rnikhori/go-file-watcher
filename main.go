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
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		log.Fatalf("resolving path: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("creating watcher: %v", err)
	}
	defer watcher.Close()

	if err := addRecursive(watcher, absRoot); err != nil {
		log.Fatalf("watching %s: %v", absRoot, err)
	}

	fmt.Printf("watching %s (recursive) — press Ctrl+C to stop\n", absRoot)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleEvent(watcher, event)
		case err, ok := <-watcher.Errors:
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

func handleEvent(watcher *fsnotify.Watcher, event fsnotify.Event) {
	fmt.Printf("[%s] %-7s %s\n", time.Now().Format("2006-01-02 15:04:05"), event.Op, event.Name)

	// A newly created directory needs its own watch, and fsnotify
	// doesn't watch subdirectories automatically.
	if event.Op&fsnotify.Create != 0 {
		info, err := os.Stat(event.Name)
		if err == nil && info.IsDir() {
			if err := addRecursive(watcher, event.Name); err != nil {
				log.Printf("watching new dir %s: %v", event.Name, err)
			}
		}
	}
}

func addRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}
