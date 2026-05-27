package trace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	traceStreamPollInterval = 500 * time.Millisecond
	traceStreamHeartbeat    = 15 * time.Second
)

type traceStreamState struct {
	UpdatedAt string
	Status    string
	Tags      string
	LogSize   int64
	LogMod    int64
	RawLines  int
}

func (s *Store) handleStream(w http.ResponseWriter, r *http.Request, id string) {
	detail, state, err := s.loadTraceStreamSnapshot(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming is not supported")
		return
	}
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")

	if err := writeSSEEvent(w, flusher, "detail", detail); err != nil {
		return
	}
	if !isTraceRunning(detail.Metadata.Status) {
		_ = writeSSEEvent(w, flusher, "done", detail.Metadata)
		return
	}

	ticker := time.NewTicker(traceStreamPollInterval)
	defer ticker.Stop()
	heartbeat := time.NewTicker(traceStreamHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			nextDetail, nextState, err := s.loadTraceStreamSnapshot(id)
			if err != nil {
				_ = writeSSEEvent(w, flusher, "trace_error", map[string]string{"error": err.Error()})
				return
			}
			if nextState != state {
				state = nextState
				if err := writeSSEEvent(w, flusher, "update", traceUpdateFromDetail(nextDetail)); err != nil {
					return
				}
			}
			if !isTraceRunning(nextDetail.Metadata.Status) {
				_ = writeSSEEvent(w, flusher, "done", nextDetail.Metadata)
				return
			}
		}
	}
}

func (s *Store) loadTraceStreamSnapshot(id string) (*AgentTraceDetail, traceStreamState, error) {
	detail, err := loadAgentTraceDetail(s.dataDir, id)
	if err != nil {
		return nil, traceStreamState{}, err
	}
	state := traceStreamState{
		UpdatedAt: detail.Metadata.UpdatedAt,
		Status:    detail.Metadata.Status,
		Tags:      strings.Join(detail.Metadata.Tags, ","),
		RawLines:  len(detail.RawLines),
	}
	if info, err := os.Stat(detail.Metadata.LogPath); err == nil {
		state.LogSize = info.Size()
		state.LogMod = info.ModTime().UnixNano()
	}
	return detail, state, nil
}

func traceUpdateFromDetail(detail *AgentTraceDetail) AgentTraceUpdate {
	if detail == nil {
		return AgentTraceUpdate{}
	}
	return AgentTraceUpdate{
		Metadata:     detail.Metadata,
		Messages:     detail.Messages,
		RawLineCount: len(detail.RawLines),
	}
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func isTraceRunning(status string) bool {
	return status == traceStatusRunning
}
