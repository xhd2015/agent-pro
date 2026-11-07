package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// rpcHost is a fake DSH Web that records every client-request and answers with
// one JSON value, so a test can assert both the request shape and the parsing.
type rpcHost struct {
	methods  []string
	requests []map[string]any
	value    string
}

func newRPCHost(t *testing.T, value string) (*rpcHost, Client) {
	t.Helper()
	host := &rpcHost{value: value}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method  string `json:"method"`
			RPCID   string `json:"rpcId"`
			Payload struct {
				Args map[string]any `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if r.URL.Path != "/api/"+req.Method || r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("invalid request envelope: %s auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		host.methods = append(host.methods, req.Method)
		request, _ := req.Payload.Args["request"].(map[string]any)
		host.requests = append(host.requests, request)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"type":"server-response","rpcId":`+strconv.Quote(req.RPCID)+`,"result":{"ok":true,"value":`+host.value+`}}`)
	}))
	t.Cleanup(srv.Close)
	return host, Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL, Token: "secret"}, nil }}
}

func TestForkSendsCutAndCheckout(t *testing.T) {
	t.Parallel()
	host, client := newRPCHost(t, `{"sessionId":"child-1"}`)
	id, err := client.Fork(context.Background(), ForkOptions{SessionID: "src", AtSeq: 42, Cwd: "/tmp/child"})
	if err != nil {
		t.Fatal(err)
	}
	if id != "child-1" {
		t.Fatalf("id=%q", id)
	}
	if strings.Join(host.methods, ",") != "session/fork" {
		t.Fatalf("methods=%v", host.methods)
	}
	request := host.requests[0]
	if request["sessionId"] != "src" || request["cwd"] != "/tmp/child" {
		t.Fatalf("request=%+v", request)
	}
	if atSeq, _ := request["atSeq"].(float64); atSeq != 42 {
		t.Fatalf("atSeq=%v", request["atSeq"])
	}
}

func TestForkKeepsLatestCutWhenSeqOmitted(t *testing.T) {
	t.Parallel()
	host, client := newRPCHost(t, `{"sessionId":"child-1"}`)
	if _, err := client.Fork(context.Background(), ForkOptions{SessionID: "src", Cwd: "/tmp/child"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := host.requests[0]["atSeq"]; ok {
		t.Fatalf("atSeq must be omitted for the latest cut: %+v", host.requests[0])
	}
}

func TestForkRejectsBadInputBeforeRequest(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		opts    ForkOptions
		value   string
		wantErr string
	}{
		{name: "missing source", opts: ForkOptions{Cwd: "/tmp/child"}, value: `{"sessionId":"c"}`, wantErr: "source session ID"},
		{name: "negative cut", opts: ForkOptions{SessionID: "src", AtSeq: -1, Cwd: "/tmp/child"}, value: `{"sessionId":"c"}`, wantErr: "must not be negative"},
		{name: "relative checkout", opts: ForkOptions{SessionID: "src", Cwd: "child"}, value: `{"sessionId":"c"}`, wantErr: "fully qualified"},
		{name: "child id missing", opts: ForkOptions{SessionID: "src", Cwd: "/tmp/child"}, value: `{}`, wantErr: "no session ID"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			host, client := newRPCHost(t, tc.value)
			_, err := client.Fork(context.Background(), tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err=%v", err)
			}
			// A rejection must cost the Host nothing; only the empty child answer
			// is allowed to reach it.
			wantCalls := 0
			if tc.name == "child id missing" {
				wantCalls = 1
			}
			if len(host.methods) != wantCalls {
				t.Fatalf("methods=%v", host.methods)
			}
		})
	}
}

func TestSetDraftNeedsNoRequestID(t *testing.T) {
	t.Parallel()
	host, client := newRPCHost(t, `{}`)
	if err := client.Send(context.Background(), "sess-1", "", "draft text", true); err != nil {
		t.Fatal(err)
	}
	if strings.Join(host.methods, ",") != "session/setDraft" {
		t.Fatalf("methods=%v", host.methods)
	}
	request := host.requests[0]
	if request["sessionId"] != "sess-1" || request["text"] != "draft text" {
		t.Fatalf("request=%+v", request)
	}
	if request["requestId"] != nil || request["content"] != nil {
		t.Fatalf("a draft must not carry delivery fields: %+v", request)
	}
	if err := client.Send(context.Background(), "sess-1", "", "text", false); err == nil || !strings.Contains(err.Error(), "request ID") {
		t.Fatalf("prompt without a request ID: %v", err)
	}
	if len(host.methods) != 1 {
		t.Fatalf("a rejected prompt must not reach the Host: %v", host.methods)
	}
}
