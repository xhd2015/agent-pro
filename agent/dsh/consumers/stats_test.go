package consumers

import "testing"

// The usage rows below are real shapes lifted from a live log: the token fields
// carry their "Tokens" suffix, and inputTokens is the uncached part, so it can
// be smaller than cacheReadTokens.
const (
	usageReadHeavy = `{"usage":{"inputTokens":3884,"outputTokens":115,"totalTokens":10143,"cacheReadTokens":6144}}`
	usageReadAgain = `{"usage":{"inputTokens":273,"outputTokens":141,"totalTokens":10398,"cacheReadTokens":9984}}`
	usageWrite     = `{"usage":{"inputTokens":100,"outputTokens":20,"cacheReadTokens":0,"cacheWriteTokens":50,"totalTokens":170}}`
)

func statsOf(t *testing.T, body string) StatsState {
	t.Helper()
	f := newFixture(t)
	f.plain("session-stats", body)
	bus := f.bus(Options{Consumers: []Consumer{SessionStats()}})
	snapshot, err := bus.Sync("session-stats")
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !snapshot.CaughtUp {
		t.Fatalf("not caught up: %+v", snapshot)
	}
	if got := snapshot.Stats(); got != snapshot.Stats() {
		t.Fatalf("fold is not stable: %+v", got)
	}
	return snapshot.Stats()
}

func TestStatsFoldCountsTurnsAndBilledTokens(t *testing.T) {
	got := statsOf(t, headerAt("session-stats", 4)+
		row("user/message", 1, `{"content":[{"type":"text","text":"hi"}]}`)+
		row("assistant/message", 2, usageReadHeavy)+
		row("user/message", 3, `{"content":[{"type":"text","text":"again"}]}`)+
		row("assistant/message", 4, usageReadAgain)+
		row("turn/start", 5, `{"turn":2}`))

	// Billed input is what the request paid for: the uncached part, plus cache
	// reads and writes. The first message: 3884 + 6144 = 10028, and its total
	// agrees (10028 + 115 = 10143). The second: 273 + 9984 = 10257.
	if got.BilledInput != 10028+10257 {
		t.Fatalf("billedInput = %d, want 20285", got.BilledInput)
	}
	if got.Tokens != 10143+10398 {
		t.Fatalf("tokens = %d, want the two messages' totals", got.Tokens)
	}
	if got.CacheRead != 6144+9984 {
		t.Fatalf("cacheRead = %d, want 16128", got.CacheRead)
	}
	if got.LastPromptTokens != 10257 {
		t.Fatalf("lastPromptTokens = %d, want the last message's billed input", got.LastPromptTokens)
	}
	// Turns are user messages, not turn rows: a turn is something the user asked
	// for, and a resumed session has turn rows the user never typed.
	if got.Turns != 2 {
		t.Fatalf("turns = %d, want 2 user messages", got.Turns)
	}
	if got.Empty() {
		t.Fatal("a session with two messages is not empty")
	}
}

func TestStatsFoldAddsCacheWritesToBilledInput(t *testing.T) {
	got := statsOf(t, headerAt("session-stats", 4)+
		row("assistant/message", 1, usageWrite))

	if got.BilledInput != 150 {
		t.Fatalf("billedInput = %d, want 100 uncached + 50 written", got.BilledInput)
	}
	if got.Tokens != 170 {
		t.Fatalf("tokens = %d, want billed plus output", got.Tokens)
	}
	if got.CacheRead != 0 {
		t.Fatalf("cacheRead = %d, want 0", got.CacheRead)
	}
}

// Some logs nest the usage under the message. Missing it would halve a long
// session's footer, so both spellings have to fold the same way.
func TestStatsFoldReadsNestedUsage(t *testing.T) {
	got := statsOf(t, headerAt("session-stats", 4)+
		row("assistant/message", 1, `{"message":{"usage":{"inputTokens":3884,"outputTokens":115,"totalTokens":10143,"cacheReadTokens":6144}}}`))

	if got.Tokens != 10143 || got.BilledInput != 10028 || got.CacheRead != 6144 {
		t.Fatalf("nested usage folded to %+v", got)
	}
}

// The window a session runs in is not in the log: the header row carries
// identity and format, and the reader drops it before any consumer sees it, so
// a context length can only come from the provider's model description. Rows
// that are not messages must leave the tally alone either way.
func TestStatsFoldIgnoresRowsThatAreNotMessages(t *testing.T) {
	got := statsOf(t, headerAt("session-stats", 4)+
		row("turn/start", 1, `{"turn":1}`)+
		row("tool/request", 2, `{"header":{"config":{"contextLength":500000}}}`)+
		row("assistant/message", 3, usageReadHeavy))

	if got.Turns != 0 || got.Tokens != 10143 {
		t.Fatalf("stats = %+v, want only the message counted", got)
	}
}

func TestStatsFoldEmptyWithoutMessages(t *testing.T) {
	got := statsOf(t, headerAt("session-stats", 4))
	if !got.Empty() {
		t.Fatalf("stats = %+v, want empty", got)
	}
}
