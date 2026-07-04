package main

import (
	"strings"
	"testing"
)

func TestSanitizeStripsCSI(t *testing.T) {
	in := "\x1b[31mRED\x1b[0m"
	got := SanitizeForPrint(in)
	if got != "RED" {
		t.Fatalf("CSI strip: got %q want RED", got)
	}
}

func TestSanitizeStripsOSC(t *testing.T) {
	in := "\x1b]0;title\x07text"
	got := SanitizeForPrint(in)
	if got != "text" {
		t.Fatalf("OSC strip: got %q want text", got)
	}
}

func TestSanitizeStripsC0KeepsNewlineTab(t *testing.T) {
	in := "line1\nline2\ttab\x07bell"
	got := SanitizeForPrint(in)
	want := "line1\nline2\ttabbell"
	if got != want {
		t.Fatalf("C0 strip: got %q want %q", got, want)
	}
}

func TestSanitizeSnapshotFixture(t *testing.T) {
	in := "\x1b[31mRED\x1b[0m\nPLAIN_LINE\n"
	got := SanitizeForPrint(in)
	if !strings.Contains(got, "PLAIN_LINE") || !strings.Contains(got, "RED") {
		t.Fatalf("fixture: got %q", got)
	}
	if strings.Contains(got, "\x1b") {
		t.Fatalf("fixture still has escapes: %q", got)
	}
}