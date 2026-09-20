package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// Message is one durable message, not a reconstruction of current model context.
type Message struct {
	SessionID string `json:"session_id"`
	Seq       int64  `json:"seq"`
	Time      int64  `json:"time"`
	Role      string `json:"role"`
	Text      string `json:"text"`
}

func eventText(event sessions.Event) (string, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return "", err
	}
	if data == nil {
		return "", fmt.Errorf("message data must be an object")
	}
	if event.Type == "assistant/message" {
		raw := data["message"]
		data = nil
		if err := json.Unmarshal(raw, &data); err != nil {
			return "", fmt.Errorf("assistant message: %w", err)
		}
		if data == nil {
			return "", fmt.Errorf("assistant message must be an object")
		}
	}
	var blocks []map[string]json.RawMessage
	if err := json.Unmarshal(data["content"], &blocks); err != nil {
		return "", fmt.Errorf("message content: %w", err)
	}
	if blocks == nil {
		return "", fmt.Errorf("message content must be an array")
	}
	var text strings.Builder
	for _, block := range blocks {
		var kind string
		if err := json.Unmarshal(block["type"], &kind); err != nil || kind == "" {
			return "", fmt.Errorf("content block requires a string type")
		}
		if kind == "text" {
			var value *string
			if err := json.Unmarshal(block["text"], &value); err != nil || value == nil {
				return "", fmt.Errorf("text block requires string text")
			}
			text.WriteString(*value)
		}
	}
	return text.String(), nil
}
func messages(in *sessions.Inspection, humanOnly bool) ([]Message, error) {
	result := []Message{}
	for _, event := range in.Events {
		role := ""
		switch event.Type {
		case "user/message":
			role = "user"
		case "assistant/message":
			if !humanOnly {
				role = "assistant"
			}
		}
		if role == "" {
			continue
		}
		if humanOnly {
			var data struct {
				Source struct {
					Kind string `json:"kind"`
				} `json:"source"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("event %d: %w", event.Seq, err)
			}
			if data.Source.Kind == "" {
				return nil, fmt.Errorf("event %d: source.kind must be a non-empty string", event.Seq)
			}
			if data.Source.Kind != "user" {
				continue
			}
		}
		text, err := eventText(event)
		if err != nil {
			return nil, fmt.Errorf("event %d: %w", event.Seq, err)
		}
		result = append(result, Message{in.ID, event.Seq, event.Time, role, text})
	}
	return result, nil
}
func stamp(ms int64) string { return time.UnixMilli(ms).UTC().Format(time.RFC3339) }
func printMessages(w io.Writer, items []Message, jsonMode bool) error {
	if jsonMode {
		return writeJSON(w, items)
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(w, "[%s] %s %s (#%d)\n%s\n\n", stamp(item.Time), item.Role, clean(item.SessionID), item.Seq, item.Text); err != nil {
			return err
		}
	}
	return nil
}
func printList(w io.Writer, rows []sessions.Summary) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tUPDATED\tOUTCOME\tCWD"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", clean(row.ID), stamp(row.UpdatedAt), clean(row.Outcome), clean(row.Cwd)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n%d sessions\n", len(rows))
	return err
}
func printInfo(w io.Writer, row sessions.Summary, status bool) error {
	if _, err := fmt.Fprintf(w, "Session:       %s\nStorage:       %s\nFormat:        v%d (%s)\nLast activity: %s\nLast turn:     %s\n", clean(row.ID), clean(row.Path), row.FormatVersion, clean(row.Encoding), stamp(row.UpdatedAt), clean(row.Outcome)); err != nil {
		return err
	}
	if !status {
		if _, err := fmt.Fprintf(w, "Created:       %s\nCwd:           %s\nParent:        %s\nOrigin:        %s\nSeeded:        %t\nProvider:      %s\nModel:         %s\nSize:          %d bytes\n", stamp(row.CreatedAt), clean(row.Cwd), clean(row.ParentSession), clean(row.Origin), row.IsSeeded, clean(row.Provider), clean(row.Model), row.SizeBytes); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Preset:        %s\nDepth:         %d\n", clean(row.AgentPreset), row.DelegationDepth); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "\nStatus scope: durable log only; process liveness is unknown.")
	return err
}
func printLog(w io.Writer, in *sessions.Inspection, tail int, jsonMode bool) error {
	if jsonMode {
		encoder := json.NewEncoder(w)
		for _, event := range last(in.Events, tail) {
			if len(event.Raw) > 0 {
				if err := encoder.Encode(event.Raw); err != nil {
					return err
				}
			} else if err := encoder.Encode(event); err != nil {
				return err
			}
		}
		return nil
	}
	entries := []string{}
	for _, event := range in.Events {
		body := ""
		switch event.Type {
		case "user/message", "assistant/message":
			text, err := eventText(event)
			if err != nil {
				return fmt.Errorf("event %d: %w", event.Seq, err)
			}
			body = text
		case "tool/call", "tool/result", "turn/start", "turn/end":
			var value any
			if err := json.Unmarshal(event.Data, &value); err != nil {
				return fmt.Errorf("event %d: %w", event.Seq, err)
			}
			data, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return err
			}
			body = string(data)
		default:
			continue
		}
		entries = append(entries, fmt.Sprintf("[%s] %s (#%d)\n%s\n", stamp(event.Time), event.Type, event.Seq, body))
	}
	for _, entry := range last(entries, tail) {
		if _, err := fmt.Fprintln(w, entry); err != nil {
			return err
		}
	}
	return nil
}
