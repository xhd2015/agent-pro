package run

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

const protocolVersionDefault = "2024-11-05"
const serverVersion = "0.0.1"

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// Serve reads MCP JSON-RPC from in and writes replies to out.
// Delay is applied after the initialize request, before the initialize result.
func Serve(in io.Reader, out io.Writer, errOut io.Writer, cfg Config) error {
	return ServeContext(context.Background(), in, out, errOut, cfg)
}

func ServeContext(ctx context.Context, in io.Reader, out io.Writer, errOut io.Writer, cfg Config) error {
	if errOut == nil {
		errOut = io.Discard
	}
	r := bufio.NewReader(in)
	var framed bool
	initialized := false
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		body, msgFramed, err := readMessage(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !initialized {
			framed = msgFramed
		}
		var req rpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return fmt.Errorf("decode json-rpc: %w", err)
		}
		switch req.Method {
		case "initialize":
			delay := cfg.chosenDelay()
			debugDelay(errOut, cfg, delay)
			if cfg.Hang {
				<-ctx.Done()
				return ctx.Err()
			}
			if err := waitDelay(ctx, delay); err != nil {
				return err
			}
			if cfg.Fail {
				resp := rpcResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error:   &rpcError{Code: -32000, Message: "mock-mcp fail"},
				}
				if err := writeRPC(out, framed, resp); err != nil {
					return err
				}
				return nil
			}
			pv := protocolVersionDefault
			var params struct {
				ProtocolVersion string `json:"protocolVersion"`
			}
			if len(req.Params) > 0 {
				_ = json.Unmarshal(req.Params, &params)
				if params.ProtocolVersion != "" {
					pv = params.ProtocolVersion
				}
			}
			resp := rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"protocolVersion": pv,
					"capabilities": map[string]any{
						"tools": map[string]any{},
					},
					"serverInfo": map[string]any{
						"name":    cfg.Name,
						"version": serverVersion,
					},
				},
			}
			if err := writeRPC(out, framed, resp); err != nil {
				return err
			}
			initialized = true
		case "notifications/initialized":
			// notification: no response
		case "tools/list":
			resp := rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]any{"tools": []any{}},
			}
			if err := writeRPC(out, framed, resp); err != nil {
				return err
			}
		case "ping":
			if len(req.ID) == 0 {
				continue
			}
			resp := rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
			if err := writeRPC(out, framed, resp); err != nil {
				return err
			}
		default:
			if len(req.ID) == 0 {
				continue
			}
			resp := rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
			}
			if err := writeRPC(out, framed, resp); err != nil {
				return err
			}
		}
	}
}

func waitDelay(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func debugDelay(errOut io.Writer, cfg Config, delay time.Duration) {
	if !envBool(EnvDebug) && os.Getenv(EnvDebug) != "1" {
		return
	}
	switch {
	case cfg.Hang:
		fmt.Fprintf(errOut, "mock-mcp: name=%s hang\n", cfg.Name)
	case cfg.DelayMin != nil && cfg.DelayMax != nil:
		fmt.Fprintf(errOut, "mock-mcp: name=%s delay=%s (%s-%s)\n", cfg.Name, delay, *cfg.DelayMin, *cfg.DelayMax)
	default:
		fmt.Fprintf(errOut, "mock-mcp: name=%s delay=%s\n", cfg.Name, delay)
	}
}

func pickDelay(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	n := int64(max - min)
	return min + time.Duration(rand.Int64N(n+1))
}

func readMessage(r *bufio.Reader) (body []byte, framed bool, err error) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				line = append(line, '\n')
			} else {
				return nil, false, err
			}
		}
		trim := bytes.TrimRight(line, "\r\n")
		if len(bytes.TrimSpace(trim)) == 0 {
			continue
		}
		lower := bytes.ToLower(trim)
		if bytes.HasPrefix(lower, []byte("content-length:")) {
			n, err := strconv.Atoi(strings.TrimSpace(string(trim[len("Content-Length:"):])))
			if err != nil {
				return nil, false, fmt.Errorf("content-length: %w", err)
			}
			for {
				h, err := r.ReadBytes('\n')
				if err != nil {
					return nil, false, err
				}
				if len(bytes.TrimSpace(h)) == 0 {
					break
				}
			}
			body := make([]byte, n)
			if _, err := io.ReadFull(r, body); err != nil {
				return nil, false, err
			}
			return body, true, nil
		}
		return trim, false, nil
	}
}

func writeRPC(w io.Writer, framed bool, resp rpcResponse) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if framed {
		_, err = fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(payload), payload)
		return err
	}
	_, err = w.Write(append(payload, '\n'))
	return err
}
