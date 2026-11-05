package agentrunapi

import (
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// prepareCodexConnection reserves a new target before a ForceNew callback can
// re-exec. Only non-secret references are stored. A child with just session-id
// can therefore recover the target without command-line JSON or env-spill.
func prepareCodexConnection(opts *Opts, meta agentstorage.SessionMeta, found bool) (agentstorage.SessionMeta, error) {
	runner := effectiveRunner(*opts, meta)
	if runner != "codex-tty" {
		if opts.CodexConnection != nil {
			return meta, fmt.Errorf("ad hoc Codex connection requires codex-tty (got %s)", runner)
		}
		return meta, nil
	}
	if found {
		if meta.CodexConnection == nil {
			if opts.CodexConnection != nil {
				opts.Model = meta.Model
				opts.ModelReasoningEffort = ""
			}
			opts.CodexConnection = nil
			return meta, nil
		}
		opts.CodexConnection = meta.CodexConnection.Clone()
	} else if opts.CodexConnection != nil {
		c := opts.CodexConnection.Clone()
		if err := c.ValidatePersistedEnv(opts.Env); err != nil {
			return meta, err
		}
		if err := c.Validate(); err != nil {
			return meta, err
		}
		meta = agentstorage.SessionMeta{Runner: runner, SessionID: opts.SessionID, Status: "prepared", InitialPrompt: strings.TrimSpace(opts.Prompt), Workspace: opts.WorkspaceDir, Model: c.Model, CodexConnection: c, PrependPaths: append([]string(nil), opts.PrependPaths...), Env: append([]string(nil), opts.Env...), AgentRunnerConfigHome: opts.RunnerConfigHome}
		if err := opts.Store.CreateSession(opts.SessionID, meta); err != nil {
			// A concurrent creator wins; never overwrite or silently change its target.
			winner, readErr := opts.Store.GetSession(opts.SessionID)
			if readErr != nil {
				return meta, err
			}
			return prepareCodexConnection(opts, winner.Meta, true)
		}
		opts.CodexConnection = c.Clone()
	}
	if opts.CodexConnection != nil {
		if err := opts.CodexConnection.ValidatePersistedEnv(opts.Env); err != nil {
			return meta, err
		}
		if err := opts.CodexConnection.Validate(); err != nil {
			return meta, err
		}
		opts.Model = opts.CodexConnection.Model
		opts.ModelReasoningEffort = opts.CodexConnection.Reasoning.Effort
	}
	return meta, nil
}
