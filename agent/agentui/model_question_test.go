package agentui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

func TestHasUnansweredQuestions(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
		},
	}
	if !m.hasUnansweredQuestions() {
		t.Error("expected hasUnansweredQuestions=true when Q2 has no answer")
	}

	m.questions[1].Answer = "A2"
	if m.hasUnansweredQuestions() {
		t.Error("expected hasUnansweredQuestions=false when all answered")
	}
}

func TestAnsweredCount(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
			{ID: "3", Question: "Q3", Answer: "A3"},
		},
	}
	if n := m.answeredCount(); n != 2 {
		t.Errorf("expected answeredCount=2, got %d", n)
	}
}

func TestCurrentPendingQuestionSkipsAnswered(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
			{ID: "3", Question: "Q3", Answer: "A3"},
		},
		selIdx: 0,
	}
	q := m.currentPendingQuestion()
	if q == nil {
		t.Fatal("expected a pending question")
	}
	if q.ID != "2" {
		t.Errorf("expected Q2 (unanswered), got %s", q.ID)
	}
}

func TestCurrentPendingQuestionAllAnswered(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
		},
		selIdx: 0,
	}
	q := m.currentPendingQuestion()
	if q != nil {
		t.Errorf("expected nil when all answered, got %v", q)
	}
}

func TestCurrentPendingQuestionEmpty(t *testing.T) {
	m := &model{selIdx: 0}
	q := m.currentPendingQuestion()
	if q != nil {
		t.Errorf("expected nil for empty questions, got %v", q)
	}
}

func TestSubmitAnswerRecordsAnswer(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
		selIdx: 0,
		logs:   []string{},
	}
	m.submitAnswer("my answer")
	if m.questions[0].Answer != "my answer" {
		t.Errorf("expected answer 'my answer', got %q", m.questions[0].Answer)
	}
	found := false
	for _, log := range m.logs {
		if strings.Contains(log, "[You] Answered #1: my answer") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected answer log entry")
	}
}

func TestSubmitAnswerClearsInput(t *testing.T) {
	input := textinput.New()
	input.SetValue("typed answer")
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
		selIdx: 0,
		logs:   []string{},
		input:  input,
	}
	m.submitAnswer("typed answer")
	if m.input.Value() != "" {
		t.Errorf("expected input cleared after submit, got %q", m.input.Value())
	}
}

func TestSubmitAnswerAllAnsweredTransitionsOutOfClarification(t *testing.T) {
	logCh := make(chan string, 64)
	doneCh := make(chan llmDoneMsg, 1)
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
			{ID: "2", Question: "Q2"},
		},
		selIdx:     1,
		llmModel:   "test-model",
		sessionID:  "test-session",
		sessionDir: t.TempDir(),
		logCh:      logCh,
		llmDoneCh:  doneCh,
		width:      80,
		logs:       []string{},
		// Avoid defaultLLMStarter: it goroutines RunLLM into sessionDir and
		// races t.TempDir cleanup ("directory not empty") in CI containers.
		llmStarter: func(string, string, string, string, string, chan<- string, chan<- llmDoneMsg) {},
	}
	m.submitAnswer("A2")
	if m.clarificationMode {
		t.Error("expected clarificationMode=false after all answered")
	}
	if m.done {
		t.Error("expected done=false after all answered (resumes LLM)")
	}
}

func TestSubmitAnswerStillHasUnanswered(t *testing.T) {
	vp := viewport.New(80, 20)
	m := &model{
		done:              true,
		clarificationMode: true,
		viewport:          vp,
		input:             textinput.New(),
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
			{ID: "2", Question: "Q2"},
		},
		selIdx: 0,
		width:  80,
		logs:   []string{},
	}
	m.submitAnswer("A1")
	if !m.clarificationMode {
		t.Error("expected clarificationMode to stay true when more unanswered")
	}
	if m.selIdx != 1 {
		t.Errorf("expected selIdx=1 after answering Q1, got %d", m.selIdx)
	}
}

func TestBuildResumePrompt(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Payment methods?", Options: []QuestionOption{{Label: "Card"}, {Label: "PayPal"}}, Answer: "Card"},
			{ID: "2", Question: "Max amount?", Answer: "5000"},
			{ID: "3", Question: "Unanswered?", Answer: ""},
		},
	}
	prompt := m.buildResumePrompt("follow up text")
	if !strings.Contains(prompt, "Payment methods?") {
		t.Error("prompt should contain answered question")
	}
	if !strings.Contains(prompt, "Card") {
		t.Error("prompt should contain answer 'Card'")
	}
	if !strings.Contains(prompt, "5000") {
		t.Error("prompt should contain answer '5000'")
	}
	if !strings.Contains(prompt, "Options: Card, PayPal") {
		t.Error("prompt should list options")
	}
	if strings.Contains(prompt, "Unanswered?") {
		t.Error("prompt should NOT contain unanswered question")
	}
	if !strings.Contains(prompt, "follow up text") {
		t.Error("prompt should contain follow-up text")
	}
	if !strings.Contains(prompt, "## Follow-up") {
		t.Error("prompt should have ## Follow-up section when followUp is provided")
	}
}

func TestBuildResumePromptNoFollowUp(t *testing.T) {
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1?", Answer: "A1"},
		},
	}
	prompt := m.buildResumePrompt("")
	if strings.Contains(prompt, "## Follow-up") {
		t.Error("prompt should NOT have ## Follow-up when followUp is empty")
	}
}

func TestLlmdoneSetsClarificationMode(t *testing.T) {
	logCh := make(chan string, 64)
	vp := viewport.New(80, 20)
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1"},
		},
		viewport: vp,
		logCh:    logCh,
		logs:     []string{},
		width:    80,
	}
	updated, _ := m.Update(llmDoneMsg{Output: "test output"})
	um := updated.(model)
	if !um.clarificationMode {
		t.Error("expected clarificationMode=true when pending questions exist")
	}
	if !um.done {
		t.Error("expected done=true after llmDoneMsg")
	}
}

func TestLlmdoneNoQuestionsGoesToDone(t *testing.T) {
	logCh := make(chan string, 64)
	vp := viewport.New(80, 20)
	m := &model{
		viewport: vp,
		logCh:    logCh,
		logs:     []string{},
		width:    80,
	}
	updated, _ := m.Update(llmDoneMsg{Output: "test output"})
	um := updated.(model)
	if um.clarificationMode {
		t.Error("expected clarificationMode=false when no pending questions")
	}
	if !um.done {
		t.Error("expected done=true after llmDoneMsg")
	}
}

func TestLlmdoneAllAnsweredGoesToDone(t *testing.T) {
	logCh := make(chan string, 64)
	vp := viewport.New(80, 20)
	m := &model{
		questions: []pendingQuestion{
			{ID: "1", Question: "Q1", Answer: "A1"},
		},
		viewport: vp,
		logCh:    logCh,
		logs:     []string{},
		width:    80,
	}
	updated, _ := m.Update(llmDoneMsg{Output: "test output"})
	um := updated.(model)
	if um.clarificationMode {
		t.Error("expected clarificationMode=false when all answered")
	}
	if !um.done {
		t.Error("expected done=true after llmDoneMsg")
	}
}
