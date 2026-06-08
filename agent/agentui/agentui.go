package agentui

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	add_pending_questions "github.com/xhd2015/agent-pro/agent/agentui/add_pending_questions"
	"github.com/xhd2015/agent-pro/agent/opencode/models"
	opencoderun "github.com/xhd2015/agent-pro/agent/opencode/run"
	lessflags "github.com/xhd2015/less-flags"

	"github.com/xhd2015/agent-pro/agent/session"
	"github.com/xhd2015/agent-pro/agent_trace/types"
	tracefmt "github.com/xhd2015/agent-pro/run"
)

type Config struct {
	AgentName     string
	SessionPrefix string
	Prompt        string
	Usage         string
	Dispatch      map[string]func() error
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
	return runMain(cfg, args)
}

type QuestionOption struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type pendingQuestion struct {
	ID       string           `json:"id"`
	Question string           `json:"question"`
	Options  []QuestionOption `json:"options"`
	Answer   string           `json:"answer,omitempty"`
}

type questionLineEntry struct {
	Type     string           `json:"type"`
	ID       string           `json:"id"`
	Question string           `json:"question,omitempty"`
	Options  []QuestionOption `json:"options,omitempty"`
	Answer   string           `json:"answer,omitempty"`
}

type logMsg string
type questionMsg pendingQuestion
type llmDoneMsg struct {
	output            string
	opencodeSessionID string
	err               error
}
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5

	case tea.KeyMsg:
		if msg.String() == "enter" && strings.TrimSpace(m.input.Value()) == "/exit" {
			return m, tea.Quit
		}

		if msg.Type == tea.KeyCtrlC {
			if m.ctrlCPending {
				return m, tea.Quit
			}
			m.ctrlCPending = true
			return m, ctrlCResetTimer()
		}

		if m.ctrlCPending {
			m.ctrlCPending = false
		}

		if m.clarificationMode {
			switch msg.String() {
			case "enter":
				answer := strings.TrimSpace(m.input.Value())
				q := m.currentPendingQuestion()
				if q != nil && len(q.Options) > 0 && m.optionHighlightIdx == len(q.Options) {
					if answer != "" {
						m.submitAnswer(answer)
					}
					return m, nil
				}
				if answer != "" {
					m.submitAnswer(answer)
					return m, nil
				}
				if q != nil && len(q.Options) > 0 && m.optionHighlightIdx < len(q.Options) {
					m.submitAnswer(q.Options[m.optionHighlightIdx].Label)
					return m, nil
				}
			case "tab":
				if len(m.questions) > 0 {
					m.selIdx = (m.selIdx + 1) % len(m.questions)
					m.input.SetValue("")
					m.optionHighlightIdx = 0
				}
			case "shift+tab":
				if len(m.questions) > 0 {
					m.selIdx = (m.selIdx - 1 + len(m.questions)) % len(m.questions)
					m.input.SetValue("")
					m.optionHighlightIdx = 0
				}
			case "up":
				q := m.currentPendingQuestion()
				if q != nil && len(q.Options) > 0 {
					m.optionHighlightIdx = (m.optionHighlightIdx - 1 + len(q.Options) + 1) % (len(q.Options) + 1)
				}
			case "down":
				q := m.currentPendingQuestion()
				if q != nil && len(q.Options) > 0 {
					m.optionHighlightIdx = (m.optionHighlightIdx + 1) % (len(q.Options) + 1)
				}
			default:
				q := m.currentPendingQuestion()
				if q != nil && len(q.Options) > 0 && m.optionHighlightIdx < len(q.Options) {
					return m, nil
				}
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		if m.done {
			switch msg.String() {
			case "enter":
				followUp := strings.TrimSpace(m.input.Value())
				if followUp == "" {
					break
				}
				m.logs = append(m.logs, "[You] "+followUp)
				m.input.SetValue("")
				m.done = false

				newDoneCh := make(chan llmDoneMsg, 1)
				m.llmDoneCh = newDoneCh

				prompt := m.buildResumePrompt(followUp)
				go runLLM(prompt, m.llmModel, m.opencodeSessionID, m.sessionDir, m.logCh, newDoneCh)

				m.viewport.SetContent(WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
				m.viewport.GotoBottom()

				return m, tea.Batch(m.waitForLLMDone(), tick())
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}

	case logMsg:
		m.logs = append(m.logs, string(msg))
		m.viewport.SetContent(WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
		return m, m.listenForLogs()

	case questionMsg:
		id := msg.ID
		question := msg.Question
		options := msg.Options
		m.questions = append(m.questions, pendingQuestion{
			ID:       id,
			Question: question,
			Options:  options,
		})
		optsDesc := ""
		if len(options) > 0 {
			var labels []string
			for _, o := range options {
				labels = append(labels, o.Label)
			}
			optsDesc = fmt.Sprintf(" [%s]", strings.Join(labels, "/"))
		}
		m.logs = append(m.logs, fmt.Sprintf("[Agent asks] #%s: %s%s", id, question, optsDesc))
		m.viewport.SetContent(WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
		return m, m.listenForQuestions()

	case llmDoneMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
			m.logs = append(m.logs, fmt.Sprintf("[Error] %v", msg.err))
		} else {
			m.llmOutput = msg.output
		}
		if msg.opencodeSessionID != "" && m.opencodeSessionID == "" {
			m.opencodeSessionID = msg.opencodeSessionID
			var meta sessionMeta
			if session.ReadJSON(m.sessionDir, "metadata.json", &meta) == nil {
				meta.OpencodeSessionID = msg.opencodeSessionID
				session.WriteJSON(m.sessionDir, "metadata.json", meta)
			}
		}

		m.clarificationMode = m.hasUnansweredQuestions()

		if m.clarificationMode {
			m.logs = append(m.logs, "[Agent] Needs clarification — answer questions below.")
		} else {
			m.logs = append(m.logs, "[Agent] Done. Review output below.")
		}
		m.viewport.SetContent(WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()

	case ctrlCResetMsg:
		m.ctrlCPending = false

	case tickMsg:
		m.spinFrame++
		if !m.done {
			cmds = append(cmds, tick())
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var view string

	if m.clarificationMode {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		answered := m.answeredCount()
		total := len(m.questions)
		status := statusStyle.Render(fmt.Sprintf("◆ Clarification needed  (%d of %d answered)", answered, total))

		questionView := ""
		q := m.currentPendingQuestion()
		if q != nil {
			questionStyle := lipgloss.NewStyle().Bold(true)
			questionView = questionStyle.Render(fmt.Sprintf("  ▸ #%s: %s", q.ID, q.Question))
			hasOptions := len(q.Options) > 0
			if hasOptions {
				highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
				for i, o := range q.Options {
					marker := "  ○"
					if i == m.optionHighlightIdx {
						marker = highlightStyle.Render("  ▸")
					}
					desc := ""
					if o.Description != "" {
						desc = fmt.Sprintf(" — %s", o.Description)
					}
					questionView += fmt.Sprintf("\n    %s %s%s", marker, o.Label, desc)
				}
				virtualMarker := "  ○"
				if m.optionHighlightIdx == len(q.Options) {
					virtualMarker = highlightStyle.Render("  ▸")
				}
				questionView += fmt.Sprintf("\n    %s ✎ Type your answer below...", virtualMarker)
			}
			questionView += "\n"
		}

		q = m.currentPendingQuestion()
		inputActive := q == nil || len(q.Options) == 0 || m.optionHighlightIdx == len(q.Options)
		var inputView string
		if !inputActive {
			m.input.Placeholder = "Select an option or move to ✎ to type..."
			inputView = m.input.View()
			inputView = lipgloss.NewStyle().Faint(true).Render(inputView)
		} else if m.input.Value() == "" {
			m.input.Placeholder = "Type your answer..."
			inputView = m.input.View()
		} else {
			m.input.Placeholder = ""
			inputView = m.input.View()
		}

		view = lipgloss.JoinVertical(lipgloss.Top,
			m.viewport.View(),
			strings.Repeat("─", m.width),
			status+"  Tab:next  Shift+Tab:prev  Enter:answer  Ctrl+C:quit",
			questionView,
			inputView,
		)
	} else if m.done {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		status := statusStyle.Render("✓ Done")
		view = lipgloss.JoinVertical(lipgloss.Top,
			m.viewport.View(),
			strings.Repeat("─", m.width),
			status+"  Enter to follow up, Ctrl+C to exit.",
			m.input.View(),
		)
	} else {
		statusLine := ""
		if len(m.questions) > 0 {
			ids := make([]string, len(m.questions))
			for i, q := range m.questions {
				ids[i] = fmt.Sprintf("#%s", q.ID)
			}
			statusLine = fmt.Sprintf("[%s] Pending: %s | Ctrl+C:quit", SpinnerChar(m.spinFrame), strings.Join(ids, ", "))
		} else {
			statusLine = fmt.Sprintf("[%s] Working...", SpinnerChar(m.spinFrame))
		}

		view = lipgloss.JoinVertical(lipgloss.Top,
			m.viewport.View(),
			strings.Repeat("─", m.width),
			statusLine,
			m.input.View(),
		)
	}

	if m.ctrlCPending {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("  ══ Press Ctrl+C again to quit ══")
		view = lipgloss.JoinVertical(lipgloss.Top, view, hint)
	}

	return view
}

func (m *model) currentPendingQuestion() *pendingQuestion {
	if len(m.questions) == 0 {
		return nil
	}
	if m.selIdx >= len(m.questions) {
		m.selIdx = 0
	}
	startIdx := m.selIdx
	for {
		q := &m.questions[m.selIdx]
		if q.Answer == "" {
			return q
		}
		m.selIdx = (m.selIdx + 1) % len(m.questions)
		if m.selIdx == startIdx {
			return nil
		}
	}
}

func (m *model) hasUnansweredQuestions() bool {
	for _, q := range m.questions {
		if q.Answer == "" {
			return true
		}
	}
	return false
}

func (m *model) answeredCount() int {
	n := 0
	for _, q := range m.questions {
		if q.Answer != "" {
			n++
		}
	}
	return n
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

		m.viewport.SetContent(WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
	} else {
		m.selIdx = (m.selIdx + 1) % len(m.questions)
		m.optionHighlightIdx = 0
		m.viewport.SetContent(WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
	}
}

func (m *model) buildResumePrompt(followUp string) string {
	var sb strings.Builder
	sb.WriteString("## Pending Questions — Answered\n\n")
	sb.WriteString("You previously asked these questions. Here are the answers:\n\n")
	for i, q := range m.questions {
		if q.Answer != "" {
			optStr := ""
			if len(q.Options) > 0 {
				var labels []string
				for _, o := range q.Options {
					labels = append(labels, o.Label)
				}
				optStr = fmt.Sprintf("\n   Options: %s", strings.Join(labels, ", "))
			}
			sb.WriteString(fmt.Sprintf("%d. **%s**%s\n   ▶ %s\n\n", i+1, q.Question, optStr, q.Answer))
		}
	}
	sb.WriteString("Continue your work incorporating these answers.\n")

	if followUp != "" {
		sb.WriteString("\n## Follow-up\n\n")
		sb.WriteString(followUp)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *model) loadQuestionsFromFile() {
	lines, err := session.ReadLines(m.sessionDir, "questions.jsonl")
	if err != nil {
		return
	}

	questionMap := make(map[string]*pendingQuestion)
	var questionOrder []string
	for _, line := range lines {
		var entry questionLineEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		switch entry.Type {
		case "question":
			if _, exists := questionMap[entry.ID]; !exists {
				questionOrder = append(questionOrder, entry.ID)
			}
			questionMap[entry.ID] = &pendingQuestion{
				ID:       entry.ID,
				Question: entry.Question,
				Options:  entry.Options,
			}
		case "answer":
			if q, ok := questionMap[entry.ID]; ok {
				q.Answer = entry.Answer
			}
		}
	}

	for _, id := range questionOrder {
		q := questionMap[id]
		m.questions = append(m.questions, *q)
		optsDesc := ""
		if len(q.Options) > 0 {
			var labels []string
			for _, o := range q.Options {
				labels = append(labels, o.Label)
			}
			optsDesc = fmt.Sprintf(" [%s]", strings.Join(labels, "/"))
		}
		answerNote := ""
		if q.Answer != "" {
			answerNote = fmt.Sprintf(" (answered: %s)", q.Answer)
		}
		m.logs = append(m.logs, fmt.Sprintf("[Agent asks] #%s: %s%s%s", q.ID, q.Question, optsDesc, answerNote))
	}
}

func runMain(cfg Config, args []string) error {
	var modelFlag *string
	var outputFlag *string
	var resumeFlag *string
	var agentRunnerFlag *string

	remainArgs, err := lessflags.String("--model", &modelFlag).
		String("-o,--output", &outputFlag).
		String("--resume", &resumeFlag).
		String("--agent-runner", &agentRunnerFlag).
		Help("-h,--help", cfg.Usage).
		Parse(args)
	if err != nil {
		return err
	}

	agentRunner := "opencode"
	if agentRunnerFlag != nil {
		agentRunner = *agentRunnerFlag
	}
	if agentRunner != "opencode" && agentRunner != "codex" {
		return fmt.Errorf("unsupported agent runner: %s (supported: opencode, codex)", agentRunner)
	}
	if agentRunner == "codex" {
		return fmt.Errorf("codex runner not yet implemented")
	}

	llmModel := ""
	if modelFlag != nil {
		llmModel = *modelFlag
	}
	outputFile := ""
	if outputFlag != nil {
		outputFile = *outputFlag
	}
	resumeID := ""
	if resumeFlag != nil {
		resumeID = *resumeFlag
	}
	feature := strings.Join(remainArgs, " ")
	var sessionID string
	var opencodeSessionID string
	var sessionDir string
	var initialLogs []string

	sid, osid, dir, feat, resumeModel, logs, err := resolveSession(cfg.AgentName, resumeID)
	if err != nil {
		return err
	}
	if resumeID != "" {
		sessionID = sid
		opencodeSessionID = osid
		sessionDir = dir
		if feature == "" {
			feature = feat
		}
		if llmModel == "" {
			llmModel = resumeModel
		}
		initialLogs = logs
	} else if feature == "" {
		fmt.Fprint(os.Stderr, cfg.Usage)
		return fmt.Errorf("missing feature description")
	}

	if llmModel == "" {
		_, preferred, err := models.ListFree()
		if err != nil {
			return fmt.Errorf("failed to list free models: %w", err)
		}
		llmModel = preferred
	}

	if sessionID == "" {
		sessionID = newSessionID(cfg.SessionPrefix)
	}
	if sessionDir == "" {
		var err error
		sessionDir, err = session.Dir(cfg.AgentName, sessionID)
		if err != nil {
			return fmt.Errorf("create session dir: %w", err)
		}
	}
	session.WriteJSON(sessionDir, "metadata.json", sessionMeta{
		SessionID: sessionID,
		Feature:   feature,
		Model:     llmModel,
	})

	tempDir, err := os.MkdirTemp("", cfg.AgentName+"-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	apqPath := filepath.Join(tempDir, "add-pending-questions")
	if out, err := exec.Command("cp", exe, apqPath).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy add-pending-questions: %w\n%s", err, string(out))
	}
	for name := range cfg.Dispatch {
		dst := filepath.Join(tempDir, name)
		if out, err := exec.Command("cp", exe, dst).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to copy %s: %w\n%s", name, err, string(out))
		}
	}

	answerDir := filepath.Join(tempDir, "answer")
	if err := os.Mkdir(answerDir, 0755); err != nil {
		return fmt.Errorf("create answer dir: %w", err)
	}

	questionFifo := filepath.Join(tempDir, "question.fifo")
	if err := syscall.Mkfifo(questionFifo, 0666); err != nil {
		return fmt.Errorf("create question fifo: %w", err)
	}

	questionsFile := filepath.Join(sessionDir, "questions.jsonl")
	os.Setenv("QUESTION_FIFO", questionFifo)
	os.Setenv("ANSWER_DIR", answerDir)
	os.Setenv("QUESTIONS_FILE", questionsFile)
	pathEntry := tempDir + string(filepath.ListSeparator)
	if !strings.Contains(os.Getenv("PATH"), pathEntry) {
		os.Setenv("PATH", pathEntry+os.Getenv("PATH"))
	}

	logCh := make(chan string, 64)
	questionCh := make(chan pendingQuestion, 16)
	llmDoneCh := make(chan llmDoneMsg, 1)

	if !isTTY(os.Stdout) {
		runPlain(cfg, feature, llmModel, outputFile)
		return nil
	}

	go startQuestionMonitor(questionFifo, questionCh)

	ti := textinput.New()
	ti.Placeholder = "Type answer and press Enter..."
	ti.Focus()
	ti.CharLimit = 500

	vp := viewport.New(80, 20)
	startContent := "Starting..."
	if len(initialLogs) > 0 {
		startContent = WrapText(strings.Join(initialLogs, "\n"), vp.Width)
	} else if feature != "" {
		initialLogs = []string{"[You] " + feature}
		startContent = WrapText("[You] "+feature, vp.Width)
	}
	vp.SetContent(startContent)
	vp.GotoBottom()

	m := &model{
		feature:           feature,
		llmModel:          llmModel,
		tempDir:           tempDir,
		answerDir:         answerDir,
		sessionID:         sessionID,
		opencodeSessionID: opencodeSessionID,
		sessionDir:        sessionDir,
		questionsFile:     questionsFile,
		input:             ti,
		viewport:          vp,
		logs:              initialLogs,
		logCh:             logCh,
		questionCh:        questionCh,
		llmDoneCh:         llmDoneCh,
	}

	if resumeID != "" {
		m.loadQuestionsFromFile()
		if m.hasUnansweredQuestions() {
			m.done = true
			m.clarificationMode = true
			m.logs = append(m.logs, "[Agent] Resumed with pending clarification questions.")
		}
	}

	if !m.clarificationMode {
		prompt := cfg.Prompt + "\n\n## Feature Description\n" + feature
		if outputFile != "" {
			prompt += "\n\n## Output File\nWrite the complete report to: " + outputFile
		}
		go runLLM(prompt, llmModel, opencodeSessionID, sessionDir, logCh, llmDoneCh)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Session %s finished.\nTo resume: %s --resume %s\n", sessionID, filepath.Base(os.Args[0]), sessionID)
	return nil
}

func WrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		for len(line) > width {
			cut := width
			for cut > 0 && line[cut] != ' ' {
				cut--
			}
			if cut == 0 {
				cut = width
			}
			lines = append(lines, line[:cut])
			if cut < len(line) && line[cut] == ' ' {
				line = line[cut+1:]
			} else {
				line = line[cut:]
			}
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
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
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			var entry questionLineEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if entry.Type == "question" {
				ch <- pendingQuestion{
					ID:       entry.ID,
					Question: entry.Question,
					Options:  entry.Options,
				}
			}
		}
		f.Close()
	}
}

func runLLM(prompt, llmModel, sessionID, sessionDir string, logCh chan<- string, doneCh chan<- llmDoneMsg) {
	sf, err := os.OpenFile(filepath.Join(sessionDir, "events.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		sf = nil
	}

	output, opencodeSID, err := opencoderun.Run(context.Background(), opencoderun.Options{
		Prompt:    prompt,
		Dir:       ".",
		Model:     llmModel,
		SessionID: sessionID,
		Logger:    &formatLogger{ch: logCh, file: sf},
	})
	if sf != nil {
		sf.Close()
	}
	doneCh <- llmDoneMsg{output: output, opencodeSessionID: opencodeSID, err: err}
}

type formatLogger struct {
	ch   chan<- string
	buf  []byte
	file *os.File
}

func (l *formatLogger) Log(msg string) {
	if msg == "" {
		return
	}
	l.buf = append(l.buf, []byte(msg)...)
	for {
		idx := indexByte(l.buf, '\n')
		if idx < 0 {
			break
		}
		line := string(l.buf[:idx])
		l.buf = l.buf[idx+1:]
		if line == "" {
			continue
		}
		if l.file != nil {
			l.file.Write([]byte(line + "\n"))
		}
		formatted := formatLogLine(line)
		if formatted == "" {
			continue
		}
		select {
		case l.ch <- formatted:
		default:
		}
	}
}

func formatLogLine(line string) string {
	parsed, ok := types.ParseAgentTraceLine(json.RawMessage(line))
	if !ok {
		return ""
	}
	if parsed.Message != nil {
		return tracefmt.FormatMessageCompact(*parsed.Message)
	}
	if parsed.Activity != nil {
		msg := types.AgentTraceMessage{
			ToolCall: parsed.Activity,
		}
		return tracefmt.FormatMessageCompact(msg)
	}
	return ""
}

func indexByte(data []byte, b byte) int {
	for i, c := range data {
		if c == b {
			return i
		}
	}
	return -1
}

type sessionMeta struct {
	SessionID         string `json:"session_id"`
	OpencodeSessionID string `json:"opencode_session_id,omitempty"`
	Feature           string `json:"feature"`
	Model             string `json:"model"`
}

func readSessionFromDir(dir, sessionID string) (string, string, string, string, []string) {
	var meta sessionMeta
	if err := session.ReadJSON(dir, "metadata.json", &meta); err != nil {
		return "", "", "", "", nil
	}
	if meta.SessionID == "" {
		meta.SessionID = sessionID
	}

	lines, err := session.ReadLines(dir, "events.jsonl")
	if err != nil {
		return meta.SessionID, meta.OpencodeSessionID, meta.Feature, meta.Model, nil
	}

	var logs []string
	for _, line := range lines {
		formatted := formatLogLine(line)
		if formatted != "" {
			logs = append(logs, formatted)
		}
	}

	return meta.SessionID, meta.OpencodeSessionID, meta.Feature, meta.Model, logs
}

func resolveSession(agentName, resumeID string) (sessionID, opencodeSessionID, sessionDir, feature, llmModel string, logs []string, err error) {
	if resumeID == "" {
		return "", "", "", "", "", nil, nil
	}
	dir, err := session.Dir(agentName, resumeID)
	if err != nil {
		return "", "", "", "", "", nil, fmt.Errorf("resume session: %w", err)
	}
	sid, osid, feat, model, eventLogs := readSessionFromDir(dir, resumeID)
	if sid == "" {
		return "", "", "", "", "", nil, fmt.Errorf("session not found: %s (check the session ID)", resumeID)
	}
	return sid, osid, dir, feat, model, eventLogs, nil
}

func newSessionID(prefix string) string {
	var b [8]byte
	rand.Read(b[:])
	return prefix + hex.EncodeToString(b[:])
}

func runPlain(cfg Config, feature, llmModel, outputFile string) {
	prompt := cfg.Prompt + "\n\n## Feature Description\n" + feature

	if outputFile != "" {
		prompt += "\n\n## Output File\nWrite the complete report to: " + outputFile
	}
	output, _, err := opencoderun.Run(context.Background(), opencoderun.Options{
		Prompt: prompt,
		Dir:    ".",
		Model:  llmModel,
		Logger: &plainLogger{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

type plainLogger struct{}

func (l *plainLogger) Log(msg string) {
	if msg == "" {
		return
	}
	fmt.Fprint(os.Stderr, msg)
}
