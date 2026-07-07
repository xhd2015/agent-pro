package agentsend

import (
	"fmt"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// SendToAgentSession resolves an agent session to a live TTY, enqueues a message,
// optionally starts the drainer, and waits per waitOpts.
func SendToAgentSession(store agentstorage.Store, runner, agentSessionID, message string, waitOpts WaitOptions) (msgID string, err error) {
	ttySess, err := agenttty.ResolveByAgentSession(store, runner, agentSessionID)
	if err != nil {
		return "", err
	}
	if !ttySess.TCPReachable {
		return "", fmt.Errorf("terminal unreachable at %s", ttySess.Registry.ListenAddr)
	}

	provider, ok := agenttty.Get(runner)
	if !ok {
		return "", fmt.Errorf("unknown tty runner: %s", runner)
	}

	sess := Session{
		Home:              store.Home(),
		Runner:            runner,
		TerminalSessionID: ttySess.TerminalSessionID,
		ListenAddr:        ttySess.Registry.ListenAddr,
	}

	enqueuedAt := time.Now()
	if waitOpts.EnqueuedAt.IsZero() {
		waitOpts.EnqueuedAt = enqueuedAt
	}

	id, err := Enqueue(store.Home(), sess, message)
	if err != nil {
		return "", err
	}

	if waitOpts.StartDrainer {
		StartDrainer(store.Home(), sess, provider)
	}

	if err := WaitForDelivery(store.Home(), sess, id, waitOpts); err != nil {
		return id, err
	}
	return id, nil
}