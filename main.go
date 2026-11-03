package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/rnikhori/go-file-watcher/internal/app"
	"github.com/rnikhori/go-file-watcher/internal/config"
	"github.com/rnikhori/go-file-watcher/internal/debounce"
	"github.com/rnikhori/go-file-watcher/internal/extfilter"
	"github.com/rnikhori/go-file-watcher/internal/ignore"
	"github.com/rnikhori/go-file-watcher/internal/output"
	"github.com/rnikhori/go-file-watcher/internal/summary"
	"github.com/rnikhori/go-file-watcher/internal/watcher"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("%v", err)
	}

	ignoreMatcher, err := buildIgnoreMatcher(cfg)
	if err != nil {
		log.Fatalf("loading ignore file: %v", err)
	}

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

	var debouncer *debounce.Coalescer
	if cfg.Debounce > 0 {
		debouncer = debounce.New(cfg.Debounce)
	}

	var counter *summary.Counter
	if cfg.Summary {
		counter = summary.NewCounter()
	}

	runner := &app.Runner{
		Watcher:   w,
		Ext:       extfilter.New(cfg.Ext),
		Formatter: formatter,
		Debounce:  debouncer,
		Once:      cfg.Once,
		Summary:   counter,
	}
	runner.Run(sigCh)

	fmt.Fprintln(os.Stderr, "\nstopping watcher")
	if counter != nil {
		fmt.Fprintf(os.Stderr, "summary: %s\n", counter)
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
