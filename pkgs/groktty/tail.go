package groktty

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// updatesTailStartOffset returns the byte offset to begin tailing. When the file
// already existed before this run (mod time well before runStart), start at EOF
// so a stale same-prompt session is not replayed from the beginning.
func updatesTailStartOffset(path string, runStart time.Time) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	cutoff := runStart.Add(-sessionDiscoveryGrace)
	if info.ModTime().Before(cutoff) {
		return info.Size()
	}
	return 0
}

const tailPollInterval = 150 * time.Millisecond

// TailUpdates polls updates.jsonl and emits AgentEvents for new lines until ctx is done.
func TailUpdates(ctx context.Context, updatesPath string, emit func(types.AgentEvent) error) error {
	return TailUpdatesFromOffset(ctx, updatesPath, 0, emit)
}

// TailUpdatesFromOffset tails updates.jsonl starting at the given byte offset.
// Use updatesTailStartOffset at discovery time to skip bytes already on disk.
func TailUpdatesFromOffset(ctx context.Context, updatesPath string, startOffset int64, emit func(types.AgentEvent) error) error {
	converter := NewACPConverter()
	offset := startOffset

	ticker := time.NewTicker(tailPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			for _, ev := range converter.Flush() {
				if err := emit(ev); err != nil {
					return err
				}
			}
			return ctx.Err()
		case <-ticker.C:
			newOffset, done, err := readNewACPUpdates(updatesPath, offset, converter, emit)
			if err != nil {
				return err
			}
			offset = newOffset
			if done {
				for _, ev := range converter.Flush() {
					if err := emit(ev); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
}

func readNewACPUpdates(path string, offset int64, converter *ACPConverter, emit func(types.AgentEvent) error) (int64, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return offset, false, nil
		}
		return offset, false, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, 0); err != nil {
		return offset, false, err
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	turnCompleted := false
	hadLines := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		hadLines = true
		if upd, ok := parseACPUpdateLine(line); ok && upd.SessionUpdate == "turn_completed" {
			turnCompleted = true
		}
		for _, ev := range converter.ProcessLine(line) {
			if err := emit(ev); err != nil {
				return offset, false, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return offset, false, err
	}
	if hadLines {
		for _, ev := range converter.Flush() {
			if err := emit(ev); err != nil {
				return offset, false, err
			}
		}
	}
	newOffset, err := f.Seek(0, 1)
	if err != nil {
		return offset, false, err
	}
	return newOffset, turnCompleted, nil
}