package main

import "fmt"

func runSnapshot(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("snapshot: requires <session-id>")
	}
	sessionID := args[0]

	home, err := TTYWatchHome()
	if err != nil {
		return err
	}
	entry, err := ReadRegistry(home, sessionID)
	if err != nil {
		return err
	}
	if !tcpReachable(entry.ListenAddr) {
		RemoveRegistryIfMatch(home, sessionID, entry.ListenAddr, entry.PID)
		return fmt.Errorf("tty-watch session %s not found", sessionID)
	}

	frame, scrollback, cols, rows, err := readSnapshot(entry.ListenAddr, sessionID)
	if err != nil {
		return err
	}
	text := renderSnapshotOutput(frame, scrollback, cols, rows)
	if text != "" {
		fmt.Println(text)
	}
	return nil
}