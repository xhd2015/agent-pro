package agentdriver

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestArgv_embeddingPrefix(t *testing.T) {
	d := Driver{Binary: "/abs/spl", Args: []string{"agent-run"}}
	got, err := d.Argv("__serve_x__", "sid-1", "/bin/grok")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/abs/spl", "agent-run", "__serve_x__", "sid-1", "/bin/grok"}
	if len(got) != len(want) {
		t.Fatalf("argv len=%d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argv[%d]=%q want %q full=%#v", i, got[i], want[i], got)
		}
	}
}

func TestArgv_emptyBinaryErrors(t *testing.T) {
	_, err := (Driver{}).Argv("run")
	if err == nil {
		t.Fatal("expected error for empty Binary without Resolve")
	}
}

func TestMustArgv_resolveSelf(t *testing.T) {
	got, err := MustArgv(Driver{}, "run", "--open")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 3 {
		t.Fatalf("got %#v", got)
	}
	if !filepath.IsAbs(got[0]) {
		t.Fatalf("binary not abs: %q", got[0])
	}
	if got[1] != "run" || got[2] != "--open" {
		t.Fatalf("remainder: %#v", got)
	}
}

func TestResolve_keepsExplicitArgsWithEmptyBinary(t *testing.T) {
	d, err := Resolve(Driver{Args: []string{"agent-run"}})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(d.Binary) {
		t.Fatalf("binary: %q", d.Binary)
	}
	if len(d.Args) != 1 || d.Args[0] != "agent-run" {
		t.Fatalf("args: %#v", d.Args)
	}
}

func TestResolve_dropsEmptyArgs(t *testing.T) {
	d, err := Resolve(Driver{Binary: "/tmp/bin", Args: []string{"", "agent-run", "  "}})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Args) != 1 || d.Args[0] != "agent-run" {
		t.Fatalf("args: %#v", d.Args)
	}
	if !strings.HasSuffix(d.Binary, "bin") && !filepath.IsAbs(d.Binary) {
		t.Fatalf("binary: %q", d.Binary)
	}
}

func TestMustArgv_followUpShape(t *testing.T) {
	got, err := MustArgv(
		Driver{Binary: "/abs/spl", Args: []string{"agent-run"}},
		"run", "--session-id=s1", "--open", "--", "hi",
	)
	if err != nil {
		t.Fatal(err)
	}
	want0 := []string{"/abs/spl", "agent-run", "run", "--session-id=s1", "--open", "--", "hi"}
	if len(got) != len(want0) {
		t.Fatalf("%#v", got)
	}
	for i := range want0 {
		if got[i] != want0[i] {
			t.Fatalf("got %#v", got)
		}
	}
}
