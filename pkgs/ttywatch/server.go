package ttywatch

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

// ServeOptions configures an embedded PTY server session.
type ServeOptions struct {
	SessionID      string
	Command        []string
	Home           string
	RegistrySubdir string
	Cwd            string
	ExtraPaths     []string
	KeepAlive      bool
}

func serveKeepAlive(opts ServeOptions) bool {
	if opts.KeepAlive {
		return true
	}
	return strings.TrimSpace(os.Getenv(envTTYWatchKeepAlive)) == "1"
}

// ServeSession runs the embedded ptywrap HTTP server until the command exits.
func ServeSession(ctx context.Context, opts ServeOptions) error {
	if opts.SessionID == "" {
		return fmt.Errorf("serve: missing session id")
	}
	if len(opts.Command) == 0 {
		return fmt.Errorf("serve: missing command")
	}

	home := opts.Home
	if home == "" {
		var err error
		home, err = TTYWatchHome()
		if err != nil {
			return err
		}
	}
	cwd := opts.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	mgr := ptywrap.NewManager()
	if extraPaths := serveExtraPaths(opts.ExtraPaths); len(extraPaths) > 0 {
		mgr.Spawn.ExtraPaths = extraPaths
	}
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		<-serverDone
	}

	if _, err := mgr.CreateCommandWithID(opts.SessionID, "tty-watch", cwd, opts.Command); err != nil {
		shutdown()
		return err
	}

	entry := RegistryEntry{
		SessionID:  opts.SessionID,
		ListenAddr: listenAddr,
		PID:        os.Getpid(),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Command:    opts.Command,
		Cwd:        cwd,
	}
	cfg := serveRegistryConfig(home, opts.RegistrySubdir)
	if err := WriteRegistry(cfg, entry); err != nil {
		mgr.Remove(opts.SessionID)
		shutdown()
		return err
	}

	waitDone := make(chan struct{})
	go func() {
		_ = mgr.Wait(opts.SessionID)
		close(waitDone)
	}()

	select {
	case <-ctx.Done():
		mgr.Remove(opts.SessionID)
		shutdown()
		RemoveRegistryIfMatch(cfg, opts.SessionID, listenAddr, entry.PID)
		return ctx.Err()
	case <-waitDone:
	}

	if serveKeepAlive(opts) {
		// Keep the ptywrap server and registry reachable after the PTY child exits
		// so web/CLI attach can replay scrollback on finished keep-tty sessions.
		<-ctx.Done()
		shutdown()
		mgr.Remove(opts.SessionID)
		RemoveRegistryIfMatch(cfg, opts.SessionID, listenAddr, entry.PID)
		return ctx.Err()
	}

	const writerAttachGrace = 2 * time.Second
	time.Sleep(writerAttachGrace)
	shutdown()
	mgr.Remove(opts.SessionID)
	RemoveRegistryIfMatch(cfg, opts.SessionID, listenAddr, entry.PID)
	return nil
}