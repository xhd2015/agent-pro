package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func runRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tty-watch run <command> [args...]")
	}

	home, err := TTYWatchHome()
	if err != nil {
		return err
	}

	sessionID, release, err := ReserveRegistrySessionID(home)
	if err != nil {
		return err
	}
	defer release()

	argv := append([]string{serveSubcommand, sessionID}, args...)
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
		sessionID, entry.ListenAddr, args, home, os.Args[0])

	// Do not wait on the serve child: on Ctrl-] detach the parent must exit
	// while the __serve__ process keeps the session alive.
	detached, err := attachWriter(entry.ListenAddr, sessionID)
	if err != nil {
		return err
	}
	if !detached {
		RemoveRegistry(home, sessionID)
	}
	return nil
}