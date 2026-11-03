// Package config parses command-line flags into a Config the rest of the
// program can act on, keeping flag definitions out of main().
package config

import (
	"flag"
	"strings"
	"time"
)

// Config holds the parsed CLI options for a single run.
type Config struct {
	// Root is the directory to watch, resolved from the first positional
	// argument (defaults to the current directory).
	Root string
	// IgnorePatterns are glob patterns supplied via --ignore, in addition to
	// whatever is loaded from an ignore file.
	IgnorePatterns []string
	// IgnoreFile is the path (relative to Root, unless absolute) of a
	// gitignore-style file to load additional patterns from.
	IgnoreFile string
	// Ext restricts reported events to files with these extensions. An
	// empty slice means no restriction.
	Ext []string
	// JSON selects line-delimited JSON output instead of plain text.
	JSON bool
	// Debounce is the window within which repeated events for the same
	// path are coalesced into one. Zero disables debouncing.
	Debounce time.Duration
}

// Parse parses args (typically os.Args[1:]) into a Config.
func Parse(args []string) (*Config, error) {
	fs := flag.NewFlagSet("file-watcher", flag.ContinueOnError)

	ignoreFlag := fs.String("ignore", "", "comma-separated glob patterns to exclude from watching and reporting")
	ignoreFile := fs.String("ignore-file", ".watchignore", "gitignore-style file (relative to the watched root) with additional ignore patterns")
	extFlag := fs.String("ext", "", "comma-separated file extensions to report (e.g. .go,.txt); all files reported if empty")
	jsonFlag := fs.Bool("json", false, "emit one JSON object per line instead of plain text")
	debounceFlag := fs.Duration("debounce", 0, "coalesce repeated events for the same path within this window into one (e.g. 300ms)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	root := "."
	if fs.NArg() > 0 {
		root = fs.Arg(0)
	}

	cfg := &Config{
		Root:       root,
		IgnoreFile: *ignoreFile,
		JSON:       *jsonFlag,
		Debounce:   *debounceFlag,
	}
	if strings.TrimSpace(*ignoreFlag) != "" {
		for _, p := range strings.Split(*ignoreFlag, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.IgnorePatterns = append(cfg.IgnorePatterns, p)
			}
		}
	}
	if strings.TrimSpace(*extFlag) != "" {
		for _, e := range strings.Split(*extFlag, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				cfg.Ext = append(cfg.Ext, e)
			}
		}
	}
	return cfg, nil
}
