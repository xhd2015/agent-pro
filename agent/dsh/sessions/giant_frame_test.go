package sessions

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// randomText returns n bytes of base64 encoded pseudo-random data. Zstandard
// cannot shrink it, so a fixture built from it has a predictable compressed
// size, which is what the suffix search budgets are about.
func randomText(rnd *rand.Rand, n int) string {
	raw := make([]byte, n)
	for i := range raw {
		raw[i] = byte(rnd.Intn(256))
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func encodeFrame(tb testing.TB, plain string) []byte {
	tb.Helper()
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		tb.Fatal(err)
	}
	defer enc.Close()
	return enc.EncodeAll([]byte(plain), nil)
}

func writeSessionLog(tb testing.TB, root, id string, blob []byte) string {
	tb.Helper()
	path := filepath.Join(root, "project", id, "session.v3.jsonl.zstd")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, blob, 0o644); err != nil {
		tb.Fatal(err)
	}
	return path
}

func bulkRow(seq int, text string) string {
	return row("user/message", seq, fmt.Sprintf(
		`{"source":{"kind":"user"},"content":[{"type":"text","text":"%s"}]}`, text))
}

// giantFrameLog writes a log shaped like a real "one bulk frame" transcript: a
// header frame, a single huge frame holding the early history, and a short run
// of small frames at the end. The frame before the suffix is larger than
// coldTailMaxWindow, so the doubling search can never cross it and every probe
// decodes the same short suffix.
// anchorRow is an assistant reply: it yields a transcript row and marks the
// tail as anchored, which is what the cold suffix acceptance requires.
func anchorRow(seq int, text string) string {
	return row("assistant/message", seq, fmt.Sprintf(
		`{"message":{"content":[{"type":"text","text":"%s"}]}}`, text))
}

func giantFrameLog(tb testing.TB, root, id string, bulkBytes, tailRows int) {
	tb.Helper()
	rnd := rand.New(rand.NewSource(7))
	blob := encodeFrame(tb, header(id))
	var bulk strings.Builder
	seq := 0
	for bulk.Len() < bulkBytes {
		bulk.WriteString(bulkRow(seq, randomText(rnd, 6<<10)))
		seq++
	}
	blob = append(blob, encodeFrame(tb, bulk.String())...)
	for i := 0; i < tailRows; i++ {
		// One frame per row keeps the trailing boundaries dense, like the
		// per-turn frames a real session appends.
		blob = append(blob, encodeFrame(tb, anchorRow(seq, randomText(rnd, 6<<10)))...)
		seq++
	}
	writeSessionLog(tb, root, id, blob)
}

// shrinkColdTailMaxWindow keeps the doubling search cheap in tests instead of
// building a fixture larger than the 8 MiB production window.
func shrinkColdTailMaxWindow(tb testing.TB, window int64) {
	tb.Helper()
	prev := coldTailMaxWindow
	coldTailMaxWindow = window
	tb.Cleanup(func() { coldTailMaxWindow = prev })
}

func TestViewTailAcceptsGiantFramePartialSuffix(t *testing.T) {
	ResetInspectCache()
	resetViewTails()
	shrinkColdTailMaxWindow(t, 1<<20)
	root := t.TempDir()
	id := "s-giant-partial"
	giantFrameLog(t, root, id, 4<<20, 15)
	store := Store{Root: root}

	view, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got := ViewTailFullFallbacks.Load(); got != 0 {
		t.Fatalf("cold ViewTail decoded the whole log: fallbacks=%d", got)
	}
	if !view.Truncated {
		t.Fatal("a suffix that holds only part of the log must stay truncated")
	}
	if len(view.Rows) != 15 {
		t.Fatalf("rows=%d want the 15 rows of the newest frames", len(view.Rows))
	}
	if got := ViewTailReads.Load(); got != 1 {
		t.Fatalf("reads=%d want one suffix decode", got)
	}
	if got := InspectReads.Load(); got != 0 {
		t.Fatalf("ViewTail must not Inspect: %d", got)
	}
}

