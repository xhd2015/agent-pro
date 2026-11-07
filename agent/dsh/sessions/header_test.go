package sessions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHeaderReadsOnlyTheFirstRow proves the cheap accessor: the events after
// the header are never decoded, so a large log costs one frame.
func TestHeaderReadsOnlyTheFirstRow(t *testing.T) {
	ResetInspectCache()
	t.Cleanup(ResetInspectCache)
	var body strings.Builder
	body.WriteString(headerAt("id-a", 3))
	for seq := 0; seq < 50; seq++ {
		body.WriteString(row("session/title", seq, `{"title":"ignored"}`))
	}
	root := writePlain(t, "a", body.String())
	store := &Store{Root: root}

	before := InspectReads.Load()
	got, err := store.Header("id-a")
	if err != nil {
		t.Fatal(err)
	}
	if InspectReads.Load() != before {
		t.Fatalf("Header decoded the event log: reads %d -> %d", before, InspectReads.Load())
	}
	if got.ID != "id-a" || got.Cwd != "/workspace" || got.FormatVersion != 3 {
		t.Fatalf("summary = %+v", got)
	}
	if got.SizeBytes != int64(statSize(t, root, "a")) {
		t.Fatalf("size = %d", got.SizeBytes)
	}
}

// TestHeaderSurvivesAppendAndTracksReplacement covers the two invalidation
// edges: appending must keep the cached header, replacing the file must not.
func TestHeaderSurvivesAppendAndTracksReplacement(t *testing.T) {
	ResetInspectCache()
	t.Cleanup(ResetInspectCache)
	root := writePlain(t, "a", headerAt("id-a", 3)+row("session/title", 0, `{"title":"one"}`))
	store := &Store{Root: root}
	if _, err := store.Header("id-a"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "project", "a", "session.v3.jsonl")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(row("session/title", 1, `{"title":"two"}`)); err != nil {
		t.Fatal(err)
	}
	file.Close()

	before := InspectReads.Load()
	appended, err := store.Header("id-a")
	if err != nil {
		t.Fatal(err)
	}
	if InspectReads.Load() != before {
		t.Fatalf("append re-read the header")
	}
	if want := int64(statSize(t, root, "a")); appended.SizeBytes != want {
		t.Fatalf("size after append = %d want %d", appended.SizeBytes, want)
	}

	// A rewrite with a different cwd keeps the identity but replaces the header.
	rewritten := strings.Replace(headerAt("id-a", 3), "/workspace", "/elsewhere", 1)
	if err := os.WriteFile(path, []byte(rewritten), 0644); err != nil {
		t.Fatal(err)
	}
	replaced, err := store.Header("id-a")
	if err != nil {
		t.Fatal(err)
	}
	if replaced.Cwd != "/elsewhere" {
		t.Fatalf("cwd after rewrite = %q", replaced.Cwd)
	}
}

// writePlain writes a v3 JSONL artifact under its session directory.
func writePlain(t *testing.T, dir, body string) string {
	t.Helper()
	root := t.TempDir()
	put(t, root, dir, "session.v3.jsonl", body)
	return root
}

func statSize(t *testing.T, root, dir string) int64 {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, "project", dir, "session.v3.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}
