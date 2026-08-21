package agentruncli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile_EmptyPath(t *testing.T) {
	t.Parallel()
	got, err := LoadEnvFile("")
	if err != nil || got != nil {
		t.Fatalf("empty path: got %v err %v", got, err)
	}
}

func TestLoadEnvFile_ParsesSkipsComments(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "env")
	body := "# comment\n\nNO_COLOR=1\nPATH=/a:/b\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"NO_COLOR=1", "PATH=/a:/b"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}

func TestLoadEnvFile_InvalidLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.env")
	if err := os.WriteFile(path, []byte("NOEQUALS\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadEnvFile(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergeEnvEntries_CLILast(t *testing.T) {
	t.Parallel()
	got := MergeEnvEntries([]string{"A=1", "B=2"}, []string{"A=9", "C=3"})
	want := []string{"A=1", "B=2", "A=9", "C=3"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}

func TestResolveRunEnv_FileThenCLI(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "e.env")
	if err := os.WriteFile(path, []byte("A=file\nB=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveRunEnv(path, []string{"A=cli"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A=file", "B=2", "A=cli"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}
