package agentruncli

import (
	"bytes"
	"io"
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
			got, err := dshWebArgv(prompt, "", workspace, "", tc.open, tc.noSubmit)
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
	got, err := dshWebArgv("", path, "", "", true, true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"web", "open", cwd, "--prompt-file", absPath, "--no-submit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
}

func TestDSHWebArgvBrowserForwardedAsTwoTokens(t *testing.T) {
	workspace := t.TempDir()
	resolved, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatal(err)
	}
	prompt := "task"
	got, err := dshWebArgv(prompt, "", workspace, "brave", true, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"web", "open", resolved, "--prompt", prompt, "--browser", "brave"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
	// Without --open the browser value still forwards beside --no-open.
	got, err = dshWebArgv(prompt, "", workspace, "firefox", false, true)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"web", "open", resolved, "--prompt", prompt, "--browser", "firefox", "--no-submit", "--no-open"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
	// File-backed prompts keep the browser beside --prompt-file.
	got, err = dshWebArgv("", filepath.Join(resolved, "task.txt"), workspace, "brave", true, false)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"web", "open", resolved, "--prompt-file", filepath.Join(resolved, "task.txt"), "--browser", "brave"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
	// Empty browser forwards no tokens.
	got, err = dshWebArgv(prompt, "", workspace, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"web", "open", resolved, "--prompt", prompt}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
}

// Parsed flags outside the consumed set announce themselves and are dropped;
// they never fail the handoff. Consumed flags stay silent.
func TestDSHWebNoticesIgnoreParsedFlags(t *testing.T) {
	consumed := flags.Flags{
		{Flag: "--agent-runner", Value: "dsh-web"},
		{Flag: "--dir", Value: "."},
		{Flag: "--prompt-file", Value: "task.txt"},
		{Flag: "--browser", Value: "brave"},
		{Flag: "--open"},
		{Flag: "--no-submit"},
		{Flag: "-v"},
		{Flag: "--verbose"},
	}
	var silent bytes.Buffer
	printDshWebNotices(&silent, consumed, false)
	if silent.Len() != 0 {
		t.Fatalf("consumed flags must stay silent, got %q", silent.String())
	}
	recorded := flags.Flags{
		{Flag: "--session-id-from-prompt"},
		{Flag: "-v"},
		{Flag: "--color"},
		{Flag: "--detach"},
		{Flag: "-e", Value: "K=V"},
		{Flag: "--env", Value: "K2=V2"},
	}
	var buf bytes.Buffer
	printDshWebNotices(&buf, recorded, false)
	want := strings.Join([]string{
		"notice: dsh-web ignores --session-id-from-prompt; it only launches a DSH Web session",
		"notice: dsh-web ignores --color; it only launches a DSH Web session",
		"notice: dsh-web ignores --detach; it only launches a DSH Web session",
		"notice: dsh-web ignores -e; it only launches a DSH Web session",
		"notice: dsh-web ignores --env; it only launches a DSH Web session",
		"",
	}, "\n")
	if got := buf.String(); got != want {
		t.Fatalf("notices=%q want %q", got, want)
	}
}

// The notice prefix is gray-wrapped only when color is enabled; the body stays plain.
func TestDSHWebNoticePrefixGray(t *testing.T) {
	if got, want := grayNoticePrefix(false), "notice:"; got != want {
		t.Fatalf("plain prefix=%q want %q", got, want)
	}
	if got, want := grayNoticePrefix(true), "\x1b[90mnotice:\x1b[0m"; got != want {
		t.Fatalf("gray prefix=%q want %q", got, want)
	}
	var buf bytes.Buffer
	printDshWebNotices(&buf, flags.Flags{{Flag: "--color"}}, true)
	want := "\x1b[90mnotice:\x1b[0m dsh-web ignores --color; it only launches a DSH Web session\n"
	if got := buf.String(); got != want {
		t.Fatalf("notices=%q want %q", got, want)
	}
}

// Prompt mutual exclusion and the missing-prompt check stay fatal.
func TestDSHWebArgvPromptRulesStayFatal(t *testing.T) {
	for _, tc := range []struct{ prompt, file, want string }{
		{"task", "task.txt", "mutually exclusive"},
		{" \n", "", "requires a prompt"},
	} {
		_, err := dshWebArgv(tc.prompt, tc.file, "", "", true, true)
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
}

// Unconsumed parsed flags print one stderr notice each, in recorded order, and
// the handoff still proceeds with a clean dsh argv. Captured stderr is a pipe,
// so the auto color policy renders the plain prefix.
func TestDSHWebCLIIgnoresFlagsWithNotices(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a POSIX executable shim")
	}
	root := t.TempDir()
	capture := filepath.Join(root, "argv")
	script := "#!/bin/sh\nprintf '%s\\0' \"$@\" > \"$DSH_TEST_ARGV\"\nexit \"${DSH_TEST_EXIT:-0}\"\n"
	if err := os.WriteFile(filepath.Join(root, "dsh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DSH_TEST_ARGV", capture)
	t.Setenv("AGENT_RUN_HOME", filepath.Join(root, "unused-agent-run-home"))
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	handleErr := Handle([]string{"run", "--agent-runner", "dsh-web", "--dir", root, "--session-id-from-prompt", "--color", "--detach", "task"})
	os.Stderr = stderr
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if handleErr != nil {
		t.Fatalf("ignored flags must not fail the handoff, got %v", handleErr)
	}
	file, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := resolveRunDir(root)
	if err != nil {
		t.Fatal(err)
	}
	gotArgs := strings.Split(strings.TrimSuffix(string(file), "\x00"), "\x00")
	wantArgs := []string{"web", "open", workspace, "--prompt", "task", "--no-open"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("argv=%q want %q", gotArgs, wantArgs)
	}
	wantNotices := strings.Join([]string{
		"notice: dsh-web ignores --session-id-from-prompt; it only launches a DSH Web session",
		"notice: dsh-web ignores --color; it only launches a DSH Web session",
		"notice: dsh-web ignores --detach; it only launches a DSH Web session",
		"",
	}, "\n")
	if got := string(data); got != wantNotices {
		t.Fatalf("stderr=%q want %q", got, wantNotices)
	}
	if strings.Contains(string(data), "\x1b[") {
		t.Fatalf("stderr must stay plain for a non-TTY writer: %q", string(data))
	}
	// Unregistered flags still fail at less-flags Parse; nothing launches.
	if err := os.Remove(capture); err != nil {
		t.Fatal(err)
	}
	err = Handle([]string{"run", "--agent-runner", "dsh-web", "--colr", "task"})
	if err == nil || !strings.Contains(err.Error(), "unrecognized flag: --colr") {
		t.Fatalf("expected parse failure for unregistered flag, got %v", err)
	}
	if _, err := os.Stat(capture); !os.IsNotExist(err) {
		t.Fatalf("unregistered flag launched dsh: %v", err)
	}
}

func TestDSHWebCLIBrowserFlag(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a POSIX executable shim")
	}
	root := t.TempDir()
	capture := filepath.Join(root, "argv")
	script := "#!/bin/sh\nprintf '%s\\0' \"$@\" > \"$DSH_TEST_ARGV\"\nexit \"${DSH_TEST_EXIT:-0}\"\n"
	if err := os.WriteFile(filepath.Join(root, "dsh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DSH_TEST_ARGV", capture)
	t.Setenv("AGENT_RUN_HOME", filepath.Join(root, "unused-agent-run-home"))
	workspace, err := resolveRunDir(root)
	if err != nil {
		t.Fatal(err)
	}
	readArgs := func() []string {
		t.Helper()
		data, err := os.ReadFile(capture)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
	}
	// Space form: browser lands in the captured dsh argv as two tokens.
	if err := Handle([]string{"run", "--agent-runner", "dsh-web", "--dir", root, "--browser", "brave", "task"}); err != nil {
		t.Fatal(err)
	}
	if got, want := readArgs(), []string{"web", "open", workspace, "--prompt", "task", "--browser", "brave", "--no-open"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
	// Equals form parses the same way.
	if err := Handle([]string{"run", "--agent-runner", "dsh-web", "--browser=firefox", "--dir", root, "--open", "task"}); err != nil {
		t.Fatal(err)
	}
	if got, want := readArgs(), []string{"web", "open", workspace, "--prompt", "task", "--browser", "firefox"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
	// Other runners reject --browser before any dispatch or launch.
	if err := os.Remove(capture); err != nil {
		t.Fatal(err)
	}
	err = Handle([]string{"run", "--agent-runner", "codex-tty", "--browser", "brave", "task"})
	if err == nil || err.Error() != "--browser requires --agent-runner dsh-web" {
		t.Fatalf("expected --browser rejection, got %v", err)
	}
	if _, err := os.Stat(capture); !os.IsNotExist(err) {
		t.Fatalf("rejected --browser launched dsh: %v", err)
	}
}
