// Package utf8util holds small UTF-8 helpers shared across agent-run packages.
package utf8util

import "strings"

// Replacement is the Unicode replacement character (U+FFFD).
const Replacement = "\uFFFD"

// ToValid returns s with each invalid UTF-8 byte sequence replaced by
// Replacement. Safe for runner argv / inject (real grok panics in
// std::env::args on invalid UTF-8).
func ToValid(s string) string {
	return strings.ToValidUTF8(s, Replacement)
}
