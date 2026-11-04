package web

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestBrowserNoneSkipsDiscoveryAndPlatformOpener(t *testing.T) {
	t.Parallel()
	client := Client{
		ReadListen: func() (Listen, error) {
			t.Fatal("No Browser tried to discover the server")
			return Listen{}, nil
		},
		OpenBrowser: func(string, string) error {
			t.Fatal("No Browser invoked the platform opener")
			return nil
		},
	}
	for _, raw := range []string{BrowserNone, " NONE ", "None"} {
		if normalized, err := NormalizeBrowser(raw); err != nil || normalized != BrowserNone {
			t.Fatalf("normalize %q: %q %v", raw, normalized, err)
		}
		if err := client.Open("existing-session", raw); err != nil {
			t.Fatal(err)
		}
	}
	if err := client.Open("", BrowserNone); err == nil {
		t.Fatal("No Browser bypassed session identity validation")
	}
}

func TestServiceBrowserNoneSubmitsWithPinnedModelAndSurvivesRestart(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failBrowser = true // Even an unavailable browser must not produce a warning.
	opts := f.opts()
	opts.Browser = " NONE "
	opts.ModelSelection = &ModelSelection{Provider: "route", Model: "reasoner", ReasoningEffort: "high"}
	first, err := f.service().OpenSession(context.Background(), opts)
	if err != nil || !first.Created || !first.Accepted || first.BrowserOpened || len(first.Warnings) != 0 {
		t.Fatalf("%+v %v", first, err)
	}
	if strings.Join(f.methods, ",") != "workspace/create,session/create,session/prompt" {
		t.Fatalf("background execution must submit, not draft: %v", f.methods)
	}
	state, found, err := f.store.Read(context.Background(), testKey)
	if err != nil || !found || !state.Created || !state.InitialDelivered || state.BrowserOpened || state.Browser != BrowserNone || !reflect.DeepEqual(state.InitialModelSelection, opts.ModelSelection) {
		t.Fatalf("%+v %v", state, err)
	}
	changed := opts
	changed.Workspace = t.TempDir()
	changed.Browser = "chrome"
	changed.ModelSelection = &ModelSelection{Provider: "other", Model: "other"}
	for i := 0; i < 2; i++ {
		repeated, err := f.service().OpenSession(context.Background(), changed)
		if err != nil || repeated.SessionID != first.SessionID || repeated.Created || !repeated.Accepted || repeated.BrowserOpened || len(repeated.Warnings) != 0 {
			t.Fatalf("retry: %+v %v", repeated, err)
		}
	}
	sent, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "followup"})
	if err != nil || !sent.Accepted || sent.SessionID != first.SessionID || sent.BrowserOpened || len(sent.Warnings) != 0 {
		t.Fatalf("followup: %+v %v", sent, err)
	}
	if len(f.creates) != 1 || len(f.opened) != 0 || strings.Join(f.prompts, ",") != "initial playbook,followup" {
		t.Fatal(f.creates, f.opened, f.prompts)
	}
	state, _, _ = f.store.Read(context.Background(), testKey)
	if state.Browser != BrowserNone || state.Workspace != opts.Workspace || !reflect.DeepEqual(state.InitialModelSelection, opts.ModelSelection) {
		t.Fatal(state)
	}
}

func TestServiceBrowserNoneKeepsPolicyAcrossInterruptedInitialDelivery(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"creation", "ambiguous prompt admission"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			f := newServiceFixture(t)
			f.failCreate = stage == "creation"
			f.failAfterAccept = stage == "ambiguous prompt admission"
			opts := f.opts()
			opts.Browser = BrowserNone
			if _, err := f.service().OpenSession(context.Background(), opts); err == nil {
				t.Fatal("expected interrupted delivery")
			}
			before, _, _ := f.store.Read(context.Background(), testKey)
			if before.Browser != BrowserNone || before.PendingRequestID == "" || len(f.opened) != 0 {
				t.Fatal(before, f.opened)
			}
			opts.Browser = "firefox"
			retried, err := f.service().OpenSession(context.Background(), opts)
			if err != nil || !retried.Accepted || retried.BrowserOpened || retried.SessionID != before.DSHSessionID || len(retried.Warnings) != 0 {
				t.Fatalf("%+v %v", retried, err)
			}
			after, _, _ := f.store.Read(context.Background(), testKey)
			if after.Browser != BrowserNone || after.PendingRequestID != "" || len(f.opened) != 0 || len(f.prompts) != 1 {
				t.Fatal(after, f.opened, f.prompts)
			}
			for _, id := range f.creates {
				if id != before.DSHSessionID {
					t.Fatal("retry replaced session identity")
				}
			}
			for _, id := range f.attempts {
				if id != before.PendingRequestID {
					t.Fatal("retry replaced input identity")
				}
			}
		})
	}
}

func TestServiceBrowserNoneManualOpenDoesNotDeliverPendingInput(t *testing.T) {
	t.Parallel()
	for _, failed := range []bool{false, true} {
		name := "opens in system browser"
		if failed {
			name = "reports platform failure"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := newServiceFixture(t)
			opts := f.opts()
			opts.Browser = BrowserNone
			created, err := f.service().OpenSession(context.Background(), opts)
			if err != nil {
				t.Fatal(err)
			}
			f.failBeforeAccept = true
			if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "pending"}); err == nil {
				t.Fatal("expected pending followup")
			}
			before, _, _ := f.store.Read(context.Background(), testKey)
			methods := len(f.methods)
			f.failBrowser = failed
			opened, err := f.service().Open(context.Background(), testKey)
			if (err != nil) != failed || opened.BrowserOpened == failed || opened.SessionID != created.SessionID {
				t.Fatalf("%+v %v", opened, err)
			}
			if len(f.opened) != 1 || f.opened[0] != "" || len(f.methods) != methods || len(f.creates) != 1 || len(f.prompts) != 1 {
				t.Fatal(f.opened, f.methods, f.creates, f.prompts)
			}
			after, _, _ := f.store.Read(context.Background(), testKey)
			if after.Browser != BrowserNone || after.BrowserOpened == failed || after.PendingRequestID != before.PendingRequestID || after.PendingPrompt != "pending" {
				t.Fatal(after)
			}
		})
	}
}

func TestServiceNoOpenRemainsAnOperationOnlyOverride(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	opts := f.opts()
	opts.NoOpen = true
	if result, err := f.service().OpenSession(context.Background(), opts); err != nil || !result.Accepted || result.BrowserOpened {
		t.Fatalf("%+v %v", result, err)
	}
	opts.NoOpen = false
	if result, err := f.service().OpenSession(context.Background(), opts); err != nil || !result.BrowserOpened {
		t.Fatalf("%+v %v", result, err)
	}
	if len(f.prompts) != 1 || len(f.opened) != 1 || f.opened[0] != "brave" {
		t.Fatal(f.prompts, f.opened)
	}
}
