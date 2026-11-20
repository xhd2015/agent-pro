package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// injectServer answers session/inject with the seqs a real DSH would return and
// records the request payload for assertions.
func injectServer(t *testing.T, seen *[]map[string]any, seqs []int64, started bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method  string
			RPCID   string `json:"rpcId"`
			Payload struct {
				Args map[string]any `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.Method != "session/inject" {
			t.Errorf("unexpected method %q", req.Method)
		}
		*seen = append(*seen, req.Payload.Args["request"].(map[string]any))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":  "server-response",
			"rpcId": req.RPCID,
			"result": map[string]any{"ok": true, "value": map[string]any{
				"seqs": seqs, "started": started,
			}},
		})
	}))
}

func TestInjectMessagesSendsOneBatch(t *testing.T) {
	t.Parallel()
	var seen []map[string]any
	srv := injectServer(t, &seen, []int64{11, 12, 13}, true)
	defer srv.Close()
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}

	result, err := c.InjectMessages(context.Background(), "session-1", "inject-1", []InjectedMessage{
		{Kind: "user", Text: "question"},
		{Kind: "thinking", Text: "considering"},
		{Kind: "assistant", Text: "answer"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Seqs) != 3 || result.Seqs[0] != 11 || result.Seqs[2] != 13 {
		t.Fatal(result.Seqs)
	}
	if !result.Started {
		t.Fatal("started")
	}
	// One call, one batch: the three messages travel together so DSH commits them
	// in a single transaction.
	if len(seen) != 1 {
		t.Fatalf("expected one inject call, got %d", len(seen))
	}
	if seen[0]["sessionId"] != "session-1" || seen[0]["requestId"] != "inject-1" {
		t.Fatal(seen[0])
	}
	messages := seen[0]["messages"].([]any)
	if len(messages) != 3 {
		t.Fatalf("expected three messages, got %d", len(messages))
	}
	first := messages[0].(map[string]any)
	if first["kind"] != "user" || first["text"] != "question" {
		t.Fatal(first)
	}
	last := messages[2].(map[string]any)
	if last["kind"] != "assistant" || last["text"] != "answer" {
		t.Fatal(last)
	}
}

func TestInjectMessagesRejectsInvalidInputBeforeCalling(t *testing.T) {
	t.Parallel()
	var seen []map[string]any
	srv := injectServer(t, &seen, nil, false)
	defer srv.Close()
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}

	cases := []struct {
		name     string
		session  string
		request  string
		messages []InjectedMessage
		want     string
	}{
		{name: "no session", request: "r", messages: []InjectedMessage{{Kind: "user", Text: "x"}}, want: "session is required"},
		{name: "no request id", session: "s", messages: []InjectedMessage{{Kind: "user", Text: "x"}}, want: "request ID is required"},
		{name: "empty batch", session: "s", request: "r", want: "at least one injected message"},
		{name: "unknown kind", session: "s", request: "r", messages: []InjectedMessage{{Kind: "system", Text: "x"}}, want: "must be user, thinking, assistant, tool-call or tool-result"},
		{name: "blank text", session: "s", request: "r", messages: []InjectedMessage{{Kind: "assistant", Text: "   "}}, want: "has no text"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.InjectMessages(context.Background(), tc.session, tc.request, tc.messages)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err.Error(), tc.want)
			}
		})
	}
	// A rejected batch never reaches the network, so nothing was committed.
	if len(seen) != 0 {
		t.Fatalf("expected no calls, got %d", len(seen))
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(needle) == 0 ||
		indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// An edit is one DSH transaction like an injection, and the wire carries the
// same shape the session/inject request does: seq plus text, or seq plus deleted.
func TestEditMessagesSendsOneBatch(t *testing.T) {
	t.Parallel()
	var seen []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method  string
			RPCID   string `json:"rpcId"`
			Payload struct {
				Args map[string]any `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.Method != "session/edit" {
			t.Errorf("unexpected method %q", req.Method)
		}
		seen = append(seen, req.Payload.Args["request"].(map[string]any))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":  "server-response",
			"rpcId": req.RPCID,
			"result": map[string]any{"ok": true, "value": map[string]any{
				"seq": 31, "started": true,
			}},
		})
	}))
	defer srv.Close()

	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}
	result, err := c.EditMessages(context.Background(), "s-1", "edit-1", []EditedMessageChange{
		{Seq: 7, Text: "rewritten"},
		{Seq: 9, Delete: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Seq != 31 || !result.Started {
		t.Fatalf("result = %+v", result)
	}
	if len(seen) != 1 {
		t.Fatalf("expected one RPC, got %d", len(seen))
	}
	edits, _ := seen[0]["edits"].([]any)
	if len(edits) != 2 {
		t.Fatalf("edits = %v", edits)
	}
	first, _ := edits[0].(map[string]any)
	if first["seq"] != float64(7) || first["text"] != "rewritten" {
		t.Fatalf("first = %+v", first)
	}
	second, _ := edits[1].(map[string]any)
	if second["seq"] != float64(9) || second["deleted"] != true {
		t.Fatalf("second = %+v", second)
	}
}

// A refusal DSH produced is a typed error, so a caller can act on the code
// without parsing a sentence.
func TestEditMessagesReportsRefusalCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RPCID string `json:"rpcId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":  "server-response",
			"rpcId": req.RPCID,
			"result": map[string]any{"ok": false, "error": map[string]any{
				"code": "session/edit-rejected",
			}},
		})
	}))
	defer srv.Close()

	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}
	_, err := c.EditMessages(context.Background(), "s-1", "edit-2", []EditedMessageChange{{Seq: 3, Text: "x"}})
	if err == nil {
		t.Fatal("expected a refusal")
	}
	var rpc *RPCError
	if !errors.As(err, &rpc) {
		t.Fatalf("error is not a typed RPC refusal: %v", err)
	}
	if rpc.Code != "session/edit-rejected" || rpc.Method != "session/edit" {
		t.Fatalf("rpc = %+v", rpc)
	}
}
