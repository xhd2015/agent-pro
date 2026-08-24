package knowledgesink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
)

// runHeadlessCLI launches codex/grok (non-TTY) via agentui — no detach/PTY.
// Caller-only path for knowledgesink; does not change agentrunapi.Run defaults.
func runHeadlessCLI(ctx context.Context, opts Opts, runOpts agentrunapi.RunOpts, schema string) (string, error) {
	storeHome := strings.TrimSpace(runOpts.StoreHome)
	store, err := agentstorage.NewFileStore(storeHome)
	if err != nil {
		return "", fmt.Errorf("headless store: %w", err)
	}

	prompt := runOpts.Prompt
	resultPath := strings.TrimSpace(runOpts.ResultFile)
	needReturnJSON := strings.TrimSpace(schema) != ""
	if needReturnJSON {
		if resultPath == "" {
			resultPath, err = newSinkJSONResultPath()
			if err != nil {
				return "", err
			}
		}
		prompt = appendSinkJSONResultInstructions(prompt, resultPath, schema)
	}

	sessionID := strings.TrimSpace(runOpts.SessionID)
	if sessionID == "" {
		sessionID = fmt.Sprintf("sink-headless-%s", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}

	stdout := io.Writer(io.Discard)
	if opts.Verbose {
		// Agent human-readable events → stderr (plan stays on stdout).
		stdout = verboseWriter(opts)
		if stdout == nil {
			stdout = os.Stderr
		}
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if from := strings.TrimSpace(runOpts.AgentRunner); opts.Verbose && from != "" {
		verboseNotice(opts, "headless runner=%s", from)
	}

	uiErr := agentui.Run(ctx, agentui.RunOptions{
		Prompt:               prompt,
		Runner:               runOpts.AgentRunner,
		Model:                runOpts.Model,
		ModelReasoningEffort: runOpts.ModelReasoningEffort,
		SessionID:            sessionID,
		Workspace:            runOpts.WorkspaceDir,
		Store:                store,
		Stdout:               stdout,
		Stderr:               stderr,
		Verbose:              opts.Verbose,
		Driver:               opts.Driver,
		JSON:                 false,
		Open:                 false,
		Detach:               false,
	})
	if uiErr != nil {
		return "", uiErr
	}

	if !needReturnJSON {
		// create-mr: host reads sink-N/result.json after return.
		return "", nil
	}
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return "", fmt.Errorf("result file %s: %w", resultPath, err)
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", fmt.Errorf("result file %s: empty (agent finished without writing JSON)", resultPath)
	}
	if !json.Valid(trimmed) {
		return "", fmt.Errorf("result file %s: not valid JSON", resultPath)
	}
	return string(data), nil
}

func newSinkJSONResultPath() (string, error) {
	dir := ""
	if st, err := os.Stat("/tmp"); err == nil && st.IsDir() {
		dir = "/tmp"
	}
	f, err := os.CreateTemp(dir, "knowledge-sink-result-*.json")
	if err != nil && dir != "" {
		f, err = os.CreateTemp("", "knowledge-sink-result-*.json")
	}
	if err != nil {
		return "", fmt.Errorf("create result temp file: %w", err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func appendSinkJSONResultInstructions(prompt, resultPath, schema string) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(prompt, "\n"))
	b.WriteString("\n\n")
	b.WriteString("The caller is blocked until this file contains valid JSON. Write it as soon as you have the object — do not delay the write for extra lookups, a closing summary, or another API call:\n  ")
	b.WriteString(resultPath)
	b.WriteString("\nMatch this example (same keys; values are illustrative):\n")
	b.WriteString(schema)
	b.WriteString("\nWrite atomically (tmp + rename). Do not write the result into the worktree.\n")
	return b.String()
}
