package run

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent/exec/tool_exec"
)

type Logger interface {
	Log(msg string)
}

type Options struct {
	Dir       string
	Model     string
	Prompt    string
	SessionID string
	Logger    Logger
}

func Run(ctx context.Context, opts Options) (string, string, error) {
	promptFile, err := os.CreateTemp("", "opencode-prompt-*.txt")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file for prompt: %w", err)
	}
	if _, err := promptFile.WriteString(opts.Prompt); err != nil {
		promptFile.Close()
		return "", "", fmt.Errorf("failed to write prompt to temp file: %w", err)
	}
	promptFile.Close()

	inlineMsg := "Read the attached file and follow the instructions in it."
	args := []string{"run", inlineMsg, "--file", promptFile.Name()}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	args = append(args, "--format", "json")
	if opts.SessionID != "" {
		args = append(args, "--session", opts.SessionID)
	}

	cmd, err := tool_exec.New("opencode", args, &tool_exec.Options{Dir: opts.Dir})
	if err != nil {
		return "", "", fmt.Errorf("failed to run opencode: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start opencode: %w", err)
	}

	resultCh := make(chan struct {
		sessionID string
		answer    string
		err       error
	}, 1)

	go func() {
		sid, answer, scanErr := scanOutputStream(stdout, nil, func(line string) {
			opts.Logger.Log(line + "\n")
		})
		resultCh <- struct {
			sessionID string
			answer    string
			err       error
		}{sid, answer, scanErr}
	}()

	stderrCh := make(chan error, 1)
	go func() {
		stderrCh <- pipeToLogger(stderr, opts.Logger)
	}()

	var (
		output    string
		sessionID string
		firstErr  error
	)
	for i := 0; i < 2; i++ {
		select {
		case result := <-resultCh:
			sessionID = result.sessionID
			output = result.answer
			if result.err != nil && firstErr == nil {
				firstErr = result.err
			}
		case err := <-stderrCh:
			if err != nil && firstErr == nil {
				firstErr = err
			}
		case <-ctx.Done():
			cmd.Process.Kill()
			return "", "", ctx.Err()
		}
	}

	waitErr := cmd.Wait()
	if firstErr != nil {
		return output, sessionID, firstErr
	}
	if waitErr != nil {
		return output, sessionID, fmt.Errorf("opencode run failed: %w", waitErr)
	}

	return output, sessionID, nil
}

func pipeToLogger(r io.Reader, logger Logger) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logger.Log(scanner.Text() + "\n")
	}
	return scanner.Err()
}

type SessionRunOpts struct {
	AgentPath        string
	Dir              string
	Env              []string
	SessionID        string
	Prompt           string
	Model            string
	Agent            string
	Command          string
	Variant          string
	File             []string
	Thinking         bool
	NoSubAgents      bool
	DisableSubAgents bool
	OnEvent          StreamCallback
}

func StartSession(ctx context.Context, opts SessionRunOpts) (string, string, error) {
	return runSession(ctx, opts)
}

func ResumeSession(ctx context.Context, opts SessionRunOpts) (string, string, error) {
	if opts.SessionID == "" {
		return "", "", fmt.Errorf("sessionID is required for resume")
	}
	return runSession(ctx, opts)
}

