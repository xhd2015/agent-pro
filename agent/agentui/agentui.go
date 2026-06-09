package agentui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	add_pending_questions "github.com/xhd2015/agent-pro/agent/agentui/add_pending_questions"
	"github.com/xhd2015/agent-pro/agent/agentui/question"
	"github.com/xhd2015/agent-pro/agent/agentui/runner"
	"github.com/xhd2015/agent-pro/agent/agentui/sessionstate"
	"github.com/xhd2015/agent-pro/agent/agentui/textutil"

	"github.com/xhd2015/agent-pro/agent/session"
)

type Config struct {
	AgentName     string
	SessionPrefix string
	Prompt        string
	Usage         string
	Dispatch      map[string]func() error
	SkillName     string
	SkillContent  string
}

func Run(cfg Config, args []string) error {
	base := filepath.Base(os.Args[0])

	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	if base == "add-pending-questions" || (suffix != "" && base == "add-pending-questions"+suffix) {
		add_pending_questions.Run()
		return nil
	}
	if handler, ok := cfg.Dispatch[base]; ok {
		return handler()
	}
	if len(args) > 0 && args[0] == "skill" {
		return runSkillCommand(cfg, args[1:])
	}
	return runMain(cfg, args)
}

type QuestionOption = question.Option
type pendingQuestion = question.Question
type questionLineEntry = question.Entry

type logMsg string
type questionMsg pendingQuestion
type llmDoneMsg = runner.Done
type tickMsg struct{}
type ctrlCResetMsg struct{}

type model struct {
	feature   string
	llmModel  string
	tempDir   string
	answerDir string

	sessionID         string
	opencodeSessionID string
	sessionDir        string

	questionsFile      string
	logs               []string
	questions          []pendingQuestion
	selIdx             int
	optionHighlightIdx int
	ctrlCPending       bool
	input              textinput.Model
	viewport           viewport.Model
	done               bool
	clarificationMode  bool
	spinFrame          int
	llmOutput          string
	err                error
	width              int
	height             int

	logCh      chan string
	questionCh chan pendingQuestion
	llmDoneCh  chan llmDoneMsg
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForLogs(),
		m.listenForQuestions(),
		m.waitForLLMDone(),
		tick(),
		textinput.Blink,
	)
}

func (m model) waitForLLMDone() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.llmDoneCh
		return msg
	}
}

var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func SpinnerChar(frame int) string {
	return SpinnerChars[frame%len(SpinnerChars)]
}

const ctrlCResetTimeout = 3 * time.Second

func ctrlCResetTimer() tea.Cmd {
	return tea.Tick(ctrlCResetTimeout, func(t time.Time) tea.Msg {
		return ctrlCResetMsg{}
	})
}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) listenForLogs() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.logCh
		if !ok {
			return nil
		}
		return logMsg(msg)
	}
}

func (m model) listenForQuestions() tea.Cmd {
	return func() tea.Msg {
		q, ok := <-m.questionCh
		if !ok {
			return nil
		}
		return questionMsg(q)
	}
}

func (m *model) currentPendingQuestion() *pendingQuestion {
	idx := question.PendingIndex(m.questions, m.selIdx)
	if idx < 0 {
		return nil
	}
	m.selIdx = idx
	return &m.questions[m.selIdx]
}

func (m *model) hasUnansweredQuestions() bool {
	return question.HasUnanswered(m.questions)
}

func (m *model) answeredCount() int {
	return question.AnsweredCount(m.questions)
}

