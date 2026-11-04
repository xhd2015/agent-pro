package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSessionLifecycle(t *testing.T) {
	t.Parallel()
	var methods []string
	var gotRequests []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("request method/auth incorrect")
		}
		var req struct {
			Type, Method string
			RPCID        string `json:"rpcId"`
			Payload      struct {
				Args map[string]any `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.Type != "client-request" || r.URL.Path != "/api/"+req.Method {
			t.Errorf("invalid request: %+v", req)
		}
		methods = append(methods, req.Method)
		gotRequests = append(gotRequests, req.Payload.Args["request"].(map[string]any))
		var value any = map[string]any{"accepted": true}
		switch req.Method {
		case "workspace/create":
			value = map[string]any{"workspace": map[string]string{"workspaceId": "ws-1"}}
		case "session/create":
			value = map[string]string{"sessionId": "stable-id"}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "server-response", "rpcId": req.RPCID, "result": map[string]any{"ok": true, "value": value}})
	}))
	defer srv.Close()
	var openedURL, openedBrowser string
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL, Token: "secret"}, nil }, OpenBrowser: func(u, b string) error { openedURL, openedBrowser = u, b; return nil }}
	id, err := c.Create(context.Background(), t.TempDir(), "stable-id")
	if err != nil {
		t.Fatal(err)
	}
	if id != "stable-id" {
		t.Fatal(id)
	}
	if err := c.Send(context.Background(), id, "delivery-1", "hello", false); err != nil {
		t.Fatal(err)
	}
	if err := c.Send(context.Background(), id, "delivery-2", "draft", true); err != nil {
		t.Fatal(err)
	}
	if err := c.Open(id, "Brave"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(methods, ",") != "workspace/create,session/create,session/prompt,session/setDraft" {
		t.Fatal(methods)
	}
	if gotRequests[1]["sessionId"] != "stable-id" || gotRequests[2]["requestId"] != "delivery-1" || gotRequests[2]["mode"] != "queue" {
		t.Fatal(gotRequests)
	}
	if gotRequests[3]["text"] != "draft" || gotRequests[3]["content"] != nil {
		t.Fatal(gotRequests[3])
	}
	for _, req := range gotRequests {
		if req["model"] != nil || req["reasoning"] != nil {
			t.Fatal("DSH must own model/reasoning")
		}
	}
	u, _ := url.Parse(openedURL)
	if openedBrowser != "brave" || u.Query().Get("session") != id || u.Query().Get("token") != "secret" {
		t.Fatalf("incorrect browser handoff")
	}
}

func TestRPCFailuresAreSafe(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"unauthorized", 401, "secret", "rejected"},
		{"unavailable", 503, "secret", "HTTP 503"},
		{"invalid JSON", 200, "not-json secret", "invalid JSON"},
		{"wrong envelope", 200, `{"type":"wrong","rpcId":"x"}`, "invalid envelope"},
		{"remote error", 200, "", "TEST_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				if tc.body != "" {
					_, _ = fmt.Fprint(w, tc.body)
					return
				}
				var req map[string]any
				_ = json.NewDecoder(r.Body).Decode(&req)
				_ = json.NewEncoder(w).Encode(map[string]any{"type": "server-response", "rpcId": req["rpcId"], "result": map[string]any{"ok": false, "error": map[string]string{"code": "TEST_ERROR", "message": "secret"}}})
			}))
			defer srv.Close()
			c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL, Token: "secret"}, nil }}
			err := c.Send(context.Background(), "s", "r", "hi", false)
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "secret") {
				t.Fatalf("unsafe/unexpected error: %v", err)
			}
		})
	}
}

func TestBrowserValidationAndRedaction(t *testing.T) {
	t.Parallel()
	for _, b := range []string{"", "default", "brave", "chrome", "firefox", "edge", "safari", "opera"} {
		if _, err := NormalizeBrowser(b); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NormalizeBrowser("netscape"); err == nil {
		t.Fatal("unknown browser accepted")
	}
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: "http://127.0.0.1:9000", Token: "secret"}, nil }, OpenBrowser: func(raw, _ string) error { return fmt.Errorf("failed: %s", raw) }}
	if err := c.Open("s", "chrome"); err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("opener credentials leaked: %v", err)
	}
	c.Home = t.TempDir()
	c.ReadListen = nil
	if err := c.Send(context.Background(), "s", "r", "text", false); err == nil || !strings.Contains(err.Error(), "dsh web") {
		t.Fatalf("missing listen: %v", err)
	}
}
