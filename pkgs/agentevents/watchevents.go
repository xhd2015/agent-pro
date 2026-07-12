package agentevents

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// WatchEvents tails events.jsonl from a byte offset and calls onLine for each new NDJSON row.
// Tailing continues until ctx is cancelled; session meta status is not consulted.
func WatchEvents(ctx context.Context, store agentstorage.Store, sessionID string, afterOffset int64, onLine func(line string) error) error {
	path := eventsPath(store, sessionID)

	if _, err := emitLinesFromOffset(path, afterOffset, onLine); err != nil {
		return err
	}

	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var watchErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchErr = logs.WatchLine(watchCtx, path, logs.WatchLineOptions{}, onLine)
	}()

	<-ctx.Done()
	cancel()
	wg.Wait()

	if ctx.Err() != nil {
		return ctx.Err()
	}
	if watchErr != nil && watchErr != context.Canceled {
		return watchErr
	}
	return nil
}

func eventsPath(store agentstorage.Store, sessionID string) string {
	return filepath.Join(store.Home(), "sessions", sessionID, "events.jsonl")
}

type assistantStreamCursor struct {
	id   string
	text string
}

func parseAssistantStreamLine(line string) (id, text string, ok bool) {
	var ev struct {
		Type  string `json:"type"`
		Role  string `json:"role"`
		ID    string `json:"id"`
		Text  string `json:"text"`
		Phase string `json:"phase"`
	}
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return "", "", false
	}
	if ev.Type != "message" || ev.Role != "assistant" || ev.Phase != "" {
		return "", "", false
	}
	return ev.ID, ev.Text, true
}

func isGrowingAssistantStream(prev assistantStreamCursor, id, text string) bool {
	if text == "" || prev.text == "" || len(text) <= len(prev.text) {
		return false
	}
	if !strings.HasPrefix(text, prev.text) {
		return false
	}
	if id != "" && prev.id != "" {
		return id == prev.id
	}
	return true
}

func emitLinesFromOffset(path string, afterOffset int64, onLine func(line string) error) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return afterOffset, nil
		}
		return 0, err
	}
	defer f.Close()

	if afterOffset > 0 {
		if _, err := f.Seek(afterOffset, io.SeekStart); err != nil {
			return 0, err
		}
	}

	reader := bufio.NewReader(f)
	var assistant assistantStreamCursor
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			trimmed = strings.TrimSpace(trimmed)
			if trimmed != "" {
				if id, text, ok := parseAssistantStreamLine(trimmed); ok {
					if isGrowingAssistantStream(assistant, id, text) {
						time.Sleep(250 * time.Millisecond)
					}
					assistant.id = id
					assistant.text = text
				}
				if err := onLine(trimmed); err != nil {
					return 0, err
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
	}

	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	return offset, nil
}