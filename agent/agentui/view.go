package agentui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	var view string
	switch {
	case m.clarificationMode:
		view = m.viewClarification()
	case m.done:
		view = m.viewDone()
	default:
		view = m.viewRunning()
	}

	if m.ctrlCPending {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("  ══ Press Ctrl+C again to quit ══")
		view = lipgloss.JoinVertical(lipgloss.Top, view, hint)
	}

	return view
}

func (m model) viewClarification() string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	answered := m.answeredCount()
	total := len(m.questions)
	status := statusStyle.Render(fmt.Sprintf("◆ Clarification needed  (%d of %d answered)", answered, total))

	questionView := m.viewCurrentQuestion()

	q := m.currentPendingQuestion()
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

	return lipgloss.JoinVertical(lipgloss.Top,
		m.viewport.View(),
		strings.Repeat("─", m.width),
		status+"  Tab:next  Shift+Tab:prev  Enter:answer  Ctrl+C:quit",
		questionView,
		inputView,
	)
}

func (m model) viewCurrentQuestion() string {
	questionView := ""
	q := m.currentPendingQuestion()
	if q == nil {
		return questionView
	}
	questionStyle := lipgloss.NewStyle().Bold(true)
	questionView = questionStyle.Render(fmt.Sprintf("  ▸ #%s: %s", q.ID, q.Question))
	if len(q.Options) > 0 {
		questionView += m.viewQuestionOptions(q)
	}
	questionView += "\n"
	return questionView
}

func (m model) viewQuestionOptions(q *pendingQuestion) string {
	var opts strings.Builder
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
		opts.WriteString(fmt.Sprintf("\n    %s %s%s", marker, o.Label, desc))
	}
	virtualMarker := "  ○"
	if m.optionHighlightIdx == len(q.Options) {
		virtualMarker = highlightStyle.Render("  ▸")
	}
	opts.WriteString(fmt.Sprintf("\n    %s ✎ Type your answer below...", virtualMarker))
	return opts.String()
}

func (m model) viewDone() string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	status := statusStyle.Render("✓ Done")
	return lipgloss.JoinVertical(lipgloss.Top,
		m.viewport.View(),
		strings.Repeat("─", m.width),
		status+"  Enter to follow up, Ctrl+C to exit.",
		m.input.View(),
	)
}

func (m model) viewRunning() string {
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

	return lipgloss.JoinVertical(lipgloss.Top,
		m.viewport.View(),
		strings.Repeat("─", m.width),
		statusLine,
		m.input.View(),
	)
}
