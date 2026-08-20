package run

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"
)

func TestServe_initializeAfterFixedDelay(t *testing.T) {
	delay := 50 * time.Millisecond
	cfg := Config{Name: "slow_01", Delay: &delay}

	clientR, serverW := io.Pipe()
	serverR, clientW := io.Pipe()
	defer clientR.Close()
	defer clientW.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(serverR, serverW, io.Discard, cfg)
	}()

	start := time.Now()
	if _, err := io.WriteString(clientW, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`+"\n"); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(clientR).ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Fatalf("initialize returned too fast: %s", elapsed)
	}

	var resp rpcResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("decode %q: %v", line, err)
	}
	raw, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(raw), `"slow_01"`) {
		t.Fatalf("result=%s", raw)
	}

	_ = clientW.Close()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not exit after stdin close")
	}
}

func TestServe_hangNeverAnswersInitialize(t *testing.T) {
	cfg := Config{Name: "hang_01", Hang: true}
	clientR, serverW := io.Pipe()
	serverR, clientW := io.Pipe()
	defer clientR.Close()
	defer clientW.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeContext(ctx, serverR, serverW, io.Discard, cfg)
	}()
	if _, err := io.WriteString(clientW, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`+"\n"); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		_, _ = bufio.NewReader(clientR).ReadBytes('\n')
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("hang answered initialize")
	case <-time.After(60 * time.Millisecond):
	}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("want ctx error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve hang did not return after ctx cancel")
	}
}

func TestServe_toolsListEmpty(t *testing.T) {
	cfg := Config{Name: "ready"}
	clientR, serverW := io.Pipe()
	serverR, clientW := io.Pipe()
	defer clientR.Close()
	defer clientW.Close()

	go func() { _ = Serve(serverR, serverW, io.Discard, cfg) }()
	r := bufio.NewReader(clientR)
	if _, err := io.WriteString(clientW, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`+"\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ReadBytes('\n'); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(clientW, `{"jsonrpc":"2.0","method":"notifications/initialized"}`+"\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(clientW, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`+"\n"); err != nil {
		t.Fatal(err)
	}
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(line), `"tools":[]`) && !strings.Contains(string(line), `"tools": []`) {
		t.Fatalf("tools/list = %s", line)
	}
	_ = clientW.Close()
}
