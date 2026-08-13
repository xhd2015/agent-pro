package agentrunapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	runMinWorkBeforeIdle = 3 * time.Second
	runIdleStable        = 1 * time.Second
)

func waitUntilDone(ctx context.Context, opts RunOpts, handle RunHandle) error {
	deadline := time.Now().Add(opts.Timeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = DefaultRunPollInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	openedAt := time.Now()
	var sawBusy bool
	var idleSince time.Time

	for {
		if resultFileReady(opts.ResultFile) {
			return nil
		}
		if store := handle.Store; store != nil && handle.SessionID != "" {
			meta, found, _ := loadRunSessionMeta(store, handle.SessionID)
			if found {
				probe := optsProbeOrDefault(opts)
				report, perr := probe(store, meta)
				if m2, ok2, _ := loadRunSessionMeta(store, handle.SessionID); ok2 {
					meta = m2
				}
				if perr == nil && report.RunnerExited != nil && *report.RunnerExited {
					return nil
				}
				if perr == nil && report.ResumeReady {
					return nil
				}
				// RunJSON: idle is not done. Grok often looks sendable between
				// tool calls; the result file (or exit) is the only completion.
				if strings.TrimSpace(opts.ResultFile) == "" {
					sendable, _, busy := probeTurnIdle(store, meta)
					switch {
					case busy:
						sawBusy = true
						idleSince = time.Time{}
					case sendable && (sawBusy || time.Since(openedAt) >= runMinWorkBeforeIdle):
						if idleSince.IsZero() {
							idleSince = time.Now()
						} else if time.Since(idleSince) >= runIdleStable {
							return nil
						}
					default:
						idleSince = time.Time{}
					}
				}
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for session %s after %s", handle.SessionID, opts.Timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func resultFileReady(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return json.Valid(bytes.TrimSpace(data))
}

func softExitAfterWait(opts RunOpts, handle RunHandle) {
	if !shouldSoftExit(opts) {
		return
	}
	store := handle.Store
	meta := agentstorage.SessionMeta{SessionID: handle.SessionID}
	if store != nil && handle.SessionID != "" {
		if m, ok, _ := loadRunSessionMeta(store, handle.SessionID); ok {
			meta = m
		}
	}
	runner := strings.TrimSpace(opts.AgentRunner)
	if runner == "" {
		runner = strings.TrimSpace(meta.Runner)
	}
	if opts.SoftExit != nil {
		opts.SoftExit(store, meta, runner)
		return
	}
	trySoftExit(store, meta, runner)
}

func shouldSoftExit(opts RunOpts) bool {
	if opts.OpenTerminal {
		return opts.ExitOnFinishTerminal
	}
	return !opts.KeepAliveDetached
}

func optsProbeOrDefault(opts RunOpts) ProbeFunc {
	// RunOpts has no Probe field; production wait always uses LifecycleProbe.
	return LifecycleProbe
}

func loadRunSessionMeta(store agentstorage.Store, sessionID string) (agentstorage.SessionMeta, bool, error) {
	if store == nil {
		return agentstorage.SessionMeta{}, false, nil
	}
	sess, err := store.GetSession(sessionID)
	if err != nil {
		msg := err.Error()
		if os.IsNotExist(err) || strings.Contains(msg, "not found") || strings.Contains(msg, "no such") {
			return agentstorage.SessionMeta{}, false, nil
		}
		return agentstorage.SessionMeta{}, false, nil
	}
	if sess == nil {
		return agentstorage.SessionMeta{}, false, nil
	}
	return sess.Meta, true, nil
}

// probeTurnIdle reports whether the TTY is sendable/idle, bound, and/or busy.
func probeTurnIdle(store agentstorage.Store, meta agentstorage.SessionMeta) (sendable, bound, busy bool) {
	bound = strings.TrimSpace(meta.RunnerSessionID) != ""
	home := ""
	if store != nil {
		home = store.Home()
	}

	var ttySess *agenttty.TTYSession
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID != "" && home != "" {
		if s, err := agenttty.ResolveByTerminalID(home, termID); err == nil {
			ttySess = s
		}
	}
	if ttySess == nil && home != "" {
		if s, err := agenttty.ResolveTerminalStatus(store, meta.Runner, meta.SessionID); err == nil {
			ttySess = s
		}
	}
	if ttySess == nil || !ttySess.TCPReachable {
		return false, bound, false
	}

	provider, ok := agenttty.Get(ttySess.RunnerID)
	if !ok || provider.CheckWritable == nil {
		return false, bound, false
	}
	scrollbackText, err := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, ttySess.TerminalSessionID)
	if err != nil || len(scrollbackText) == 0 {
		return false, bound, false
	}
	st := provider.CheckWritable([]byte(scrollbackText))
	if st.Ready {
		return true, bound, false
	}
	return false, bound, true
}

func trySoftExit(store agentstorage.Store, meta agentstorage.SessionMeta, runner string) {
	if store == nil {
		return
	}
	home := store.Home()
	var ttySess *agenttty.TTYSession
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID != "" {
		if s, err := agenttty.ResolveByTerminalID(home, termID); err == nil {
			ttySess = s
		}
	}
	if ttySess == nil {
		if s, err := agenttty.ResolveTerminalStatus(store, meta.Runner, meta.SessionID); err == nil {
			ttySess = s
		}
	}
	if ttySess == nil || !ttySess.TCPReachable {
		return
	}
	r := strings.TrimSpace(runner)
	if r == "" {
		r = strings.TrimSpace(meta.Runner)
	}
	if r == "" {
		r = ttySess.RunnerID
	}
	_ = agenttty.InjectMessage(ttySess.Registry.ListenAddr, ttySess.TerminalSessionID, r, "/exit", true)
}
