package agenttty

import (
	"time"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

func fetchSnapshotBytes(listenAddr, sessionID string) ([]byte, error) {
	text, err := ttywatch.SnapshotText(listenAddr, sessionID)
	if err != nil {
		return nil, err
	}
	return []byte(text), nil
}

// WaitUntilWritable polls CheckWritable until ready or timeout.
func WaitUntilWritable(provider Provider, listenAddr, sessionID string, timeout time.Duration) WritableStatus {
	return ttywatch.WaitUntilWritable(provider.CheckWritable, listenAddr, sessionID, timeout)
}