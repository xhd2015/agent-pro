package agentui

import (
	"fmt"
	"strings"
)

func resolveCodexConnection(opts *RunOptions, id, runner string) error {
	if runner != "codex-tty" {
		if opts.CodexConnection != nil {
			return fmt.Errorf("ad hoc Codex connection requires codex-tty")
		}
		return nil
	}
	if sess, err := opts.Store.GetSession(id); err == nil {
		if sess.Meta.CodexConnection == nil {
			if opts.CodexConnection != nil {
				opts.Model = sess.Meta.Model
				opts.ModelReasoningEffort = ""
			}
			opts.CodexConnection = nil
		} else {
			opts.CodexConnection = sess.Meta.CodexConnection.Clone()
		}
	} else if !strings.Contains(err.Error(), "session not found") {
		return err
	}
	if c := opts.CodexConnection; c != nil {
		if err := c.ValidatePersistedEnv(opts.Env); err != nil {
			return err
		}
		if err := c.Validate(); err != nil {
			return err
		}
		opts.Model = c.Model
		opts.ModelReasoningEffort = c.Reasoning.Effort
	}
	return nil
}
