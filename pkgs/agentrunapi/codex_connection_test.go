package agentrunapi_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/codex/adhoc"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func adhocConnection(port string) *adhoc.Connection {
	return &adhoc.Connection{Version: 1, Provider: "gateway", BaseURL: "http://127.0.0.1:" + port + "/v1", Model: "exact--upstream", Reasoning: adhoc.ReasoningPolicy{Requested: "default", Effort: "medium", Supported: []string{"medium", "high"}, Origin: "configured"}, Auth: adhoc.Credential{PublicToken: "PROXY_MANAGED"}}
}

func TestAdhocTargetIsBoundBeforeReexecAndSurvivesRetrySendAndResume(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	first := adhocConnection("9991")
	sentinel := errors.New("terminal admission failed")
	opts := agentrunapi.Opts{SessionID: "conversation", Prompt: "initial", AgentRunner: "codex-tty", WorkspaceDir: t.TempDir(), CodexConnection: first, Store: store, Stdout: io.Discard, Stderr: io.Discard}
	opts.RunSession = func(ctx context.Context, o agentrunapi.Opts, m agentstorage.SessionMeta, found bool) error {
		saved, err := store.GetSession(o.SessionID)
		if err != nil || saved.Meta.CodexConnection == nil || saved.Meta.CodexConnection.BaseURL != first.BaseURL {
			t.Fatalf("target not bound before dispatch: %v", err)
		}
		if o.Model != first.Model || o.ModelReasoningEffort != "medium" {
			t.Fatal("wire model/effort was not projected")
		}
		return sentinel
	}
	if err := agentrunapi.AutoSendOrResume(context.Background(), opts); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	// Reconstructed launcher with a new preference must not retarget this conversation.
	opts.CodexConnection = adhocConnection("9992")
	opts.Prompt = "retry"
	opts.RunSession = func(ctx context.Context, o agentrunapi.Opts, m agentstorage.SessionMeta, found bool) error {
		if !found || o.CodexConnection.BaseURL != first.BaseURL {
			t.Fatal("retry applied current preference")
		}
		return nil
	}
	if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	live := false
	opts.Probe = func(agentstorage.Store, agentstorage.SessionMeta) (agentrunapi.ProbeReport, error) {
		return agentrunapi.ProbeReport{RunnerExited: &live}, nil
	}
	opts.SendLive = func(ctx context.Context, o agentrunapi.Opts, m agentstorage.SessionMeta) error {
		if o.CodexConnection.BaseURL != first.BaseURL {
			t.Fatal("live target changed")
		}
		return nil
	}
	if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateSessionRunnerSessionID("conversation", "codex-session"); err != nil {
		t.Fatal(err)
	}
	opts.CodexConnection = nil
	opts.Probe = func(agentstorage.Store, agentstorage.SessionMeta) (agentrunapi.ProbeReport, error) {
		return agentrunapi.ProbeReport{ResumeReady: true}, nil
	}
	opts.ResumeSession = func(ctx context.Context, o agentrunapi.Opts, m agentstorage.SessionMeta) error {
		if o.CodexConnection == nil || o.CodexConnection.BaseURL != first.BaseURL {
			t.Fatal("restart lost connection")
		}
		return nil
	}
	if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(home, "sessions", "conversation", "meta.json"))
	if err != nil || !strings.Contains(string(raw), "codex_connection") {
		t.Fatalf("missing durable binding: %v", err)
	}
}

func TestExistingNativeConversationDoesNotAdoptNewAisPreference(t *testing.T) {
	store, err := agentstorage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("native", agentstorage.SessionMeta{Runner: "codex-tty", SessionID: "native", Model: "old-native", Status: "finished"}); err != nil {
		t.Fatal(err)
	}
	opts := agentrunapi.Opts{SessionID: "native", AgentRunner: "codex-tty", Prompt: "followup", Model: "new-ais-label", CodexConnection: adhocConnection("9991"), Store: store, Stdout: io.Discard, Stderr: io.Discard}
	opts.RunSession = func(ctx context.Context, o agentrunapi.Opts, m agentstorage.SessionMeta, found bool) error {
		if o.CodexConnection != nil || o.Model != "old-native" {
			t.Fatal("existing native session was retargeted")
		}
		return nil
	}
	if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
}

func TestAdhocCredentialsCannotEnterPersistedEnvOrAnotherRunner(t *testing.T) {
	for _, runner := range []string{"codex-tty", "grok-tty"} {
		t.Run(runner, func(t *testing.T) {
			store, err := agentstorage.NewFileStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			c := adhocConnection("9991")
			c.Auth = adhoc.Credential{EnvKey: "PRIVATE_KEY"}
			opts := agentrunapi.Opts{SessionID: "secret", AgentRunner: runner, Prompt: "hi", CodexConnection: c, Env: []string{"PRIVATE_KEY=secret-value"}, Store: store, Stdout: io.Discard, Stderr: io.Discard}
			called := false
			opts.RunSession = func(context.Context, agentrunapi.Opts, agentstorage.SessionMeta, bool) error {
				called = true
				return nil
			}
			if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err == nil || called {
				t.Fatal("unsafe launch was not rejected")
			}
			if _, err := store.GetSession("secret"); err == nil {
				t.Fatal("unsafe credential request was persisted")
			}
		})
	}
}

func TestFollowUpCarriesOnlyBindingGuardAndStoreHome(t *testing.T) {
	line, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{SessionID: "id", AgentRunner: "codex-tty", RequireCodexConnection: true, StoreHome: "/tmp/store with spaces"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(line, "--require-codex-connection-v1") || !strings.Contains(line, "AGENT_RUN_HOME=") {
		t.Fatal("reexec lost guard/store home")
	}
	if strings.Contains(line, "PROXY_MANAGED") || strings.Contains(line, "base_url") {
		t.Fatal("reexec serialized a connection instead of restoring binding")
	}
}
