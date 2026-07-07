package agentsend

import (
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

// StartDrainer launches a background loop that elects a drainer via flock and
// delivers pending messages FIFO when the terminal becomes writable.
func StartDrainer(home string, sess Session, provider agenttty.Provider) {
	go func() {
		for {
			if !drainStep(home, sess, provider) {
				return
			}
		}
	}()
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
	if sess.Runner == "grok-tty" && !strings.Contains(payload, "\n") {
		payload += "\n"
	}
	if err := ttywatch.SendMessage(sess.ListenAddr, sess.TerminalSessionID, payload, true); err != nil {
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