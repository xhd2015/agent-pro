package web

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type faultStore struct {
	Store
	saves  atomic.Int32
	failAt int32
}
type faultRecord struct {
	Record
	owner *faultStore
}

func (s *faultStore) WithSession(ctx context.Context, key string, fn func(Record) error) error {
	return s.Store.WithSession(ctx, key, func(r Record) error { return fn(faultRecord{Record: r, owner: s}) })
}
func (r faultRecord) Save(state State) error {
	if r.owner.saves.Add(1) == r.owner.failAt {
		return errors.New("injected checkpoint write failure")
	}
	return r.Record.Save(state)
}

func TestServiceAdmissionCheckpointFailureReusesRequestID(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	store := &faultStore{Store: f.store, failAt: 4}
	service := Service{Client: f.client, Store: store}
	if _, err := service.OpenSession(context.Background(), f.opts()); err == nil {
		t.Fatal("want checkpoint failure")
	}
	state, _, _ := f.store.Read(context.Background(), testKey)
	if state.PendingRequestID == "" || state.InitialDelivered {
		t.Fatal("last confirmed checkpoint lost")
	}
	if _, err := service.OpenSession(context.Background(), f.opts()); err != nil {
		t.Fatal(err)
	}
	if len(f.prompts) != 1 || len(f.creates) != 1 || len(f.attempts) != 2 || f.attempts[0] != f.attempts[1] {
		t.Fatal(f.prompts, f.creates, f.attempts)
	}
}

func TestServiceBrowserCheckpointFailureDoesNotFailAdmission(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	store := &faultStore{Store: f.store, failAt: 5}
	result, err := (Service{Client: f.client, Store: store}).OpenSession(context.Background(), f.opts())
	if err != nil || !result.Accepted || !result.BrowserOpened || len(result.Warnings) != 1 || result.Warnings[0].Code != "browser_checkpoint_failed" {
		t.Fatalf("%+v %v", result, err)
	}
	if _, err := f.service().OpenSession(context.Background(), f.opts()); err != nil {
		t.Fatal(err)
	}
	if len(f.prompts) != 1 || len(f.creates) != 1 {
		t.Fatal("browser bookkeeping caused duplicate input")
	}
}

func TestFileStoreCancellationDoesNotStealDeliveryLock(t *testing.T) {
	t.Parallel()
	store := FileStore{Dir: t.TempDir()}
	held := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- store.WithSession(context.Background(), "key", func(Record) error { close(held); <-release; return nil })
	}()
	<-held
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	entered := false
	err := store.WithSession(ctx, "key", func(Record) error { entered = true; return nil })
	close(release)
	if releaseErr := <-done; releaseErr != nil {
		t.Fatal(releaseErr)
	}
	if !errors.Is(err, context.DeadlineExceeded) || entered {
		t.Fatalf("lock stolen or cancellation ignored: %v %v", entered, err)
	}
}

func TestServiceRejectsMissingTerminalAndCorruptBindings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, raw string }{
		{"terminal", `{"runner":"codex-tty"}`},
		{"corrupt JSON", `{broken`},
		{"missing discriminator", `{}`},
		{"missing provider", `{"runner":"dsh-web","created":true}`},
		{"missing pending prompt", `{"runner":"dsh-web","pending_request_id":"pending"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newServiceFixture(t)
			path, _ := f.store.path(testKey)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.raw), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := f.service().OpenSession(context.Background(), f.opts()); err == nil {
				t.Fatal("invalid binding accepted")
			}
			if len(f.creates) != 0 || len(f.opened) != 0 || len(f.attempts) != 0 {
				t.Fatal("invalid binding caused external effects")
			}
		})
	}
	f := newServiceFixture(t)
	if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "text"}); err == nil {
		t.Fatal("Send created missing binding")
	}
	for _, key := range []string{"", ".", "..", "../escape", "/absolute", "a/b", "a\\b"} {
		opts := f.opts()
		opts.Key = key
		if _, err := f.service().OpenSession(context.Background(), opts); err == nil {
			t.Fatalf("key accepted: %q", key)
		}
	}
	opts := f.opts()
	opts.Browser = "unknown-browser"
	if _, err := f.service().OpenSession(context.Background(), opts); err == nil {
		t.Fatal("unknown browser accepted")
	}
	if len(f.creates) != 0 || len(f.opened) != 0 {
		t.Fatal("validation had external effects")
	}
}

func TestServiceLegacyPendingInputRecoversWithoutRebinding(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	legacy := State{Runner: "dsh", Workspace: f.store.Dir, Browser: "firefox", DSHSessionID: "old-session", Created: true, PendingRequestID: "old-request", PendingHash: promptHash("old playbook", false), PendingPrompt: "old playbook"}
	if err := f.store.WithSession(context.Background(), testKey, func(r Record) error { return r.Save(legacy) }); err != nil {
		t.Fatal(err)
	}
	result, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "new input"})
	if err != nil || result.SessionID != "old-session" || len(f.creates) != 0 || f.attempts[0] != "old-request" || strings.Join(f.prompts, ",") != "old playbook,new input" {
		t.Fatalf("%+v %v prompts=%v requests=%v", result, err, f.prompts, f.attempts)
	}
}
