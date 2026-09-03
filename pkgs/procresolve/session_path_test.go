package procresolve

import (
	"strings"
	"testing"
)

func TestGrokSessionDirFromPath(t *testing.T) {
	path := "/Users/x/.grok/sessions/%2Ftmp%2Fproj/019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01/events.jsonl"
	dir, sid, ok := GrokSessionDirFromPath(path)
	if !ok {
		t.Fatal("expected ok")
	}
	if sid != "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01" {
		t.Fatalf("sid=%q", sid)
	}
	wantDir := "/Users/x/.grok/sessions/%2Ftmp%2Fproj/019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01"
	if dir != wantDir {
		t.Fatalf("dir=%q want %q", dir, wantDir)
	}
}

func TestParseSessionFromPath_CodexThreadWriterLock(t *testing.T) {
	const want = "01a064e5-b881-7bb0-b44b-ef6ddd27874d"
	cases := []struct {
		name string
		path string
		ok   bool
		id   string
	}{
		{
			name: "abs lock",
			path: "/Users/x/.codex/thread-writer-locks/" + want + ".lock",
			ok:   true,
			id:   want,
		},
		{
			name: "relative lock",
			path: ".codex/thread-writer-locks/" + want + ".lock",
			ok:   true,
			id:   want,
		},
		{
			name: "uppercase hex lowercased",
			path: "/Users/x/.codex/thread-writer-locks/" + strings.ToUpper(want) + ".lock",
			ok:   true,
			id:   want,
		},
		{
			name: "not a uuid",
			path: "/Users/x/.codex/thread-writer-locks/not-a-uuid.lock",
		},
		{
			name: "extra suffix",
			path: "/Users/x/.codex/thread-writer-locks/" + want + ".lock.bak",
		},
		{
			name: "prefixed basename",
			path: "/Users/x/.codex/thread-writer-locks/prefix-" + want + ".lock",
		},
		{
			name: "rollout still works",
			path: "/Users/x/.codex/sessions/2026/09/03/rollout-2026-09-03T09-32-51-" + want + ".jsonl",
			ok:   true,
			id:   want,
		},
		{
			name: "grok unchanged",
			path: "/Users/x/.grok/sessions/%2Ftmp/019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01/events.jsonl",
			ok:   true,
			id:   "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, sid, ok := ParseSessionFromPath(tc.path)
			if ok != tc.ok {
				t.Fatalf("ok=%v want %v kind=%q sid=%q", ok, tc.ok, kind, sid)
			}
			if !tc.ok {
				return
			}
			if strings.Contains(tc.path, ".grok/") {
				if kind != "grok" {
					t.Fatalf("kind=%q want grok", kind)
				}
			} else if kind != "codex" {
				t.Fatalf("kind=%q want codex", kind)
			}
			if sid != tc.id {
				t.Fatalf("sid=%q want %q", sid, tc.id)
			}
		})
	}
}

func TestResolveFromPID_GrokIgnoresDirectoryScanOpens(t *testing.T) {
	const want = "01a06639-3c96-7ff0-9256-fd0daa519985"
	const stale = "01a0464c-08e4-7061-a893-2773ce974845"
	staleDir := "/Users/x/.grok/sessions/%2Fold/" + stale
	staleSub := staleDir + "/subagents"
	primary := "/Users/x/.grok/sessions/%2Ftmp/" + want + "/events.jsonl"
	res, err := ResolveFromPID(200, Options{
		ListProcs: func() []Proc {
			return []Proc{
				{PID: 100, PPID: 1, Cmd: "agent-run"},
				{PID: 200, PPID: 100, Cmd: "/usr/local/bin/grok"},
			}
		},
		Lsof: func(pid int) []string {
			if pid == 200 {
				// Scan dirs first (old first-hit bug), then primary artifact.
				return []string{staleDir, staleSub, primary}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Kind != "grok" || res.SessionID != want || res.Confidence != "hard" {
		t.Fatalf("res=%+v want session %s", res, want)
	}
}

func TestResolveFromPID_GrokDirectoryOnlyIsSoftMiss(t *testing.T) {
	const stale = "01a0464c-08e4-7061-a893-2773ce974845"
	staleDir := "/Users/x/.grok/sessions/%2Fold/" + stale
	res, err := ResolveFromPID(200, Options{
		ListProcs: func() []Proc {
			return []Proc{
				{PID: 200, PPID: 1, Cmd: "/usr/local/bin/grok"},
			}
		},
		Lsof: func(pid int) []string {
			if pid == 200 {
				return []string{staleDir, staleDir + "/subagents"}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Kind != "none" || res.SessionID != "" || res.Confidence != "" {
		t.Fatalf("directory-only opens must soft-miss; got %+v", res)
	}
}

func TestIsGrokPrimarySessionOpenPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/x/.grok/sessions/%2Fa/01a06639-3c96-7ff0-9256-fd0daa519985/events.jsonl", true},
		{"/x/.grok/sessions/%2Fa/01a06639-3c96-7ff0-9256-fd0daa519985/updates.jsonl", true},
		{"/x/.grok/sessions/%2Fa/01a06639-3c96-7ff0-9256-fd0daa519985", false},
		{"/x/.grok/sessions/%2Fa/01a06639-3c96-7ff0-9256-fd0daa519985/summary.json", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isGrokPrimarySessionOpenPath(tc.path); got != tc.want {
			t.Fatalf("path %q: got %v want %v", tc.path, got, tc.want)
		}
	}
}

func TestResolveFromPID_CodexLockPath(t *testing.T) {
	const want = "01a064e5-b881-7bb0-b44b-ef6ddd27874d"
	lock := "/Users/x/.codex/thread-writer-locks/" + want + ".lock"
	res, err := ResolveFromPID(200, Options{
		ListProcs: func() []Proc {
			return []Proc{
				{PID: 100, PPID: 1, Cmd: "agent-run"},
				{PID: 200, PPID: 100, Cmd: "/usr/local/bin/codex"},
			}
		},
		Lsof: func(pid int) []string {
			if pid == 200 {
				return []string{lock}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Kind != "codex" || res.SessionID != want || res.Confidence != "hard" {
		t.Fatalf("res=%+v", res)
	}
}

func TestParseLsofFnByPID(t *testing.T) {
	raw := []byte("p100\nf1\nn/tmp/a\np200\nn/tmp/b\nn/tmp/a\n")
	got := parseLsofFnByPID(raw)
	if len(got[100]) != 1 || got[100][0] != "/tmp/a" {
		t.Fatalf("pid 100: %#v", got[100])
	}
	if len(got[200]) != 2 {
		t.Fatalf("pid 200: %#v", got[200])
	}
}
