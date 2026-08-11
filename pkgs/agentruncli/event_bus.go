package agentruncli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/dot-pkgs/go-pkgs/eventbus"
)

// Wire vocabulary for agent.tty.restarted / reason (kept as string literals so
// consumers compile against published eventbus that may not yet export them;
// values match go-pkgs/eventbus constants).
const (
	typeAgentTTYRestarted = "agent.tty.restarted"
	reasonTTYNew          = "new"
	reasonTTYFollowup     = "followup"
	reasonTTYResume       = "resume"
)

// EventBusOpts configures best-effort publish of agent.tty.* events.
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
	// AlreadyNotified, when non-nil, is an at-most-once guard for agent.tty.started only
	// (ForceNew NotifyOnOpenPath + library WireOnTTYStarted). Never used for restarted.
	// If *true, NotifyTTYStarted skips. After publishing with a non-empty URL, sets *true.
	// Empty URL does not set the flag.
	AlreadyNotified *bool
}

// NotifyTTYStarted publishes type=agent.tty.started source=agent-run with payload
// {session_id, runner, workspace, reason}. reason defaults to "new".
// Best-effort: empty URL is a no-op; publish errors write a "warning:" line.
// When AlreadyNotified is non-nil and *true, skips; after a non-empty-URL publish
// attempt, sets *true (empty URL does not set).
func NotifyTTYStarted(opts EventBusOpts, sessionID, runner, workspace string) {
	NotifyTTYStartedReason(opts, sessionID, runner, workspace, reasonTTYNew)
}

// NotifyTTYStartedReason is NotifyTTYStarted with an explicit reason field.
func NotifyTTYStartedReason(opts EventBusOpts, sessionID, runner, workspace, reason string) {
	if opts.AlreadyNotified != nil && *opts.AlreadyNotified {
		return
	}
	url := strings.TrimSpace(opts.URL)
	if url == "" {
		return
	}
	if opts.AlreadyNotified != nil {
		*opts.AlreadyNotified = true
	}
	if strings.TrimSpace(reason) == "" {
		reason = reasonTTYNew
	}
	publishTTYEvent(opts, eventbus.TypeAgentTTYStarted, sessionID, runner, workspace, reason)
}

// NotifyTTYRestarted publishes type=agent.tty.restarted source=agent-run with
// payload {session_id, runner, workspace, reason}. reason should be followup|resume.
// No AlreadyNotified guard — each restart may re-open local attach.
func NotifyTTYRestarted(opts EventBusOpts, sessionID, runner, workspace, reason string) {
	url := strings.TrimSpace(opts.URL)
	if url == "" {
		return
	}
	if strings.TrimSpace(reason) == "" {
		reason = reasonTTYFollowup
	}
	publishTTYEvent(opts, typeAgentTTYRestarted, sessionID, runner, workspace, reason)
}

func publishTTYEvent(opts EventBusOpts, eventType, sessionID, runner, workspace, reason string) {
	payload, err := json.Marshal(map[string]string{
		"session_id": sessionID,
		"runner":     runner,
		"workspace":  workspace,
		"reason":     reason,
	})
	if err != nil {
		warnEventBus(opts.WarnWriter, err)
		return
	}
	ctx := context.Background()
	if opts.PublishHook != nil {
		err = opts.PublishHook(ctx, eventType, eventbus.SourceAgentRun, payload)
	} else {
		pub := eventbus.NewPublisher(opts.URL, eventbus.WithToken(opts.Token))
		err = pub.Publish(ctx, eventbus.Event{
			Type:    eventType,
			Source:  eventbus.SourceAgentRun,
			Payload: payload,
		})
	}
	if err != nil {
		warnEventBus(opts.WarnWriter, err)
	}
}

// WireOnTTYStarted returns a callback for agentrunapi.Opts.OnTTYStarted.
// Empty URL → nil. Uses info.Reason when set, else "new".
func WireOnTTYStarted(opts EventBusOpts) func(agentrunapi.TTYStartedInfo) {
	if strings.TrimSpace(opts.URL) == "" {
		return nil
	}
	return func(info agentrunapi.TTYStartedInfo) {
		reason := info.Reason
		if reason == "" {
			reason = reasonTTYNew
		}
		NotifyTTYStartedReason(opts, info.SessionID, info.Runner, info.Workspace, reason)
	}
}

// WireOnTTYRestarted returns a callback for agentrunapi.Opts.OnTTYRestarted.
// Empty URL → nil. Does not share AlreadyNotified with started.
func WireOnTTYRestarted(opts EventBusOpts) func(agentrunapi.TTYStartedInfo) {
	if strings.TrimSpace(opts.URL) == "" {
		return nil
	}
	return func(info agentrunapi.TTYStartedInfo) {
		reason := info.Reason
		if reason == "" {
			reason = reasonTTYFollowup
		}
		NotifyTTYRestarted(opts, info.SessionID, info.Runner, info.Workspace, reason)
	}
}

// NotifyOnOpenPath dispatches open-path publish policy:
//   - "new-terminal" — after successful ForceNew / open-profile → NotifyTTYStarted once
//   - "send" — live send path → no-op here (library OnTTYRestarted handles follow-up)
func NotifyOnOpenPath(kind string, opts EventBusOpts, sessionID, runner, workspace string) {
	switch kind {
	case "new-terminal":
		NotifyTTYStarted(opts, sessionID, runner, workspace)
	case "send":
		return
	default:
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
