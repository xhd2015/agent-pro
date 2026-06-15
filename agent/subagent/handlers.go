package subagent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func HandleYieldPendingQuestions(args []string) error {
	questionFifo := os.Getenv("QUESTION_FIFO")
	if questionFifo == "" {
		return fmt.Errorf("QUESTION_FIFO must be set")
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: yield-pending-questions '<json>' '<json>' ...")
	}

	f, err := os.OpenFile(questionFifo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open questions file: %w", err)
	}
	defer f.Close()

	for i, arg := range args {
		var input struct {
			ID       string `json:"id"`
			Question string `json:"question"`
			Options  []struct {
				Option      string `json:"option"`
				Explanation string `json:"explanation"`
			} `json:"options"`
		}
		if err := json.Unmarshal([]byte(arg), &input); err != nil || input.Question == "" {
			continue
		}
		id := input.ID
		if id == "" {
			id = fmt.Sprintf("%d", i+1)
		}
		entry := map[string]any{
			"type":     "question",
			"id":       id,
			"question": input.Question,
		}
		if len(input.Options) > 0 {
			entry["options"] = input.Options
		}
		data, _ := json.Marshal(entry)
		fmt.Fprintf(f, "%s\n", string(data))
	}
	return nil
}

func HandleReportProgress(args []string) error {
	progressFile := os.Getenv("PROGRESS_FILE")
	if progressFile == "" {
		return fmt.Errorf("PROGRESS_FILE must be set")
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: report-progress <description>")
	}

	description := strings.Join(args, " ")

	entry := map[string]string{
		"type":        "progress",
		"description": description,
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(entry)

	f, err := os.OpenFile(progressFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open progress file: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "%s\n", string(data))
	return nil
}
