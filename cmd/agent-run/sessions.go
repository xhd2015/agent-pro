package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/less-gen/flags"
)

const sessionsHelp = `
Usage: agent-run sessions [--json] [--limit N]
       agent-run sessions --clear
       agent-run sessions <session_id> --print
       agent-run sessions --print --grok-session-id ID

Options:
  --json               list sessions as JSON (list mode only)
  --limit N            max sessions to show (default 10; 0 = all)
  --clear              delete all stored sessions under agent-run home
  --print              print formatted session events (required with <session_id> or --grok-session-id)
  --grok-session-id ID resolve print target by provider runner_session_id (meta.runner grok|grok-tty);
                       mutually exclusive with positional <session_id>; requires --print
  -h, --help           show help
`

const defaultSessionsListLimit = 10

func runSessions(args []string) error {
	var jsonFlag bool
	var printFlag bool
	var clearFlag bool
	var grokSessionID *string
	limit := defaultSessionsListLimit
	remaining, err := flags.Bool("--json", &jsonFlag).
		Bool("--print", &printFlag).
		Bool("--clear", &clearFlag).
		Int("--limit", &limit).
		String("--grok-session-id", &grokSessionID).
		Help("-h,--help", sessionsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}
	if clearFlag {
		if len(remaining) > 0 {
			return fmt.Errorf("--clear cannot be used with a session reference")
		}
		if printFlag {
			return fmt.Errorf("--clear cannot be used with --print")
		}
		if jsonFlag {
			return fmt.Errorf("--clear cannot be used with --json")
		}
		if grokSessionID != nil {
			return fmt.Errorf("--clear cannot be used with --grok-session-id")
		}
		return store.ClearAllSessions()
	}
	if grokSessionID != nil {
		if len(remaining) > 0 {
			return fmt.Errorf("--grok-session-id and positional <session-id> are mutually exclusive; cannot use both")
		}
		if !printFlag {
			return fmt.Errorf("--grok-session-id requires --print")
		}
		meta, err := resolveSessionMetaByGrokSessionID(store, *grokSessionID)
		if err != nil {
			return err
		}
		return runSessionsPrint(store, meta.SessionID)
	}
	if len(remaining) > 0 {
		if len(remaining) != 1 {
			return fmt.Errorf("expected a single session id")
		}
		if !printFlag {
			return fmt.Errorf("sessions with a session id requires --print; see agent-run sessions --help for usage")
		}
		sessionID, err := parseBareSessionID(remaining[0])
		if err != nil {
			return err
		}
		return runSessionsPrint(store, sessionID)
	}
	if printFlag {
		return fmt.Errorf("--print requires a session id or --grok-session-id")
	}
	list, err := listAllSessions(store)
	if err != nil {
		return err
	}
	sortSessionsNewestFirst(list)
	total := len(list)
	list = applySessionLimit(list, limit)

	type item struct {
		Runner    string `json:"runner"`
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at,omitempty"`
		UpdatedAt string `json:"updated_at,omitempty"`
	}
	all := make([]item, 0, len(list))
	for _, s := range list {
		all = append(all, item{
			Runner:    s.Runner,
			SessionID: s.SessionID,
			Status:    s.Status,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}
	if jsonFlag {
		out := map[string]any{"sessions": all}
		enc := json.NewEncoder(os.Stdout)
		return enc.Encode(out)
	}
	return printSessionsListHuman(list, total, limit)
}

func printSessionsListHuman(list []agentstorage.SessionMeta, total, limit int) error {
	now := time.Now()
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "SESSION_ID\tRUNNER\tSTATUS\tUPDATED")
	for _, s := range list {
		ts := s.UpdatedAt
		if ts == "" {
			ts = s.CreatedAt
		}
		updated := agentstorage.FormatRelativeAge(now, parseSessionTime(ts))
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.SessionID, s.Runner, s.Status, updated)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if limit > 0 && total > limit {
		fmt.Fprintf(os.Stdout, "(showing %d of %d; use --limit N or --limit 0 for all)\n", limit, total)
	}
	return nil
}
