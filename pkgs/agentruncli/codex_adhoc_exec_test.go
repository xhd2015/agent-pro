//go:build !windows

package agentruncli

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/codex/adhoc"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// e2e: the receiver reloads a binding and resolves the key only after the
// tty-watch/Terminal process boundary. A shell fixture avoids real inference.
func TestAdhocExecReceiverLoadsSecretOnlyInFinalChild(t *testing.T) {
	home := t.TempDir()
	const secret = "receiver-fixture-private-key"
	token := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(token, []byte(secret), 0600); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(t.TempDir(), "codex")
	script := "#!/bin/sh\n[ \"$AGENT_PRO_CODEX_AUTH_TOKEN\" = '" + secret + "' ] || exit 13\nprintf FINAL_AUTH_OK\n"
	if err := os.WriteFile(fake, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	c := &adhoc.Connection{Version: 1, Provider: "fixture", BaseURL: "http://127.0.0.1:1/v1", Model: "raw-model", Reasoning: adhoc.ReasoningPolicy{Requested: "high", Effort: "high", Supported: []string{"high"}, Origin: "api"}, Auth: adhoc.Credential{File: token}}
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("bound", agentstorage.SessionMeta{SessionID: "bound", Runner: "codex-tty", Status: "prepared", CodexConnection: c}); err != nil {
		t.Fatal(err)
	}
	argv, err := c.Apply([]string{fake}, true)
	if err != nil {
		t.Fatal(err)
	}
	args := append([]string{"-test.run=^TestAdhocExecReceiverProcess$", "--", "__codex-adhoc-exec", home, "bound", "--"}, argv...)
	env := []string{}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, adhoc.TokenEnv+"=") && !strings.HasPrefix(e, "AGENT_PRO_RECEIVER_TEST_CHILD=") {
			env = append(env, e)
		}
	}
	env = append(env, "AGENT_PRO_RECEIVER_TEST_CHILD=1")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], args...)
	command.Env = env
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "FINAL_AUTH_OK") {
		t.Fatalf("final receiver: %v %s", err, output)
	}
	meta, err := os.ReadFile(filepath.Join(home, "sessions", "bound", "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	launch, _ := json.Marshal(struct{ Args, Env []string }{args, env})
	if strings.Contains(string(meta)+string(launch)+string(output), secret) {
		t.Fatal("key escaped final-child environment into launch artifacts")
	}
	if err := os.WriteFile(token, nil, 0600); err != nil {
		t.Fatal(err)
	}
	command = exec.CommandContext(ctx, os.Args[0], args...)
	command.Env = env
	output, err = command.CombinedOutput()
	if err == nil || strings.Contains(string(output), "FINAL_AUTH_OK") || strings.Contains(string(output), secret) {
		t.Fatal("missing key did not fail closed")
	}
}

func TestAdhocExecReceiverProcess(t *testing.T) {
	if os.Getenv("AGENT_PRO_RECEIVER_TEST_CHILD") != "1" {
		t.Skip("receiver subprocess helper")
	}
	at := -1
	for i, arg := range os.Args {
		if arg == "--" {
			at = i
			break
		}
	}
	if at < 0 {
		os.Exit(2)
	}
	if err := Handle(os.Args[at+1:]); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
