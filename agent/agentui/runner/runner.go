package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	opencoderun "github.com/xhd2015/agent-pro/agent/opencode/run"
	"github.com/xhd2015/agent-pro/agent_trace/types"
	tracefmt "github.com/xhd2015/agent-pro/run"
)

type Done struct {
	Output            string
	OpencodeSessionID string
	Err               error
}

func RunLLM(prompt, llmModel, sessionID, sessionDir string, logCh chan<- string, doneCh chan<- Done) {
	sf, err := os.OpenFile(filepath.Join(sessionDir, "events.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		sf = nil
	}

	output, opencodeSID, err := opencoderun.Run(context.Background(), opencoderun.Options{
		Prompt:    prompt,
		Dir:       ".",
		Model:     llmModel,
		SessionID: sessionID,
		Logger:    NewLogger(logCh, sf),
	})
	if sf != nil {
		sf.Close()
	}
	doneCh <- Done{Output: output, OpencodeSessionID: opencodeSID, Err: err}
}

func RunPlain(prompt, llmModel string) {
	output, _, err := opencoderun.Run(context.Background(), opencoderun.Options{
		Prompt: prompt,
		Dir:    ".",
		Model:  llmModel,
		Logger: &PlainLogger{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

func FormatLogLine(line string) string {
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

type PlainLogger struct{}

func (l *PlainLogger) Log(msg string) {
	if msg == "" {
		return
	}
	fmt.Fprint(os.Stderr, msg)
}
