package acp

import (
	"context"
	"io"
	"log/slog"

	acpsdk "github.com/coder/acp-go-sdk"
)

type agentBridge struct {
	agent Agent
	conn  *acpsdk.AgentSideConnection
}

func ServeStdio(ctx context.Context, agent Agent, logger *slog.Logger) error {
	return serve(ctx, agent, discardWriter{}, &nopReader{}, logger)
}

func serve(ctx context.Context, agent Agent, w io.Writer, r io.Reader, logger *slog.Logger) error {
	bridge := &agentBridge{agent: agent}
	conn := acpsdk.NewAgentSideConnection(bridge, w, r)
	bridge.conn = conn
	if logger != nil {
		conn.SetLogger(logger)
	}
	<-conn.Done()
	return nil
}

type discardWriter struct{}
func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

type nopReader struct{}
func (nopReader) Read(p []byte) (int, error) { return 0, io.EOF }
