package main

import (
	"fmt"
	"os"
)

func runWatch(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("watch: requires <session-id>")
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
		RemoveRegistry(home, sessionID)
		return fmt.Errorf("tty-watch session %s not found", sessionID)
	}

	return streamObserver(entry.ListenAddr, sessionID, os.Stdout)
}