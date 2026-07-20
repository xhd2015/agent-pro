package view

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// PrintEvents writes human-readable events to w using agent/event/print
// FormatState for assistant/tool/think streams. User messages are rendered
// as separate USER blocks so they are not coalesced into ASSISTANT output.
func PrintEvents(w io.Writer, events []types.AgentEvent) int {
	if w == nil {
		w = os.Stdout
	}
	n := 0
	var state print.FormatState
	for _, ev := range events {
		if ev.Type == types.ActionMessage && ev.Role == "user" {
			state.Flush()
			text := strings.TrimSpace(ev.Text)
			if text == "" {
				continue
			}
			n++
			fmt.Fprintf(w, "[%d]  👤   USER\n", n)
			for _, line := range strings.Split(text, "\n") {
				fmt.Fprintf(w, "  %s\n", line)
			}
			fmt.Fprintln(w)
			continue
		}
		line, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		header, body, isMsg := state.FormatLine(string(line))
		if header == "" && body == "" && !isMsg {
			continue
		}
		if isMsg {
			if header != "" {
				n++
				fmt.Fprintf(w, "[%d]  %s\n", n, header)
			}
			if body != "" {
				fmt.Fprint(w, body)
				if body[len(body)-1] != '\n' {
					fmt.Fprintln(w)
				}
			}
		} else {
			n++
			fmt.Fprintf(w, "[%d]  %s\n", n, header)
			if body != "" {
				fmt.Fprint(w, body)
				if len(body) > 0 && body[len(body)-1] != '\n' {
					fmt.Fprintln(w)
				}
			}
		}
	}
	state.Flush()
	return n
}

// PrintSnapshot prints a header and all events currently in the viewer.
func PrintSnapshot(w io.Writer, v *Viewer) int {
	if w == nil {
		w = os.Stdout
	}
	events := v.Events()
	id := ""
	if v.Info != nil {
		id = v.Info.ID
	}
	fmt.Fprintf(w, "Session: %s\n", id)
	if v.Info != nil && v.Info.CWD != "" {
		fmt.Fprintf(w, "Workspace: %s\n", v.Info.CWD)
	}
	fmt.Fprintf(w, "Events: %d\n\n", len(events))
	if len(events) == 0 {
		fmt.Fprintln(w, "(no events)")
		return 0
	}
	return PrintEvents(w, events)
}

// PrintFollow prints the current snapshot, then streams new events until ctx is done.
func PrintFollow(ctx context.Context, w io.Writer, v *Viewer) error {
	if w == nil {
		w = os.Stdout
	}
	PrintSnapshot(w, v)

	ch, cancelSub := v.Subscribe()
	defer cancelSub()

	// Track how many events already printed.
	printed := len(v.Events())
	fmt.Fprintln(w, "Following new events (Ctrl+C to stop)...")

	followCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- v.Follow(followCtx)
	}()

	for {
		select {
		case <-followCtx.Done():
			// Drain final events after follow stops.
			events := v.Events()
			if len(events) > printed {
				PrintEvents(w, events[printed:])
			}
			<-errCh
			fmt.Fprintln(w, "\nDone.")
			return nil
		case <-ch:
			events := v.Events()
			if len(events) > printed {
				PrintEvents(w, events[printed:])
				printed = len(events)
			}
		case err := <-errCh:
			events := v.Events()
			if len(events) > printed {
				PrintEvents(w, events[printed:])
			}
			if err != nil && err != context.Canceled {
				return err
			}
			fmt.Fprintln(w, "\nDone.")
			return nil
		}
	}
}
