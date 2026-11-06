package sessions

import (
	"os"
	"sync"
	"testing"
)

func TestInspectCachesUnchangedArtifact(t *testing.T) {
	resetInspectBodies()
	root := t.TempDir()
	id := "s-cache"
	path := put(t, root, id, "session.v3.jsonl", header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`))
	store := Store{Root: root}
	first, err := store.Inspect(id)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Inspect(id)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("unchanged Inspect should return the cached pointer")
	}
	if got := InspectReads.Load(); got != 1 {
		t.Fatalf("reads=%d want 1", got)
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
	third, err := store.Inspect(id)
	if err != nil {
		t.Fatal(err)
	}
	if third == first {
		t.Fatal("grown artifact should decode a new Inspection")
	}
	if len(third.Events) != 2 {
		t.Fatalf("events=%d", len(third.Events))
	}
	if got := InspectReads.Load(); got != 2 {
		t.Fatalf("reads after grow=%d want 2", got)
	}
}

func TestInspectSingleflight(t *testing.T) {
	resetInspectBodies()
	root := t.TempDir()
	id := "s-flight"
	put(t, root, id, "session.v3.jsonl", header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`))
	store := Store{Root: root}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.Inspect(id); err != nil {
				t.Errorf("inspect: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := InspectReads.Load(); got != 1 {
		t.Fatalf("parallel Inspect reads=%d want 1", got)
	}
}
