package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// archiveHost is the smallest DSH host that answers the two archive verbs. It
// records every call, so the *sequence* a caller issues is assertable: archiving
// a thread with a verifier must be two calls in a known order, and a test that
// only checked the final set would pass on the wrong order.
type archiveHost struct {
	methods  []string
	requests []map[string]any
	// refuse, when set, makes that session's archive fail with refuseCode, which
	// is how DSH says "this one is still working".
	refuse     map[string]string
	refuseCode string
	archived   []string
}

func (h *archiveHost) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method  string `json:"method"`
			RPCID   string `json:"rpcId"`
			Payload struct {
				Args map[string]any `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		if r.URL.Path != "/api/"+req.Method {
			t.Errorf("path %s does not name method %s", r.URL.Path, req.Method)
		}
		h.methods = append(h.methods, req.Method)
		request, _ := req.Payload.Args["request"].(map[string]any)
		h.requests = append(h.requests, request)

		sessionID, _ := request["sessionId"].(string)
		if _, ok := h.refuse[sessionID]; ok {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":  "server-response",
				"rpcId": req.RPCID,
				"result": map[string]any{"ok": false, "error": map[string]any{
					"code": h.refuseCode, "message": "refused",
				}},
			})
			return
		}
		if req.Method == "workspace/archiveSession" {
			h.archived = append(h.archived, sessionID)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":  "server-response",
			"rpcId": req.RPCID,
			"result": map[string]any{"ok": true, "value": map[string]any{
				"archivedSessionIds": h.archived,
			}},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestArchiveSendsTheDocumentedRequest pins the wire shape DSH's
// workspace/archiveSession expects, `stopActivity` included: the flag is what
// chooses between "refuse because it is working" and "stop it first".
func TestArchiveSendsTheDocumentedRequest(t *testing.T) {
	host := &archiveHost{}
	srv := host.server(t)
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL, Token: "secret"}, nil }}

	if _, err := c.Archive(context.Background(), "session-a", ArchiveOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Archive(context.Background(), "session-b", ArchiveOptions{StopActivity: true}); err != nil {
		t.Fatal(err)
	}
	if got := host.methods[0]; got != "workspace/archiveSession" {
		t.Fatalf("method = %q", got)
	}
	if host.requests[0]["sessionId"] != "session-a" {
		t.Fatalf("request = %+v", host.requests[0])
	}
	// Absent, not false: the flag is omitted so an older host reads what it
	// always did rather than a new explicit "do not stop".
	if _, ok := host.requests[0]["stopActivity"]; ok {
		t.Fatalf("stopActivity must be omitted when unset: %+v", host.requests[0])
	}
	if host.requests[1]["stopActivity"] != true {
		t.Fatalf("stopActivity was dropped: %+v", host.requests[1])
	}
}

// TestArchiveReturnsTheWholeSet: DSH answers with the registry-wide set, and the
// SDK must hand it over intact so a caller can replace its mirror in one move.
func TestArchiveReturnsTheWholeSet(t *testing.T) {
	host := &archiveHost{}
	srv := host.server(t)
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}

	if _, err := c.Archive(context.Background(), "session-a", ArchiveOptions{}); err != nil {
		t.Fatal(err)
	}
	value, err := c.Archive(context.Background(), "session-b", ArchiveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(value.ArchivedSessionIDs) != 2 {
		t.Fatalf("set = %v", value.ArchivedSessionIDs)
	}
}

// TestArchiveRefusalIsTyped is the rule a caller branches on: a Session with
// running work comes back as *RPCError carrying DSH's own code, not as a
// transport failure, so "it is still working" can be told from "the host is
// down" without parsing prose.
func TestArchiveRefusalIsTyped(t *testing.T) {
	host := &archiveHost{refuse: map[string]string{"busy": "workspace/session-active"}, refuseCode: "workspace/session-active"}
	srv := host.server(t)
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}

	_, err := c.Archive(context.Background(), "busy", ArchiveOptions{})
	if err == nil {
		t.Fatal("expected a refusal")
	}
	var rpc *RPCError
	if !errors.As(err, &rpc) {
		t.Fatalf("not a typed refusal: %v", err)
	}
	if rpc.Code != "workspace/session-active" || rpc.Method != "workspace/archiveSession" {
		t.Fatalf("rpc = %+v", rpc)
	}
}

// TestUnarchiveUsesTheRestoreVerb keeps the two verbs apart: restoring through
// archiveSession would look like success while changing nothing.
func TestUnarchiveUsesTheRestoreVerb(t *testing.T) {
	host := &archiveHost{}
	srv := host.server(t)
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL}, nil }}

	if _, err := c.Unarchive(context.Background(), "session-a"); err != nil {
		t.Fatal(err)
	}
	if host.methods[0] != "workspace/unarchiveSession" {
		t.Fatalf("method = %q", host.methods[0])
	}
	if host.requests[0]["sessionId"] != "session-a" {
		t.Fatalf("request = %+v", host.requests[0])
	}
}

// TestArchiveRejectsAnEmptySession is the one guard worth having locally: an
// empty id would archive nothing and look like a success.
func TestArchiveRejectsAnEmptySession(t *testing.T) {
	c := Client{ReadListen: func() (Listen, error) { return Listen{Origin: "http://127.0.0.1:1"}, nil }}
	if _, err := c.Archive(context.Background(), "  ", ArchiveOptions{}); err == nil {
		t.Fatal("an empty session must be refused")
	}
	if _, err := c.Unarchive(context.Background(), ""); err == nil {
		t.Fatal("an empty session must be refused")
	}
}
