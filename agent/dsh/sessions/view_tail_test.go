package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
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

func TestViewTailProjectionMatchesStoreView(t *testing.T) {
	resetInspectBodies()
	ResetInspectCache()
	resetViewTails()
	root := t.TempDir()
	id := "s-tail-proj"
	put(t, root, id, "session.v3.jsonl", header(id)+
		row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"confirm?"}]}`)+
		row("tool/call", 1, `{"callId":"c1","name":"ask_user_question","arguments":{"questions":[{"id":"kind","question":"哪一种？","options":[{"label":"金色标记"}]}]}}`)+
		row("tool/result", 2, `{"callId":"c1","name":"ask_user_question","content":[{"type":"text","text":"{\"answers\":[{\"id\":\"kind\",\"selected\":[\"金色标记\"]}]}"}]}`)+
		row("tool/call", 3, `{"name":"todo_write","arguments":{"todos":[{"content":"a","status":"in_progress"},{"content":" ","status":"pending"},{"content":"b"}]}}`)+
		row("request/header", 4, `{"header":{"config":{"provider":"xai","model":"grok","reasoningEffort":"high"}}}`)+
		row("tool/call", 5, `{"name":"present","arguments":{"files":[{"path":"a.md","description":"doc"}]}}`)+
		row("assistant/message", 6, `{"message":{"content":[{"type":"thinking","text":"think"},{"type":"text","text":"done"}]}}`))
	store := Store{Root: root}
	tail, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	full, err := store.View(id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tail.Rows, full.Rows) {
		t.Fatalf("incremental projection drifted:\ntail=%s\nfull=%s", mustJSON(tail.Rows), mustJSON(full.Rows))
	}
	if tail.ReasoningEffort != full.ReasoningEffort || !reflect.DeepEqual(tail.Todos, full.Todos) {
		t.Fatalf("effort/todos: tail=%q %+v full=%q %+v", tail.ReasoningEffort, tail.Todos, full.ReasoningEffort, full.Todos)
	}
}

func TestViewTailBeforeGrowsWindowToWholeLog(t *testing.T) {
	ResetInspectCache()
	resetViewTails()
	root := t.TempDir()
	id := "s-tail-before"
	var blob []byte
	blob = append(blob, zstdFrameTB(t, header(id))...)
	pad := strings.Repeat("0123456789abcdef", 8)
	for i := 0; i < 4000; i++ {
		blob = append(blob, zstdFrameTB(t, row("user/message", i, fmt.Sprintf(
			`{"source":{"kind":"user"},"content":[{"type":"text","text":"msg-%d-%s"}]}`, i, pad)))...)
	}
	put(t, root, id, "session.v3.jsonl.zstd", string(blob))
	store := Store{Root: root}
	head, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if !head.Truncated || len(head.Rows) == 0 {
		t.Fatalf("expected a suffix window, got truncated=%v rows=%d", head.Truncated, len(head.Rows))
	}
	reads := ViewTailReads.Load()
	wide, err := store.ViewTailBefore(id, head.Rows[0].Seq, 50, nil)
	if err != nil {
		t.Fatal(err)
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
	if grew := ViewTailReads.Load() - reads; grew < 1 || grew > beforeSeqGrowSteps {
		t.Fatalf("growth decodes=%d want 1..%d", grew, beforeSeqGrowSteps)
	}
	if got := InspectReads.Load(); got != 0 {
		t.Fatalf("earlier page must not decode the whole log: InspectReads=%d", got)
	}
	cached := ViewTailReads.Load()
	again, err := store.ViewTailBefore(id, head.Rows[0].Seq, 50, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again != wide || ViewTailReads.Load() != cached {
		t.Fatalf("repeat earlier page re-decoded: reads=%d want %d", ViewTailReads.Load(), cached)
	}
	for i := 0; i < 32 && wide.Truncated; i++ {
		wide, err = store.ViewTailBefore(id, wide.Rows[0].Seq, 50, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
	if wide.Truncated {
		t.Fatal("window never reached the file head")
	}
	full, err := store.View(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(wide.Rows) != len(full.Rows) {
		t.Fatalf("paged rows=%d want %d", len(wide.Rows), len(full.Rows))
	}
	if !reflect.DeepEqual(wide.Rows, full.Rows) {
		for i := range wide.Rows {
			if !reflect.DeepEqual(wide.Rows[i], full.Rows[i]) {
				t.Fatalf("row[%d]=%s want %s", i, mustJSON(wide.Rows[i]), mustJSON(full.Rows[i]))
			}
		}
	}
}

func TestViewTailCacheEvictsLeastRecentlyUsed(t *testing.T) {
	ResetInspectCache()
	resetViewTails()
	root := t.TempDir()
	store := Store{Root: root}
	keys := make([]string, 0, viewTailsMaxEntries+4)
	for i := 0; i < viewTailsMaxEntries+4; i++ {
		id := fmt.Sprintf("s-evict-%d", i)
		keys = append(keys, put(t, root, id, "session.v3.jsonl", header(id)+
			row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)))
		if _, err := store.ViewTail(id, 50); err != nil {
			t.Fatal(err)
		}
	}
	viewTails.mu.Lock()
	size := len(viewTails.cache)
	_, newest := viewTails.cache[keys[len(keys)-1]]
	_, oldest := viewTails.cache[keys[0]]
	viewTails.mu.Unlock()
	if size > viewTailsMaxEntries {
		t.Fatalf("cache=%d want <=%d", size, viewTailsMaxEntries)
	}
	if !newest || oldest {
		t.Fatalf("eviction kept the wrong cursors: newest=%v oldest=%v", newest, oldest)
	}
	reads := ViewTailReads.Load()
	if _, err := store.ViewTail(fmt.Sprintf("s-evict-%d", viewTailsMaxEntries+3), 50); err != nil {
		t.Fatal(err)
	}
	if ViewTailReads.Load() != reads {
		t.Fatal("recently used cursor should stay cached")
	}
	if _, err := store.ViewTail("s-evict-0", 50); err != nil {
		t.Fatal(err)
	}
	if ViewTailReads.Load() == reads {
		t.Fatal("evicted cursor should decode again")
	}
}

func mustJSON(v any) string {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(body)
}

// An earlier page can expose the ask row whose answer arrived in the window
// that was already decoded; the prepend must carry those answers over.
func TestViewTailBeforeCarriesAnswersToPrependedAsk(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-tail-ask"
	var blob []byte
	blob = append(blob, zstdFrameTB(t, header(id))...)
	blob = append(blob, zstdFrameTB(t, row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"go"}]}`))...)
	blob = append(blob, zstdFrameTB(t, row("tool/call", 1, `{"callId":"c1","name":"ask_user_question","arguments":{"questions":[{"id":"kind","question":"哪一种？","options":[{"label":"金色标记"}]}]}}`))...)
	// The filler must not compress away, otherwise the log never exceeds
	// coldTailMin and the request reads it whole instead of windowing.
	chunk := ""
	for seq := 2; seq < 4000; seq++ {
		words := make([]string, 32)
		for i := range words {
			words[i] = fmt.Sprintf("%08x", (uint64(seq)*2654435761+uint64(i)*40503)&0xffffffff)
		}
		chunk += row("user/message", seq, fmt.Sprintf(`{"source":{"kind":"user"},"content":[{"type":"text","text":"%s-%d"}]}`, strings.Join(words, " "), seq))
		if seq%100 == 0 {
			blob = append(blob, zstdFrameTB(t, chunk)...)
			chunk = ""
		}
	}
	if chunk != "" {
		blob = append(blob, zstdFrameTB(t, chunk)...)
	}
	blob = append(blob, zstdFrameTB(t, row("tool/result", 4000, `{"callId":"c1","name":"ask_user_question","content":[{"type":"text","text":"{\"answers\":[{\"id\":\"kind\",\"selected\":[\"金色标记\"]}]}"}]}`))...)
	if int64(len(blob)) <= coldTailMin {
		t.Fatalf("log=%d bytes should exceed the cold tail window", len(blob))
	}
	put(t, root, id, "session.v3.jsonl.zstd", string(blob))
	store := Store{Root: root}
	head, err := store.ViewTail(id, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !head.Truncated {
		t.Fatalf("first window should be a suffix, rows=%d", len(head.Rows))
	}
	for _, message := range head.Rows {
		if message.Ask != nil && message.Ask.CallID == "c1" {
			t.Fatal("test needs the ask row outside the first window")
		}
	}
	wide, err := store.ViewTailBefore(id, head.Rows[0].Seq, 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range wide.Rows {
		if message.Ask == nil || message.Ask.CallID != "c1" {
			continue
		}
		if len(message.Ask.Answers) == 0 {
			t.Fatalf("prepended ask row lost its answers: %s", mustJSON(message.Ask))
		}
		return
	}
	t.Fatal("ask row never appeared in the widened window")
}

// Logs below coldTailMin are decoded whole, but the payload is still zstd
// framed: reading them as plain JSONL parsed compressed bytes as a header.
func TestViewTailSmallZstdLogIsDecodedWhole(t *testing.T) {
	ResetInspectCache()
	root := t.TempDir()
	id := "s-tail-small-zstd"
	plain := header(id) +
		row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`) +
		row("assistant/message", 1, `{"message":{"content":[{"type":"text","text":"yo"}]}}`)
	blob := zstdFrameTB(t, plain)
	if int64(len(blob)) > coldTailMin {
		t.Fatal("fixture should stay below the cold tail floor")
	}
	put(t, root, id, "session.v3.jsonl.zstd", string(blob))
	store := Store{Root: root}
	view, err := store.ViewTail(id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if view.Truncated || len(view.Rows) != 2 {
		t.Fatalf("view=%+v", view)
	}
}
