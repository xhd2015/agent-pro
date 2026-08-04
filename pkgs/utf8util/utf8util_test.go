package utf8util

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestToValid_InvalidUTF8(t *testing.T) {
	bad := "hello" + string([]byte{0xe5, 0x92}) + "world"
	if utf8.ValidString(bad) {
		t.Fatal("fixture expected invalid UTF-8")
	}
	got := ToValid(bad)
	if !utf8.ValidString(got) {
		t.Fatalf("ToValid left invalid UTF-8: %q", got)
	}
	if !strings.Contains(got, Replacement) {
		t.Fatalf("expected %q replacement, got %q", Replacement, got)
	}
	if !strings.HasPrefix(got, "hello") || !strings.HasSuffix(got, "world") {
		t.Fatalf("surrounding text lost: %q", got)
	}
}

func TestToValid_ValidUnchanged(t *testing.T) {
	good := "在checkout场景 SPL"
	if got := ToValid(good); got != good {
		t.Fatalf("got %q want %q", got, good)
	}
}

func TestToValid_Empty(t *testing.T) {
	if got := ToValid(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
