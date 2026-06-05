package faketools

import (
	"math/rand"
	"testing"
)

func TestExecToolCall_ReturnsCommandAndOutput(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	r := ExecToolCall(rng, "echo hello")
	if r.Command != "echo hello" {
		t.Fatalf("Command = %q, want %q", r.Command, "echo hello")
	}
}

func TestExecToolCall_Deterministic(t *testing.T) {
	r1 := rand.New(rand.NewSource(99))
	r2 := rand.New(rand.NewSource(99))
	a := ExecToolCall(r1, "test")
	b := ExecToolCall(r2, "test")
	if a.Output != b.Output || a.Command != b.Command {
		t.Fatalf("deterministic: %+v vs %+v", a, b)
	}
}

func TestExecToolCall_NoChanges(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	r := ExecToolCall(rng, "test")
	if len(r.Changes) != 0 {
		t.Fatalf("expected no changes for tool call, got %d", len(r.Changes))
	}
}

func TestExecFileRead_ProducesCatCommand(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	r := ExecFileRead(rng, "/tmp/log.txt")
	if r.Command != "cat /tmp/log.txt" {
		t.Fatalf("Command = %q, want %q", r.Command, "cat /tmp/log.txt")
	}
}

func TestExecFileRead_NonEmptyOutput(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		rng := rand.New(rand.NewSource(seed))
		r := ExecFileRead(rng, "/tmp/log.txt")
		if r.Output == "" {
			t.Fatalf("seed %d: file read returned empty output", seed)
		}
	}
}

func TestExecFileRead_NoChanges(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	r := ExecFileRead(rng, "/tmp/log.txt")
	if len(r.Changes) != 0 {
		t.Fatalf("expected no changes for file read, got %d", len(r.Changes))
	}
}

func TestExecFileWrite_HasChanges(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	r := ExecFileWrite(rng, "/tmp/out.txt")
	if len(r.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(r.Changes))
	}
	if r.Changes[0].Path != "/tmp/out.txt" {
		t.Fatalf("Path = %q, want /tmp/out.txt", r.Changes[0].Path)
	}
	kind := r.Changes[0].Kind
	if kind != "add" && kind != "modify" {
		t.Fatalf("Kind = %q, want add or modify", kind)
	}
}

func TestExecFileWrite_NoCommand(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	r := ExecFileWrite(rng, "/tmp/out.txt")
	if r.Command != "" {
		t.Fatalf("expected no command for file write, got %q", r.Command)
	}
}

func TestExecSearch_ProducesGrepCommand(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	r := ExecSearch(rng, "TODO")
	if r.Command != "grep -rn TODO ." {
		t.Fatalf("Command = %q, want %q", r.Command, "grep -rn TODO .")
	}
}

func TestExecSearch_NoChanges(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	r := ExecSearch(rng, "TODO")
	if len(r.Changes) != 0 {
		t.Fatalf("expected no changes for search, got %d", len(r.Changes))
	}
}

func TestExecSearch_MayBeEmpty(t *testing.T) {
	emptyFound := false
	for seed := int64(0); seed < 100; seed++ {
		rng := rand.New(rand.NewSource(seed))
		r := ExecSearch(rng, "NOMATCH")
		if r.Output == "" {
			emptyFound = true
			break
		}
	}
	if !emptyFound {
		t.Fatal("expected at least one empty search result")
	}
}

func TestExecRandom_Deterministic(t *testing.T) {
	r1 := rand.New(rand.NewSource(77))
	r2 := rand.New(rand.NewSource(77))
	a := ExecRandom(r1)
	b := ExecRandom(r2)
	if a.Command != b.Command || a.Output != b.Output {
		t.Fatalf("deterministic: %+v vs %+v", a, b)
	}
}
