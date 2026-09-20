package main

import (
	"strings"
	"testing"
)

func TestDSHRegistered(t *testing.T) {
	if !strings.Contains(help, "dsh               inspect stored DeepSeek Harness sessions") {
		t.Fatal("root help lacks dsh")
	}
	if err := handle([]string{"dsh", "session", "send"}); err == nil || !strings.Contains(err.Error(), "unknown dsh session command: send") {
		t.Fatalf("unexpected dispatch: %v", err)
	}
}
