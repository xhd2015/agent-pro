package agentruncli

import (
	"strings"
	"testing"
)

func TestAssetsHelp_noPanic(t *testing.T) {
	// Help paths must not panic and must return nil.
	for _, args := range [][]string{
		nil,
		{},
		{"--help"},
		{"-h"},
		{"status", "--help"},
		{"ensure", "--help"},
	} {
		if err := runAssets(args); err != nil {
			t.Fatalf("runAssets(%v): %v", args, err)
		}
	}
}

func TestAssetsUnknownCommand(t *testing.T) {
	err := runAssets([]string{"nope"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown assets command") {
		t.Fatalf("err = %v", err)
	}
}

func TestHandleAssetsHelp(t *testing.T) {
	if err := Handle([]string{"assets", "--help"}); err != nil {
		t.Fatalf("Handle assets --help: %v", err)
	}
	if err := Handle([]string{"assets"}); err != nil {
		t.Fatalf("Handle assets: %v", err)
	}
}

func TestAssetsStatus_noNetwork(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("AGENT_PRO_ASSET_BASE_URL", "")
	if err := runAssetsStatus(nil); err != nil {
		t.Fatal(err)
	}
}
