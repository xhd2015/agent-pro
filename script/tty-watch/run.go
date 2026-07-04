package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func parseRunArgs(args []string) (customSessionID string, commandArgs []string, err error) {
	i := 0
	for i < len(args) {
		if args[i] == "--session-id" {
			if customSessionID != "" {
				return "", nil, fmt.Errorf("run: duplicate --session-id flag")
			}
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("run: --session-id requires an argument")
			}
			customSessionID = args[i+1]
			i += 2
			continue
		}
		break
	}
	commandArgs = args[i:]
	if len(commandArgs) == 0 {
		return "", nil, fmt.Errorf("usage: tty-watch run [--session-id <id>] <command> [args...]")
	}
	return customSessionID, commandArgs, nil
}

func runRun(args []string) error {
	customSessionID, commandArgs, err := parseRunArgs(args)
	if err != nil {
		return err
	}

	home, err := TTYWatchHome()
	if err != nil {
		return err
	}

	var sessionID string
	var release func()
	if customSessionID != "" {
		release, err = ReserveCustomSessionID(home, customSessionID)
		if err != nil {
			return err
		}
		sessionID = customSessionID
	} else {
		sessionID, release, err = ReserveRegistrySessionID(home)
		if err != nil {
			return err
		}
	}
	defer release()

	argv := append([]string{serveSubcommand, sessionID}, commandArgs...)
	cmd := exec.Command(os.Args[0], argv...)
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	entry, err := waitForRegistryEntry(home, sessionID, 15*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		return err
	}
	release()

	debugLogf("run start session=%s listen=%s command=%v home=%s argv0=%s",
		sessionID, entry.ListenAddr, commandArgs, home, os.Args[0])

	// Do not wait on the serve child: on Ctrl-] detach the parent must exit
	// while the __serve__ process keeps the session alive.
	detached, err := attachWriter(entry.ListenAddr, sessionID, "screen")
	if err != nil {
		return err
	}
	if !detached {
		RemoveRegistryIfMatch(home, sessionID, entry.ListenAddr, entry.PID)
	}
	return nil
}