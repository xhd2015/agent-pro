package agentdriver

import (
	"context"
	"errors"
	"strings"
	"time"
)

const CodexAdhocGuardFlag = "--require-codex-connection-v1"

// VerifyCodexAdhoc checks the actual re-exec binary, not the calling daemon's
// linked SDK version. The command is read-only and bounded; no provider starts.
func VerifyCodexAdhoc(ctx context.Context, d Driver) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd, err := CommandContext(ctx, d, "run", "--help")
	if err != nil {
		return err
	}
	output, err := cmd.Output()
	if err != nil {
		return errors.New("cannot inspect the agent-run receiver; update/restart the actual Marcus CLI")
	}
	if !strings.Contains(string(output), CodexAdhocGuardFlag) {
		return errors.New("the agent-run receiver does not support ad hoc Codex connections; update/restart the actual Marcus CLI")
	}
	return nil
}
