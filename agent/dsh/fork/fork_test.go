package fork_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/dsh/fork"
	"github.com/xhd2015/agent-pro/agent/dsh/web"
)

// host is a fake DSH Web that records every client-request the fork sends and
// can fail one method to exercise the partial-delivery path.
type host struct {
	methods  []string
	requests []map[string]any
	childID  string
	failWith string
	opened   []string
}

func newHost(t *testing.T, childID string) (*host, web.Client) {
	t.Helper()
	h := &host{childID: childID}
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
		h.methods = append(h.methods, req.Method)
		request, _ := req.Payload.Args["request"].(map[string]any)
		h.requests = append(h.requests, request)
		w.Header().Set("Content-Type", "application/json")
		if req.Method == h.failWith {
			_, _ = io.WriteString(w, `{"type":"server-response","rpcId":`+strconv.Quote(req.RPCID)+`,"result":{"ok":false,"error":{"code":"boom","message":"host refused"}}}`)
			return
		}
		value := `{}`
		if req.Method == "session/fork" {
			value = `{"sessionId":` + strconv.Quote(h.childID) + `}`
		}
		_, _ = io.WriteString(w, `{"type":"server-response","rpcId":`+strconv.Quote(req.RPCID)+`,"result":{"ok":true,"value":`+value+`}}`)
	}))
	t.Cleanup(srv.Close)
	client := web.Client{
		ReadListen: func() (web.Listen, error) { return web.Listen{Origin: srv.URL, Token: "secret"}, nil },
		OpenBrowser: func(raw, app string) error {
			h.opened = append(h.opened, raw+"#"+app)
			return nil
		},
	}
	return h, client
}

// reader serves checkouts from a map; a missing id is an unknown session and an
// empty checkout means the reader cannot tell.
func reader(workspaces map[string]string) func(string) (fork.SessionInfo, error) {
	return func(id string) (fork.SessionInfo, error) {
		dir, ok := workspaces[id]
		if !ok {
			return fork.SessionInfo{}, fmt.Errorf("unknown session %q", id)
		}
		return fork.SessionInfo{ID: id, Cwd: dir}, nil
	}
}

func TestRunForksIntoOtherCheckoutAndSeedsDraft(t *testing.T) {
	source, child := t.TempDir(), t.TempDir()
	h, client := newHost(t, "child-1")
	result, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(map[string]string{"src": source, "child-1": child}),
		SourceID:    "src",
		Cwd:         child,
		AtSeq:       9,
		Draft:       "/brainstorm child plan",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != (fork.Result{SessionID: "child-1", Forked: true, Drafted: true}) {
		t.Fatalf("result=%+v", result)
	}
	if strings.Join(h.methods, ",") != "session/fork,session/setDraft" {
		t.Fatalf("methods=%v", h.methods)
	}
	forkRequest := h.requests[0]
	if forkRequest["sessionId"] != "src" || forkRequest["cwd"] != child {
		t.Fatalf("fork request=%+v", forkRequest)
	}
	if atSeq, _ := forkRequest["atSeq"].(float64); atSeq != 9 {
		t.Fatalf("fork atSeq=%v", forkRequest["atSeq"])
	}
	draftRequest := h.requests[1]
	if draftRequest["sessionId"] != "child-1" || draftRequest["text"] != "/brainstorm child plan" {
		t.Fatalf("draft request=%+v", draftRequest)
	}
	if draftRequest["content"] != nil {
		t.Fatalf("a fork draft must never be submitted: %+v", draftRequest)
	}
	if len(h.opened) != 0 {
		t.Fatalf("the browser must stay closed unless asked: %v", h.opened)
	}
}

func TestRunRefusesSourceCheckout(t *testing.T) {
	checkout := t.TempDir()
	h, client := newHost(t, "child-1")
	_, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(map[string]string{"src": checkout}),
		SourceID:    "src",
		Cwd:         checkout,
		Draft:       "hi",
	})
	if !errors.Is(err, fork.ErrSameCheckout) {
		t.Fatalf("err=%v", err)
	}
	if len(h.methods) != 0 {
		t.Fatalf("a refused fork must not reach the Host: %v", h.methods)
	}
}

func TestRunTreatsSymlinkedCheckoutSpellingsAsOne(t *testing.T) {
	dir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil || resolved == dir {
		t.Skip("temp root has no symlinked spelling on this system")
	}
	_, client := newHost(t, "child-1")
	_, err = fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(map[string]string{"src": resolved}),
		SourceID:    "src",
		Cwd:         dir,
	})
	if !errors.Is(err, fork.ErrSameCheckout) {
		t.Fatalf("err=%v", err)
	}
}

func TestRunRefusesUnknownSource(t *testing.T) {
	h, client := newHost(t, "child-1")
	_, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(nil),
		SourceID:    "gone",
		Cwd:         t.TempDir(),
	})
	if !errors.Is(err, fork.ErrSourceMissing) {
		t.Fatalf("err=%v", err)
	}
	if len(h.methods) != 0 {
		t.Fatalf("methods=%v", h.methods)
	}
}

