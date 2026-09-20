package agentruncli

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	flags "github.com/xhd2015/less-flags"
)

func TestDSHWebArgv(t *testing.T) {
	workspace := t.TempDir()
	resolved, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatal(err)
	}
	prompt := "  /brainstorm 'quoted' $HOME\n你好\n"
	for _, tc := range []struct {
		name           string
		open, noSubmit bool
		wantFlags      []string
	}{
		{"open draft", true, true, []string{"--no-submit"}},
		{"open submit", true, false, nil},
		{"submit without browser", false, false, []string{"--no-open"}},
		{"draft without browser", false, true, []string{"--no-submit", "--no-open"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := dshWebArgv(nil, prompt, "", workspace, tc.open, tc.noSubmit)
			if err != nil {
				t.Fatal(err)
			}
			want := append([]string{"web", "open", resolved, "--prompt", prompt}, tc.wantFlags...)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("argv=%q want %q", got, want)
			}
		})
	}
}

func TestDSHWebArgvPromptFileAndDefaultDirectory(t *testing.T) {
	// The caller's path is forwarded, not read or expanded into argv.
	path := filepath.Join("relative", "task file.txt")
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	cwd, err := resolveRunDir(".")
	if err != nil {
		t.Fatal(err)
	}
	got, err := dshWebArgv(nil, "", path, "", true, true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"web", "open", cwd, "--prompt-file", absPath, "--no-submit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
}

func TestDSHWebRejectsUnsupportedFlagsBeforeLaunch(t *testing.T) {
	for _, flag := range []string{
		"--json", "--model", "--model-reasoning-effort", "--session", "--session-id",
		"--session-id-from-prompt", "--auto-send-or-resume", "--new-terminal",
		"--exec-in-terminal", "--no-exec-in-terminal", "--keep-tty", "--detach",
		"--allow-relocate-resume-session-dir", "--resume-from-grok-session", "--fork",
		"--agent-runner-binary", "--agent-runner-config-home", "--prepend-path",
		"-e", "--env", "--env-file", "--color", "--exit-on-idle", "--idle-timeout",
		"--event-bus-url", "--event-bus-token",
	} {
		t.Run(flag, func(t *testing.T) {
			_, err := dshWebArgv(flags.Flags{{Flag: flag}}, "task", "", "", true, true)
			if err == nil || !strings.Contains(err.Error(), "dsh-web does not support "+flag) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	for _, tc := range []struct{ prompt, file, want string }{
		{"task", "task.txt", "mutually exclusive"},
		{" \n", "", "requires a prompt"},
	} {
		_, err := dshWebArgv(nil, tc.prompt, tc.file, "", true, true)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("err=%v want %q", err, tc.want)
		}
	}
}

func TestDSHWebCLIHandoff(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a POSIX executable shim")
	}
	root := t.TempDir()
	capture := filepath.Join(root, "argv")
	home := filepath.Join(root, "unused-agent-run-home")
	script := "#!/bin/sh\nprintf '%s\\0' \"$@\" > \"$DSH_TEST_ARGV\"\nexit \"${DSH_TEST_EXIT:-0}\"\n"
	if err := os.WriteFile(filepath.Join(root, "dsh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DSH_TEST_ARGV", capture)
	t.Setenv("AGENT_RUN_HOME", home)
	prompt := " /brainstorm 'quoted'\n你好\n"
	if err := Handle([]string{"run", "--agent-runner=dsh-web", "--dir", root, "--open", "--no-submit", prompt}); err != nil {
		t.Fatal(err)
	}
	assertArgs := func(want []string) {
		t.Helper()
		data, err := os.ReadFile(capture)
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("argv=%q want %q", got, want)
		}
	}
	workspace, err := resolveRunDir(root)
	if err != nil {
		t.Fatal(err)
	}
	assertArgs([]string{"web", "open", workspace, "--prompt", prompt, "--no-submit"})

	file := filepath.Join(root, "long task.txt")
	body := strings.Repeat("Unicode 你好\n", 100000)
	if err := os.WriteFile(file, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Handle([]string{"--agent-runner", "dsh-web", "run", "--dir", root, "--prompt-file", file}); err != nil {
		t.Fatal(err)
	}
	assertArgs([]string{"web", "open", workspace, "--prompt-file", file, "--no-open"})
	if data, err := os.ReadFile(file); err != nil || string(data) != body {
		t.Fatalf("prompt file changed: %v", err)
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("browser handoff must not create agent-run storage: %v", err)
	}
	t.Setenv("DSH_TEST_EXIT", "7")
	if err := Handle([]string{"run", "--agent-runner", "dsh-web", prompt}); err == nil || !strings.Contains(err.Error(), "dsh-web launch failed") {
		t.Fatalf("expected launch failure, got %v", err)
	}
	if err := os.Remove(capture); err != nil {
		t.Fatal(err)
	}
	if err := Handle([]string{"run", "--agent-runner", "dsh-web", "--color", prompt}); err == nil || !strings.Contains(err.Error(), "does not support --color") {
		t.Fatalf("expected incompatible flag failure, got %v", err)
	}
	if _, err := os.Stat(capture); !os.IsNotExist(err) {
		t.Fatalf("unsupported flag launched dsh: %v", err)
	}
}
