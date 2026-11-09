package consumers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// --- fixtures ---------------------------------------------------------------

const testDir = "project"

func headerAt(id string, version int) string {
	b, _ := json.Marshal(map[string]any{
		"type":            "session",
		"version":         version,
		"id":              id,
		"createdAt":       10,
		"cwd":             "/workspace",
		"isSeeded":        false,
		"delegationDepth": 0,
		"agentPreset":     "std-no-subagent",
	})
	return string(b) + "\n"
}

func row(typ string, seq int, data string) string {
	b, _ := json.Marshal(map[string]any{"type": typ, "seq": seq, "time": 20 + seq, "data": json.RawMessage(data)})
	return string(b) + "\n"
}

func zstdFrame(t testing.TB, plain string) []byte {
	t.Helper()
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	return enc.EncodeAll([]byte(plain), nil)
}

// writeArtifact writes a plain JSONL log, which the reader treats exactly like
// a zstd one except that it has no frame boundaries.
func writeArtifact(t testing.TB, root, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(root, testDir, dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	sessions.ResetInspectCache()
	return path
}

func appendBytes(t testing.TB, path string, blob []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(blob); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

// logFixture is a session store over a temp home plus a cache directory.
type logFixture struct {
	t     testing.TB
	root  string
	cache string
	store *sessions.Store
}

func newFixture(t testing.TB) *logFixture {
	t.Helper()
	root := t.TempDir()
	sessions.ResetInspectCache()
	t.Cleanup(sessions.ResetInspectCache)
	return &logFixture{
		t:     t,
		root:  root,
		cache: filepath.Join(t.TempDir(), "cache"),
		store: &sessions.Store{Root: root},
	}
}

// plain writes an uncompressed log under the fixture home.
func (f *logFixture) plain(id, body string) string {
	f.t.Helper()
	return writeArtifact(f.t, f.root, id, "session.v3.jsonl", body)
}

func (f *logFixture) bus(opts Options) *Bus {
	f.t.Helper()
	opts.Dir = f.cache
	if opts.Consumers == nil {
		opts.Consumers = Defaults()
	}
	bus, err := New(f.store, opts)
	if err != nil {
		f.t.Fatal(err)
	}
	return bus
}

func mustSync(t testing.TB, bus *Bus, id string) Snapshot {
	t.Helper()
	snapshot, err := bus.Sync(id)
	if err != nil {
		t.Fatalf("Sync(%s): %v", id, err)
	}
	return snapshot
}

// --- consumer folds ---------------------------------------------------------

func titleData(title, kind string) string {
	b, _ := json.Marshal(map[string]any{
		"messageSeqs": []int{0},
		"source":      map[string]any{"kind": kind},
		"title":       title,
	})
	return string(b)
}

func TestFoldTitle(t *testing.T) {
	f := newFixture(t)
	id := "s-title"
	f.plain(id, headerAt(id, 3)+
		row("session/title", 0, `{"title":"   ","source":{"kind":"fallback"}}`)+
		row("session/title", 1, titleData("first prompt", "fallback"))+
		row("unrelated", 2, `{"ignored":true}`)+
		row("session/title", 3, titleData("Migrate blog to xdev-agent", "provider"))+
		row("session/title", 4, titleData("Renamed by hand", "user")))

	bus := f.bus(Options{})
	snapshot := mustSync(t, bus, id)
	got := snapshot.Title()
	if got != "Renamed by hand" {
		t.Fatalf("title = %q", got)
	}
	if title := snapshot.TitleState(); title.Seq != 4 || title.Source != "user" {
		t.Fatalf("title state = %+v", title)
	}
	if !snapshot.CaughtUp {
		t.Fatalf("snapshot not caught up: %+v", snapshot)
	}
}

func TestFoldOutcome(t *testing.T) {
	f := newFixture(t)
	id := "s-outcome"
	f.plain(id, headerAt(id, 3)+
		row("turn/start", 0, `{"turn":0}`)+
		row("turn/end", 1, `{"turn":0,"reason":{"kind":"completed"}}`)+
		row("turn/start", 2, `{"turn":1}`)+
		row("llm/retry", 3, `{"turn":1,"retry":1,"failure":{"message":"terminated","code":"TRANSPORT"}}`)+
		row("turn/end", 4, `{"turn":1,"reason":{"kind":"error","error":{"message":"boom","code":"SERVER"}}}`))

	bus := f.bus(Options{})
	snapshot := mustSync(t, bus, id)
	outcome := snapshot.Outcome()
	if outcome.Running {
		t.Fatalf("running after turn/end: %+v", outcome)
	}
	if outcome.Turns != 2 || outcome.Errors != 1 || outcome.Retries != 1 {
		t.Fatalf("outcome = %+v", outcome)
	}
	if outcome.Last != OutcomeError || outcome.Turn != 1 {
		t.Fatalf("last = %q turn = %d", outcome.Last, outcome.Turn)
	}
	if failure := snapshot.Failure(); failure == nil || failure.Message != "boom" || failure.Code != "SERVER" {
		t.Fatalf("failure = %+v", failure)
	}

	// An unfinished turn keeps the session running until it is closed.
	appendBytes(t, filepath.Join(f.root, testDir, id, "session.v3.jsonl"),
		[]byte(row("turn/start", 5, `{"turn":2}`)))
	sessions.ResetInspectCache()
	if running := mustSync(t, bus, id).Outcome(); !running.Running || running.Turn != 2 {
		t.Fatalf("outcome after new turn = %+v", running)
	}
}

func TestFoldQuestion(t *testing.T) {
	f := newFixture(t)
	id := "s-question"
	askArguments := `{"questions":[{"header":"Pick","id":"q1","question":"Which one?","multiSelect":false,"options":[{"label":"A","description":"first"},{"label":"B"}]}]}`
	askData, _ := json.Marshal(map[string]any{
		"turn": 0, "step": 1, "callId": "call-1", "name": "ask_user_question",
		"arguments": askArguments,
	})
	resultData, _ := json.Marshal(map[string]any{
		"turn": 0, "step": 1,
		"message": map[string]any{"role": "tool", "toolCallId": "call-1", "isError": false},
	})
	f.plain(id, headerAt(id, 3)+
		row("turn/start", 0, `{"turn":0}`)+
		row("tool/call", 1, string(askData))+
		row("approval/asked", 2, `{"id":"app-1","toolName":"bash","callId":"call-2","reason":"needs network"}`))

	bus := f.bus(Options{})
	snapshot := mustSync(t, bus, id)
	question := snapshot.Question()
	if !snapshot.Pending() {
		t.Fatalf("no pending prompt: %+v", question)
	}
	if question.Asked != 1 || question.Answered != 0 || len(question.Asks) != 1 {
		t.Fatalf("asks = %+v", question)
	}
	ask := question.Asks[0]
	if ask.CallID != "call-1" || len(ask.Items) != 1 || ask.Items[0].Question != "Which one?" {
		t.Fatalf("ask = %+v", ask)
	}
	if options := ask.Items[0].Options; len(options) != 2 || options[0] != "A" || options[1] != "B" {
		t.Fatalf("options = %v", options)
	}
	if len(question.Approvals) != 1 || question.Approvals[0].Tool != "bash" {
		t.Fatalf("approvals = %+v", question.Approvals)
	}

	appendBytes(t, filepath.Join(f.root, testDir, id, "session.v3.jsonl"),
		[]byte(row("tool/result", 3, string(resultData))+
			row("approval/decided", 4, `{"id":"app-1","outcome":"allowed-once"}`)))
	sessions.ResetInspectCache()
	resolved := mustSync(t, bus, id)
	if resolved.Pending() {
		t.Fatalf("prompt still pending: %+v", resolved.Question())
	}
	if resolved.Question().Answered != 1 || resolved.Question().ApprovalsDecided != 1 {
		t.Fatalf("counts = %+v", resolved.Question())
	}

	// A turn that ends with an ask still open means the ask was abandoned.
	appendBytes(t, filepath.Join(f.root, testDir, id, "session.v3.jsonl"),
		[]byte(row("tool/call", 5, string(askData))+row("turn/end", 6, `{"turn":0,"reason":{"kind":"aborted"}}`)))
	sessions.ResetInspectCache()
	abandoned := mustSync(t, bus, id)
	if abandoned.Pending() {
		t.Fatalf("abandoned ask still pending: %+v", abandoned.Question())
	}
}

func TestFoldUsage(t *testing.T) {
	f := newFixture(t)
	id := "s-usage"
	usage := func(input, output, cacheRead int) string {
		return `{"turn":0,"step":0,"usage":{"inputTokens":` + itoa(input) +
			`,"outputTokens":` + itoa(output) + `,"cacheReadTokens":` + itoa(cacheRead) + `}}`
	}
	f.plain(id, headerAt(id, 3)+
		row("turn/start", 0, `{"turn":0}`)+
		row("assistant/message", 1, usage(3884, 115, 6144))+
		row("assistant/message", 2, usage(1000, 20, 0))+
		row("assistant/message", 3, `{"turn":0,"step":2,"message":{"content":[]}}`))

	bus := f.bus(Options{})
	got := mustSync(t, bus, id).Usage()
	if got.Messages != 2 || got.Turns != 1 {
		t.Fatalf("usage = %+v", got)
	}
	if got.Input != 4884 || got.Output != 135 || got.CacheRead != 6144 {
		t.Fatalf("usage = %+v", got)
	}
	if got.Total != 4884+135+6144 {
		t.Fatalf("total = %d", got.Total)
	}
}

func itoa(v int) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// --- incremental reads ------------------------------------------------------

func TestSyncReadsOnlyNewBytes(t *testing.T) {
	f := newFixture(t)
	id := "s-incremental"
	path := f.plain(id, headerAt(id, 3)+
		row("session/title", 0, titleData("one", "user"))+
		row("turn/start", 1, `{"turn":0}`))

	bus := f.bus(Options{})
	first := mustSync(t, bus, id)
	if first.ReadBytes == 0 || first.Folded != 4 {
		t.Fatalf("first sync = %+v", first)
	}
	if !first.CaughtUp {
		t.Fatalf("first sync not caught up")
	}
	afterFirst := bus.Stats()

	// An unchanged log must not be read at all.
	second := mustSync(t, bus, id)
	if second.ReadBytes != 0 || second.Folded != 0 {
		t.Fatalf("second sync read bytes: %+v", second)
	}
	if got := bus.Stats().CacheHits - afterFirst.CacheHits; got != 1 {
		t.Fatalf("cache hits grew by %d", got)
	}

	// One appended row costs that row, not the log.
	appended := row("session/title", 2, titleData("two", "user"))
	appendBytes(t, path, []byte(appended))
	sessions.ResetInspectCache()
	third := mustSync(t, bus, id)
	if third.ReadBytes != int64(len(appended)) {
		t.Fatalf("read %d bytes for a %d-byte append", third.ReadBytes, len(appended))
	}
	if third.Folded != 1 || third.Title() != "two" {
		t.Fatalf("third sync = %+v title %q", third, third.Title())
	}
}

func TestSyncResumesFromDiskCache(t *testing.T) {
	f := newFixture(t)
	id := "s-resume"
	f.plain(id, headerAt(id, 3)+row("session/title", 0, titleData("cached", "user")))

	first := mustSync(t, f.bus(Options{}), id)
	if first.ReadBytes == 0 {
		t.Fatalf("cold sync read nothing")
	}

	// A new bus over the same cache directory reads no bytes.
	warm := f.bus(Options{})
	second := mustSync(t, warm, id)
	if second.ReadBytes != 0 {
		t.Fatalf("warm bus read %d bytes", second.ReadBytes)
	}
	if second.Title() != "cached" {
		t.Fatalf("title = %q", second.Title())
	}
	if got := warm.Stats().CacheHits; got != 1 {
		t.Fatalf("cache hits = %d", got)
	}
}

func TestCacheCorruptionIsAMiss(t *testing.T) {
	f := newFixture(t)
	id := "s-corrupt"
	f.plain(id, headerAt(id, 3)+row("session/title", 0, titleData("intact", "user")))
	mustSync(t, f.bus(Options{}), id)

	records, err := os.ReadDir(f.cache)
	if err != nil || len(records) != 1 {
		t.Fatalf("records = %v err %v", records, err)
	}
	if err := os.WriteFile(filepath.Join(f.cache, records[0].Name()), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	bus := f.bus(Options{})
	snapshot := mustSync(t, bus, id)
	if snapshot.Title() != "intact" {
		t.Fatalf("title = %q", snapshot.Title())
	}
	if snapshot.ReadBytes == 0 {
		t.Fatalf("corrupt cache was not re-read")
	}
}

func TestVersionBumpReplaysOnlyThatConsumer(t *testing.T) {
	f := newFixture(t)
	id := "s-version"
	f.plain(id, headerAt(id, 3)+
		row("session/title", 0, titleData("stable", "user"))+
		row("turn/start", 1, `{"turn":0}`)+
		row("assistant/message", 2, `{"turn":0,"step":0,"usage":{"inputTokens":7,"outputTokens":3,"cacheReadTokens":0}}`)+
		row("turn/end", 3, `{"turn":0,"reason":{"kind":"completed"}}`))
	mustSync(t, f.bus(Options{}), id)

	// Bump the title consumer only: it restarts from the origin, everything
	// else must skip the rows it already folded.
	bumped := []Consumer{
		&Typed[TitleState]{
			Name:  TitleID,
			Ver:   2,
			Watch: []string{"session/title"},
			Apply: func(state *TitleState, ev Event, _ Point) error {
				fields := decodeObject(ev.Data)
				state.Title = text(fields, "title")
				state.Source = text(objectField(fields, "source"), "kind")
				state.Seq = ev.Seq
				return nil
			},
		},
		Outcome(), Question(), Usage(),
	}
	bus := f.bus(Options{Consumers: bumped})
	snapshot := mustSync(t, bus, id)
	if snapshot.Title() != "stable" {
		t.Fatalf("title after bump = %q", snapshot.Title())
	}
	if snapshot.Skipped == 0 {
		t.Fatalf("unchanged consumers did not skip: %+v", snapshot)
	}
	if usage := snapshot.Usage(); usage.Input != 7 || usage.Total != 10 {
		t.Fatalf("usage was folded twice: %+v", usage)
	}
	if outcome := snapshot.Outcome(); outcome.Turns != 1 {
		t.Fatalf("outcome was folded twice: %+v", outcome)
	}
	if got := bus.Stats().Resets; got == 0 {
		t.Fatalf("version bump was not counted as a reset")
	}

	// A third run must be back to zero reads.
	if again := mustSync(t, bus, id); again.ReadBytes != 0 {
		t.Fatalf("post-replay sync read %d bytes", again.ReadBytes)
	}
}

func TestTornTailIsNotConsumed(t *testing.T) {
	f := newFixture(t)
	id := "s-torn"
	first := zstdFrame(t, headerAt(id, 3)+row("session/title", 0, titleData("durable", "user")))
	secondFrame := zstdFrame(t, row("session/title", 1, titleData("partial", "user")))
	path := filepath.Join(f.root, testDir, id, "session.v4.jsonl.zstd")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append([]byte{}, first...), secondFrame[:len(secondFrame)/2]...), 0o644); err != nil {
		t.Fatal(err)
	}
	sessions.ResetInspectCache()

	bus := f.bus(Options{})
	snapshot := mustSync(t, bus, id)
	if snapshot.CaughtUp {
		t.Fatalf("torn tail reported caught up: %+v", snapshot)
	}
	if snapshot.Title() != "durable" {
		t.Fatalf("title = %q", snapshot.Title())
	}
	if snapshot.LastSeq != 0 {
		t.Fatalf("last seq = %d", snapshot.LastSeq)
	}
	// The reader must stop at the frame boundary, not past it: a second read
	// picks up the completed frame and nothing else.

	appendBytes(t, path, secondFrame[len(secondFrame)/2:])
	sessions.ResetInspectCache()
	completed := mustSync(t, bus, id)
	if !completed.CaughtUp || completed.Title() != "partial" {
		t.Fatalf("completed = %+v title %q", completed, completed.Title())
	}
	if completed.ReadBytes != int64(len(secondFrame)) {
		t.Fatalf("read %d bytes for a %d-byte frame", completed.ReadBytes, len(secondFrame))
	}
}

func TestShrinkRestartsTheStream(t *testing.T) {
	f := newFixture(t)
	id := "s-shrink"
	path := f.plain(id, headerAt(id, 3)+
		row("session/title", 0, titleData("long history", "user"))+
		row("turn/start", 1, `{"turn":0}`)+
		row("turn/end", 2, `{"turn":0,"reason":{"kind":"completed"}}`))
	bus := f.bus(Options{})
	mustSync(t, bus, id)

	if err := os.WriteFile(path, []byte(headerAt(id, 3)+row("session/title", 0, titleData("rewritten", "user"))), 0o644); err != nil {
		t.Fatal(err)
	}
	sessions.ResetInspectCache()
	snapshot := mustSync(t, bus, id)
	if snapshot.Title() != "rewritten" {
		t.Fatalf("title = %q", snapshot.Title())
	}
	if outcome := snapshot.Outcome(); outcome.Turns != 0 || outcome.Running {
		t.Fatalf("stale outcome survived the rewrite: %+v", outcome)
	}
}

func TestSizeEqualRewriteRestarts(t *testing.T) {
	f := newFixture(t)
	id := "s-rewrite"
	path := f.plain(id, headerAt(id, 3)+row("session/title", 0, titleData("aaaa", "user")))
	bus := f.bus(Options{})
	mustSync(t, bus, id)

	// Same length, different content and a newer mtime.
	body := headerAt(id, 3) + row("session/title", 0, titleData("bbbb", "user"))
	if len(body) != len(headerAt(id, 3)+row("session/title", 0, titleData("aaaa", "user"))) {
		t.Fatalf("fixture bodies differ in length")
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	sessions.ResetInspectCache()
	if got := mustSync(t, bus, id).Title(); got != "bbbb" {
		t.Fatalf("title = %q", got)
	}
}

func TestEventBudgetResumes(t *testing.T) {
	f := newFixture(t)
	id := "s-budget"
	var body strings.Builder
	body.WriteString(headerAt(id, 3))
	for seq := 0; seq < 12; seq++ {
		body.WriteString(row("session/title", seq, titleData("t"+itoa(seq), "user")))
	}
	path := filepath.Join(f.root, testDir, id, "session.v4.jsonl.zstd")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// One frame per row, so the event budget can stop on a frame boundary.
	for _, line := range strings.Split(strings.TrimSuffix(body.String(), "\n"), "\n") {
		appendBytes(t, path, zstdFrame(t, line+"\n"))
	}
	sessions.ResetInspectCache()

	bus := f.bus(Options{MaxEvents: 1, MaxBatches: 1, Budget: time.Hour})
	var folded int64
	snapshot := mustSync(t, bus, id)
	for attempt := 0; !snapshot.CaughtUp; attempt++ {
		if attempt > 20 {
			t.Fatalf("never caught up: %+v", snapshot)
		}
		folded += snapshot.Folded
		snapshot = mustSync(t, bus, id)
	}
	folded += snapshot.Folded
	if folded != 12 {
		t.Fatalf("folded %d rows", folded)
	}
	if snapshot.LastSeq != 11 || snapshot.Title() != "t11" {
		t.Fatalf("snapshot = %+v title %q", snapshot, snapshot.Title())
	}
}

func TestLargeFrameCrossesWindow(t *testing.T) {
	f := newFixture(t)
	id := "s-large-frame"
	var body strings.Builder
	body.WriteString(headerAt(id, 3))
	for seq := 0; seq < 200; seq++ {
		body.WriteString(row("session/title", seq, titleData("title-"+itoa(seq), "user")))
	}
	path := filepath.Join(f.root, testDir, id, "session.v4.jsonl.zstd")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, zstdFrame(t, body.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	sessions.ResetInspectCache()

	// A window far smaller than the single frame forces the growth path.
	bus := f.bus(Options{Window: 256, MaxWindow: 4 << 20, MaxBatches: 8})
	snapshot := mustSync(t, bus, id)
	for attempt := 0; !snapshot.CaughtUp; attempt++ {
		if attempt > 8 {
			t.Fatalf("never caught up: %+v", snapshot)
		}
		snapshot = mustSync(t, bus, id)
	}
	if snapshot.Title() != "title-199" || snapshot.LastSeq != 199 {
		t.Fatalf("snapshot = %+v title %q", snapshot, snapshot.Title())
	}
}

func TestSeqGapIsReported(t *testing.T) {
	f := newFixture(t)
	id := "s-gap"
	path := f.plain(id, headerAt(id, 3)+
		row("session/title", 0, titleData("one", "user"))+
		row("session/title", 1, titleData("two", "user")))
	bus := f.bus(Options{})
	mustSync(t, bus, id)

	appendBytes(t, path, []byte(row("session/title", 5, titleData("skipped", "user"))))
	sessions.ResetInspectCache()
	_, err := bus.Sync(id)
	if err == nil {
		t.Fatalf("gap was not reported")
	}
	if !strings.Contains(err.Error(), "where 2 was expected") {
		t.Fatalf("error = %v", err)
	}
}

func TestForgetRewinds(t *testing.T) {
	f := newFixture(t)
	id := "s-forget"
	f.plain(id, headerAt(id, 3)+row("session/title", 0, titleData("one", "user")))
	bus := f.bus(Options{})
	if read := mustSync(t, bus, id).ReadBytes; read == 0 {
		t.Fatalf("cold sync read nothing")
	}
	if err := bus.Forget(id); err != nil {
		t.Fatal(err)
	}
	if read := mustSync(t, bus, id).ReadBytes; read == 0 {
		t.Fatalf("forget did not rewind")
	}
	if _, err := os.Stat(filepath.Join(f.cache, encodeID(id)+".json")); err != nil {
		t.Fatalf("record was not rewritten: %v", err)
	}
}

func TestSyncAllConcurrent(t *testing.T) {
	f := newFixture(t)
	ids := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		id := "s-all-" + itoa(i)
		ids = append(ids, id)
		f.plain(id, headerAt(id, 3)+
			row("session/title", 0, titleData("session "+itoa(i), "user"))+
			row("turn/start", 1, `{"turn":0}`))
	}
	bus := f.bus(Options{Parallelism: 4, MaxOpenStreams: 2})
	snapshots := bus.SyncAll(ids)
	if len(snapshots) != len(ids) {
		t.Fatalf("snapshots = %d want %d", len(snapshots), len(ids))
	}
	for i, id := range ids {
		snapshot, ok := snapshots[id]
		if !ok {
			t.Fatalf("session %s missing", id)
		}
		if snapshot.Title() != "session "+itoa(i) {
			t.Fatalf("%s title = %q", id, snapshot.Title())
		}
		if !snapshot.Outcome().Running {
			t.Fatalf("%s not running", id)
		}
	}
	for _, snapshot := range bus.SyncAll(ids) {
		if snapshot.ReadBytes != 0 {
			t.Fatalf("warm sync read %d bytes", snapshot.ReadBytes)
		}
	}
}

// --- validation -------------------------------------------------------------

func TestNewValidatesConsumers(t *testing.T) {
	store := &sessions.Store{Root: t.TempDir()}
	cases := []struct {
		name      string
		consumers []Consumer
		want      string
	}{
		{"version zero", []Consumer{&Typed[OutcomeState]{Name: "x", Ver: 0}}, "version"},
		{"no id", []Consumer{&Typed[OutcomeState]{Ver: 1}}, "id"},
		{"duplicate", []Consumer{Outcome(), Outcome()}, "duplicate"},
	}
	for _, tc := range cases {
		_, err := New(store, Options{Dir: t.TempDir(), Consumers: tc.consumers})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: err = %v", tc.name, err)
		}
	}
	if _, err := New(nil, Options{}); err == nil {
		t.Fatalf("nil store accepted")
	}
	if got := Defaults(); len(got) != 6 {
		t.Fatalf("defaults = %d consumers", len(got))
	}
}

func TestCacheDirDefaultsUnderUserCache(t *testing.T) {
	store := &sessions.Store{Root: t.TempDir()}
	bus, err := New(store, Options{Consumers: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	base, err := os.UserCacheDir()
	if err != nil {
		t.Skipf("no user cache dir: %v", err)
	}
	want := filepath.Join(base, filepath.FromSlash(CacheSuffix))
	if bus.CacheDir() != want {
		t.Fatalf("cache dir = %q want %q", bus.CacheDir(), want)
	}
}

func TestSyncUnknownSession(t *testing.T) {
	f := newFixture(t)
	_, err := f.bus(Options{}).Sync("s-missing")
	if err == nil || !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestSnapshotValue(t *testing.T) {
	f := newFixture(t)
	id := "s-value"
	f.plain(id, headerAt(id, 3)+row("session/title", 0, titleData("decoded", "user")))
	snapshot := mustSync(t, f.bus(Options{}), id)

	var title TitleState
	if !snapshot.Value(TitleID, &title) {
		t.Fatalf("value missing: %v", snapshot.Values)
	}
	if title.Title != "decoded" || title.Seq != 0 {
		t.Fatalf("title = %+v", title)
	}
	var missing UsageState
	if snapshot.Value("absent", &missing) {
		t.Fatalf("absent consumer reported a value")
	}
}

// --- benchmarks -------------------------------------------------------------

func benchmarkLog(tb testing.TB, rows int) *logFixture {
	tb.Helper()
	f := newFixture(tb)
	var body strings.Builder
	body.WriteString(headerAt("s-bench", 3))
	for seq := 0; seq < rows; seq++ {
		if seq%7 == 0 {
			body.WriteString(row("turn/start", seq, `{"turn":`+itoa(seq/7)+`}`))
			continue
		}
		body.WriteString(row("session/title", seq, titleData("title-"+itoa(seq), "user")))
	}
	f.plain("s-bench", body.String())
	return f
}

func BenchmarkSyncCold(b *testing.B) {
	f := benchmarkLog(b, 2000)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bus := f.bus(Options{})
		b.StartTimer()
		if _, err := bus.Sync("s-bench"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSyncWarm(b *testing.B) {
	f := benchmarkLog(b, 2000)
	bus := f.bus(Options{})
	if _, err := bus.Sync("s-bench"); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := bus.Sync("s-bench"); err != nil {
			b.Fatal(err)
		}
	}
}

// TestRealSessionLog is opt-in: CONSUMERS_REAL_DSH=1 go test ./agent/dsh/consumers -run RealSession -v
func TestRealSessionLog(t *testing.T) {
	store := realStore(t)
	summaries, _, err := store.ListSummaries()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) == 0 {
		t.Skip("no sessions under the DSH home")
	}
	// The report is about the biggest log, which is also the worst case.
	summary := summaries[0]
	bus := realBus(t, store)
	cold, err := bus.Sync(summary.ID)
	if err != nil {
		t.Fatalf("cold sync %s: %v", summary.ID, err)
	}
	t.Logf("cold: %d/%d bytes, last seq %d, %d folds, caught up %v",
		cold.ReadBytes, cold.Size, cold.LastSeq, cold.Folded, cold.CaughtUp)

	rounds := 1
	for !cold.CaughtUp && rounds < 50 {
		cold, err = bus.Sync(summary.ID)
		if err != nil {
			t.Fatalf("sync round %d: %v", rounds, err)
		}
		rounds++
	}
	if !cold.CaughtUp {
		t.Fatalf("never caught up after %d rounds", rounds)
	}
	t.Logf("caught up after %d rounds; title %q, usage %+v, outcome %+v",
		rounds, cold.Title(), cold.Usage(), cold.Outcome())

	warm, err := bus.Sync(summary.ID)
	if err != nil {
		t.Fatal(err)
	}
	if warm.ReadBytes != 0 || warm.Folded != 0 {
		t.Fatalf("warm sync re-read: %d bytes, %d folds", warm.ReadBytes, warm.Folded)
	}
	inspection, err := store.Inspect(summary.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cold.Title() != inspection.Title {
		t.Fatalf("title = %q, inspect says %q", cold.Title(), inspection.Title)
	}
}

// TestRealTitlesMatchInspect checks the title fold against the full reader on
// real logs. A late rename lives far from the end, so ListSummaries (which
// samples the prefix and the tail window) can miss it while Inspect and the
// fold both see it; Inspect is therefore the reference.
func TestRealTitlesMatchInspect(t *testing.T) {
	store := realStore(t)
	summaries, _, err := store.ListSummaries()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) == 0 {
		t.Skip("no sessions under the DSH home")
	}
	limit := len(summaries)
	if limit > 24 {
		limit = 24
	}
	ids := make([]string, 0, limit)
	stale := 0
	for _, summary := range summaries[:limit] {
		ids = append(ids, summary.ID)
	}
	bus := realBus(t, store)
	started := time.Now()
	snapshots := bus.SyncAll(ids)
	for round := 0; round < 40; round++ {
		var pending []string
		for _, id := range ids {
			if snapshot, ok := snapshots[id]; ok && !snapshot.CaughtUp {
				pending = append(pending, id)
			}
		}
		if len(pending) == 0 {
			break
		}
		for id, snapshot := range bus.SyncAll(pending) {
			snapshots[id] = snapshot
		}
	}
	elapsed := time.Since(started)
	stats := bus.Stats()

	matched, mismatched := 0, 0
	for _, id := range ids {
		snapshot, ok := snapshots[id]
		if !ok {
			continue
		}
		if !snapshot.CaughtUp {
			// A live log keeps growing; compare only a settled stream.
			t.Logf("%s: skipped, still %d/%d bytes", id, snapshot.ReadBytes, snapshot.Size)
			continue
		}
		inspection, err := store.Inspect(id)
		if err != nil {
			t.Errorf("%s: inspect: %v", id, err)
			continue
		}
		if summary := findSummary(summaries, id); summary != nil && summary.Title != inspection.Title {
			stale++
		}
		if snapshot.Title() == inspection.Title {
			matched++
			continue
		}
		mismatched++
		t.Errorf("%s: fold title %q, inspect title %q (caught up %v)", id, snapshot.Title(), inspection.Title, snapshot.CaughtUp)
	}
	t.Logf("%d sessions in %s: %d titles matched, %d mismatched; %d summaries carry a stale sample title; %d bytes read, %d events, %d folds, %d errors",
		len(ids), elapsed.Round(time.Millisecond), matched, mismatched, stale,
		stats.BytesRead, stats.Events, stats.Folded, stats.Errors)
}

func findSummary(summaries []sessions.Summary, id string) *sessions.Summary {
	for i := range summaries {
		if summaries[i].ID == id {
			return &summaries[i]
		}
	}
	return nil
}

func realStore(t *testing.T) *sessions.Store {
	t.Helper()
	if os.Getenv("CONSUMERS_REAL_DSH") != "1" {
		t.Skip("set CONSUMERS_REAL_DSH=1 to follow the real ~/.dsh sessions")
	}
	store, err := sessions.NewStore(sessions.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func realBus(t *testing.T, store *sessions.Store) *Bus {
	t.Helper()
	bus, err := New(store, Options{Dir: t.TempDir(), Consumers: Defaults(), Parallelism: 8})
	if err != nil {
		t.Fatal(err)
	}
	return bus
}
