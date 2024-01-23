package prosemirror

import (
	"testing"
	"unicode/utf16"
)

// FuzzUtf16Len tests the utf16Len function for correctness by comparing
// its output to the standard library's utf16.Encode function.
// fuzz: elapsed: 9m6s, execs: 27876807 (55584/sec), new interesting: 16 (total: 16)
func FuzzUtf16Len(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string) {
		expected := len(utf16.Encode([]rune(s)))
		actual := utf16Len(s)

		if expected != actual {
			t.Errorf("utf16Len(%q) = %d; want %d", s, actual, expected)
		}
	})
}
