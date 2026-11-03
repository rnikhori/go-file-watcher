// Package output formats filesystem events for either human-readable text
// or line-delimited JSON.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event is the JSON representation of a reported filesystem event.
type Event struct {
	Time string `json:"time"`
	Op   string `json:"op"`
	Path string `json:"path"`
}

// Formatter writes a single event to an underlying stream.
type Formatter interface {
	Write(event fsnotify.Event, when time.Time) error
}

// TextFormatter writes events as timestamped plain text lines.
type TextFormatter struct {
	w io.Writer
}

// NewText creates a TextFormatter writing to w.
func NewText(w io.Writer) *TextFormatter {
	return &TextFormatter{w: w}
}

// Write implements Formatter.
func (f *TextFormatter) Write(event fsnotify.Event, when time.Time) error {
	_, err := fmt.Fprintf(f.w, "[%s] %-7s %s\n", when.Format("2006-01-02 15:04:05"), event.Op, event.Name)
	return err
}

// JSONFormatter writes one JSON object per line, suitable for machine
// consumption.
type JSONFormatter struct {
	enc *json.Encoder
}

// NewJSON creates a JSONFormatter writing to w.
func NewJSON(w io.Writer) *JSONFormatter {
	return &JSONFormatter{enc: json.NewEncoder(w)}
}

// Write implements Formatter.
func (f *JSONFormatter) Write(event fsnotify.Event, when time.Time) error {
	return f.enc.Encode(Event{
		Time: when.Format(time.RFC3339),
		Op:   event.Op.String(),
		Path: event.Name,
	})
}
