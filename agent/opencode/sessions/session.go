package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/xhd2015/agent-pro/agent/opencode"
)

func Export(ctx context.Context, agent *opencode.OpencodeAgent, sessionID string) (*opencode.SessionExport, error) {
	agentPath, err := agent.ResolvePath()
	if err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp("", "acp-session-export-*.json")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, agentPath, "export", sessionID)
	cmd.Env = agent.BuildEnv()
	cmd.Stdout = tmpFile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("export session: %s: %w", stderr.String(), err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("seek temp file: %w", err)
	}
	var export opencode.SessionExport
	if err := json.NewDecoder(tmpFile).Decode(&export); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("parse session export: %w", err)
	}
	tmpFile.Close()
	return &export, nil
}