func TestRunResumesOnlyChildInCheckout(t *testing.T) {
	source, child := t.TempDir(), t.TempDir()
	workspaces := map[string]string{"src": source, "child-1": child}
	h, client := newHost(t, "new-child")
	result, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(workspaces),
		SourceID:    "src",
		Cwd:         child,
		Resume:      "child-1",
		Draft:       "retry",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Forked || result.SessionID != "child-1" || !result.Drafted {
		t.Fatalf("result=%+v", result)
	}
	if strings.Join(h.methods, ",") != "session/setDraft" {
		t.Fatalf("a resume must skip session/fork: %v", h.methods)
	}

	elsewhere := t.TempDir()
	workspaces["child-1"] = elsewhere
	h.methods = nil
	_, err = fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(workspaces),
		SourceID:    "src",
		Cwd:         child,
		Resume:      "child-1",
		Draft:       "retry",
	})
	if !errors.Is(err, fork.ErrChildElsewhere) {
		t.Fatalf("err=%v", err)
	}
	if len(h.methods) != 0 {
		t.Fatalf("a foreign child must not receive the draft: %v", h.methods)
	}
}

func TestRunRefusesChildThatKeptSourceCheckout(t *testing.T) {
	source, child := t.TempDir(), t.TempDir()
	h, client := newHost(t, "child-1")
	// The Host dropped the cwd field and forked into the source directory.
	result, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(map[string]string{"src": source, "child-1": source}),
		SourceID:    "src",
		Cwd:         child,
		Draft:       "hi",
	})
	if !errors.Is(err, fork.ErrCwdNotHonored) {
		t.Fatalf("err=%v", err)
	}
	if result.SessionID != "child-1" || !result.Forked || result.Drafted {
		t.Fatalf("result=%+v", result)
	}
	if strings.Join(h.methods, ",") != "session/fork" {
		t.Fatalf("a stray child must not receive the draft: %v", h.methods)
	}
}

func TestRunKeepsChildWhenDraftFails(t *testing.T) {
	source, child := t.TempDir(), t.TempDir()
	h, client := newHost(t, "child-1")
	h.failWith = "session/setDraft"
	result, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(map[string]string{"src": source, "child-1": child}),
		SourceID:    "src",
		Cwd:         child,
		Draft:       "hi",
	})
	if err == nil || !strings.Contains(err.Error(), "DSH fork draft") {
		t.Fatalf("err=%v", err)
	}
	// The caller checkpoints this ID and retries with Resume.
	if result.SessionID != "child-1" || !result.Forked || result.Drafted {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunOpensChildWhenRequested(t *testing.T) {
	source, child := t.TempDir(), t.TempDir()
	workspaces := map[string]string{"src": source, "child-1": child}
	h, client := newHost(t, "child-1")
	result, err := fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(workspaces),
		SourceID:    "src",
		Cwd:         child,
		Draft:       "hi",
		Open:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Opened || len(h.opened) != 1 {
		t.Fatalf("result=%+v opened=%v", result, h.opened)
	}
	if !strings.Contains(h.opened[0], "session=child-1") || !strings.Contains(h.opened[0], "token=secret") {
		t.Fatalf("opened=%q", h.opened[0])
	}

	h.opened = nil
	result, err = fork.Run(context.Background(), fork.Options{
		Client:      client,
		ReadSession: reader(workspaces),
		SourceID:    "src",
		Cwd:         child,
		Draft:       "hi",
		Open:        true,
		Browser:     web.BrowserNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Opened || len(h.opened) != 0 {
		t.Fatalf("browser none must still finish: result=%+v opened=%v", result, h.opened)
	}
}

func TestRunValidatesOptionsBeforeRequest(t *testing.T) {
	source := t.TempDir()
	h, client := newHost(t, "child-1")
	missing := filepath.Join(t.TempDir(), "gone")
	cases := []struct {
		name    string
		opts    fork.Options
		wantErr string
	}{
		{name: "no reader", opts: fork.Options{Client: client, SourceID: "src", Cwd: source}, wantErr: "session reader"},
		{name: "no source", opts: fork.Options{Client: client, ReadSession: reader(nil), Cwd: source}, wantErr: "source session ID"},
		{name: "no checkout", opts: fork.Options{Client: client, ReadSession: reader(nil), SourceID: "src"}, wantErr: "checkout directory"},
		{name: "relative checkout", opts: fork.Options{Client: client, ReadSession: reader(nil), SourceID: "src", Cwd: "eluc"}, wantErr: "fully qualified"},
		{name: "missing checkout", opts: fork.Options{Client: client, ReadSession: reader(nil), SourceID: "src", Cwd: missing}, wantErr: "fork cwd"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fork.Run(context.Background(), tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	if len(h.methods) != 0 {
		t.Fatalf("invalid options must not reach the Host: %v", h.methods)
	}
}
