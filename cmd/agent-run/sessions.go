package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xhd2015/less-gen/flags"
)

const sessionsHelp = `
Usage: agent-run sessions [--json]
       agent-run sessions <runner>/<session_id> --print

Options:
  --json      list sessions as JSON (list mode only)
  --print     print formatted session events (required with <runner>/<session_id>)
  -h, --help  show help
`

func runSessions(args []string) error {
	var jsonFlag bool
	var printFlag bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		Bool("--print", &printFlag).
		Help("-h,--help", sessionsHelp).
		Parse(args)
	if err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		if len(remaining) != 1 {
			return fmt.Errorf("expected a single session reference runner/session_id")
		}
		if !printFlag {
			return fmt.Errorf("sessions with a session reference requires --print; see agent-run sessions --help for usage")
		}
		runner, sessionID, err := parseSessionRef(remaining[0])
		if err != nil {
			return err
		}
		return runSessionsPrint(store, runner, sessionID)
	}
	if printFlag {
		return fmt.Errorf("--print requires a session reference runner/session_id")
	}
	list, err := listAllSessions(store)
	if err != nil {
		return err
	}
	type item struct {
		Runner    string `json:"runner"`
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}
	var all []item
	for _, s := range list {
		all = append(all, item{Runner: s.Runner, SessionID: s.SessionID, Status: s.Status})
	}
	if jsonFlag {
		out := map[string]any{"sessions": all}
		enc := json.NewEncoder(os.Stdout)
		return enc.Encode(out)
	}
	for _, s := range all {
		fmt.Printf("%s/%s %s\n", s.Runner, s.SessionID, s.Status)
	}
	return nil
}