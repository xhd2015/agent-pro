package agentsend

import (
	"context"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const sessionDrainerIdlePoll = 150 * time.Millisecond

// StartDrainer launches a background loop that elects a drainer via flock and
// delivers pending messages FIFO when the terminal becomes writable.
// The loop exits when the queue is empty or a drain step fails unrecoverably.
// Prefer RunSessionDrainer for session-owned consumers that must survive empty queues.
func StartDrainer(home string, sess Session, provider agenttty.Provider) {
	go func() {
		for {
			if !drainStep(home, sess, provider) {
				return
			}
		}
	}()
}

// StartSessionDrainer starts RunSessionDrainer in a new goroutine.
func StartSessionDrainer(ctx context.Context, home string, sess Session, provider agenttty.Provider) {
	go RunSessionDrainer(ctx, home, sess, provider)
}

// RunSessionDrainer is a durable session-owned queue consumer. It reuses drainStep
// (flock + WaitUntilWritable + SendMessage + FIFO dequeue) and keeps polling while
// the queue is empty until ctx is cancelled. Intended to run inside the live TTY
// serve process so --no-wait enqueues deliver without a blocking CLI send.
func RunSessionDrainer(ctx context.Context, home string, sess Session, provider agenttty.Provider) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !queueHasWork(home, sess) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(sessionDrainerIdlePoll):
			}
			continue
		}

		more := drainStep(home, sess, provider)
		if more {
			continue
		}
		// Empty after delivery, lock failure, or transient error — brief backoff.
		select {
		case <-ctx.Done():
			return
		case <-time.After(sessionDrainerIdlePoll):
		}
	}
}

func drainStep(home string, sess Session, provider agenttty.Provider) bool {
	path := queuePath(home, sess.Runner, sess.TerminalSessionID)

	release, err := acquireLock(home, sess.Runner, sess.TerminalSessionID)
	if err != nil {
		return false
	}
	head, err := loadHead(path)
	if err != nil || head == nil {
		release()
		return false
	}
	release()

	writable := agenttty.WaitUntilWritable(provider, sess.ListenAddr, sess.TerminalSessionID, 0)
	if !writable.Ready {
		time.Sleep(200 * time.Millisecond)
		return queueHasWork(home, sess)
	}

	release, err = acquireLock(home, sess.Runner, sess.TerminalSessionID)
	if err != nil {
		return false
	}
	defer release()

	current, err := loadHead(path)
	if err != nil || current == nil {
		return false
	}
	if current.ID != head.ID {
		return true
	}

	payload := current.Text
	submit := !current.NoSubmit
	// Grok prompt box expects a trailing newline before Enter when auto-submitting.
	// When NoSubmit, leave text as-is (no forced \n) and do not send trailing Enter.
	if submit && sess.Runner == "grok-tty" && !strings.Contains(payload, "\n") {
		payload += "\n"
	}
	// codex-tty: one-shot text+\r only types into the composer on real Codex TUI;
	// InjectMessage types then sends a separate Enter.
	if err := agenttty.InjectMessage(sess.ListenAddr, sess.TerminalSessionID, sess.Runner, payload, submit); err != nil {
		return true
	}
	if err := dequeueHead(path); err != nil {
		return true
	}
	return queueHasWork(home, sess)
}

func queueHasWork(home string, sess Session) bool {
	entries, err := readEntries(queuePath(home, sess.Runner, sess.TerminalSessionID))
	if err != nil {
		return false
	}
	return len(entries) > 0
}
