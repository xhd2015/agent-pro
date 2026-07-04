package main

import (
	"fmt"
)

func runAttach(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("attach: requires <session-id>")
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

	_, err = attachWriter(entry.ListenAddr, sessionID, "attach")
	return err
}