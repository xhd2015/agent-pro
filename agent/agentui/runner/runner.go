package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	opencoderun "github.com/xhd2015/agent-pro/agent/opencode/run"
	"github.com/xhd2015/agent-pro/agent_trace/types"
	tracefmt "github.com/xhd2015/agent-pro/run"
)

type Done struct {
	Output            string
	OpencodeSessionID string
	Err               error
}

func RunLLM(prompt, llmModel, sessionID, sessionDir, agentRunner string, logCh chan<- string, doneCh chan<- Done) {
	if strings.TrimSpace(agentRunner) != "" && agentRunner != "opencode" {
		runAgentCLI(prompt, llmModel, sessionID, sessionDir, agentRunner, logCh, doneCh)
		return
	}

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

func RunPlain(prompt, llmModel, agentRunner string) {
	if strings.TrimSpace(agentRunner) != "" && agentRunner != "opencode" {
		output, err := askAgentCLI(prompt, llmModel, "", "", agentRunner, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "agent error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
		return
	}

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

func runAgentCLI(prompt, llmModel, sessionID, sessionDir, agentRunner string, logCh chan<- string, doneCh chan<- Done) {
	output, err := askAgentCLI(prompt, llmModel, sessionID, sessionDir, agentRunner, logCh)
	doneCh <- Done{Output: output, Err: err}
}

func askAgentCLI(prompt, llmModel, sessionID, sessionDir, agentRunner string, logCh chan<- string) (string, error) {
	var sf *os.File
	if strings.TrimSpace(sessionDir) != "" {
		var err error
		sf, err = os.OpenFile(filepath.Join(sessionDir, "events.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			sf = nil
		}
	}
	if sf != nil {
		defer sf.Close()
	}

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
	runner, err := agentprovider.Build(registry.AgentRunnerID(agentRunner), "", ".", env)
	if err != nil {
		return "", err
	}
	opts := &registry.AskOptions{
		Model:     llmModel,
		Workspace: ".",
		RawLog:    sf,
		SessionID: sessionID,
		OnToolCall: func(event registry.ToolCallEvent) {
			if logCh == nil {
				return
			}
			msg := strings.TrimSpace(event.Summary)
			if msg == "" {
				msg = event.ToolName
			}
			if msg == "" {
				return
			}
			select {
			case logCh <- msg:
			default:
			}
		},
	}
	return runner.Agent.Ask(context.Background(), prompt, opts, func(delta string) {
		if logCh == nil {
			return
		}
		delta = strings.TrimSpace(delta)
		if delta == "" {
			return
		}
		select {
		case logCh <- delta:
		default:
		}
	})
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
