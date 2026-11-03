package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/rnikhori/go-file-watcher/internal/watcher"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	w, err := watcher.New(watcher.Options{Root: root})
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
			handleEvent(w, event)
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

func handleEvent(w *watcher.Watcher, event fsnotify.Event) {
	fmt.Printf("[%s] %-7s %s\n", time.Now().Format("2006-01-02 15:04:05"), event.Op, event.Name)

	if event.Op&fsnotify.Create != 0 {
		info, err := os.Stat(event.Name)
		if err == nil && info.IsDir() {
			if err := w.AddRecursive(event.Name); err != nil {
				log.Printf("watching new dir %s: %v", event.Name, err)
			}
		}
	}
}
