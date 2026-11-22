package web

import (
	"context"
	"testing"
)

func TestCompactIntoNewSendsTheSourceAndReturnsTheChild(t *testing.T) {
	host, client := newRPCHost(t, `{"sessionId":"session-child","compacted":true,"compactedTokens":18240}`)
	child, err := client.CompactIntoNew(context.Background(), "session-source")
	if err != nil {
		t.Fatal(err)
	}
	if child != "session-child" {
		t.Fatalf("child = %q", child)
	}
	if len(host.methods) != 1 || host.methods[0] != "session/compactIntoNew" {
		t.Fatalf("methods = %v", host.methods)
	}
	if host.requests[0]["sessionId"] != "session-source" {
		t.Fatalf("request = %v", host.requests[0])
	}
}

func TestCompactIntoNewRequiresASource(t *testing.T) {
	_, client := newRPCHost(t, `{"sessionId":"session-child"}`)
	if _, err := client.CompactIntoNew(context.Background(), "  "); err == nil {
		t.Fatal("an empty source session must fail")
	}
}

func TestCompactIntoNewRejectsAnEmptyChild(t *testing.T) {
	_, client := newRPCHost(t, `{}`)
	if _, err := client.CompactIntoNew(context.Background(), "session-source"); err == nil {
		t.Fatal("a response without a session ID must fail")
	}
}
