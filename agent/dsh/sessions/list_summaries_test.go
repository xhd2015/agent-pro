package sessions

import (
	"os"
	"reflect"
	"testing"
)

func TestListSummariesMatchesList(t *testing.T) {
	root := t.TempDir()
	body := header("s-sum") +
		row("session/title", 0, `{"title":"Audit the server"}`) +
		row("request/header", 1, `{"header":{"config":{"provider":"deepseek","model":"v4"}}}`) +
		row("turn/start", 2, `{"turn":1}`) +
		row("turn/end", 3, `{"turn":1,"reason":{"kind":"completed"}}`)
	put(t, root, "s-sum", "session.v3.jsonl", body)
	store := Store{Root: root}
	want, _, err := store.List()
	if err != nil || len(want) != 1 {
		t.Fatalf("List %+v %v", want, err)
	}
	got, warnings, err := store.ListSummaries()
	if err != nil || len(warnings) != 0 || len(got) != 1 {
		t.Fatalf("ListSummaries %+v %v %v", got, warnings, err)
	}
	if !reflect.DeepEqual(got[0], want[0]) {
		t.Fatalf("ListSummaries row\n got %+v\nwant %+v", got[0], want[0])
	}
}

func TestListSummariesFoldsAppendedFrames(t *testing.T) {
	root := t.TempDir()
	id := "s-sum-grow"
	first := header(id) + row("session/title", 0, `{"title":"Faster agents view"}`) + row("turn/start", 1, `{"turn":1}`)
	path := put(t, root, id, "session.v3.jsonl.zstd", string(encodeZstdFrame(t, first)))
	store := Store{Root: root}
	started, warnings, err := store.ListSummaries()
	if err != nil || len(warnings) != 0 || len(started) != 1 {
		t.Fatalf("%+v %v %v", started, warnings, err)
	}
	if started[0].Outcome != "incomplete" || started[0].Title != "Faster agents view" {
		t.Fatalf("first %+v", started[0])
	}
	second := row("turn/end", 2, `{"turn":1,"reason":{"kind":"completed"}}`)
	appended := append(mustRead(t, path), encodeZstdFrame(t, second)...)
	if err := os.WriteFile(path, appended, 0o644); err != nil {
		t.Fatal(err)
	}
	ended, _, err := store.ListSummaries()
	if err != nil || len(ended) != 1 {
		t.Fatalf("%+v %v", ended, err)
	}
	if ended[0].ID != id || ended[0].Outcome != "completed" || ended[0].Title != "Faster agents view" {
		t.Fatalf("second %+v", ended[0])
	}
	if ended[0].UpdatedAt <= started[0].UpdatedAt || ended[0].SizeBytes <= started[0].SizeBytes {
		t.Fatalf("appended frame not folded: %+v -> %+v", started[0], ended[0])
	}
}

func TestListSummariesReadsLogAgainWhenItShrinks(t *testing.T) {
	root := t.TempDir()
	path := put(t, root, "s-sum-rewrite", "session.v3.jsonl", header("s-sum-rewrite")+row("turn/start", 0, `{"turn":1}`)+row("turn/end", 1, `{"turn":1,"reason":{"kind":"aborted"}}`))
	store := Store{Root: root}
	if rows, _, err := store.ListSummaries(); err != nil || len(rows) != 1 || rows[0].Outcome != "aborted" {
		t.Fatalf("%+v %v", rows, err)
	}
	// A shorter log is a different session, not an append.
	if err := os.WriteFile(path, []byte(header("s-sum-other")), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, _, err := store.ListSummaries()
	if err != nil || len(rows) != 1 || rows[0].ID != "s-sum-other" || rows[0].Outcome != "empty" {
		t.Fatalf("%+v %v", rows, err)
	}
}

func TestListSummariesIsolatesUnreadableLog(t *testing.T) {
	root := t.TempDir()
	put(t, root, "s-good", "session.v3.jsonl", header("s-good"))
	put(t, root, "s-bad", "session.v3.jsonl", "not-json\n")
	store := Store{Root: root}
	rows, warnings, err := store.ListSummaries()
	if err != nil || len(rows) != 1 || rows[0].ID != "s-good" || len(warnings) != 1 {
		t.Fatalf("%+v %v %v", rows, warnings, err)
	}
}
