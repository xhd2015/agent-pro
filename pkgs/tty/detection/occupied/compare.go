// Package occupied classifies live composer occupancy by injecting a space
// and comparing before/after snapshots.
package occupied

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
// space (0x20) inserted. Both sides are newline-stripped first, then compared
// byte-by-byte.
//
// Occupied when: len(after)==len(before)+1, the extra byte is ' ', and removing
// that byte yields before. Empty before is never occupied.
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
	return true
}
