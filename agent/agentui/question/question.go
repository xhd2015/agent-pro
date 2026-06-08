package question

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Option struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type Question struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Options  []Option `json:"options"`
	Answer   string   `json:"answer,omitempty"`
}

type Entry struct {
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Question string   `json:"question,omitempty"`
	Options  []Option `json:"options,omitempty"`
	Answer   string   `json:"answer,omitempty"`
}

func ReplayLines(lines []string) []Question {
	questionMap := make(map[string]*Question)
	var questionOrder []string
	for _, line := range lines {
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		switch entry.Type {
		case "question":
			if _, exists := questionMap[entry.ID]; !exists {
				questionOrder = append(questionOrder, entry.ID)
			}
			questionMap[entry.ID] = &Question{
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

	var questions []Question
	for _, id := range questionOrder {
		questions = append(questions, *questionMap[id])
	}
	return questions
}

func HasUnanswered(questions []Question) bool {
	for _, q := range questions {
		if q.Answer == "" {
			return true
		}
	}
	return false
}

func AnsweredCount(questions []Question) int {
	n := 0
	for _, q := range questions {
		if q.Answer != "" {
			n++
		}
	}
	return n
}

func PendingIndex(questions []Question, selected int) int {
	if len(questions) == 0 {
		return -1
	}
	if selected >= len(questions) || selected < 0 {
		selected = 0
	}
	startIdx := selected
	for {
		if questions[selected].Answer == "" {
			return selected
		}
		selected = (selected + 1) % len(questions)
		if selected == startIdx {
			return -1
		}
	}
}

func BuildResumePrompt(questions []Question, followUp string) string {
	var sb strings.Builder
	sb.WriteString("## Pending Questions — Answered\n\n")
	sb.WriteString("You previously asked these questions. Here are the answers:\n\n")
	for i, q := range questions {
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

func FormatAskLog(q Question) string {
	optsDesc := ""
	if len(q.Options) > 0 {
		var labels []string
		for _, o := range q.Options {
			labels = append(labels, o.Label)
		}
		optsDesc = fmt.Sprintf(" [%s]", strings.Join(labels, "/"))
	}
	return fmt.Sprintf("[Agent asks] #%s: %s%s", q.ID, q.Question, optsDesc)
}

func FormatReplayLog(q Question) string {
	answerNote := ""
	if q.Answer != "" {
		answerNote = fmt.Sprintf(" (answered: %s)", q.Answer)
	}
	return FormatAskLog(q) + answerNote
}

func ReadQuestions(r io.Reader, ch chan<- Question) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Type == "question" {
			ch <- Question{
				ID:       entry.ID,
				Question: entry.Question,
				Options:  entry.Options,
			}
		}
	}
}
