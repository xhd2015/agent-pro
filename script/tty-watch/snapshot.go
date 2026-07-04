package main

import (
	"fmt"
	"strings"
)

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
		RemoveRegistry(home, sessionID)
		return fmt.Errorf("tty-watch session %s not found", sessionID)
	}

	raw, err := readSnapshot(entry.ListenAddr, sessionID)
	if err != nil {
		return err
	}
	text := SanitizeForPrint(raw)
	text = strings.TrimRight(text, "\n")
	if text != "" {
		fmt.Println(text)
	}
	return nil
}