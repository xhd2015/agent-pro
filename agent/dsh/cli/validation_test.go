package cli

import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

func TestMessagePayloadValidation(t *testing.T) {
	for _, payload := range []string{`null`, `{}`, `{"content":null}`, `{"content":{}}`, `{"content":[null]}`, `{"content":[{}]}`, `{"content":[{"type":2}]}`, `{"content":[{"type":"text"}]}`, `{"content":[{"type":"text","text":null}]}`, `{"content":[{"type":"text","text":7}]}`} {
		for _, kind := range []string{"user/message", "assistant/message"} {
			data := payload
			if kind == "assistant/message" {
				data = `{"message":` + payload + `}`
			}
			if _, err := eventText(sessions.Event{Type: kind, Data: json.RawMessage(data)}); err == nil {
				t.Errorf("accepted %s: %s", kind, data)
			}
		}
	}
	for _, data := range []string{`null`, `{}`, `{"message":null}`} {
		if _, err := eventText(sessions.Event{Type: "assistant/message", Data: json.RawMessage(data)}); err == nil {
			t.Errorf("accepted assistant %s", data)
		}
	}
	for _, data := range []string{`{"content":[]}`, `{"content":[{"type":"text","text":""}]}`} {
		if text, err := eventText(sessions.Event{Type: "user/message", Data: json.RawMessage(data)}); err != nil || text != "" {
			t.Fatalf("empty text rejected: %q %v", text, err)
		}
	}
}
func TestHumanSourceValidation(t *testing.T) {
	for _, data := range []string{`null`, `{}`, `{"source":null}`, `{"source":{}}`, `{"source":{"kind":null}}`, `{"source":{"kind":7}}`} {
		in := &sessions.Inspection{Events: []sessions.Event{{Type: "user/message", Data: json.RawMessage(data)}}}
		if _, err := messages(in, true); err == nil {
			t.Errorf("accepted %s", data)
		}
	}
}
