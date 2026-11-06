package sessions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func benchBody(id string, events int) string {
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

func putBenchSession(b *testing.B, events int, zstdEnc bool) (Store, string) {
	b.Helper()
	root := b.TempDir()
	id := "s-bench-" + strconv.Itoa(events)
	body := benchBody(id, events)
	name := "session.v3.jsonl"
	payload := []byte(body)
	if zstdEnc {
		name = "session.v3.jsonl.zstd"
		enc, err := zstd.NewWriter(nil, zstd.WithEncoderCRC(true))
		if err != nil {
			b.Fatal(err)
		}
		payload = enc.EncodeAll(payload, nil)
		_ = enc.Close()
	}
	path := filepath.Join(root, "project", id, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		b.Fatal(err)
	}
	return Store{Root: root}, id
}

func BenchmarkInspectUnchanged(b *testing.B) {
	resetInspectBodies()
	store, id := putBenchSession(b, 4000, false)
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

func BenchmarkInspectCold(b *testing.B) {
	store, id := putBenchSession(b, 4000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetInspectBodies()
		if _, err := store.Inspect(id); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInspectUnchangedZstd(b *testing.B) {
	resetInspectBodies()
	store, id := putBenchSession(b, 4000, true)
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

func BenchmarkInspectConcurrent(b *testing.B) {
	resetInspectBodies()
	store, id := putBenchSession(b, 4000, false)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := store.Inspect(id); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkViewUnchanged(b *testing.B) {
	resetInspectBodies()
	store, id := putBenchSession(b, 4000, false)
	if _, err := store.View(id); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.View(id); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkViewCold(b *testing.B) {
	store, id := putBenchSession(b, 20000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetInspectBodies()
		if _, err := store.View(id); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInspectSingleflight(b *testing.B) {
	store, id := putBenchSession(b, 4000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetInspectBodies()
		var wg sync.WaitGroup
		for n := 0; n < 8; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := store.Inspect(id); err != nil {
					b.Error(err)
				}
			}()
		}
		wg.Wait()
	}
}
