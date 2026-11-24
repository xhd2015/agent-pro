package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestReadLogBudgetKeepsEveryEvent reads a plain JSONL log through a window that
// holds far more rows than the batch budget, and requires every seq to come back.
//
// The cursor used to advance past the whole window while the batch carried only
// the first `maxEvents` rows, so the rest were never returned to anyone: a reader
// resuming from Next skipped them, saw a seq gap, and started the log over. On a
// large enough log that repeats forever and the fold never finishes.
func TestReadLogBudgetKeepsEveryEvent(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, "sessions")
	const total = 500
	var body string
	body += header("session-budget")
	for seq := 0; seq < total; seq++ {
		body += bulkRow(seq, fmt.Sprintf("row %d", seq))
	}
	id := "session-budget"
	dir := filepath.Join(root, "project", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "session.v3.jsonl")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(Options{Home: home})
	if err != nil {
		t.Fatal(err)
	}

	// A window far larger than the batch budget, which is the shape that lost rows.
	var got []int64
	from := int64(0)
	for round := 0; round < 100; round++ {
		batch, err := store.ReadLog(id, ReadLogOptions{
			From:      from,
			Window:    1 << 20,
			MaxWindow: 32 << 20,
			MaxEvents: 16,
		})
		if err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		for _, event := range batch.Events {
			got = append(got, event.Seq)
		}
		if batch.Done {
			break
		}
		if batch.Next == from {
			t.Fatalf("round %d made no progress", round)
		}
		from = batch.Next
	}
	if len(got) != total {
		t.Fatalf("read %d events, want %d", len(got), total)
	}
	for i, seq := range got {
		if seq != int64(i) {
			t.Fatalf("event %d has seq %d, want %d (a row was skipped)", i, seq, i)
		}
	}
}
