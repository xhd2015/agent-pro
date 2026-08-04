package agentsend

import (
	"fmt"
	"time"
)

// EnqueueOptions configures Enqueue behavior.
type EnqueueOptions struct {
	// NoSubmit stores Entry.NoSubmit so drain does not auto-submit (no forced \n / Enter).
	NoSubmit bool
}

// Enqueue appends a message to the session queue under flock and returns its id.
func Enqueue(home string, sess Session, text string) (string, error) {
	return EnqueueWith(home, sess, text, EnqueueOptions{})
}

// EnqueueWith is Enqueue with options (e.g. NoSubmit from send --no-submit).
func EnqueueWith(home string, sess Session, text string, opts EnqueueOptions) (string, error) {
	release, err := acquireLock(home, sess.Runner, sess.TerminalSessionID)
	if err != nil {
		return "", err
	}
	defer release()

	path := queuePath(home, sess.Runner, sess.TerminalSessionID)
	entries, err := readEntries(path)
	if err != nil {
		return "", err
	}
	id, err := nextMessageID(home, sess.Runner, sess.TerminalSessionID, entries)
	if err != nil {
		return "", err
	}
	entry := Entry{
		ID:                id,
		Text:              normalizeSendText(text),
		TerminalSessionID: sess.TerminalSessionID,
		Runner:            sess.Runner,
		EnqueuedAt:        time.Now().UTC(),
		NoSubmit:          opts.NoSubmit,
	}
	if err := appendEntryLocked(path, entry); err != nil {
		return "", err
	}
	return id, nil
}

// Cancel removes a pending message from the queue.
func Cancel(home string, sess Session, msgID string) error {
	release, err := acquireLock(home, sess.Runner, sess.TerminalSessionID)
	if err != nil {
		return err
	}
	defer release()

	path := queuePath(home, sess.Runner, sess.TerminalSessionID)
	found, err := removeEntryByID(path, msgID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("msg cancel: message %s not found", msgID)
	}
	return nil
}