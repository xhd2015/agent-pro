package agentsend

import (
	"fmt"
	"time"
)

// WaitMode controls how long the caller blocks after enqueue.
type WaitMode int

const (
	WaitDefault WaitMode = iota
	WaitNoWait
	WaitMaxWait
)

// WaitOptions configures post-enqueue behavior.
type WaitOptions struct {
	Mode         WaitMode
	MaxWait      time.Duration
	EnqueuedAt   time.Time
	StartDrainer bool
}

// WaitForDelivery polls until msgID is absent from the queue or a max-wait
// deadline expires.
func WaitForDelivery(home string, sess Session, msgID string, opts WaitOptions) error {
	if opts.Mode == WaitNoWait {
		return nil
	}

	deadline := time.Time{}
	if opts.Mode == WaitMaxWait {
		start := opts.EnqueuedAt
		if start.IsZero() {
			start = time.Now()
		}
		deadline = start.Add(opts.MaxWait)
	}

	for {
		present, err := queueContains(home, sess.Runner, sess.TerminalSessionID, msgID)
		if err != nil {
			return err
		}
		if !present {
			return nil
		}
		if opts.Mode == WaitMaxWait && !deadline.IsZero() && time.Now().After(deadline) {
			release, err := acquireLock(home, sess.Runner, sess.TerminalSessionID)
			if err != nil {
				return err
			}
			_, _ = removeEntryByID(queuePath(home, sess.Runner, sess.TerminalSessionID), msgID)
			release()
			return fmt.Errorf("send: %s not delivered within %s", msgID, formatDuration(opts.MaxWait))
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func formatDuration(d time.Duration) string {
	if d%(time.Second) == 0 {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	return d.String()
}