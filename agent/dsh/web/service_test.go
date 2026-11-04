package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const testKey = "conversation-1"

type serviceFixture struct {
	mu               sync.Mutex
	store            FileStore
	client           Client
	modelSelections  []*ModelSelection
	noModelEcho      bool
	echoOverride     *ModelSelection
	methods          []string
	creates          []string
	attempts         []string
	prompts          []string
	opened           []string
	seen             map[string]bool
	failAfterAccept  bool
	failBeforeAccept bool
	failBrowser      bool
	failCreate       bool
	origin           string
}

func newServiceFixture(t *testing.T) *serviceFixture {
	t.Helper()
	f := &serviceFixture{store: FileStore{Dir: t.TempDir()}, seen: map[string]bool{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer private-launch-token" {
			t.Error("missing auth")
		}
		var req struct {
			Method  string
			RPCID   string `json:"rpcId"`
			Payload struct {
				Args struct {
					Request map[string]any `json:"request"`
				} `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		request := req.Payload.Args.Request
		f.methods = append(f.methods, req.Method)
		value := any(map[string]any{"accepted": true})
		switch req.Method {
		case "workspace/create":
			value = map[string]any{"workspace": map[string]string{"workspaceId": "ws-1"}}
		case "session/create":
			reserved := request["sessionId"].(string)
			state, found, err := f.store.Read(context.Background(), testKey)
			if err != nil || !found || state.DSHSessionID != reserved {
				t.Errorf("provider identity not reserved before creation: %+v %v", state, err)
			}
			var selected *ModelSelection
			if raw := request["modelSelection"]; raw != nil {
				data, _ := json.Marshal(raw)
				if err := json.Unmarshal(data, &selected); err != nil {
					t.Error(err)
				}
			}
			if !reflect.DeepEqual(state.InitialModelSelection, selected) {
				t.Errorf("model selection not pinned before creation: %+v != %+v", state.InitialModelSelection, selected)
			}
			f.modelSelections = append(f.modelSelections, selected)
			f.creates = append(f.creates, reserved)
			if f.failCreate {
				f.failCreate = false
				w.WriteHeader(503)
				return
			}
			value = map[string]any{"sessionId": reserved}
			if selected != nil && !f.noModelEcho {
				if f.echoOverride != nil {
					selected = f.echoOverride
				}
				value.(map[string]any)["modelSelection"] = selected
			}
		case "session/prompt", "session/setDraft":
			state, found, err := f.store.Read(context.Background(), testKey)
			if err != nil || !found || !state.Created || state.DSHSessionID != request["sessionId"] || state.PendingRequestID == "" {
				t.Errorf("missing durable creation/prompt checkpoint: %+v %v", state, err)
			}
			id := state.PendingRequestID
			text, _ := request["text"].(string)
			if req.Method == "session/prompt" {
				id = request["requestId"].(string)
				text = request["content"].([]any)[0].(map[string]any)["text"].(string)
			}
			f.attempts = append(f.attempts, id)
			if f.failBeforeAccept {
				w.WriteHeader(503)
				return
			}
			if !f.seen[id] {
				f.seen[id] = true
				f.prompts = append(f.prompts, text)
			}
			if f.failAfterAccept {
				f.failAfterAccept = false
				w.WriteHeader(503)
				return
			}
		default:
			t.Errorf("unexpected method: %s", req.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "server-response", "rpcId": req.RPCID, "result": map[string]any{"ok": true, "value": value}})
	}))
	t.Cleanup(server.Close)
	f.origin = server.URL
	f.client = Client{ReadListen: func() (Listen, error) { return Listen{Origin: server.URL, Token: "private-launch-token"}, nil }, OpenBrowser: func(raw, browser string) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.opened = append(f.opened, browser)
		if f.failBrowser {
			return fmt.Errorf("platform error contains %s", raw)
		}
		return nil
	}}
	return f
}
func (f *serviceFixture) service() Service { return Service{Client: f.client, Store: f.store} }
func (f *serviceFixture) opts() SessionOptions {
	return SessionOptions{Key: testKey, Workspace: f.store.Dir, Browser: "brave", Prompt: "initial playbook"}
}

func TestServiceCreatesOnceAndReusesPinnedSession(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	first, err := f.service().OpenSession(context.Background(), f.opts())
	if err != nil {
		t.Fatal(err)
	}
	if !first.Accepted || !first.Created || !first.BrowserOpened || len(first.Warnings) != 0 {
		t.Fatalf("%+v", first)
	}
	changed := f.opts()
	changed.Workspace = t.TempDir()
	changed.Browser = "firefox"
	repeated, err := f.service().OpenSession(context.Background(), changed)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.Created || repeated.BrowserOpened || !repeated.Accepted || repeated.SessionID != first.SessionID {
		t.Fatalf("%+v", repeated)
	}
	sent, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "followup"})
	if err != nil {
		t.Fatal(err)
	}
	if !sent.Accepted || sent.SessionID != first.SessionID || sent.BrowserOpened {
		t.Fatalf("%+v", sent)
	}
	if len(f.creates) != 1 || strings.Join(f.prompts, ",") != "initial playbook,followup" || strings.Join(f.opened, ",") != "brave" {
		t.Fatal(f.creates, f.prompts, f.opened)
	}
	state, _, _ := f.store.Read(context.Background(), testKey)
	if state.Workspace != f.store.Dir || state.Browser != "brave" {
		t.Fatal(state)
	}
	path, _ := f.store.path(testKey)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "private-launch-token") || strings.Contains(string(raw), "initial playbook") || strings.Contains(string(raw), "pending_request_id") {
		t.Fatalf("completed input/credentials retained: %s", raw)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("unsafe state permissions: %v %v", info, err)
	}
}

func TestServiceBrowserFailureIsStructuredAndNeverRedelivers(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failBrowser = true
	result, err := f.service().OpenSession(context.Background(), f.opts())
	if err != nil || !result.Accepted || result.BrowserOpened || len(result.Warnings) != 1 || result.Warnings[0].Code != "browser_open_failed" {
		t.Fatalf("%+v %v", result, err)
	}
	if strings.Contains(result.Warnings[0].Message, "private-launch-token") {
		t.Fatal("token leaked")
	}
	f.failBrowser = false
	// A follow-up must not pop a browser just because the first opener failed.
	if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "next"}); err != nil {
		t.Fatal(err)
	}
	if len(f.opened) != 1 {
		t.Fatal(f.opened)
	}
	reopened, err := f.service().Open(context.Background(), testKey)
	if err != nil || !reopened.BrowserOpened || len(f.creates) != 1 || len(f.prompts) != 2 {
		t.Fatalf("%+v %v", reopened, err)
	}
}

func TestServiceReservesPrefixedDSHSessionID(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err != nil {
		t.Fatal(err)
	}
	if len(f.creates) != 1 || !strings.HasPrefix(f.creates[0], "session-") {
		t.Fatalf("reserved %q", f.creates)
	}
}

func TestServiceAmbiguousDeliveryReusesRequestID(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failAfterAccept = true
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err == nil {
		t.Fatal("expected ambiguous HTTP failure")
	}
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err != nil {
		t.Fatal(err)
	}
	if len(f.creates) != 1 || len(f.prompts) != 1 || len(f.attempts) != 2 || f.attempts[0] != f.attempts[1] {
		t.Fatal(f.creates, f.prompts, f.attempts)
	}
}

func TestServiceNewMessageSettlesInterruptedInitialContext(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failBeforeAccept = true
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err == nil {
		t.Fatal("want failure")
	}
	f.failBeforeAccept = false
	if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "newer message"}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(f.prompts, ",") != "initial playbook,newer message" || f.attempts[0] != f.attempts[1] || f.attempts[1] == f.attempts[2] {
		t.Fatal(f.prompts, f.attempts)
	}
}

func TestServiceCreationRetryKeepsReservedIdentity(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failCreate = true
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err == nil {
		t.Fatal("want creation error")
	}
	if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "new input"}); err != nil {
		t.Fatal(err)
	}
	if len(f.creates) != 2 || f.creates[0] != f.creates[1] || strings.Join(f.prompts, ",") != "initial playbook,new input" {
		t.Fatal(f.creates, f.prompts)
	}
}

func TestServiceOpenDoesNotCreateOrFlushPendingInput(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failBeforeAccept = true
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err == nil {
		t.Fatal("want error")
	}
	attempts := len(f.attempts)
	if _, err := f.service().Open(context.Background(), testKey); err != nil {
		t.Fatal(err)
	}
	if len(f.creates) != 1 || len(f.attempts) != attempts {
		t.Fatal("pure open changed input/provider sessions")
	}
	if _, err := f.service().Open(context.Background(), "missing"); err == nil {
		t.Fatal("open must not create a missing session")
	}
}

func TestServiceDraftAndNoOpen(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	opts := f.opts()
	opts.NoSubmit = true
	opts.NoOpen = true
	result, err := f.service().OpenSession(context.Background(), opts)
	if err != nil || !result.Accepted || result.BrowserOpened || len(f.opened) != 0 {
		t.Fatalf("%+v %v", result, err)
	}
	state, _, _ := f.store.Read(context.Background(), testKey)
	if !state.InitialDelivered || state.BrowserOpened {
		t.Fatal(state)
	}
	if _, err := f.service().Open(context.Background(), testKey); err != nil {
		t.Fatal(err)
	}
	if len(f.prompts) != 1 || len(f.opened) != 1 {
		t.Fatal(f.prompts, f.opened)
	}
}

func TestServiceConcurrentInstancesCreateOneSession(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := f.service().OpenSession(context.Background(), f.opts()); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if len(f.creates) != 1 || len(f.prompts) != 1 || len(f.opened) != 1 {
		t.Fatal(f.creates, f.prompts, f.opened)
	}
}

// This sparse process-boundary test verifies the OS lock rather than relying on
// a package-global mutex. Children run this test binary, never agent-run or dsh.
func TestServiceConcurrentProcessesCreateOneSession(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var jobs []*exec.Cmd
	for i := 0; i < 3; i++ {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestServiceProcessHelper$")
		cmd.Env = append(os.Environ(), "DSH_SERVICE_HELPER=1", "DSH_SERVICE_STORE="+f.store.Dir, "DSH_SERVICE_ORIGIN="+f.origin)
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, cmd)
	}
	for _, cmd := range jobs {
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.creates) != 1 || len(f.prompts) != 1 {
		t.Fatal(f.creates, f.prompts)
	}
}
func TestServiceProcessHelper(t *testing.T) {
	if os.Getenv("DSH_SERVICE_HELPER") != "1" {
		return
	}
	root := os.Getenv("DSH_SERVICE_STORE")
	origin := os.Getenv("DSH_SERVICE_ORIGIN")
	service := Service{Client: Client{ReadListen: func() (Listen, error) { return Listen{Origin: origin, Token: "private-launch-token"}, nil }}, Store: FileStore{Dir: root}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := service.OpenSession(ctx, SessionOptions{Key: testKey, Workspace: root, Prompt: "initial playbook", NoOpen: true}); err != nil {
		t.Fatal(err)
	}
}

func TestServiceReadsLegacyFlatBinding(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	path, _ := f.store.path(testKey)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	legacy := `{"runner":"dsh-web","browser":"firefox","workspace":"/original","dsh_session_id":"existing-provider","created":true,"initial_delivered":true,"browser_opened":true,"app_metadata":"keep-me"}`
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "followup"})
	if err != nil || result.SessionID != "existing-provider" || len(f.creates) != 0 {
		t.Fatalf("%+v %v", result, err)
	}
	raw, _ := os.ReadFile(path)
	if !strings.Contains(string(raw), "keep-me") {
		t.Fatal("application metadata lost")
	}
}

func TestServiceIdenticalFollowupsAreDistinctAfterConfirmation(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "same text"}); err != nil {
			t.Fatal(err)
		}
	}
	if len(f.prompts) != 3 || len(f.attempts) != 3 || f.attempts[1] == f.attempts[2] {
		t.Fatal("confirmed repeated messages were incorrectly deduplicated", f.prompts, f.attempts)
	}
}
