package main

import (
	"fmt"

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
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

	_, err = ttywatch.AttachWriter(entry.ListenAddr, sessionID, "attach")
	return err
}