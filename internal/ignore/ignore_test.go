package ignore

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string
		path     string
		want     bool
	}{
		{"empty matcher never matches", nil, "foo.go", false},
		{"exact basename match", []string{"foo.go"}, "foo.go", true},
		{"extension glob", []string{"*.log"}, "debug.log", true},
		{"extension glob no match", []string{"*.log"}, "debug.txt", false},
		{"directory name anywhere in path", []string{"node_modules"}, "a/node_modules/b.js", true},
		{"directory name at root", []string{"node_modules"}, "node_modules/b.js", true},
		{"nested relative glob", []string{"build/*"}, "build/output.bin", true},
		{"comment lines are ignored", []string{"# a comment", ""}, "foo.go", false},
		{"whitespace-only pattern ignored", []string{"   "}, "", false},
		{"trailing slash directory pattern matches at root", []string{"build/"}, "build/output.bin", true},
		{"trailing slash directory pattern matches nested", []string{"cache/"}, "a/cache/file.txt", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := New(tc.patterns)
			got := m.Match(tc.path)
			if got != tc.want {
				t.Errorf("Match(%q) with patterns %v = %v, want %v", tc.path, tc.patterns, got, tc.want)
			}
		})
	}
}

func TestMatchNilMatcher(t *testing.T) {
	var m *Matcher
	if m.Match("anything") {
		t.Fatal("nil matcher should never match")
	}
	if m.Patterns() != nil {
		t.Fatal("nil matcher should report no patterns")
	}
}

func TestPatternsNormalizesWhitespaceCommentsAndTrailingSlash(t *testing.T) {
	m := New([]string{" *.tmp ", "# skip me", "", "cache/"})
	got := m.Patterns()
	want := []string{"*.tmp", "cache"}
	if len(got) != len(want) {
		t.Fatalf("Patterns() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Patterns()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
