package sessions

import (
	"os"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func zstdFrameTB(t *testing.T, plain string) []byte {
	t.Helper()
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	return enc.EncodeAll([]byte(plain), nil)
}

func TestViewTailReusesUnchangedFile(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-tail-hit"
	put(t, root, id, "session.v3.jsonl", header(id)+
		row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)+
		row("assistant/message", 1, `{"message":{"content":[{"type":"text","text":"yo"}]}}`))
	store := Store{Root: root}
	first, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Rows) < 2 || first.Truncated {
		t.Fatalf("view=%+v", first)
	}
	if got := ViewTailReads.Load(); got != 1 {
		t.Fatalf("reads=%d", got)
	}
	second, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("unchanged ViewTail should return the cached pointer")
	}
	if got := ViewTailReads.Load(); got != 1 {
		t.Fatalf("second reads=%d", got)
	}
	if got := InspectReads.Load(); got != 0 {
		t.Fatalf("ViewTail must not Inspect: %d", got)
	}
}

func TestViewTailAppendDoesNotRereadPrefix(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-tail-append"
	path := put(t, root, id, "session.v3.jsonl", header(id)+row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`))
	store := Store{Root: root}
	if _, err := store.ViewTail(id, 50); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(row("assistant/message", 1, `{"message":{"content":[{"type":"text","text":"later"}]}}`)); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	view, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Rows) < 2 || view.Rows[len(view.Rows)-1].Text != "later" {
		t.Fatalf("rows=%+v", view.Rows)
	}
	if got := ViewTailReads.Load(); got != 2 {
		t.Fatalf("append reads=%d want 2", got)
	}
}

func TestViewTailZstdSuffixSkipsPrefix(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-tail-zstd"
	var blob []byte
	blob = append(blob, zstdFrameTB(t, header(id))...)
	for i := 0; i < 3000; i++ {
		blob = append(blob, zstdFrameTB(t, row("tool/call", i, `{"name":"read","arguments":{"path":"f"}}`))...)
	}
	last := 3000
	blob = append(blob, zstdFrameTB(t, row("user/message", last, `{"source":{"kind":"user"},"content":[{"type":"text","text":"tail-user"}]}`)+
		row("assistant/message", last+1, `{"message":{"content":[{"type":"text","text":"tail-asst"}]}}`))...)
	if int64(len(blob)) <= coldTailMin {
		t.Fatalf("fixture too small to take suffix path: %d", len(blob))
	}
	put(t, root, id, "session.v3.jsonl.zstd", string(blob))
	store := Store{Root: root}
	view, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if !view.Truncated {
		t.Fatal("large zstd should be a suffix")
	}
	if len(view.Rows) == 0 || view.Rows[len(view.Rows)-1].Text != "tail-asst" {
		t.Fatalf("missing tail rows: %+v", view.Rows)
	}
	if got := InspectReads.Load(); got != 0 {
		t.Fatalf("suffix path InspectReads=%d", got)
	}
	foundUser := false
	for _, message := range view.Rows {
		if message.Text == "tail-user" {
			foundUser = true
		}
	}
	if !foundUser {
		t.Fatalf("expected tail-user in %+v", kindsOf(view.Rows))
	}
}
