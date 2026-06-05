package explain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/opencode/models"
	"github.com/xhd2015/agent-pro/agent/opencode/run"
	"github.com/xhd2015/less-gen/flags"
)

var explainHelp = `Usage: explain [options] <message> [follow-up messages...]

Ask an AI agent a question and get an answer. Sessions are reused when the
positional arguments match a prefix of a previous session's user messages.

Options:
  --model MODEL
              Model to use for generation
  --agent-runner RUNNER
              Agent runner to use (opencode or codex, default: opencode)
  -v, --verbose
              Show details about session creation/reuse
  -h, --help  Show this help message
`

type Runner interface {
	Start(ctx context.Context, model string, prompt string) (sessionID string, output string, err error)
	Resume(ctx context.Context, model string, prompt string, meta json.RawMessage) (output string, err error)
}

type Runtime struct {
	AgentRunner string
}

func (r *Runtime) Start(ctx context.Context, model string, prompt string) (string, string, error) {
	if r.AgentRunner == "codex" {
		return "", "", fmt.Errorf("codex runner not yet implemented")
	}
	agentPath := agentPathFromEnv("opencode")
	sessionID, answer, err := run.StartSession(ctx, run.SessionRunOpts{
		AgentPath: agentPath,
		Prompt:    prompt,
		Model:     model,
	})
	return sessionID, answer, err
}

func (r *Runtime) Resume(ctx context.Context, model string, prompt string, meta json.RawMessage) (string, error) {
	if r.AgentRunner == "codex" {
		return "", fmt.Errorf("codex runner not yet implemented")
	}
	var opencodeMeta struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(meta, &opencodeMeta); err != nil {
		return "", fmt.Errorf("parse opencode meta: %w", err)
	}
	if opencodeMeta.SessionID == "" {
		return "", fmt.Errorf("session_id not found in opencode meta")
	}
	agentPath := agentPathFromEnv("opencode")
	_, answer, err := run.ResumeSession(ctx, run.SessionRunOpts{
		AgentPath: agentPath,
		SessionID: opencodeMeta.SessionID,
		Prompt:    prompt,
		Model:     model,
	})
	return answer, err
}

func agentPathFromEnv(defaultPath string) string {
	if p := os.Getenv("EXPLAIN_AGENT_PATH"); p != "" {
		return p
	}
	return defaultPath
}

func RunExplain(args []string) error {
	return RunExplainWithRunner(args, &Runtime{})
}

func RunExplainWithRunner(rawArgs []string, runner Runner) error {
	var model string
	var agentRunner string
	var verbose bool
	remainArgs, err := flags.
		String("--model", &model).
		String("--agent-runner", &agentRunner).
		Bool("-v,--verbose", &verbose).
		Help("-h,--help", explainHelp).
		Parse(rawArgs)
	if err != nil {
		return err
	}

	if agentRunner == "" {
		agentRunner = "opencode"
	}
	if agentRunner != "opencode" && agentRunner != "codex" {
		return fmt.Errorf("unsupported agent runner: %s (supported: opencode, codex)", agentRunner)
	}

	if len(remainArgs) == 0 {
		return fmt.Errorf("missing message argument\n\n%s", explainHelp)
	}

	firstMsg := remainArgs[0]
	var followUp string
	if len(remainArgs) >= 2 {
		followUp = strings.Join(remainArgs[1:], " ")
	}

	if r, ok := runner.(*Runtime); ok {
		r.AgentRunner = agentRunner
	}

	ctx := context.Background()

	if model == "" {
		_, preferredModel, listErr := models.ListFree()
		if listErr == nil && preferredModel != "" {
			model = preferredModel
		}
	}

	match, err := findMatchingSession(remainArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to search sessions: %v\n", err)
	}

	if match != nil {
		agentRunnersMeta := match.Data.AgentRunnersMeta
		if agentRunnersMeta == nil {
			agentRunnersMeta = make(RunnerMeta)
		}
		runnerMetaBytes, ok := agentRunnersMeta[agentRunner]
		if !ok || len(runnerMetaBytes) == 0 {
			return fmt.Errorf("session found but no %s runner meta available", agentRunner)
		}

		prompt := firstMsg
		if followUp != "" {
			followUpArgs := remainArgs[match.MatchedCount:]
			prompt = strings.Join(followUpArgs, " ")
		}

		if verbose {
			matchedMsgs := userMessageSlice(match.Data)[:match.MatchedCount]
			fmt.Fprintf(os.Stderr, "[explain] matched session %s (%d msg prefix: %v)\n", filepath.Base(match.SessionDir), match.MatchedCount, matchedMsgs)
			fmt.Fprintf(os.Stderr, "[explain] resuming %s session...\n", agentRunner)
		}

		output, resumeErr := runner.Resume(ctx, model, prompt, runnerMetaBytes)
		if resumeErr != nil {
			return fmt.Errorf("resume session failed: %w", resumeErr)
		}
		if output == "" {
			return fmt.Errorf("agent returned empty response")
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[explain] done\n")
		}

		match.Data.Messages = append(match.Data.Messages, Message{Role: "user", Message: prompt})
		match.Data.Messages = append(match.Data.Messages, Message{Role: "assistant", Message: output})
		if model != "" {
			match.Data.Model = model
		}
		if err := updateSession(match.SessionDir, match.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
		}

		fmt.Print(output)
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[explain] no matching session\n")
		fmt.Fprintf(os.Stderr, "[explain] starting new %s session...\n", agentRunner)
	}

	sessionID, output, startErr := runner.Start(ctx, model, firstMsg)
	if startErr != nil {
		return fmt.Errorf("start session failed: %w", startErr)
	}
	if output == "" {
		return fmt.Errorf("agent returned empty response")
	}

	runnerMeta := make(RunnerMeta)
	switch agentRunner {
	case "opencode":
		runnerMeta["opencode"] = mustMarshalJSON(map[string]string{"session_id": sessionID})
	case "codex":
		runnerMeta["codex"] = mustMarshalJSON(map[string]string{"codex_thread_id": sessionID})
	}

	data := SessionData{
		AgentRunner:      agentRunner,
		Model:            model,
		AgentRunnersMeta: runnerMeta,
		Messages: []Message{
			{Role: "user", Message: firstMsg},
			{Role: "assistant", Message: output},
		},
	}

	sessionDir, saveErr := saveSession(firstMsg, data)
	if saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[explain] saved to %s\n", filepath.Base(sessionDir))
	}

	if followUp != "" {
		m := &MatchResult{
			SessionDir: sessionDir,
			Data:       data,
		}
		followUpOutput, resumeErr := runner.Resume(ctx, model, followUp, runnerMeta[agentRunner])
		if resumeErr != nil {
			return fmt.Errorf("follow-up resume failed: %w", resumeErr)
		}
		if followUpOutput == "" {
			return fmt.Errorf("agent returned empty response")
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[explain] follow-up done\n")
		}

		m.Data.Messages = append(m.Data.Messages, Message{Role: "user", Message: followUp})
		m.Data.Messages = append(m.Data.Messages, Message{Role: "assistant", Message: followUpOutput})
		if err := updateSession(sessionDir, m.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", saveErr)
		}

		fmt.Print(followUpOutput)
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[explain] done\n")
	}
	fmt.Print(output)
	return nil
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal json: %v", err))
	}
	return json.RawMessage(bytes)
}
