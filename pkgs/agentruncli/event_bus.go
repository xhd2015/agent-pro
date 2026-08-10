package agentruncli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/eventbus"
)

// EventBusOpts configures best-effort publish of agent.tty.started events.
// Empty URL disables publish (no HTTP, no warning).
// PublishHook is the L2 inject seam; when nil, production uses eventbus.NewPublisher.
type EventBusOpts struct {
	URL   string
	Token string
	// PublishHook, when set, replaces the production HTTP publisher (tests).
	// Signature matches wire type/source + payload only (no eventbus import in harness).
	PublishHook func(ctx context.Context, eventType, source string, payload json.RawMessage) error
	// WarnWriter receives best-effort failure lines (prefix "warning:"). Nil → os.Stderr.
	WarnWriter io.Writer
}

// NotifyTTYStarted publishes type=agent.tty.started source=agent-run with payload
// {session_id, runner, workspace}. Best-effort: empty URL is a no-op; publish
// errors write a "warning:" line and never return to the caller.
func NotifyTTYStarted(opts EventBusOpts, sessionID, runner, workspace string) {
	url := strings.TrimSpace(opts.URL)
	if url == "" {
		return
	}

	payload, err := json.Marshal(map[string]string{
		"session_id": sessionID,
		"runner":     runner,
		"workspace":  workspace,
	})
	if err != nil {
		warnEventBus(opts.WarnWriter, err)
		return
	}

	ctx := context.Background()
	if opts.PublishHook != nil {
		err = opts.PublishHook(ctx, eventbus.TypeAgentTTYStarted, eventbus.SourceAgentRun, payload)
	} else {
		pub := eventbus.NewPublisher(url, eventbus.WithToken(opts.Token))
		err = pub.Publish(ctx, eventbus.Event{
			Type:    eventbus.TypeAgentTTYStarted,
			Source:  eventbus.SourceAgentRun,
			Payload: payload,
		})
	}
	if err != nil {
		warnEventBus(opts.WarnWriter, err)
	}
}

// NotifyOnOpenPath dispatches open-path publish policy:
//   - "new-terminal" — after successful ForceNew / open-profile → NotifyTTYStarted once
//   - "send" — live send path → no-op (never publishes)
func NotifyOnOpenPath(kind string, opts EventBusOpts, sessionID, runner, workspace string) {
	switch kind {
	case "new-terminal":
		NotifyTTYStarted(opts, sessionID, runner, workspace)
	case "send":
		// Live send never publishes agent.tty.started.
		return
	default:
		// Unknown kinds are ignored (best-effort policy helper).
		return
	}
}

// AppendEventBusFlags appends --event-bus-url / --event-bus-token to args when
// URL is non-empty. Empty URL returns a copy of args unchanged (no flags added).
// Token is appended only when non-empty.
func AppendEventBusFlags(args []string, url, token string) []string {
	out := append([]string(nil), args...)
	if strings.TrimSpace(url) == "" {
		return out
	}
	out = append(out, "--event-bus-url", url)
	if token != "" {
		out = append(out, "--event-bus-token", token)
	}
	return out
}

// RunHelpText returns the `agent-run run -h` help body (same text flags.Help uses).
// Always ends with a trailing newline.
func RunHelpText() string {
	return runHelp
}

func warnEventBus(w io.Writer, err error) {
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintf(w, "warning: event bus publish failed: %v\n", err)
}
