package agentruncli

import (
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/less-gen/flags"
)

const focusHelp = `
Usage: agent-run focus <session-id> [OPTIONS]
       agent-run focus --session-id ID [OPTIONS]

Focus the iTerm2 window/tab that hosts an agent-run session by resolving
session meta → registry PID → process-tree TTYs → iTerm FindByTTY.

Arguments:
  <session-id>          agent-run session id (or use --session-id)

Options:
  --session-id ID       session id (alternative to positional)
  --index N             0-based candidate index when multiple iTerm matches
  --dry-run             resolve and print chosen candidate without focusing
  -h, --help            show help
`

// RunFocus implements `agent-run focus` with injectable writers for L2 tests.
// CLI surface: focus <session-id> [--index N] [--dry-run] [--session-id ID] [-h]
func RunFocus(args []string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	var dryRun bool
	var index *int
	var sessionIDFlag *string
	var wantHelp bool

	// Manual help so output goes to the injectable stdout (flags.Help prints to os.Stdout).
	remaining, err := flags.Bool("--dry-run", &dryRun).
		Int("--index", &index).
		String("--session-id", &sessionIDFlag).
		Bool("-h,--help", &wantHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if wantHelp {
		txt := strings.TrimPrefix(focusHelp, "\n")
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, _ = io.WriteString(stdout, txt)
		return nil
	}

	sessionID := ""
	if sessionIDFlag != nil {
		sessionID = strings.TrimSpace(*sessionIDFlag)
	}
	if len(remaining) > 0 {
		pos := strings.TrimSpace(remaining[0])
		if sessionID != "" && pos != "" && sessionID != pos {
			return fmt.Errorf("--session-id and positional <session-id> are mutually exclusive; cannot use both")
		}
		if sessionID == "" {
			sessionID = pos
		}
		if len(remaining) > 1 {
			return fmt.Errorf("focus accepts at most one positional <session-id>")
		}
	}
	if sessionID == "" {
		return fmt.Errorf("focus requires <session-id> (or --session-id)")
	}

	store, err := openStore()
	if err != nil {
		return err
	}

	opts := agentrunapi.FocusOpts{
		Store:     store,
		SessionID: sessionID,
		Index:     index,
		DryRun:    dryRun,
	}
	chosen, err := agentrunapi.FocusSession(opts)
	if err != nil {
		// Surface multi/none guidance on stderr with Error: prefix (DOCTEST policy).
		msg := err.Error()
		fmt.Fprintf(stderr, "Error: %s\n", msg)
		return err
	}

	if dryRun {
		fmt.Fprintf(stdout, "dry-run: would focus index=%d window=%s tab=%d tty=%s session=%s\n",
			chosen.Index, chosen.Ref.WindowID, chosen.Ref.TabIndex, chosen.TTY, chosen.Ref.SessionID)
		return nil
	}
	fmt.Fprintf(stdout, "focused index=%d window=%s tab=%d tty=%s\n",
		chosen.Index, chosen.Ref.WindowID, chosen.Ref.TabIndex, chosen.TTY)
	return nil
}
