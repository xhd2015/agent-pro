package sessionactivity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

func benchSessionBody(id string, events int) string {
	header, _ := json.Marshal(map[string]any{
		"type": "session", "version": 3, "id": id, "createdAt": 10,
		"cwd": "/workspace", "isSeeded": false, "delegationDepth": 0,
		"agentPreset": "std-no-subagent",
	})
	var b []byte
	b = append(b, header...)
	b = append(b, '\n')
	for i := 0; i < events; i++ {
		row, _ := json.Marshal(map[string]any{
			"type": "tool/call", "seq": i, "time": 20 + i,
			"data": json.RawMessage(`{"name":"read","arguments":{"path":"f"}}`),
		})
		b = append(b, row...)
		b = append(b, '\n')
	}
	return string(b)
}

func writeBenchSession(b *testing.B, events int) string {
	b.Helper()
	home := b.TempDir()
	id := "s-bench-" + strconv.Itoa(events)
	path := filepath.Join(home, "sessions", "p", id, "session.v3.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(benchSessionBody(id, events)), 0o644); err != nil {
		b.Fatal(err)
	}
	b.Setenv("DSH_HOME", home)
	b.Setenv("HOME", home)
	return id
}

// BenchmarkInspectRepeat is the pre-cache cost: every poll re-parses the JSONL.
func BenchmarkInspectRepeat(b *testing.B) {
	id := writeBenchSession(b, 4000)
	store, err := dshsessions.NewStore(dshsessions.Options{})
	if err != nil {
		b.Fatal(err)
	}
	if _, err := store.Inspect(id); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.Inspect(id); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReadUnchanged is the /activity poll after Read caches an unchanged file.
func BenchmarkReadUnchanged(b *testing.B) {
	id := writeBenchSession(b, 4000)
	if _, err := Read(RunnerDSH, id); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Read(RunnerDSH, id); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFollowUnchanged is the poll path after Follow caches an unchanged file.
func BenchmarkFollowUnchanged(b *testing.B) {
	id := writeBenchSession(b, 4000)
	if _, _, err := Follow(RunnerDSH, id, 0, 24); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := Follow(RunnerDSH, id, 0, 24); err != nil {
			b.Fatal(err)
		}
	}
}
