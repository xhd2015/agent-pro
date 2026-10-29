package sessions

import (
	"os"
	"testing"
	"time"
)

func TestListRecentSkipsOlderLogs(t *testing.T) {
	root := t.TempDir()
	oldBody := header("s-old") + row("tool/call", 0, `{"name":"read"}`) + "not-json\n"
	put(t, root, "s-old", "session.v3.jsonl", oldBody)
	put(t, root, "s-mid", "session.v3.jsonl", header("s-mid"))
	put(t, root, "s-new", "session.v3.jsonl", header("s-new"))
	old := root + "/project/s-old/session.v3.jsonl"
	mid := root + "/project/s-mid/session.v3.jsonl"
	newp := root + "/project/s-new/session.v3.jsonl"
	base := time.Now().Add(-3 * time.Hour)
	if err := os.Chtimes(old, base, base); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(mid, base.Add(time.Hour), base.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newp, base.Add(2*time.Hour), base.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: root}
	got, warnings, err := store.ListRecent(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings=%v", warnings)
	}
	if len(got) != 2 || got[0].ID != "s-new" || got[1].ID != "s-mid" {
		t.Fatalf("%+v", got)
	}
	all, allWarn, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(allWarn) == 0 {
		t.Fatalf("List should warn on the corrupt older log, got %d rows %v", len(all), allWarn)
	}
}

func TestListRecentReadsTitleFromPrefix(t *testing.T) {
	root := t.TempDir()
	body := header("s-named") + row("session/title", 0, `{"title":"Floorplan G"}`) + row("turn/start", 1, `{"turn":1}`)
	put(t, root, "s-named", "session.v3.jsonl", body)
	store := Store{Root: root}
	got, _, err := store.ListRecent(1)
	if err != nil || len(got) != 1 || got[0].Title != "Floorplan G" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestListRecentReadsTitleFromLaterZstdFrame(t *testing.T) {
	root := t.TempDir()
	id := "s-frame-title"
	first := encodeZstdFrame(t, header(id)+row("permission/preset", 0, `{"preset":"std"}`))
	second := encodeZstdFrame(t, row("session/title", 1, `{"title":"/brainstorm there is a regression"}`))
	put(t, root, id, "session.v3.jsonl.zstd", string(append(first, second...)))
	store := Store{Root: root}
	got, _, err := store.ListRecent(1)
	if err != nil || len(got) != 1 || got[0].Title != "/brainstorm there is a regression" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestListRecentKeepsLastSessionTitle(t *testing.T) {
	root := t.TempDir()
	body := header("s-last") +
		row("session/title", 0, `{"title":"/brainstorm the server is running","source":{"kind":"fallback"}}`) +
		row("session/title", 1, `{"title":"Orchestrator agents and todo tracker","source":{"kind":"provider"}}`)
	put(t, root, "s-last", "session.v3.jsonl", body)
	store := Store{Root: root}
	got, _, err := store.ListRecent(1)
	if err != nil || len(got) != 1 || got[0].Title != "Orchestrator agents and todo tracker" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestExistsLooksUpSessionID(t *testing.T) {
	root := t.TempDir()
	put(t, root, "s-here", "session.v3.jsonl", header("s-here"))
	store := Store{Root: root}
	ok, err := store.Exists("s-here")
	if err != nil || !ok {
		t.Fatalf("exists s-here: %v %v", ok, err)
	}
	ok, err = store.Exists("s-missing")
	if err != nil || ok {
		t.Fatalf("exists s-missing: %v %v", ok, err)
	}
}