func TestViewTailBeforeCrossesGiantFrame(t *testing.T) {
	ResetInspectCache()
	resetViewTails()
	shrinkColdTailMaxWindow(t, 1<<20)
	root := t.TempDir()
	id := "s-giant-before"
	giantFrameLog(t, root, id, 4<<20, 15)
	store := Store{Root: root}

	head, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if !head.Truncated || len(head.Rows) == 0 {
		t.Fatalf("expected a truncated suffix, got truncated=%v rows=%d", head.Truncated, len(head.Rows))
	}
	reads := ViewTailReads.Load()
	wide, err := store.ViewTailBefore(id, head.Rows[0].Seq, 50, nil)
	if err != nil {
		t.Fatal(err)
	}
	if wide.Truncated {
		t.Fatal("paging past a giant frame must reach the log head")
	}
	older := 0
	for _, message := range wide.Rows {
		if message.Seq < head.Rows[0].Seq {
			older++
		}
	}
	if older < 50 {
		t.Fatalf("older rows=%d want >=50", older)
	}
	if grew := ViewTailReads.Load() - reads; grew != 1 {
		t.Fatalf("growth decodes=%d want a single jump to the head", grew)
	}
	if got := InspectReads.Load(); got != 0 {
		t.Fatalf("earlier page must not Inspect: %d", got)
	}
}

func TestTailAcceptsGiantFramePartialSuffix(t *testing.T) {
	ResetInspectCache()
	shrinkColdTailMaxWindow(t, 1<<20)
	root := t.TempDir()
	id := "s-giant-tail"
	giantFrameLog(t, root, id, 4<<20, 15)
	store := Store{Root: root}

	tail, err := store.SyncTail(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := TailFullFallbacks.Load(); got != 0 {
		t.Fatalf("cold SyncTail decoded the whole log: fallbacks=%d", got)
	}
	if len(tail.Rows) != 15 {
		t.Fatalf("rows=%d want the 15 rows of the newest frames", len(tail.Rows))
	}
	if !tail.sawAnchor {
		t.Fatal("accepted suffix lost its anchor")
	}
	if !tail.windowed || tail.Size == 0 {
		t.Fatalf("accepted suffix must keep a windowed cursor: windowed=%v size=%d", tail.windowed, tail.Size)
	}
}

func TestViewTailSmallLogUnchanged(t *testing.T) {
	ResetInspectCache()
	resetViewTails()
	root := t.TempDir()
	id := "s-small"
	put(t, root, id, "session.v3.jsonl", header(id)+
		row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)+
		row("assistant/message", 1, `{"message":{"content":[{"type":"text","text":"yo"}]}}`))
	store := Store{Root: root}

	view, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if view.Truncated {
		t.Fatal("a small log is decoded whole")
	}
	if len(view.Rows) != 2 {
		t.Fatalf("rows=%d want 2", len(view.Rows))
	}
	if ViewTailFullFallbacks.Load() != 0 || TailFullFallbacks.Load() != 0 {
		t.Fatalf("fallbacks view=%d tail=%d", ViewTailFullFallbacks.Load(), TailFullFallbacks.Load())
	}
}

func TestColdTailCountsFullFallback(t *testing.T) {
	ResetInspectCache()
	resetViewTails()
	shrinkColdTailMaxWindow(t, 1<<20)
	root := t.TempDir()
	id := "s-giant-single-frame"
	rnd := rand.New(rand.NewSource(11))
	var plain strings.Builder
	plain.WriteString(header(id))
	seq := 0
	for plain.Len() < 4<<20 {
		plain.WriteString(bulkRow(seq, randomText(rnd, 6<<10)))
		seq++
	}
	// One frame for the whole log: no boundary exists inside any probe window,
	// so even the widened search keeps landing on the same mid-frame start.
	writeSessionLog(t, root, id, encodeFrame(t, plain.String()))
	store := Store{Root: root}

	view, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got := ViewTailFullFallbacks.Load(); got != 1 {
		t.Fatalf("ViewTail fallbacks=%d want 1", got)
	}
	if view.Truncated || len(view.Rows) != seq {
		t.Fatalf("fallback must decode the whole log: truncated=%v rows=%d want %d", view.Truncated, len(view.Rows), seq)
	}
	tail, err := store.SyncTail(id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := TailFullFallbacks.Load(); got != 1 {
		t.Fatalf("SyncTail fallbacks=%d want 1", got)
	}
	// Tail keeps the newest tailKeep rows, so a whole-log fallback is capped.
	keep := seq
	if keep > tailKeep {
		keep = tailKeep
	}
	if len(tail.Rows) != keep {
		t.Fatalf("tail rows=%d want %d", len(tail.Rows), keep)
	}
}

func BenchmarkViewTailGiantFrame(b *testing.B) {
	ResetInspectCache()
	shrinkColdTailMaxWindow(b, 1<<20)
	root := b.TempDir()
	id := "bench-giant"
	giantFrameLog(b, root, id, 4<<20, 15)
	store := Store{Root: root}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetViewTails()
		if _, err := store.ViewTail(id, 50); err != nil {
			b.Fatal(err)
		}
	}
}
