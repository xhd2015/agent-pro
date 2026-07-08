package agentsync

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"

	grok_session "github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

func runGrokSyncWorker(ctx context.Context, opts GrokSyncOptions) error {
	checkpoint, err := opts.Sink.LoadCheckpoint()
	if err != nil {
		return err
	}

	offset := int64(0)
	turnIndex := 0
	grokSessionID := strings.TrimSpace(opts.GrokSessionID)
	updatesPath := strings.TrimSpace(opts.UpdatesPath)
	if strings.TrimSpace(checkpoint.UpdatesPath) != "" {
		offset = checkpoint.UpdatesOffset
		turnIndex = checkpoint.TurnIndex
		if grokSessionID == "" {
			grokSessionID = checkpoint.GrokSessionID
		}
		if updatesPath == "" {
			updatesPath = checkpoint.UpdatesPath
		}
	}

	converter := grok_session.NewConverter()
	converter.SetTurnIndex(turnIndex)

	emitEvent := func(ev types.AgentEvent) error {
		if ev.Type == types.ActionMessage && strings.TrimSpace(ev.Role) == "" {
			ev.Role = "assistant"
		}
		if ev.Type == types.ActionMessage && ev.Timestamp == 0 {
			ev.Timestamp = time.Now().UnixMilli()
		}
		return opts.Sink.AppendEvent(trimACPAgentEventNewlines(ev))
	}

	if updatesPath == "" {
		if err := emitEvent(types.AgentEvent{
			Type:      types.ActionThink,
			Text:      "Resolve session id...",
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}
		id, path, discErr := bootstrapGrokSession(ctx, opts, grokSessionID)
		if discErr != nil {
			if prompt := strings.TrimSpace(opts.InitialPrompt); prompt != "" {
				if err := emitEvent(types.AgentEvent{
					Type:      types.ActionMessage,
					Role:      "user",
					Text:      prompt,
					Timestamp: time.Now().UnixMilli(),
				}); err != nil {
					return err
				}
			}
			if err := emitEvent(types.AgentEvent{
				Type:      types.ActionError,
				Text:      "Cannot resolve session id: " + discErr.Error(),
				Timestamp: time.Now().UnixMilli(),
			}); err != nil {
				return err
			}
			return opts.Sink.OnTurnCompleted()
		}
		if id != "" {
			grokSessionID = id
		}
		if path != "" {
			updatesPath = path
			if offset == 0 && checkpoint.UpdatesPath == "" {
				offset = updatesTailStartOffset(path, discoveryRunStart(opts))
			}
		}
	}
	if strings.TrimSpace(updatesPath) == "" {
		return nil
	}
	if size, statErr := updatesFileSize(updatesPath); statErr == nil && offset > size {
		offset = 0
	}

	saveCheckpoint := func(newOffset int64) error {
		return opts.Sink.SaveCheckpoint(GrokSyncState{
			GrokSessionID: grokSessionID,
			UpdatesPath:   updatesPath,
			UpdatesOffset: newOffset,
			TurnIndex:     converter.TurnIndex(),
		})
	}

	processLine := func(line string, newOffset int64) error {
		turnCompleted := isACPTurnCompletedLine(line)
		if err := processACPUpdateLine(line, converter, emitEvent); err != nil {
			return err
		}
		if err := saveCheckpoint(newOffset); err != nil {
			return err
		}
		if turnCompleted {
			return opts.Sink.OnTurnCompleted()
		}
		return nil
	}

	offset, err = readACPUpdatesFromOffset(updatesPath, offset, processLine)
	if err != nil {
		return err
	}

	watchErr := logs.WatchLine(ctx, updatesPath, logs.WatchLineOptions{DisableDebounce: true}, func(line string) error {
		line = strings.TrimRight(line, "\n")
		if strings.TrimSpace(line) == "" {
			return nil
		}
		newOffset, sizeErr := updatesFileSize(updatesPath)
		if sizeErr != nil {
			return sizeErr
		}
		if newOffset < offset {
			offset = 0
		}
		return processLine(line, newOffset)
	})

	for _, ev := range converter.Flush() {
		if err := emitEvent(ev); err != nil {
			return err
		}
	}

	if watchErr != nil {
		return watchErr
	}
	return ctx.Err()
}

func isACPTurnCompletedLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	upd, ok := grok_session.ParseLine(line)
	return ok && upd.SessionUpdate == "turn_completed"
}

func updatesFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func readACPUpdatesFromOffset(path string, offset int64, onLine func(line string, newOffset int64) error) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return offset, nil
		}
		return offset, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return offset, err
	}

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\n")
			newOffset, seekErr := f.Seek(0, io.SeekCurrent)
			if seekErr != nil {
				return offset, seekErr
			}
			if strings.TrimSpace(trimmed) != "" {
				if err := onLine(trimmed, newOffset); err != nil {
					return offset, err
				}
			}
			offset = newOffset
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return offset, err
		}
	}
	return offset, nil
}

func trimACPAgentEventNewlines(ev types.AgentEvent) types.AgentEvent {
	ev.Text = strings.TrimRight(ev.Text, "\n")
	ev.Output = strings.TrimRight(ev.Output, "\n")
	return ev
}

func processACPUpdateLine(line string, converter *grok_session.Converter, emit func(types.AgentEvent) error) error {
	line = strings.TrimRight(line, "\n")
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}
	if _, ok := grok_session.ParseLine(trimmed); !ok {
		return nil
	}
	for _, ev := range converter.ProcessLine(line) {
		if err := emit(ev); err != nil {
			return err
		}
	}
	for _, ev := range converter.Flush() {
		if err := emit(ev); err != nil {
			return err
		}
	}
	return nil
}