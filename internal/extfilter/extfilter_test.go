package extfilter

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		name string
		exts []string
		path string
		want bool
	}{
		{"empty filter matches everything", nil, "main.go", true},
		{"empty filter matches extensionless files", nil, "Makefile", true},
		{"dot-prefixed extension matches", []string{".go"}, "main.go", true},
		{"bare extension matches", []string{"go"}, "main.go", true},
		{"case insensitive", []string{".GO"}, "main.go", true},
		{"no match", []string{".go"}, "main.txt", false},
		{"multiple extensions", []string{".go", ".txt"}, "notes.txt", true},
		{"extensionless file against configured filter", []string{".go"}, "Makefile", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := New(tc.exts)
			if got := f.Match(tc.path); got != tc.want {
				t.Errorf("Match(%q) with exts %v = %v, want %v", tc.path, tc.exts, got, tc.want)
			}
		})
	}
}

func TestMatchNilFilter(t *testing.T) {
	var f *Filter
	if !f.Match("anything.go") {
		t.Fatal("nil filter should match everything")
	}
}
