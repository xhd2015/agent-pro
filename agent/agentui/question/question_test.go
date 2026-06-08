package question

import (
	"strings"
	"testing"
)

func TestReplayLinesEmpty(t *testing.T) {
	if got := ReplayLines(nil); len(got) != 0 {
		t.Fatalf("ReplayLines(nil) returned %d questions, want 0", len(got))
	}
}

func TestReplayLinesQuestionThenAnswer(t *testing.T) {
	got := ReplayLines([]string{
		`{"type":"question","id":"q1","question":"What platform?","options":[{"label":"Web"}]}`,
		`{"type":"answer","id":"q1","answer":"Web"}`,
	})
	if len(got) != 1 {
		t.Fatalf("got %d questions, want 1", len(got))
	}
	if got[0].ID != "q1" || got[0].Question != "What platform?" || got[0].Answer != "Web" {
		t.Fatalf("unexpected question replay: %+v", got[0])
	}
	if len(got[0].Options) != 1 || got[0].Options[0].Label != "Web" {
		t.Fatalf("options not replayed: %+v", got[0].Options)
	}
}

func TestReplayLinesPreservesOrderAndUpdatesDuplicateQuestion(t *testing.T) {
	got := ReplayLines([]string{
		`{"type":"question","id":"q2","question":"Q2"}`,
		`{"type":"question","id":"q1","question":"old"}`,
		`{"type":"question","id":"q1","question":"new"}`,
		`{"type":"answer","id":"q1","answer":"A1"}`,
	})
	if len(got) != 2 {
		t.Fatalf("got %d questions, want 2", len(got))
	}
	if got[0].ID != "q2" || got[1].ID != "q1" {
		t.Fatalf("order = %s,%s; want q2,q1", got[0].ID, got[1].ID)
	}
	if got[1].Question != "new" || got[1].Answer != "A1" {
		t.Fatalf("duplicate question was not updated with answer: %+v", got[1])
	}
}

func TestReplayLinesIgnoresInvalidJSONAndUnknownAnswer(t *testing.T) {
	got := ReplayLines([]string{
		`not-json`,
		`{"type":"answer","id":"missing","answer":"A"}`,
		`{"type":"question","id":"q1","question":"Q1"}`,
	})
	if len(got) != 1 || got[0].ID != "q1" || got[0].Answer != "" {
		t.Fatalf("unexpected replay result: %+v", got)
	}
}

func TestQuestionCountsAndPendingIndex(t *testing.T) {
	questions := []Question{
		{ID: "1", Answer: "A1"},
		{ID: "2"},
		{ID: "3", Answer: "A3"},
	}
	if !HasUnanswered(questions) {
		t.Fatal("expected unanswered question")
	}
	if got := AnsweredCount(questions); got != 2 {
		t.Fatalf("AnsweredCount = %d, want 2", got)
	}
	if got := PendingIndex(questions, 0); got != 1 {
		t.Fatalf("PendingIndex = %d, want 1", got)
	}
	if got := PendingIndex([]Question{{ID: "1", Answer: "A1"}}, 0); got != -1 {
		t.Fatalf("PendingIndex all answered = %d, want -1", got)
	}
	if got := PendingIndex(nil, 0); got != -1 {
		t.Fatalf("PendingIndex empty = %d, want -1", got)
	}
}

func TestBuildResumePrompt(t *testing.T) {
	prompt := BuildResumePrompt([]Question{
		{ID: "1", Question: "Payment methods?", Options: []Option{{Label: "Card"}, {Label: "PayPal"}}, Answer: "Card"},
		{ID: "2", Question: "Max amount?", Answer: "5000"},
		{ID: "3", Question: "Unanswered?"},
	}, "follow up text")
	for _, want := range []string{"Payment methods?", "Card", "5000", "Options: Card, PayPal", "## Follow-up", "follow up text"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
	if strings.Contains(prompt, "Unanswered?") {
		t.Fatalf("prompt should not include unanswered question: %s", prompt)
	}
	if strings.Contains(BuildResumePrompt([]Question{{Question: "Q1?", Answer: "A1"}}, ""), "## Follow-up") {
		t.Fatal("empty follow-up should not render follow-up section")
	}
}

func TestFormatLogs(t *testing.T) {
	q := Question{ID: "1", Question: "Payment?", Options: []Option{{Label: "Card"}, {Label: "PayPal"}}, Answer: "Card"}
	if got := FormatAskLog(q); got != "[Agent asks] #1: Payment? [Card/PayPal]" {
		t.Fatalf("FormatAskLog = %q", got)
	}
	if got := FormatReplayLog(q); got != "[Agent asks] #1: Payment? [Card/PayPal] (answered: Card)" {
		t.Fatalf("FormatReplayLog = %q", got)
	}
}

func TestReadQuestions(t *testing.T) {
	ch := make(chan Question, 4)
	ReadQuestions(strings.NewReader(strings.Join([]string{
		`{"type":"question","id":"q1","question":"Q1","options":[{"label":"A"}]}`,
		`{"type":"answer","id":"q1","answer":"A"}`,
		`not-json`,
	}, "\n")), ch)
	close(ch)

	var got []Question
	for q := range ch {
		got = append(got, q)
	}
	if len(got) != 1 || got[0].ID != "q1" || len(got[0].Options) != 1 {
		t.Fatalf("unexpected questions read: %+v", got)
	}
}
