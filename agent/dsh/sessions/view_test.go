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
