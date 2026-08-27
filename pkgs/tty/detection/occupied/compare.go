// Package occupied classifies live composer occupancy by injecting a space
// and comparing before/after snapshots.
package occupied

import "unicode/utf8"

// StripNewlines removes \n and \r so compares ignore line breaks.
func StripNewlines(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if c == '\n' || c == '\r' {
			continue
		}
		out = append(out, c)
	}
	return out
}

// ExactlyOneMoreSpace reports whether after is before with exactly one ASCII
// space (0x20) inserted into real draft text. Both sides are newline-stripped
// first, then compared byte-by-byte.
//
// Not occupied when:
//   - before is empty
//   - the space is inserted into an existing space run (padding)
//   - the space is inserted immediately after a composer/chrome glyph (›»❯│…)
func ExactlyOneMoreSpace(before, after []byte) bool {
	b := StripNewlines(before)
	a := StripNewlines(after)
	if len(b) == 0 {
		return false
	}
	if len(a) != len(b)+1 {
		return false
	}
	// Find the single insertion index of ' '.
	i := 0
	for i < len(b) && b[i] == a[i] {
		i++
	}
	if a[i] != ' ' {
		return false
	}
	// Remainder of after (after the inserted space) must equal remainder of before.
	if string(a[i+1:]) != string(b[i:]) {
		return false
	}
	// Inserted into a space run (padding) → not a draft keystroke.
	if i > 0 && a[i-1] == ' ' {
		return false
	}
	// Inserted right after a prompt/box glyph → empty composer accepted the probe.
	if r, ok := leftRune(a, i); ok && isChromeRune(r) {
		return false
	}
	return true
}

func leftRune(a []byte, i int) (rune, bool) {
	if i <= 0 {
		return 0, false
	}
	r, size := utf8.DecodeLastRune(a[:i])
	if r == utf8.RuneError && size == 1 {
		return 0, false
	}
	return r, true
}

func isChromeRune(r rune) bool {
	switch r {
	case '›', '»', '❯', '❱', '|', '·', '╭', '╮', '╰', '╯':
		return true
	}
	// Box drawing block.
	if r >= 0x2500 && r <= 0x257F {
		return true
	}
	return false
}
