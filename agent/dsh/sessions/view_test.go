package sessions

import (
	"encoding/json"
	"os"
	"testing"
)

func kindsOf(messages []Message) []string {
	out := make([]string, len(messages))
	for i, message := range messages {
		out[i] = message.Kind
	}
	return out
}

func TestProjectThinkingKeepsLogOrder(t *testing.T) {
	events := []Event{
		{Type: "user/message", Seq: 1, Time: 1, Data: json.RawMessage(`{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)},
		{Type: "assistant/message", Seq: 2, Time: 2, Data: json.RawMessage(`{"message":{"content":[{"type":"reasoning","text":"plan A"},{"type":"text","text":"calling tools"}]}}`)},
		{Type: "tool/call", Seq: 3, Time: 3, Data: json.RawMessage(`{"name":"read","arguments":{"path":"a.go"}}`)},
		{Type: "assistant/attempt", Seq: 4, Time: 4, Data: json.RawMessage(`{"message":{"content":[{"type":"reasoning","text":"partial"}]}}`)},
		{Type: "assistant/message", Seq: 5, Time: 5, Data: json.RawMessage(`{"message":{"content":[{"type":"text","text":"first"},{"type":"reasoning","text":"later thought"},{"type":"text","text":"done"}]}}`)},
		{Type: "assistant/message", Seq: 6, Time: 6, Data: json.RawMessage(`{"message":{"content":[{"type":"reasoning","text":"only thinking"}]}}`)},
	}
	got := Project(events)
	want := []string{"user", "thinking", "assistant", "tool", "assistant", "thinking", "assistant", "thinking"}
	if len(got) != len(want) {
		t.Fatalf("rows=%d kinds=%v want %v", len(got), kindsOf(got), want)
	}
	for i, kind := range want {
		if got[i].Kind != kind {
			t.Fatalf("kind[%d]=%q want %q (%v)", i, got[i].Kind, kind, kindsOf(got))
		}
	}
}

func TestProjectGrokThinkingBlocks(t *testing.T) {
	events := []Event{
		{Type: "user/message", Seq: 1, Time: 1, Data: json.RawMessage(`{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)},
		{Type: "assistant/message", Seq: 2, Time: 2, Data: json.RawMessage(`{"message":{"content":[{"type":"thinking","text":"grok thought"},{"type":"text","text":"ok"}]}}`)},
		{Type: "assistant/message", Seq: 3, Time: 3, Data: json.RawMessage(`{"message":{"content":[{"type":"reasoning","reasoning":"field only"},{"type":"text","text":"done"}]}}`)},
		{Type: "assistant/message", Seq: 4, Time: 4, Data: json.RawMessage(`{"message":{"content":[{"type":"thinking","text":"   "},{"type":"text","text":"skip empty"}]}}`)},
		{Type: "assistant/attempt", Seq: 5, Time: 5, Data: json.RawMessage(`{"message":{"content":[{"type":"thinking","text":"partial"}]}}`)},
	}
	got := Project(events)
	want := []string{"user", "thinking", "assistant", "thinking", "assistant", "assistant"}
	if len(got) != len(want) {
		t.Fatalf("rows=%d kinds=%v want %v", len(got), kindsOf(got), want)
	}
	for i, kind := range want {
		if got[i].Kind != kind {
			t.Fatalf("kind[%d]=%q want %q (%v)", i, got[i].Kind, kind, kindsOf(got))
		}
	}
	if got[1].Text != "grok thought" || got[3].Text != "field only" || got[5].Text != "skip empty" {
		t.Fatalf("texts: %+v", got)
	}
}

func TestProjectAskExtractsQuestionsAndAnswers(t *testing.T) {
	events := []Event{
		{Type: "assistant/message", Seq: 1, Time: 1, Data: json.RawMessage(`{"message":{"content":[{"type":"text","text":"confirm:"}]}}`)},
		{Type: "tool/call", Seq: 2, Time: 2, Data: json.RawMessage(`{"callId":"c1","name":"ask_user_question","arguments":{"questions":[{"id":"kind","question":"哪一种？","options":[{"label":"金色标记","description":"徽章"},{"label":"文案"}],"multi_select":false}]}}`)},
		{Type: "tool/result", Seq: 3, Time: 3, Data: json.RawMessage(`{"callId":"c1","name":"ask_user_question","content":[{"type":"text","text":"{\"answers\":[{\"id\":\"kind\",\"selected\":[\"金色标记\"]}]}"}]}`)},
	}
	got := Project(events)
	var ask *Ask
	for i := range got {
		if got[i].Ask != nil {
			ask = got[i].Ask
			break
		}
	}
	if ask == nil || ask.CallID != "c1" || len(ask.Questions) != 1 || ask.Questions[0].Question != "哪一种？" {
		t.Fatalf("ask: %+v rows=%+v", ask, got)
	}
	if len(ask.Answers) != 1 || len(ask.Answers[0].Selected) != 1 || ask.Answers[0].Selected[0] != "金色标记" {
		t.Fatalf("answers: %+v", ask.Answers)
	}
}

func TestProjectAskExtractsNestedV4ToolResult(t *testing.T) {
	// Real DSH v4 tool/result nests callId and content under data.message;
	// the flattened fixture in TestProjectAskExtractsQuestionsAndAnswers is
	// not what the log writes.
	events := []Event{
		{Type: "tool/call", Seq: 104423, Time: 1, Data: json.RawMessage(`{"callId":"call_00_lkmfd8t9awzd5f1hc685e8gm","name":"ask_user_question","arguments":{"questions":[{"id":"resume-thread","question":"请确认要 resume 哪条：","options":[{"label":"开始实现 Task 4"}]}]}}`)},
		{Type: "tool/result", Seq: 104426, Time: 2, Data: json.RawMessage(`{"turn":329,"step":10,"message":{"role":"tool","source":{"kind":"tool","callId":"call_00_lkmfd8t9awzd5f1hc685e8gm"},"toolCallId":"call_00_lkmfd8t9awzd5f1hc685e8gm","content":[{"type":"text","text":"{\"answers\":[{\"id\":\"resume-thread\",\"selected\":[],\"custom\":\"更新你的mockup: spec里面应该是tree\"}]}"}]}}`)},
	}
	got := Project(events)
	var ask *Ask
	for i := range got {
		if got[i].Ask != nil {
			ask = got[i].Ask
			break
		}
	}
	if ask == nil || ask.CallID != "call_00_lkmfd8t9awzd5f1hc685e8gm" {
		t.Fatalf("ask: %+v rows=%+v", ask, got)
	}
	if len(ask.Answers) != 1 || ask.Answers[0].Custom != "更新你的mockup: spec里面应该是tree" || len(ask.Answers[0].Selected) != 0 {
		t.Fatalf("answers: %+v", ask.Answers)
	}
	for _, message := range got {
		if message.Kind == "tool" && message.Ask == nil && message.Tool == "ask_user_question" {
			t.Fatalf("result leaked as tool row: %+v", message)
		}
	}
}

func TestProjectTurnEndError(t *testing.T) {
	event := Event{Type: "turn/end", Seq: 9, Data: json.RawMessage(`{"turn":1,"reason":{"kind":"error","error":{"message":"502 status code (no body)","code":"SERVER"}}}`)}
	rows := Project([]Event{event})
	if len(rows) != 1 || rows[0].Kind != "error" || rows[0].Text != "502 status code (no body)" || rows[0].Sum != "SERVER" {
		t.Fatalf("%+v", rows)
	}
}

func TestViewCachesProjection(t *testing.T) {
	resetInspectBodies()
	root := t.TempDir()
	id := "s-view"
	path := put(t, root, id, "session.v3.jsonl", header(id)+
		row("user/message", 0, `{"source":{"kind":"user"},"content":[{"type":"text","text":"hello"}]}`)+
		row("tool/call", 1, `{"name":"present","arguments":{"files":[{"path":"a.md"}]}}`)+
		row("request/header", 2, `{"header":{"config":{"provider":"xai","model":"grok","reasoningEffort":"high"}}}`))
	store := Store{Root: root}
	first, err := store.View(id)
	if err != nil {
		t.Fatal(err)
	}
	if first.ReasoningEffort != "high" || len(first.Rows) != 2 || first.Rows[1].Present == nil {
		t.Fatalf("view: %+v", first)
	}
	second, err := store.View(id)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("unchanged View should return the cached pointer")
	}
	if got := InspectReads.Load(); got != 1 {
		t.Fatalf("reads=%d", got)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(row("user/message", 3, `{"source":{"kind":"user"},"content":[{"type":"text","text":"again"}]}`)); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	third, err := store.View(id)
	if err != nil {
		t.Fatal(err)
	}
	if third == first || len(third.Rows) != 3 {
		t.Fatalf("grown view rows=%d ptr-equal=%v", len(third.Rows), third == first)
	}
}

// A /compact keeps its pending row and gains a settled one at the done event's
// seq: a reader polling afterSeq has already seen the run's seq, so a settled
// row that kept it would never be delivered. The two rows share a Ref so the
// reader replaces the first with the second instead of stacking both.
func TestProjectCompactRunAndDoneShareRef(t *testing.T) {
	events := []Event{
		{Type: "user/message", Seq: 1, Time: 1, Data: json.RawMessage(`{"source":{"kind":"user"},"content":[{"type":"text","text":"hi"}]}`)},
		{Type: "command/run", Seq: 2, Time: 2, Data: json.RawMessage(`{"commandId":"c1","name":"compact","source":{"kind":"user"}}`)},
		{Type: "command/done", Seq: 3, Time: 3, Data: json.RawMessage(`{"commandId":"c1","kind":"success","text":"Compacted 42 history items (~12k tokens)."}`)},
	}
	got := Project(events)
	if len(got) != 3 || got[0].Kind != "user" || got[1].Kind != TranscriptCommand || got[2].Kind != TranscriptCommand {
		t.Fatalf("kinds=%v", kindsOf(got))
	}
	pending, settled := got[1], got[2]
	if pending.Seq != 2 || pending.Tool != "compact" || pending.Ref != "c1" || pending.Text != "" || pending.Sum != "" {
		t.Fatalf("pending=%+v", pending)
	}
	if settled.Seq != 3 || settled.Tool != "compact" || settled.Ref != "c1" {
		t.Fatalf("settled=%+v", settled)
	}
	if settled.Sum != "success" || settled.Text != "Compacted 42 history items (~12k tokens)." {
		t.Fatalf("settled text=%+v", settled)
	}
}

func TestProjectCompactRunAloneIsPending(t *testing.T) {
	got := Project([]Event{
		{Type: "command/run", Seq: 8, Time: 8, Data: json.RawMessage(`{"commandId":"c2","name":"compact","source":{"kind":"user"}}`)},
	})
	if len(got) != 1 || got[0].Kind != TranscriptCommand || got[0].Tool != "compact" || got[0].Text != "" || got[0].Sum != "" || got[0].Ref != "c2" {
		t.Fatalf("pending=%+v", got)
	}
}

func TestProjectCompactDoneError(t *testing.T) {
	got := Project([]Event{
		{Type: "command/run", Seq: 1, Time: 1, Data: json.RawMessage(`{"commandId":"c3","name":"compact"}`)},
		{Type: "command/done", Seq: 2, Time: 2, Data: json.RawMessage(`{"commandId":"c3","kind":"error","text":"No compactable history yet."}`)},
	})
	if len(got) != 2 {
		t.Fatalf("kinds=%v", kindsOf(got))
	}
	settled := got[1]
	if settled.Sum != "error" || settled.Text != "No compactable history yet." || settled.Seq != 2 || settled.Tool != "compact" {
		t.Fatalf("error=%+v", settled)
	}
}

// The Messages tab reads the cached cursor, not Project: this is the path the
// compaction rows must actually reach.
func TestViewTailCompactionRows(t *testing.T) {
	cursor := &viewTailCursor{}
	feed := func(typ string, seq int64, data string) {
		cursor.applyEvent(Event{Type: typ, Seq: seq, Time: seq, Data: json.RawMessage(data)})
	}

	// A command other than /compact has no transcript row: the Messages tab
	// shows history folds, not every slash command.
	feed("command/run", 1, `{"commandId":"g1","name":"goal","source":{"kind":"user"}}`)
	feed("command/done", 2, `{"commandId":"g1","kind":"success","text":"Goal set."}`)
	if len(cursor.rows) != 0 {
		t.Fatalf("non-compact command rows=%+v", cursor.rows)
	}

	// A manual /compact shows a pending card and then a settled one sharing its
	// ref, so a live reader replaces the first instead of stacking both.
	feed("command/run", 3, `{"commandId":"c1","name":"compact","source":{"kind":"user"}}`)
	if len(cursor.rows) != 1 || cursor.rows[0].Sum != "" || cursor.rows[0].Ref != "c1" {
		t.Fatalf("pending=%+v", cursor.rows)
	}
	feed("command/done", 4, `{"commandId":"c1","kind":"success","text":"Compacted 42 history items (~12k tokens)."}`)
	if len(cursor.rows) != 2 {
		t.Fatalf("rows=%+v", cursor.rows)
	}
	settled := cursor.rows[1]
	if settled.Seq != 4 || settled.Sum != "success" || settled.Ref != "c1" || settled.Tool != "compact" {
		t.Fatalf("settled=%+v", settled)
	}
	if settled.Text != "Compacted 42 history items (~12k tokens)." {
		t.Fatalf("settled text=%q", settled.Text)
	}

	// An automatic compaction surfaces once, at its checkpoint, priced by the
	// summary that preceded it.
	feed("compaction/start", 5, `{"compactionId":"k1","turn":2}`)
	feed("compaction/summary", 6, `{"compactionId":"k1","shadowedSeqs":[1,2,3],"shadowedTokenCount":4242}`)
	if len(cursor.rows) != 2 {
		t.Fatalf("summary row=%+v", cursor.rows)
	}
	feed("user/message", 7, `{"source":{"kind":"compact-checkpoint","compactionId":"k1"},"content":[]}`)
	feed("compaction/end", 8, `{"compactionId":"k1","turn":2}`)
	if len(cursor.rows) != 3 {
		t.Fatalf("rows=%+v", cursor.rows)
	}
	auto := cursor.rows[2]
	if auto.Seq != 7 || auto.Kind != TranscriptCommand || auto.Sum != "success" || auto.Ref != "k1" || auto.Tool != "" {
		t.Fatalf("auto=%+v", auto)
	}
	if auto.Text != "3 history items folded · ~4242 tokens" {
		t.Fatalf("auto text=%q", auto.Text)
	}
}

// An automatic compaction has no command and no summary text on screen: the
// checkpoint is its only visible surface, and the metering comes from the
// summary event that preceded it.
func TestProjectAutoCompactionCheckpoint(t *testing.T) {
	got := Project([]Event{
		{Type: "compaction/start", Seq: 1, Time: 1, Data: json.RawMessage(`{"compactionId":"k1","turn":1}`)},
		{Type: "compaction/summary", Seq: 2, Time: 2, Data: json.RawMessage(`{"compactionId":"k1","shadowedSeqs":[1,2,3],"shadowedTokenCount":4242,"summary":[{"type":"text","text":"model facing"}]}`)},
		{Type: "user/message", Seq: 3, Time: 3, Data: json.RawMessage(`{"source":{"kind":"compact-checkpoint","compactionId":"k1"},"content":[{"type":"text","text":"model facing"}]}`)},
		{Type: "compaction/end", Seq: 4, Time: 4, Data: json.RawMessage(`{"compactionId":"k1","turn":1}`)},
	})
	if len(got) != 1 {
		t.Fatalf("rows=%+v", got)
	}
	row := got[0]
	if row.Kind != TranscriptCommand || row.Seq != 3 || row.Sum != "success" || row.Ref != "k1" {
		t.Fatalf("row=%+v", row)
	}
	if row.Text != "3 history items folded · ~4242 tokens" {
		t.Fatalf("text=%q", row.Text)
	}
}

// A manual /compact appends the same kind of checkpoint. The command row speaks
// for it, so the checkpoint must not add a second card.
func TestProjectManualCheckpointStaysHidden(t *testing.T) {
	got := Project([]Event{
		{Type: "command/run", Seq: 1, Time: 1, Data: json.RawMessage(`{"commandId":"c4","name":"compact"}`)},
		{Type: "compaction/start", Seq: 2, Time: 2, Data: json.RawMessage(`{"compactionId":"k2","sourceCommandId":"c4"}`)},
		{Type: "compaction/summary", Seq: 3, Time: 3, Data: json.RawMessage(`{"compactionId":"k2","shadowedSeqs":[1,2],"shadowedTokenCount":900}`)},
		{Type: "user/message", Seq: 4, Time: 4, Data: json.RawMessage(`{"source":{"kind":"compact-checkpoint","compactionId":"k2","sourceCommandId":"c4"},"content":[]}`)},
		{Type: "command/done", Seq: 5, Time: 5, Data: json.RawMessage(`{"commandId":"c4","kind":"success","text":"Compacted 2 history items (~900 tokens)."}`)},
	})
	if len(got) != 2 {
		t.Fatalf("rows=%+v", got)
	}
	if got[1].Seq != 5 || got[1].Text != "Compacted 2 history items (~900 tokens)." || got[1].Ref != "c4" {
		t.Fatalf("settled=%+v", got[1])
	}
}

// A window that opens after the summary keeps the checkpoint but cannot price
// it. The row stays text-free rather than inventing a count.
func TestProjectCheckpointWithoutSummaryKeepsQuiet(t *testing.T) {
	got := Project([]Event{
		{Type: "user/message", Seq: 9, Time: 9, Data: json.RawMessage(`{"source":{"kind":"compact-checkpoint","compactionId":"k3"},"content":[]}`)},
	})
	if len(got) != 1 || got[0].Text != "" || got[0].Sum != "success" || got[0].Ref != "k3" {
		t.Fatalf("rows=%+v", got)
	}
}

// compaction/prune has no surface of its own: its replacement is an ordinary
// tool/result that already projects.
func TestProjectCompactionPruneIsInvisible(t *testing.T) {
	got := Project([]Event{
		{Type: "compaction/prune", Seq: 1, Time: 1, Data: json.RawMessage(`{"shadowedSeqs":[1,2],"shadowedTokenCount":10}`)},
	})
	if len(got) != 0 {
		t.Fatalf("rows=%+v", got)
	}
}