func runSession(ctx context.Context, opts SessionRunOpts) (string, string, error) {
	agentPath := opts.AgentPath
	workspace := opts.Dir
	if workspace == "" {
		workspace, _ = os.Getwd()
	}

	args := []string{
		"run",
		"--format", "json",
	}
	if opts.SessionID != "" {
		args = append(args, "--session", opts.SessionID)
	}
	args = append(args, "--dir", workspace)
	args = append(args, "--dangerously-skip-permissions")

	if opts.Model != "" {
		model := opts.Model
		if !strings.Contains(model, "/") {
			model = "opencode/" + model
		}
		args = append(args, "--model", model)
	}
	if opts.Agent != "" {
		args = append(args, "--agent", opts.Agent)
	}
	if opts.Dir != "" {
		args = append(args, "--dir", opts.Dir)
	}
	if opts.Command != "" {
		args = append(args, "--command", opts.Command)
	}
	if opts.Variant != "" {
		args = append(args, "--variant", opts.Variant)
	}
	if opts.Thinking {
		args = append(args, "--thinking")
	}
	for _, f := range opts.File {
		args = append(args, "--file", f)
	}

	prompt := opts.Prompt
	if opts.DisableSubAgents || opts.NoSubAgents {
		prompt += "\n\n# CRITICAL RULE: DO NOT USE SUB-AGENTS\nYou MUST NOT use the Task tool (sub-agents/subagents) under any circumstances. Perform all work directly yourself without delegating to sub-agents."
	}

	args = append(args, prompt)

	cmd, err := tool_exec.New(agentPath, args, &tool_exec.Options{
		Dir: workspace,
		Env: envMap(opts.Env),
	})
	if err != nil {
		return "", "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("start opencode: %w", err)
	}

	type resultMsg struct {
		sessionID string
		answer    string
		err       error
	}
	resultCh := make(chan resultMsg, 1)
	stderrCh := make(chan error, 1)

	go func() {
		sid, answer, scanErr := scanOutputStream(stdout, opts.OnEvent, nil)
		resultCh <- resultMsg{sid, answer, scanErr}
	}()
	go func() {
		stderrBuf := strings.Builder{}
		io.Copy(&stderrBuf, stderr)
		if stderrBuf.Len() > 0 {
			stderrCh <- fmt.Errorf("%s", strings.TrimSpace(stderrBuf.String()))
		} else {
			stderrCh <- nil
		}
	}()

	var (
		sessionIDResult string
		fullAnswer      strings.Builder
		firstErr        error
	)
	for i := 0; i < 2; i++ {
		select {
		case result := <-resultCh:
			if result.sessionID != "" && sessionIDResult == "" {
				sessionIDResult = result.sessionID
			}
			fullAnswer.WriteString(result.answer)
			if result.err != nil && firstErr == nil {
				firstErr = result.err
			}
		case err := <-stderrCh:
			if err != nil && firstErr == nil {
				firstErr = err
			}
		case <-ctx.Done():
			cmd.Process.Kill()
			return "", "", ctx.Err()
		}
	}

	if sessionIDResult == "" {
		sessionIDResult = opts.SessionID
	}

	waitErr := cmd.Wait()
	if firstErr != nil {
		return sessionIDResult, fullAnswer.String(), firstErr
	}
	if waitErr != nil {
		return sessionIDResult, fullAnswer.String(), fmt.Errorf("opencode exited: %w", waitErr)
	}

	if opts.OnEvent != nil {
		opts.OnEvent(StreamEvent{Type: "done", Done: true})
	}
	return sessionIDResult, fullAnswer.String(), nil
}

func scanOutputStream(r io.Reader, onEvent StreamCallback, onRawLine func(string)) (sessionID, answer string, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	var text strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if onRawLine != nil {
			onRawLine(line)
		}
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		event := parseStreamEvent(line, onEvent)
		if event.SessionID != "" && sessionID == "" {
			sessionID = event.SessionID
		}
		if event.Text != "" {
			text.WriteString(event.Text)
		}
	}
	return sessionID, text.String(), scanner.Err()
}

func parseStreamEvent(line string, onEvent StreamCallback) StreamEvent {
	var raw struct {
		Type      string          `json:"type"`
		Timestamp int64           `json:"timestamp,omitempty"`
		SessionID string          `json:"sessionID,omitempty"`
		Part      json.RawMessage `json:"part,omitempty"`
		Error     json.RawMessage `json:"error,omitempty"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return StreamEvent{}
	}

	event := StreamEvent{
		Type:      raw.Type,
		Timestamp: raw.Timestamp,
		SessionID: raw.SessionID,
	}

	switch raw.Type {
	case "text":
		var part struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(raw.Part, &part) == nil {
			event.Text = part.Text
		}
	case "tool_use":
		var part struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Tool   string `json:"tool"`
			CallID string `json:"callID,omitempty"`
			State  *struct {
				Status string `json:"status"`
				Title  string `json:"title"`
				Output string `json:"output"`
				Error  string `json:"error"`
			} `json:"state,omitempty"`
		}
		if json.Unmarshal(raw.Part, &part) == nil {
			toolUse := &ToolUseEvent{
				ID:   part.ID,
				Tool: part.Tool,
			}
			if part.State != nil {
				toolUse.Status = part.State.Status
				toolUse.Summary = part.State.Title
				toolUse.Output = part.State.Output
				toolUse.Error = part.State.Error
			}
			if toolUse.Status == "" {
				toolUse.Status = "started"
			}
			event.ToolUse = toolUse
		}
	case "error":
		var errData struct {
			Name string `json:"name"`
			Data struct {
				Message string `json:"message"`
			} `json:"data"`
		}
		if json.Unmarshal(raw.Error, &errData) == nil && errData.Name != "" {
			event.Error = errData.Name
			if errData.Data.Message != "" {
				event.Error = errData.Data.Message
			}
		}
		if event.Error == "" {
			event.Error = "opencode error"
		}
	case "reasoning":
		var part struct {
			Text string `json:"text"`
			Time *struct {
				Start int64 `json:"start"`
				End   int64 `json:"end"`
			} `json:"time,omitempty"`
		}
		if json.Unmarshal(raw.Part, &part) == nil {
			event.Reasoning = part.Text
			if part.Time != nil {
				event.ReasoningTime = &StreamEventTime{
					Start: part.Time.Start,
					End:   part.Time.End,
				}
			}
		}
	}

	if onEvent != nil {
		onEvent(event)
	}

	return event
}

func envMap(env []string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	m := make(map[string]string, len(env))
	for _, e := range env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				m[e[:i]] = e[i+1:]
				break
			}
		}
	}
	return m
}
