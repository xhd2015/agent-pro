package sessions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func header(id string) string {
	return headerAt(id, 3)
}
func headerAt(id string, version int) string {
	b, _ := json.Marshal(map[string]any{"type": "session", "version": version, "id": id, "createdAt": 10, "cwd": "/workspace", "isSeeded": false, "delegationDepth": 0, "agentPreset": "std-no-subagent"})
	return string(b) + "\n"
}
func row(typ string, seq int, data string) string {
	b, _ := json.Marshal(map[string]any{"type": typ, "seq": seq, "time": 20 + seq, "data": json.RawMessage(data)})
	return string(b) + "\n"
}
func put(t *testing.T, root, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(root, "project", dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func requireError(t *testing.T, err error, part string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), part) {
		t.Fatalf("error = %v, want %q", err, part)
	}
}

func TestRootPrecedence(t *testing.T) {
	home := t.TempDir()
	env := t.TempDir()
	explicit := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DSH_HOME", env)
	for _, tc := range []struct {
		opts Options
		want string
	}{{Options{}, filepath.Join(env, "sessions")}, {Options{Home: explicit}, filepath.Join(explicit, "sessions")}, {Options{Home: explicit, SessionsDir: home}, home}, {Options{Home: "~/other"}, filepath.Join(home, "other", "sessions")}, {Options{Home: `~\other`}, filepath.Join(home, "other", "sessions")}, {Options{SessionsDir: "~"}, home}} {
		got, err := ResolveRoot(tc.opts)
		if err != nil || got != tc.want {
			t.Fatalf("%+v: %q %v want %q", tc.opts, got, err, tc.want)
		}
	}
	t.Setenv("DSH_HOME", "  ")
	got, err := NewStore(Options{})
	if err != nil || got.Root != filepath.Join(home, ".dsh", "sessions") {
		t.Fatalf("%+v %v", got, err)
	}
}
func TestInspectLastSessionTitle(t *testing.T) {
	root := t.TempDir()
	body := header("s-title") + row("session/title", 0, `{"title":"first"}`) + row("turn/start", 1, `{"turn":1}`) + row("session/title", 2, `{"title":"Orchestrator agents and todo tracker"}`)
	put(t, root, "s-title", "session.v3.jsonl", body)
	store := Store{Root: root}
	got, err := store.Inspect("s-title")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Orchestrator agents and todo tracker" {
		t.Fatalf("title=%q", got.Title)
	}
}

