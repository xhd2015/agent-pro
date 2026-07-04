package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

const serveSubcommand = "__serve__"

func runServeSession(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("%s: missing session id or command", serveSubcommand)
	}
	sessionID := args[0]
	command := args[1:]

	home, err := TTYWatchHome()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	mgr := ptywrap.NewManager()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	listenAddr := ln.Addr().String()

	mux := http.NewServeMux()
	registerPrepareInjectAPI(mux, mgr)
	ptywrap.RegisterAPIWithManager(mux, mgr)
	srv := &http.Server{Handler: mux}
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- srv.Serve(ln)
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		<-serverDone
	}

	if _, err := mgr.CreateCommandWithID(sessionID, "tty-watch", cwd, command); err != nil {
		shutdown()
		return err
	}

	entry := RegistryEntry{
		SessionID:  sessionID,
		ListenAddr: listenAddr,
		PID:        os.Getpid(),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Command:    command,
		Cwd:        cwd,
	}
	if err := WriteRegistry(home, entry); err != nil {
		mgr.Remove(sessionID)
		shutdown()
		return err
	}

	_ = mgr.Wait(sessionID)
	// Short-lived commands can exit before the parent run process attaches.
	// Keep the session registered and listener open briefly so writer attach
	// can receive scrollback instead of dialing a removed session (which would
	// spawn a fresh interactive shell and hang).
	const writerAttachGrace = 2 * time.Second
	time.Sleep(writerAttachGrace)
	shutdown()
	mgr.Remove(sessionID)
	RemoveRegistryIfMatch(home, sessionID, listenAddr, entry.PID)
	return nil
}