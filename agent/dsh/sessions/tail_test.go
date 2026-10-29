package sessions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func encodeZstdFrame(t *testing.T, plain string) []byte {
	t.Helper()
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	return enc.EncodeAll([]byte(plain), nil)
}

func TestSyncTailReadsNewZstdFramesOnly(t *testing.T) {
	root := t.TempDir()
	id := "s-tail-zstd"
	first := header(id) + row("session/title", 0, `{"title":"one"}`) + row("turn/start", 1, `{"turn":1}`)
	path := put(t, root, id, "session.v3.jsonl.zstd", string(encodeZstdFrame(t, first)))
	store := Store{Root: root}
	a, err := store.SyncTail(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.Summary.Title != "one" || a.fold.kind != ActivityThinking || len(a.Rows) != 0 {
		t.Fatalf("first %+v kind=%s rows=%d", a.Summary, a.fold.kind, len(a.Rows))
	}
	offset := a.Offset
	second := row("tool/call", 2, `{"name":"read","arguments":{"path":"f"}}`)
	appended := append(mustRead(t, path), encodeZstdFrame(t, second)...)
	if err := os.WriteFile(path, appended, 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := store.SyncTail(id, a)
	if err != nil {
		t.Fatal(err)
	}
	if b.Offset <= offset {
		t.Fatalf("offset did not advance %d -> %d", offset, b.Offset)
	}
	if b.fold.kind != ActivityToolCalling || len(b.Rows) != 1 || b.Rows[0].Tool != "read" {
		t.Fatalf("second kind=%s rows=%+v", b.fold.kind, b.Rows)
	}
	c, err := store.SyncTail(id, b)
	if err != nil || c != b {
		t.Fatalf("unchanged should reuse cursor %v %v", c, err)
	}
}

func TestSyncTailCapsTranscriptRows(t *testing.T) {
	root := t.TempDir()
	id := "s-tail-cap"
	body := header(id)
	for i := 0; i < 250; i++ {
		body += row("tool/call", i, `{"name":"read","arguments":{"path":"f"}}`)
	}
	put(t, root, id, "session.v3.jsonl", body)
	store := Store{Root: root}
	got, err := store.SyncTail(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != tailKeep {
		t.Fatalf("rows=%d want %d", len(got.Rows), tailKeep)
	}
}

func TestScanZstdFramesTorn(t *testing.T) {
	full := encodeZstdFrame(t, "hello\n")
	frames, tornStart, torn, err := scanZstdFrames(full[:len(full)/2])
	if err != nil || !torn || tornStart != 0 || len(frames) != 0 {
		t.Fatalf("frames=%v torn=%v start=%d err=%v", frames, torn, tornStart, err)
	}
}

func BenchmarkSyncTailAppend(b *testing.B) {
	root := b.TempDir()
	id := "s-tail-bench"
	body := header(id)
	for i := 0; i < 2000; i++ {
		body += row("tool/call", i, `{"name":"read","arguments":{"path":"f"}}`)
	}
	path := putBench(b, root, id, "session.v3.jsonl", body)
	store := Store{Root: root}
	cur, err := store.SyncTail(id, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seq := 2000 + i
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := f.WriteString(row("tool/call", seq, `{"name":"read","arguments":{"path":"f"}}`)); err != nil {
			b.Fatal(err)
		}
		f.Close()
		cur, err = store.SyncTail(id, cur)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func putBench(b *testing.B, root, dir, name, body string) string {
	b.Helper()
	path := filepath.Join(root, "project", dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		b.Fatal(err)
	}
	return path
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
