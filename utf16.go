package prosemirror

// utf16Len returns the length of a utf-8 encoded string in utf-16 code units.
// it is equivalent to `len(utf16.Encode([]rune(s)))` except it doesn't allocate the whole string thrice.
func utf16Len(s string) int {
	// we have an utf-8 encoded string, and we want to know the length in utf-16 code units.
	// https://en.wikipedia.org/wiki/UTF-16#U+010000_to_U+10FFFF

	const surrSelf = 0x10000 // maximum allowed code point for a surrogate pair.
	l := 0
	for _, r := range s {
		if r >= surrSelf {
			l += 2
			continue
		}
		l++
	}

	return l
}
