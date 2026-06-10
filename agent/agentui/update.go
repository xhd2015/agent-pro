package agentui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xhd2015/agent-pro/agent/agentui/question"
	"github.com/xhd2015/agent-pro/agent/agentui/sessionstate"
	"github.com/xhd2015/agent-pro/agent/agentui/textutil"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
	case tea.KeyMsg:
		return m.updateKey(msg)
	case logMsg:
		return m.updateLog(msg)
	case questionMsg:
		return m.updateQuestion(msg)
	case llmDoneMsg:
		return m.updateLLMDone(msg)
	case ctrlCResetMsg:
		m.ctrlCPending = false
	case tickMsg:
		m.spinFrame++
		if !m.done {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, tea.Batch(tick(), cmd)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.updateClarificationKey(msg)
	}

	if m.done {
		return m.updateDoneKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateClarificationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m model) updateDoneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.startLLM(prompt, newDoneCh)

		m.viewport.SetContent(textutil.WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()

		return m, tea.Batch(m.waitForLLMDone(), tick())
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) updateLog(msg logMsg) (tea.Model, tea.Cmd) {
	m.logs = append(m.logs, string(msg))
	m.viewport.SetContent(textutil.WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
	m.viewport.GotoBottom()
	return m, m.listenForLogs()
}

func (m model) updateQuestion(msg questionMsg) (tea.Model, tea.Cmd) {
	m.questions = append(m.questions, pendingQuestion(msg))
	m.logs = append(m.logs, question.FormatAskLog(question.Question(msg)))
	m.viewport.SetContent(textutil.WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
	m.viewport.GotoBottom()
	return m, m.listenForQuestions()
}

func (m model) updateLLMDone(msg llmDoneMsg) (tea.Model, tea.Cmd) {
	m.done = true
	if msg.Err != nil {
		m.err = msg.Err
		m.logs = append(m.logs, fmt.Sprintf("[Error] %v", msg.Err))
	} else {
		m.llmOutput = msg.Output
	}
	if msg.OpencodeSessionID != "" && m.opencodeSessionID == "" {
		m.opencodeSessionID = msg.OpencodeSessionID
		sessionstate.UpdateOpencodeSessionID(m.sessionDir, msg.OpencodeSessionID)
	}

	m.clarificationMode = m.hasUnansweredQuestions()

	if m.clarificationMode {
		m.logs = append(m.logs, "[Agent] Needs clarification — answer questions below.")
	} else {
		m.logs = append(m.logs, "[Agent] Done. Review output below.")
	}
	m.viewport.SetContent(textutil.WrapText(strings.Join(m.logs, "\n"), m.viewport.Width))
	m.viewport.GotoBottom()

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