func TestReadTimelineAndExactHeaderIdentity(t *testing.T) {
	root := t.TempDir()
	body := header("../opaque") + row("turn/start", 0, `{"turn":1}`) + row("request/header", 1, `{"header":{"config":{"provider":"p","model":"m"}}}`) + row("turn/end", 2, `{"turn":1,"reason":{"kind":"completed"}}`) + row("external/opaque", 3, `null`)
	path := put(t, root, "not-the-id", "session.v3.jsonl", body+`{"partial":`)
	store := Store{Root: root}
	got, err := store.Inspect("../opaque")
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != path || got.ID != "../opaque" || got.Outcome != "completed" || got.Provider != "p" || got.Model != "m" || got.UpdatedAt != 23 || len(got.Events) != 4 {
		t.Fatalf("%+v", got)
	}
	if string(got.Events[3].Raw) != strings.TrimSpace(row("external/opaque", 3, `null`)) {
		t.Fatal("raw changed")
	}
	_, err = store.Inspect("not-the-id")
	requireError(t, err, "not found")
	after, _ := os.ReadFile(path)
	if string(after) != body+`{"partial":` {
		t.Fatal("reader changed file")
	}
}
func TestGenerationSelectionAndWarnings(t *testing.T) {
	root := t.TempDir()
	put(t, root, "good", "session.v3.jsonl", header("good"))
	put(t, root, "good", "session.v03.jsonl", "bad")
	put(t, root, "future", "session.v3.jsonl", header("future"))
	put(t, root, "future", "session.v10.jsonl", strings.Replace(header("future"), `"version":3`, `"version":10`, 1))
	put(t, root, "bad", "session.v3.jsonl", header("bad")+"oops\n")
	store := Store{Root: root}
	list, warnings, err := store.List()
	if err != nil || len(list) != 1 || len(warnings) != 2 {
		t.Fatalf("%+v %v %v", list, warnings, err)
	}
	_, err = store.Inspect("future")
	requireError(t, err, "v10")
	_, err = store.Inspect("bad")
	requireError(t, err, "event row")
}
func TestMixedEncodingAcrossGenerationsAndSessions(t *testing.T) {
	for _, dir := range []string{"same", "other"} {
		t.Run(dir, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "same", "session.v3.jsonl", header("a"))
			put(t, root, dir, "session.v1.jsonl.zstd", "")
			store := Store{root}
			_, _, err := store.List()
			requireError(t, err, "mixed")
			_, err = store.Inspect("a")
			requireError(t, err, "mixed")
		})
	}
}
func TestDuplicateIdentityIncludingCorruptLog(t *testing.T) {
	root := t.TempDir()
	put(t, root, "a", "session.v3.jsonl", header("same"))
	put(t, root, "b", "session.v3.jsonl", header("same")+"bad\n")
	store := Store{root}
	_, _, err := store.List()
	requireError(t, err, "duplicate")
	_, err = store.Inspect("same")
	requireError(t, err, "duplicate")
}
func TestMissingRoot(t *testing.T) {
	store := Store{filepath.Join(t.TempDir(), "absent")}
	list, warnings, err := store.List()
	if err != nil || list == nil || len(list) != 0 || len(warnings) != 0 {
		t.Fatal(list, warnings, err)
	}
	_, err = store.Inspect("x")
	requireError(t, err, "not found")
}
func TestRejectMalformedCompleteRows(t *testing.T) {
	cases := []string{`null`, `{}`, `{"type":"x","seq":null,"time":0,"data":{}}`, `{"type":"x","seq":0,"data":{}}`, `{"type":"x","seq":0,"time":null,"data":{}}`, `{"type":"x","seq":1,"time":0,"data":{}}`, `{"type":"x","seq":0,"time":0}`, strings.TrimSpace(row("turn/start", 0, `null`)), strings.TrimSpace(row("turn/start", 0, `{"turn":null}`)), strings.TrimSpace(row("turn/end", 0, `{"turn":1,"reason":null}`)), strings.TrimSpace(row("turn/end", 0, `{"turn":1,"reason":{"kind":2}}`)), strings.TrimSpace(row("request/header", 0, `{"header":{"config":{"provider":"p","model":null}}}`)), strings.TrimSpace(row("request/header", 0, `{"header":{}}`))}
	for _, bad := range cases {
		t.Run(bad, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "a", "session.v3.jsonl", header("a")+bad+"\n")
			_, err := (&Store{root}).Inspect("a")
			requireError(t, err, "event")
		})
	}
}
func TestHeaderMismatch(t *testing.T) {
	root := t.TempDir()
	put(t, root, "a", "session.v3.jsonl", strings.Replace(header("a"), `"version":3`, `"version":2`, 1))
	_, err := (&Store{root}).Inspect("a")
	requireError(t, err, "header version")
}
func TestZstdConcatenatedFramesAndCorruption(t *testing.T) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	frames := enc.EncodeAll([]byte(header("a")), nil)
	frames = append(frames, enc.EncodeAll([]byte(row("turn/start", 0, `{"turn":1}`)), nil)...)
	for _, kind := range []string{"valid", "truncated", "checksum", "incomplete-line"} {
		t.Run(kind, func(t *testing.T) {
			data := append([]byte(nil), frames...)
			switch kind {
			case "truncated":
				data = data[:len(data)-2]
			case "checksum":
				data[len(data)-1] ^= 1
			case "incomplete-line":
				data = append(data, enc.EncodeAll([]byte("{}"), nil)...)
			}
			root := t.TempDir()
			put(t, root, "a", "session.v3.jsonl.zstd", string(data))
			got, err := (&Store{root}).Inspect("a")
			if kind != "valid" {
				if err == nil {
					t.Fatal("accepted damaged stream")
				}
				return
			}
			if err != nil || got.Outcome != "incomplete" || len(got.Events) != 1 || got.Encoding != "zstd" {
				t.Fatal(got, err)
			}
		})
	}
}
func TestLinksAndNonregularArtifactsAreRejected(t *testing.T) {
	root := t.TempDir()
	path := put(t, root, "a", "session.v3.jsonl", header("a"))
	target := filepath.Join(t.TempDir(), "target")
	if err := os.Rename(path, target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	store := Store{root}
	list, warnings, err := store.List()
	if err != nil || len(list) != 0 || len(warnings) != 1 {
		t.Fatal(list, warnings, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	_, err = store.Inspect("a")
	requireError(t, err, "regular")
}

func TestNewestFirstAndRawExtensions(t *testing.T) {
	root := t.TempDir()
	put(t, root, "b", "session.v3.jsonl", header("b"))
	put(t, root, "a", "session.v3.jsonl", header("a"))
	raw := `{"type":"external/opaque","seq":0,"time":99,"data":{},"sourceEventSeqs":[[1,3]],"ignorable":true}`
	put(t, root, "c", "session.v3.jsonl", header("c")+raw+"\n")
	store := Store{root}
	list, _, err := store.List()
	if err != nil || len(list) != 3 || list[0].ID != "c" || list[1].ID != "a" {
		t.Fatal(list, err)
	}
	got, err := store.Inspect("c")
	if err != nil || string(got.Events[0].Raw) != raw {
		t.Fatal(got, err)
	}
}

func TestReadV4TimelineAndActivity(t *testing.T) {
	root := t.TempDir()
	body := headerAt("v4-sess", 4) +
		row("turn/start", 0, `{"turn":1}`) +
		`{"type":"user/message","seq":1,"time":21,"data":{"content":[{"type":"text","text":"hi"}],"source":{"kind":"user"}},"surfaceOp":"append"}` + "\n" +
		row("tool/call", 2, `{"name":"bash"}`) +
		row("request/header", 3, `{"header":{"config":{"provider":"p","model":"m"}}}`)
	path := put(t, root, "v4-sess", "session.v4.jsonl", body)
	got, err := (&Store{root}).Inspect("v4-sess")
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != path || got.FormatVersion != 4 || got.Provider != "p" || got.Model != "m" || got.Outcome != "incomplete" || got.AgentPreset != "std-no-subagent" || len(got.Events) != 4 {
		t.Fatalf("%+v", got)
	}
	act := Classify(got.Events, true)
	if act.Kind != ActivityToolCalling || act.Detail != "bash" || !act.Live {
		t.Fatalf("activity=%+v", act)
	}
}

func TestReadV4Zstd(t *testing.T) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	data := enc.EncodeAll([]byte(headerAt("z4", 4)+row("turn/start", 0, `{"turn":1}`)), nil)
	root := t.TempDir()
	put(t, root, "z4", "session.v4.jsonl.zstd", string(data))
	got, err := (&Store{root}).Inspect("z4")
	if err != nil || got.FormatVersion != 4 || got.Encoding != "zstd" || got.Outcome != "incomplete" {
		t.Fatal(got, err)
	}
}

func TestV4PreferredOverV3InSameDirectory(t *testing.T) {
	root := t.TempDir()
	put(t, root, "both", "session.v3.jsonl", header("both"))
	put(t, root, "both", "session.v4.jsonl", headerAt("both", 4)+row("turn/start", 0, `{"turn":1}`))
	got, err := (&Store{root}).Inspect("both")
	if err != nil || got.FormatVersion != 4 || got.Outcome != "incomplete" {
		t.Fatal(got, err)
	}
}

func TestUnsupportedGenerationsStayRefused(t *testing.T) {
	root := t.TempDir()
	put(t, root, "old", "session.v2.jsonl", headerAt("old", 2))
	put(t, root, "new", "session.v5.jsonl", headerAt("new", 5))
	store := Store{root}
	_, err := store.Inspect("old")
	requireError(t, err, "v2")
	_, err = store.Inspect("new")
	requireError(t, err, "v5")
	if err == nil || !strings.Contains(err.Error(), "v3 and v4") {
		t.Fatalf("want supports v3 and v4: %v", err)
	}
}
