package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

type listTableRow struct {
	sessionID string
	command   string
	uptime    string
	watch     int
	attached  int
}

func runList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("list: unexpected arguments %v", args)
	}
	home, err := TTYWatchHome()
	if err != nil {
		return err
	}

	entries, err := ListRegistryEntries(home, true)
	if err != nil {
		return err
	}
	rows := make([]listTableRow, 0, len(entries))
	for _, entry := range entries {
		cmdLine := strings.Join(entry.Command, " ")
		if cmdLine == "" {
			cmdLine = "(shell)"
		}
		watch, attached := fetchSessionClientCounts(entry.ListenAddr, entry.SessionID)
		rows = append(rows, listTableRow{
			sessionID: entry.SessionID,
			command:   truncateCommand(cmdLine, 64),
			uptime:    formatUptime(entry.CreatedAt),
			watch:     watch,
			attached:  attached,
		})
	}
	printListTable(rows)
	return nil
}

func printListTable(rows []listTableRow) {
	if len(rows) == 0 {
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SESSION\tUPTIME\tWATCH\tATTACHED\tCOMMAND")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
			row.sessionID, row.uptime, row.watch, row.attached, row.command)
	}
	_ = w.Flush()
}

func truncateCommand(cmd string, max int) string {
	if len(cmd) <= max {
		return cmd
	}
	const ellipsis = "..."
	if max <= len(ellipsis) {
		return cmd[:max]
	}
	return cmd[:max-len(ellipsis)] + ellipsis
}

func formatUptime(createdAt string) string {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
}