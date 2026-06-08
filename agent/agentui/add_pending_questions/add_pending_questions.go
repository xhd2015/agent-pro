package add_pending_questions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

type QuestionOption struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type QuestionInput struct {
	ID       string           `json:"id,omitempty"`
	Question string           `json:"question"`
	Options  []QuestionOption `json:"options,omitempty"`
}

type QuestionEntry struct {
	Type     string           `json:"type"`
	ID       string           `json:"id"`
	Question string           `json:"question,omitempty"`
	Options  []QuestionOption `json:"options,omitempty"`
}

func Run() {
	questionFifo := os.Getenv("QUESTION_FIFO")
	if questionFifo == "" {
		fmt.Fprintf(os.Stderr, "QUESTION_FIFO must be set\n")
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: add-pending-questions '<json>' '<json>' ...\n")
		os.Exit(1)
	}

	answerDir := os.Getenv("ANSWER_DIR")
	counterFile := ""
	if answerDir != "" {
		counterFile = filepath.Join(answerDir, ".counter")
	}

	var entries []QuestionEntry
	for _, arg := range args {
		var input QuestionInput
		if err := json.Unmarshal([]byte(arg), &input); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse question JSON: %v\n%s\n", err, arg)
			os.Exit(1)
		}
		if input.Question == "" {
			fmt.Fprintf(os.Stderr, "question field is required in JSON: %s\n", arg)
			os.Exit(1)
		}
		id := nextID(counterFile)
		entries = append(entries, QuestionEntry{
			Type:     "question",
			ID:       id,
			Question: input.Question,
			Options:  input.Options,
		})
	}

	questionsFile := os.Getenv("QUESTIONS_FILE")

	fifo, err := os.OpenFile(questionFifo, os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open question fifo: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal question: %v\n", err)
			os.Exit(1)
		}
		line := string(data)

		fmt.Fprintf(fifo, "%s\n", line)

		if questionsFile != "" {
			appendLine(questionsFile, line)
		}
	}
	fifo.Close()

	fmt.Printf("%d questions recorded; you can suspend the chat now.\n", len(entries))
}

func appendLine(path, line string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open questions file: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", line)
}

func nextID(counterFile string) string {
	if counterFile == "" {
		return "unknown"
	}
	f, err := os.OpenFile(counterFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "counter file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, _ := io.ReadAll(f)
	n := 1
	if len(data) > 0 {
		fmt.Sscanf(string(data), "%d", &n)
	}
	f.Seek(0, 0)
	f.Truncate(0)
	fmt.Fprintf(f, "%d", n+1)
	return fmt.Sprintf("%d", n)
}
