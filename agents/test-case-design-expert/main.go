package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/xhd2015/agent-pro/agent/opencode/models"
	opencoderun "github.com/xhd2015/agent-pro/agent/opencode/run"
	"github.com/xhd2015/agent-pro/agent_trace/types"
	ask_user "github.com/xhd2015/agent-pro/agents/test-case-design-expert/ask_user/lib"
	tracefmt "github.com/xhd2015/agent-pro/run"
)

//go:embed PROMPT.md
var promptTemplate string

const usage = `Usage: test-case-design-expert [--model MODEL] [--output FILE] <feature description>

Generate user-facing end-to-end test cases for the given feature description.
The expert will brainstorm the idea, ask clarifying questions interactively,
and produce a complete test plan.

Arguments:
  <feature description>   The feature to design tests for (positional args joined with spaces)

Options:
  --model MODEL    Model identifier to use (defaults to the first free model from opencode)
  -o, --output FILE  File to write the test report (default: auto-generated <name>-tests-design.md)
  -h, --help       Show this help message
`

type pendingQuestion struct {
	id       string
	question string
}

type logMsg string
type questionMsg pendingQuestion
type llmDoneMsg struct {
	output string
	err    error
}
type tickMsg struct{}

type model struct {
	feature   string
	llmModel  string
	tempDir   string
	answerDir string

	sessionID   string
	sessionFile *os.File

	logs      []string
	questions []pendingQuestion
	selIdx    int
	input     textinput.Model
	viewport  viewport.Model
	done      bool
	spinFrame int
	llmOutput string
	err       error
	width     int
	height    int

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

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func spinnerChar(frame int) string {
	return spinnerChars[frame%len(spinnerChars)]
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
		if m.done {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
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
				go runLLM(followUp, m.llmModel, m.sessionID, "", m.logCh, newDoneCh)

				m.viewport.SetContent(wrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
				m.viewport.GotoBottom()

				return m, tea.Batch(m.waitForLLMDone(), tick())
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if len(m.questions) > 0 {
				m.selIdx = (m.selIdx + 1) % len(m.questions)
			}
		case "enter":
			answer := m.input.Value()
			if answer == "" {
				break
			}
			q := m.currentQuestion()
			if q == nil {
				break
			}
			af := filepath.Join(m.answerDir, q.id+".fifo")
			f, err := os.OpenFile(af, os.O_WRONLY, 0)
			if err == nil {
				f.Write([]byte(answer))
				f.Close()
			}
			m.logs = append(m.logs, fmt.Sprintf("[You] Answered #%s", q.id))
			m.removeQuestion(q.id)
			m.input.SetValue("")
			m.input.Focus()
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}

	case logMsg:
		m.logs = append(m.logs, string(msg))
		m.viewport.SetContent(wrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
		return m, m.listenForLogs()

	case questionMsg:
		id := msg.id
		question := msg.question
		m.questions = append(m.questions, pendingQuestion{id: id, question: question})
		m.logs = append(m.logs, fmt.Sprintf("[Agent asks] #%s: %s", id, question))
		m.viewport.SetContent(wrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()
		return m, m.listenForQuestions()

	case llmDoneMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
			m.logs = append(m.logs, fmt.Sprintf("[Error] %v", msg.err))
		} else {
			m.llmOutput = msg.output
			m.logs = append(m.logs, "[Agent] Done. Review output below.")
		}
		m.viewport.SetContent(wrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()

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
	if m.done {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		status := statusStyle.Render("✓ Done")
		return lipgloss.JoinVertical(lipgloss.Top,
			m.viewport.View(),
			strings.Repeat("─", m.width),
			status+"  Enter to follow up, Ctrl+C to exit.",
			m.input.View(),
		)
	}

	statusLine := ""
	if len(m.questions) > 0 {
		ids := make([]string, len(m.questions))
		for i, q := range m.questions {
			marker := " "
			if i == m.selIdx {
				marker = "▶"
			}
			ids[i] = fmt.Sprintf("%s#%s", marker, q.id)
		}
		statusLine = fmt.Sprintf("[%s] Pending: %s | Tab:switch  Enter:answer  Ctrl+C:quit", spinnerChar(m.spinFrame), strings.Join(ids, ", "))
	} else {
		statusLine = fmt.Sprintf("[%s] Waiting for agent questions...", spinnerChar(m.spinFrame))
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.viewport.View(),
		strings.Repeat("─", m.width),
		statusLine,
		m.input.View(),
	)
}

func (m *model) currentQuestion() *pendingQuestion {
	if len(m.questions) == 0 {
		return nil
	}
	if m.selIdx >= len(m.questions) {
		m.selIdx = 0
	}
	return &m.questions[m.selIdx]
}

func (m *model) removeQuestion(id string) {
	for i, q := range m.questions {
		if q.id == id {
			m.questions = append(m.questions[:i], m.questions[i+1:]...)
			break
		}
	}
	if m.selIdx >= len(m.questions) && len(m.questions) > 0 {
		m.selIdx = len(m.questions) - 1
	}
}

func main() {
	base := filepath.Base(os.Args[0])
	if base == "ask_user" || base == "ask_user.exe" {
		ask_user.Run()
		return
	}

	if err := runMain(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runMain(args []string) error {
	var llmModel string
	var outputFile string
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			fmt.Fprint(os.Stderr, usage)
			os.Exit(0)
		case arg == "--model":
			if i+1 >= len(args) {
				return fmt.Errorf("--model requires a value\n\n%s", usage)
			}
			llmModel = args[i+1]
			i++
		case strings.HasPrefix(arg, "--model="):
			llmModel = strings.TrimPrefix(arg, "--model=")
		case arg == "-o" || arg == "--output":
			if i+1 >= len(args) {
				return fmt.Errorf("--output requires a file path\n\n%s", usage)
			}
			outputFile = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output="):
			outputFile = strings.TrimPrefix(arg, "--output=")
		default:
			positional = append(positional, arg)
		}
	}

	feature := strings.Join(positional, " ")
	var sessionID string
	var initialLogs []string

	if data, err := os.ReadFile(".session.jsonl"); err == nil {
		sid, feat, model, logs, err := parseSessionFile(data)
		if err != nil {
			return fmt.Errorf("parse session file: %w", err)
		}
		sessionID = sid
		if feature == "" {
			feature = feat
		}
		if llmModel == "" {
			llmModel = model
		}
		initialLogs = logs
	} else if feature == "" {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("missing feature description")
	}

	if llmModel == "" {
		_, preferred, err := models.ListFree()
		if err != nil {
			return fmt.Errorf("failed to list free models: %w", err)
		}
		llmModel = preferred
	}

	tempDir, err := os.MkdirTemp("", "test-case-design-expert-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	askUserPath := filepath.Join(tempDir, "ask_user")
	if out, err := exec.Command("cp", exe, askUserPath).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy ask_user: %w\n%s", err, string(out))
	}

	answerDir := filepath.Join(tempDir, "answer")
	if err := os.Mkdir(answerDir, 0755); err != nil {
		return fmt.Errorf("create answer dir: %w", err)
	}

	questionFifo := filepath.Join(tempDir, "question.fifo")
	if err := syscall.Mkfifo(questionFifo, 0666); err != nil {
		return fmt.Errorf("create question fifo: %w", err)
	}

	os.Setenv("QUESTION_FIFO", questionFifo)
	os.Setenv("ANSWER_DIR", answerDir)
	pathEntry := tempDir + string(filepath.ListSeparator)
	if !strings.Contains(os.Getenv("PATH"), pathEntry) {
		os.Setenv("PATH", pathEntry+os.Getenv("PATH"))
	}

	logCh := make(chan string, 64)
	questionCh := make(chan pendingQuestion, 16)
	llmDoneCh := make(chan llmDoneMsg, 1)

	if !isTTY(os.Stdout) {
		runPlain(feature, llmModel, outputFile)
		return nil
	}

	go startQuestionMonitor(questionFifo, questionCh)

	prompt := promptTemplate + "\n\n## Feature Description\n" + feature
	if outputFile != "" {
		prompt += "\n\n## Output File\nWrite the complete report to: " + outputFile
	}
	go runLLM(prompt, llmModel, sessionID, feature, logCh, llmDoneCh)

	ti := textinput.New()
	ti.Placeholder = "Type answer and press Enter..."
	ti.Focus()
	ti.CharLimit = 500

	vp := viewport.New(80, 20)
	startContent := "Starting..."
	if len(initialLogs) > 0 {
		startContent = wrapText(strings.Join(initialLogs, "\n"), vp.Width)
	}
	vp.SetContent(startContent)
	vp.GotoBottom()

	m := &model{
		feature:    feature,
		llmModel:   llmModel,
		tempDir:    tempDir,
		answerDir:  answerDir,
		sessionID:  sessionID,
		input:      ti,
		viewport:   vp,
		logs:       initialLogs,
		logCh:      logCh,
		questionCh: questionCh,
		llmDoneCh:  llmDoneCh,
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}
	return nil
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		for len(line) > width {
			// find last space within width
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
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				ch <- pendingQuestion{id: parts[0], question: parts[1]}
			}
		}
		f.Close()
	}
}

func runLLM(prompt, llmModel, sessionID, feature string, logCh chan<- string, doneCh chan<- llmDoneMsg) {
	var sf *os.File
	if f, err := os.OpenFile(".session.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		// write metadata if new file
		if fi, _ := f.Stat(); fi.Size() == 0 {
			meta := fmt.Sprintf("{\"session_id\":\"%s\",\"feature\":%q,\"model\":%q}\n",
				sessionID, feature, llmModel)
			f.Write([]byte(meta))
		}
		sf = f
	}

	output, err := opencoderun.Run(opencoderun.Options{
		Prompt:    prompt,
		Dir:       ".",
		Model:     llmModel,
		SessionID: sessionID,
		Logger:    &formatLogger{ch: logCh, file: sf},
	})
	if sf != nil {
		sf.Close()
	}
	doneCh <- llmDoneMsg{output: output, err: err}
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
	SessionID string `json:"session_id"`
	Feature   string `json:"feature"`
	Model     string `json:"model"`
}

func parseSessionFile(data []byte) (sessionID, feature, model string, logs []string, err error) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return "", "", "", nil, fmt.Errorf("empty session file")
	}

	var meta sessionMeta
	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "{") {
		return "", "", "", nil, fmt.Errorf("invalid session file: first line is not JSON")
	}
	if err := json.Unmarshal([]byte(firstLine), &meta); err != nil {
		return "", "", "", nil, fmt.Errorf("invalid session metadata: %w", err)
	}

	sessionID = meta.SessionID

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if sessionID == "" {
			var ev struct {
				SessionID string `json:"sessionID"`
			}
			if err := json.Unmarshal([]byte(line), &ev); err == nil && ev.SessionID != "" {
				sessionID = ev.SessionID
			}
		}
		formatted := formatLogLine(line)
		if formatted != "" {
			logs = append(logs, formatted)
		}
	}

	if sessionID == "" {
		return "", "", "", nil, fmt.Errorf("session file has no session_id")
	}

	return sessionID, meta.Feature, meta.Model, logs, nil
}

func generateOutputName(feature string) string {
	const maxPrefixLen = 50
	const suffix = "-tests-design.md"

	s := strings.ToLower(strings.TrimSpace(feature))

	var buf strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			buf.WriteRune(r)
		}
	}
	s = buf.String()

	s = strings.ReplaceAll(s, " ", "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")

	if s == "" {
		return "tests-design.md"
	}
	if len(s) > maxPrefixLen {
		s = s[:maxPrefixLen]
	}
	return s + suffix
}

func runPlain(feature, llmModel, outputFile string) {
	prompt := promptTemplate + "\n\n## Feature Description\n" + feature

	if outputFile != "" {
		prompt += "\n\n## Output File\nWrite the complete report to: " + outputFile
	}
	output, err := opencoderun.Run(opencoderun.Options{
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