func (m *model) submitAnswer(answer string) {
	q := m.currentPendingQuestion()
	if q == nil {
		return
	}
	q.Answer = answer
	m.logs = append(m.logs, fmt.Sprintf("[You] Answered #%s: %s", q.ID, answer))

	entry := questionLineEntry{
		Type:   "answer",
		ID:     q.ID,
		Answer: answer,
	}
	data, _ := json.Marshal(entry)
	if m.questionsFile != "" {
		session.AppendLine(m.sessionDir, "questions.jsonl", string(data))
	}

	m.input.SetValue("")

	if !m.hasUnansweredQuestions() {
		m.clarificationMode = false
		m.optionHighlightIdx = 0
		m.logs = append(m.logs, "[Agent] All questions answered. Resuming...")

		prompt := m.buildResumePrompt("")
		newDoneCh := make(chan llmDoneMsg, 1)
		m.llmDoneCh = newDoneCh
		m.done = false
		go runLLM(prompt, m.llmModel, m.opencodeSessionID, m.sessionDir, m.logCh, newDoneCh)

		m.viewport.SetContent(textutil.WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
	} else {
		m.selIdx = (m.selIdx + 1) % len(m.questions)
		m.optionHighlightIdx = 0
		m.viewport.SetContent(textutil.WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
	}
}

func (m *model) buildResumePrompt(followUp string) string {
	return question.BuildResumePrompt(m.questions, followUp)
}

func (m *model) loadQuestionsFromFile() {
	lines, err := session.ReadLines(m.sessionDir, "questions.jsonl")
	if err != nil {
		return
	}

	questions := question.ReplayLines(lines)
	for _, q := range questions {
		m.questions = append(m.questions, pendingQuestion(q))
		m.logs = append(m.logs, question.FormatReplayLog(q))
	}
}

func runMain(cfg Config, args []string) error {
	opts, err := parseRunOptions(cfg, args)
	if err != nil {
		return err
	}

	resolved, err := resolveRunSession(cfg, opts)
	if err != nil {
		return err
	}

	paths, cleanup, err := prepareRuntimePaths(cfg, resolved.SessionDir)
	if err != nil {
		return err
	}
	defer cleanup()
	configureQuestionEnvironment(paths)

	channels := newRuntimeChannels()

	if !isTTY(os.Stdout) {
		runPlain(cfg, resolved.Feature, resolved.Model, opts.OutputFile)
		return nil
	}

	go startQuestionMonitor(paths.QuestionFIFO, channels.QuestionCh)

	m := newRuntimeModel(cfg, opts, resolved, paths, channels)

	if !m.clarificationMode {
		prompt := cfg.Prompt + "\n\n## Feature Description\n" + resolved.Feature
		if opts.OutputFile != "" {
			prompt += "\n\n## Output File\nWrite the complete report to: " + opts.OutputFile
		}
		go runLLM(prompt, resolved.Model, resolved.OpencodeSessionID, resolved.SessionDir, channels.LogCh, channels.DoneCh)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Session %s finished.\nTo resume: %s --resume %s\n", resolved.SessionID, filepath.Base(os.Args[0]), resolved.SessionID)
	return nil
}

func WrapText(s string, width int) string {
	return textutil.WrapText(s, width)
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func startQuestionMonitor(fifo string, ch chan<- pendingQuestion) {
	for {
		f, err := os.Open(fifo)
		if err != nil {
			return
		}
		question.ReadQuestions(f, ch)
		f.Close()
	}
}

func runLLM(prompt, llmModel, sessionID, sessionDir string, logCh chan<- string, doneCh chan<- llmDoneMsg) {
	runner.RunLLM(prompt, llmModel, sessionID, sessionDir, logCh, doneCh)
}

func formatLogLine(line string) string {
	return runner.FormatLogLine(line)
}

func indexByte(data []byte, b byte) int {
	return textutil.IndexByte(data, b)
}

type sessionMeta = sessionstate.Meta

func readSessionFromDir(dir, sessionID string) (string, string, string, string, []string) {
	return sessionstate.ReadFromDir(dir, sessionID)
}

func resolveSession(agentName, resumeID string) (sessionID, opencodeSessionID, sessionDir, feature, llmModel string, logs []string, err error) {
	return sessionstate.Resolve(agentName, resumeID)
}

func newSessionID(prefix string) string {
	return sessionstate.NewID(prefix)
}

func runPlain(cfg Config, feature, llmModel, outputFile string) {
	prompt := cfg.Prompt + "\n\n## Feature Description\n" + feature

	if outputFile != "" {
		prompt += "\n\n## Output File\nWrite the complete report to: " + outputFile
	}
	runner.RunPlain(prompt, llmModel)
}
