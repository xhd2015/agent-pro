package agentui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type TUITestQuestion struct {
	ID       string
	Question string
	Options  []QuestionOption
	Answer   string
}

type TUITestLLMCall struct {
	Prompt      string
	Model       string
	SessionID   string
	SessionDir  string
	AgentRunner string
}

type TUITestLLMResult struct {
	Output string
	Err    error
}

type TUITestDriverOptions struct {
	Width       int
	Feature     string
	Model       string
	AgentRunner string
	SessionID   string
	SessionDir  string
	Done        bool
	Questions   []TUITestQuestion
	OnLLMStart  func(TUITestLLMCall) TUITestLLMResult
}

type TUITestSnapshot struct {
	Done              bool
	ClarificationMode bool
	CtrlCPending      bool
	Logs              []string
	Questions         []TUITestQuestion
	SelectedQuestion  string
	Input             string
	LLMOutput         string
	Error             string
	View              string
}

type TUITestDriver struct {
	model model
}

func NewTUITestDriver(opts TUITestDriverOptions) *TUITestDriver {
	width := opts.Width
	if width <= 0 {
		width = 80
	}
	input := textinput.New()
	input.Placeholder = "Type answer and press Enter..."
	input.Focus()
	vp := viewport.New(width, 20)
	logs := []string{}
	if strings.TrimSpace(opts.Feature) != "" {
		logs = append(logs, "[You] "+opts.Feature)
		vp.SetContent(WrapText("[You] "+opts.Feature, width))
	}
	questions := make([]pendingQuestion, 0, len(opts.Questions))
	for _, q := range opts.Questions {
		questions = append(questions, pendingQuestion{
			ID:       q.ID,
			Question: q.Question,
			Options:  q.Options,
			Answer:   q.Answer,
		})
	}
	driver := &TUITestDriver{
		model: model{
			feature:     opts.Feature,
			llmModel:    opts.Model,
			agentRunner: opts.AgentRunner,
			sessionID:   opts.SessionID,
			sessionDir:  opts.SessionDir,
			done:        opts.Done,
			input:       input,
			viewport:    vp,
			width:       width,
			logs:        logs,
			questions:   questions,
			logCh:       make(chan string, 64),
			llmDoneCh:   make(chan llmDoneMsg, 1),
		},
	}
	if opts.OnLLMStart != nil {
		driver.model.llmStarter = func(prompt, llmModel, sessionID, sessionDir, agentRunner string, logCh chan<- string, doneCh chan<- llmDoneMsg) {
			result := opts.OnLLMStart(TUITestLLMCall{
				Prompt:      prompt,
				Model:       llmModel,
				SessionID:   sessionID,
				SessionDir:  sessionDir,
				AgentRunner: agentRunner,
			})
			doneCh <- llmDoneMsg{Output: result.Output, Err: result.Err}
		}
	}
	return driver
}

func (d *TUITestDriver) Send(msg tea.Msg) {
	updated, _ := d.model.Update(msg)
	d.model = updated.(model)
}

func (d *TUITestDriver) Press(key string) {
	d.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func (d *TUITestDriver) Enter() {
	d.Send(tea.KeyMsg{Type: tea.KeyEnter})
}

func (d *TUITestDriver) Tab() {
	d.Send(tea.KeyMsg{Type: tea.KeyTab})
}

func (d *TUITestDriver) TypeText(text string) {
	if q := d.model.currentPendingQuestion(); q != nil && len(q.Options) > 0 && d.model.optionHighlightIdx < len(q.Options) {
		d.model.optionHighlightIdx = len(q.Options)
	}
	for _, r := range text {
		d.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func (d *TUITestDriver) DeliverLLMDone(output string, err error) {
	d.Send(llmDoneMsg{Output: output, Err: err})
}

func (d *TUITestDriver) FlushLLM() bool {
	select {
	case msg := <-d.model.llmDoneCh:
		d.Send(msg)
		return true
	default:
		return false
	}
}

func (d *TUITestDriver) DeliverQuestion(q TUITestQuestion) {
	d.Send(questionMsg(pendingQuestion{
		ID:       q.ID,
		Question: q.Question,
		Options:  q.Options,
		Answer:   q.Answer,
	}))
}

func (d *TUITestDriver) Snapshot() TUITestSnapshot {
	questions := make([]TUITestQuestion, 0, len(d.model.questions))
	for _, q := range d.model.questions {
		questions = append(questions, TUITestQuestion{
			ID:       q.ID,
			Question: q.Question,
			Options:  q.Options,
			Answer:   q.Answer,
		})
	}
	selected := ""
	for _, q := range d.model.questions {
		if q.Answer == "" {
			selected = q.ID
			break
		}
	}
	errMsg := ""
	if d.model.err != nil {
		errMsg = d.model.err.Error()
	}
	return TUITestSnapshot{
		Done:              d.model.done,
		ClarificationMode: d.model.clarificationMode,
		CtrlCPending:      d.model.ctrlCPending,
		Logs:              append([]string(nil), d.model.logs...),
		Questions:         questions,
		SelectedQuestion:  selected,
		Input:             d.model.input.Value(),
		LLMOutput:         d.model.llmOutput,
		Error:             errMsg,
		View:              d.model.View(),
	}
}
