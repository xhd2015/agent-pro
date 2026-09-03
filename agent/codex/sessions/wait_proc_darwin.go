//go:build darwin

package sessions

import (
	"context"
	"fmt"

	"golang.org/x/sys/unix"
)

// waitProcessesExit blocks until any of pids exits, or ctx is cancelled.
// Uses kqueue EVFILT_PROC / NOTE_EXIT (event-driven; no PID polling).
func waitProcessesExit(ctx context.Context, pids []int) error {
	uniq := uniquePositivePIDs(pids)
	if len(uniq) == 0 {
		<-ctx.Done()
		return ctx.Err()
	}
	// If every pid is already gone, abandon immediately (registering
	// EVFILT_PROC on a dead pid can fail or never fire NOTE_EXIT).
	if allPIDsDead(uniq) {
		return nil
	}

	kq, err := unix.Kqueue()
	if err != nil {
		return fmt.Errorf("kqueue: %w", err)
	}
	defer unix.Close(kq)

	// Register each pid individually so one failure does not drop the rest
	// (some ancestors like iTerm helpers may reject EVFILT_PROC).
	registered := 0
	for _, pid := range uniq {
		if !pidAlive(pid) {
			return nil // exited between snapshot and register
		}
		ev := []unix.Kevent_t{{
			Ident:  uint64(pid),
			Filter: unix.EVFILT_PROC,
			Flags:  unix.EV_ADD | unix.EV_ENABLE,
			Fflags: unix.NOTE_EXIT,
		}}
		if _, err := unix.Kevent(kq, ev, nil, nil); err != nil {
			continue
		}
		registered++
	}
	if registered == 0 {
		return nil
	}
	// User event to wake kevent on ctx cancel (ident distinct from real pids).
	const wakeIdent = ^uint64(0) // max uint64
	if _, err := unix.Kevent(kq, []unix.Kevent_t{{
		Ident:  wakeIdent,
		Filter: unix.EVFILT_USER,
		Flags:  unix.EV_ADD | unix.EV_CLEAR,
		Fflags: unix.NOTE_FFNOP,
	}}, nil, nil); err != nil {
		return fmt.Errorf("kevent register wake: %w", err)
	}

	abort := make(chan struct{})
	defer close(abort)
	go func() {
		select {
		case <-ctx.Done():
			_, _ = unix.Kevent(kq, []unix.Kevent_t{{
				Ident:  wakeIdent,
				Filter: unix.EVFILT_USER,
				Flags:  unix.EV_ADD | unix.EV_CLEAR,
				Fflags: unix.NOTE_TRIGGER,
			}}, nil, nil)
		case <-abort:
		}
	}()

	events := make([]unix.Kevent_t, len(uniq)+1)
	for {
		n, err := unix.Kevent(kq, nil, events, nil)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("kevent wait: %w", err)
		}
		for i := 0; i < n; i++ {
			ev := events[i]
			if ev.Filter == unix.EVFILT_PROC && ev.Fflags&unix.NOTE_EXIT != 0 {
				return nil
			}
			if ev.Filter == unix.EVFILT_USER && ctx.Err() != nil {
				return ctx.Err()
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
