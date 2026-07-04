package main

import (
	"strings"
	"testing"
)

func TestObserverTTYDetachCleanupSequence(t *testing.T) {
	for _, seq := range []string{
		"\x1b[?1049l",
		"\x1b[<u",
		"\x1b[?1000l",
		"\x1b[?1002l",
		"\x1b[?1003l",
		"\x1b[?1006l",
	} {
		if !strings.Contains(observerTTYDetachCleanup, seq) {
			t.Fatalf("observerTTYDetachCleanup missing %q", seq)
		}
	}
}