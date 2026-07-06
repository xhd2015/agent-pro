package ttywatch

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

func TestContainsKittyCtrlC(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "etx ctrl", in: "\x1b[3;5u", want: true},
		{name: "letter c ctrl", in: "\x1b[99;5u", want: true},
		{name: "letter c ctrl release", in: "\x1b[99;5:3u", want: true},
		{name: "letter c no ctrl", in: "\x1b[99;1u", want: false},
		{name: "arrow ctrl", in: "\x1b[1;5C", want: false},
		{name: "partial", in: "\x1b[99;5", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := containsKittyCtrlC([]byte(tc.in)); got != tc.want {
				t.Fatalf("containsKittyCtrlC(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}