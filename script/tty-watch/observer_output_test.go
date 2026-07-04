package main

import (
	"strings"
	"testing"
)

func TestRenderObserverFrameGrokLike(t *testing.T) {
	data := []byte("\x1b[?1049h\x1b[2J\x1b[H\x1b[?25lGrok Build \xe2\x80\xba \x1b[?25h")
	got := string(renderObserverFrame(data, 80, 24))
	if !strings.Contains(got, "Grok Build") {
		t.Fatalf("missing prompt, got %q", got)
	}
	if strings.Contains(got, "\x1b") {
		t.Fatalf("leaked escape, got %q", got)
	}
}

func TestRenderObserverFramePlainText(t *testing.T) {
	data := []byte("WATCH_MARKER\n")
	got := string(renderObserverFrame(data, 80, 24))
	if got != "WATCH_MARKER\n" {
		t.Fatalf("plain text changed, got %q", got)
	}
}