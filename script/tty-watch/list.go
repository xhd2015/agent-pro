package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

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
	for _, entry := range entries {
		uptime := formatUptime(entry.CreatedAt)
		cmdLine := strings.Join(entry.Command, " ")
		if cmdLine == "" {
			cmdLine = "(shell)"
		}
		fmt.Fprintf(os.Stdout, "%s  %s  %s\n", entry.SessionID, cmdLine, uptime)
	}
	return nil
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