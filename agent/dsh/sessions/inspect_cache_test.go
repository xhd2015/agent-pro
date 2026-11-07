package sessions

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func expireInspectIndex() {
	inspectMemo.mu.Lock()
	inspectMemo.index = nil
	inspectMemo.mu.Unlock()
}

func TestBuildIndexReusesIDWhenLogGrows(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-grow"
	path := put(t, root, id, "session.v3.jsonl", header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`))
	store := Store{Root: root}
	if _, err := store.lookupArtifact(id); err != nil {
		t.Fatal(err)
	}
	if got := ArtifactIDReads.Load(); got != 1 {
		t.Fatalf("first index reads=%d", got)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(row("assistant/message", 1, `{"message":{"content":[{"type":"text","text":"yo"}]}}`)); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	expireInspectIndex()
	got, err := store.lookupArtifact(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.path != path {
		t.Fatalf("path=%s", got.path)
	}
	if n := ArtifactIDReads.Load(); n != 1 {
		t.Fatalf("append must not re-decode header: reads=%d", n)
	}
}

func TestBuildIndexRereadsIDWhenLogShrinks(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-shrink"
	path := put(t, root, id, "session.v3.jsonl", header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)+row("assistant/message", 1, `{"message":{"content":[{"type":"text","text":"yo"}]}}`))
	store := Store{Root: root}
	if _, err := store.lookupArtifact(id); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)), 0o644); err != nil {
		t.Fatal(err)
	}
	expireInspectIndex()
	if _, err := store.lookupArtifact(id); err != nil {
		t.Fatal(err)
	}
	if n := ArtifactIDReads.Load(); n != 2 {
		t.Fatalf("shrink must re-decode header: reads=%d", n)
	}
}

func TestBuildIndexPicksUpNewSession(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	store := Store{Root: root}
	put(t, root, "s-old", "session.v3.jsonl", header("s-old"))
	if _, err := store.lookupArtifact("s-old"); err != nil {
		t.Fatal(err)
	}
	put(t, root, "s-new", "session.v3.jsonl", header("s-new"))
	expireInspectIndex()
	if _, err := store.lookupArtifact("s-new"); err != nil {
		t.Fatal(err)
	}
	if n := ArtifactIDReads.Load(); n != 2 {
		t.Fatalf("new session should decode once: reads=%d", n)
	}
}

func TestBuildIndexRepeatDoesNotReread(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	store := Store{Root: root}
	for i, id := range []string{"s-a", "s-b", "s-c"} {
		_ = i
		put(t, root, id, "session.v3.jsonl", header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`))
	}
	if _, err := store.lookupArtifact("s-a"); err != nil {
		t.Fatal(err)
	}
	first := ArtifactIDReads.Load()
	if first != 3 {
		t.Fatalf("first index reads=%d want 3", first)
	}
	expireInspectIndex()
	if _, err := store.lookupArtifact("s-b"); err != nil {
		t.Fatal(err)
	}
	if n := ArtifactIDReads.Load(); n != first {
		t.Fatalf("second index reads=%d want %d", n, first)
	}
}

func BenchmarkBuildIndexRepeat(b *testing.B) {
	ResetInspectCache()
	root := b.TempDir()
	store := Store{Root: root}
	for i := 0; i < 40; i++ {
		id := "s-bench-" + strconv.Itoa(i)
		dir := filepath.Join(root, "project", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatal(err)
		}
		body := header(id) + row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)
		if err := os.WriteFile(filepath.Join(dir, "session.v3.jsonl"), []byte(body), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	if _, err := store.lookupArtifact("s-bench-0"); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expireInspectIndex()
		if _, err := store.lookupArtifact("s-bench-1"); err != nil {
			b.Fatal(err)
		}
	}
}
