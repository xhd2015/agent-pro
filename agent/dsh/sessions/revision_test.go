package sessions

import (
	"os"
	"path/filepath"
	"testing"
)

// A revision is the one answer a content cache needs: the file it read is still
// the file this store resolves, unchanged. An append changes it; an idle session
// keeps it.
func TestRevisionChangesOnlyWhenTheArtifactChanges(t *testing.T) {
	ResetInspectCache()
	t.Cleanup(ResetInspectCache)
	root := writePlain(t, "rev", headerAt("id-rev", 3)+
		row("session/title", 0, `{"title":"one"}`))
	store := &Store{Root: root}

	first, ok := store.Revision("id-rev")
	if !ok || first == "" {
		t.Fatalf("revision = %q ok=%v", first, ok)
	}
	again, ok := store.Revision("id-rev")
	if !ok || again != first {
		t.Fatalf("an idle session changed revision: %q -> %q", first, again)
	}

	path := filepath.Join(root, "project", "rev", "session.v3.jsonl")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(row("user/message", 1, `{"content":[{"type":"text","text":"hi"}]}`)); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	moved, ok := store.Revision("id-rev")
	if !ok || moved == first {
		t.Fatalf("an append kept the revision: %q", moved)
	}
	if _, ok := store.Revision("id-missing"); ok {
		t.Fatal("a missing session reported a revision")
	}
}

// Title answers from the log's tail, and the latest title wins.
func TestTitleReadsTheLatestTitleFromTheTail(t *testing.T) {
	ResetInspectCache()
	t.Cleanup(ResetInspectCache)
	root := writePlain(t, "title", headerAt("id-title", 3)+
		row("session/title", 0, `{"title":"one"}`)+
		row("user/message", 1, `{"content":[{"type":"text","text":"hi"}]}`)+
		row("session/title", 2, `{"title":"two"}`))
	store := &Store{Root: root}

	title, ok := store.Title("id-title")
	if !ok || title != "two" {
		t.Fatalf("title = %q ok=%v, want two", title, ok)
	}
	if title, ok := store.Title("id-missing"); ok || title != "" {
		t.Fatalf("missing session title = %q ok=%v", title, ok)
	}
}
