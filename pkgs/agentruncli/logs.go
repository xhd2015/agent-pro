package agentruncli

import (
	"fmt"
	"io"
	"os"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/less-gen/flags"
)

const logsHelp = `
Usage: agent-run logs [--json] <session-id>

Print session-scoped runtime errors. A successful session may have no logs.

Options:
  --json       print logs.jsonl unchanged
  -h, --help   show help
`

// RunLogs implements `agent-run logs` with injected storage and output.
func RunLogs(args []string, store agentstorage.Store, stdout io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	var jsonFlag bool
	var wantHelp bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		Bool("-h,--help", &wantHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if wantHelp {
		writeCLIHelp(stdout, logsHelp)
		return nil
	}
	if len(remaining) != 1 {
		return fmt.Errorf("logs requires a single session id")
	}
	sessionID, err := parseBareSessionID(remaining[0])
	if err != nil {
		return err
	}
	if _, err := store.GetSession(sessionID); err != nil {
		return err
	}
	records, raw, err := agentstorage.ReadLogs(store.Home(), sessionID)
	if err != nil {
		return err
	}
	if jsonFlag {
		_, err := stdout.Write(raw)
		return err
	}
	if len(records) == 0 {
		_, err := fmt.Fprintf(stdout, "No logs recorded for %s.\n", sessionID)
		return err
	}
	for _, record := range records {
		if _, err := fmt.Fprintf(stdout, "%s %s %s: %s\n", record.Timestamp, record.Level, record.Component, record.Message); err != nil {
			return err
		}
	}
	return nil
}
