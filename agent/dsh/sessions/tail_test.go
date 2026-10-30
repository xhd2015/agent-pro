package sessions

import (
	"crypto/rand"
	"encoding/hex"
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

func TestFindFrameStartSkipsPartial(t *testing.T) {
	frame := encodeZstdFrame(t, "hello\n")
	buf := append([]byte{0, 1, 2, 3}, frame...)
	if got := findFrameStart(buf); got != 4 {
		t.Fatalf("partial start=%d", got)
	}
	if got := findFrameStart(frame); got != 0 {
		t.Fatalf("aligned start=%d", got)
	}
}

// A large concatenated log must classify and show its tail without parsing
// the middle. The seq jump would fail a full read; the suffix resyncs.
func TestSyncTailLargeZstdSkipsMiddle(t *testing.T) {
	root := t.TempDir()
	id := "s-tail-window"
	head := encodeZstdFrame(t, header(id)+row("session/title", 0, `{"title":"Interview design"}`)+row("turn/start", 1, `{"turn":1}`))
	var body []byte
	seq := 10000
	for len(body) <= coldTailWindow+4096 {
		raw := make([]byte, 32*1024)
		if _, err := rand.Read(raw); err != nil {
			t.Fatal(err)
		}
		payload := `{"name":"pad","arguments":{"blob":"` + hex.EncodeToString(raw) + `"}}`
		body = append(body, encodeZstdFrame(t, row("tool/call", seq, payload))...)
		seq++
	}
	const smallTools = tailKeep
	firstSmall := seq
	for i := 0; i < smallTools; i++ {
		body = append(body, encodeZstdFrame(t, row("tool/call", seq, `{"name":"read","arguments":{"path":"f"}}`))...)
		seq++
	}
	lastTool := seq - 1
	body = append(body, encodeZstdFrame(t, row("turn/end", seq, `{"turn":1,"reason":{"kind":"completed"}}`))...)
	turnSeq := seq
	seq++
	path := put(t, root, id, "session.v3.jsonl.zstd", string(append(head, body...)))
	store := Store{Root: root}
	got, err := store.SyncTail(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary.Title != "Interview design" || got.Summary.Cwd != "/workspace" || got.Summary.Outcome != "completed" {
		t.Fatalf("summary %+v", got.Summary)
	}
	if got.fold.kind != ActivityIdle || len(got.Rows) != tailKeep || got.Rows[len(got.Rows)-1].Seq != int64(lastTool) {
		t.Fatalf("kind=%s rows=%d last=%+v want seq %d", got.fold.kind, len(got.Rows), got.Rows[len(got.Rows)-1], lastTool)
	}
	if got.Offset != info.Size() || got.Offset <= int64(len(head)) || got.nextSeq != int64(turnSeq+1) {
		t.Fatalf("offset %d size %d head %d next %d", got.Offset, info.Size(), len(head), got.nextSeq)
	}
	if got.Rows[0].Seq < int64(firstSmall) && got.Rows[0].Seq < 10000 {
		t.Fatalf("row seq restarted at %d", got.Rows[0].Seq)
	}
	extra := encodeZstdFrame(t, row("tool/call", seq, `{"name":"edit","arguments":{"path":"f"}}`))
	if err := os.WriteFile(path, append(mustRead(t, path), extra...), 0o644); err != nil {
		t.Fatal(err)
	}
	before := got.Offset
	next, err := store.SyncTail(id, got)
	if err != nil {
		t.Fatal(err)
	}
	if next.Offset-before != int64(len(extra)) || next.fold.kind != ActivityToolCalling || next.Rows[len(next.Rows)-1].Tool != "edit" {
		t.Fatalf("delta %d frame %d kind %s row %+v", next.Offset-before, len(extra), next.fold.kind, next.Rows[len(next.Rows)-1])
	}
	same, err := store.SyncTail(id, next)
	if err != nil || same != next {
		t.Fatalf("unchanged cursor %v %v", same, err)
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
