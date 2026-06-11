package implementer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/agent/session"
)

type Options struct {
	Prompt      string
	AgentRunner string
	MockConfig  string
}

func Run(opts Options) error {
	prompt := strings.TrimSpace(opts.Prompt)
	if prompt == "" {
		return fmt.Errorf("agent implement requires <prompt>")
	}

	agentRunner := strings.TrimSpace(opts.AgentRunner)
	if agentRunner == "" {
		agentRunner = "opencode"
	}

	threadID := os.Getenv("CODEX_THREAD_ID")
	var sessionDir string
	if threadID == "" {
		threadID = fmt.Sprintf("impl_%d", time.Now().UnixNano())
		os.Setenv("CODEX_THREAD_ID", threadID)
		var err error
		sessionDir, err = session.Dir("doctest-agent", threadID)
		if err != nil {
			return fmt.Errorf("create session dir: %w", err)
		}
	} else {
		var err error
		sessionDir, err = session.Dir("doctest-agent", threadID)
		if err != nil {
			return fmt.Errorf("resolve session dir: %w", err)
		}
	}

	tempDir, err := os.MkdirTemp("", "doctest-agent-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}
	ypqPath := filepath.Join(tempDir, "yield-pending-questions")
	if out, err := exec.Command("cp", exe, ypqPath).CombinedOutput(); err != nil {
		return fmt.Errorf("copy yield-pending-questions: %w\n%s", err, string(out))
	}

	pathEntry := tempDir + string(filepath.ListSeparator)
	os.Setenv("PATH", pathEntry+os.Getenv("PATH"))

	questionFile := filepath.Join(tempDir, "questions.jsonl")
	os.Setenv("QUESTION_FIFO", questionFile)

	if opts.MockConfig != "" {
		os.Setenv("FAKE_CODEX_MOCK_CONFIG", opts.MockConfig)
	}

	fullPrompt := prompt
	output, err := runAgent(agentRunner, fullPrompt, threadID, sessionDir)
	if err != nil {
		return fmt.Errorf("sub-agent failed: %w", err)
	}

	f, fErr := os.Open(questionFile)
	if fErr == nil {
		defer f.Close()
		var buf bytes.Buffer
		buf.ReadFrom(f)
		if buf.Len() > 0 {
			fmt.Print(buf.String())
			return nil
		}
	}

	fmt.Print(output)
	return nil
}

func runAgent(agentRunner, prompt, threadID, sessionDir string) (string, error) {
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
		Workspace: ".",
		SessionID: threadID,
	}

	output, err := runner.Agent.Ask(context.Background(), prompt, opts, func(delta string) {})
	if err != nil {
		return output, err
	}
	return output, nil
}

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
		data, _ := json.Marshal(entry)
		fmt.Fprintf(f, "%s\n", string(data))
	}
	return nil
}
