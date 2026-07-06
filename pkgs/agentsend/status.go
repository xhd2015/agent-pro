package agentsend

import (
	"fmt"
	"strings"
)

// MessageStatus reports whether a session-local message is still queued.
// Returns "pending" when the message is in the queue, "delivered" when the id
// was assigned but is no longer queued (delivered, cancelled, or timed out).
func MessageStatus(home string, sess Session, msgID string) (string, error) {
	if !strings.HasPrefix(msgID, "msg_") {
		return "", fmt.Errorf("msg status: invalid message id %s", msgID)
	}
	var n int
	if _, err := fmt.Sscanf(msgID, "msg_%d", &n); err != nil || n < 1 {
		return "", fmt.Errorf("msg status: invalid message id %s", msgID)
	}

	inQueue, err := queueContains(home, sess.Runner, sess.TerminalSessionID, msgID)
	if err != nil {
		return "", err
	}
	if inQueue {
		return "pending", nil
	}

	seq, err := readSeq(seqPath(home, sess.Runner, sess.TerminalSessionID))
	if err != nil {
		return "", err
	}
	if n > seq {
		return "", fmt.Errorf("msg status: message %s not found", msgID)
	}
	return "delivered", nil
}