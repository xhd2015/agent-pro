package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type httpLogger struct {
	w     io.Writer
	mu    sync.Mutex
	index int
}

func newHTTPLogger(path string) (*httpLogger, error) {
	if !strings.HasSuffix(path, ".jsonl") {
		return nil, fmt.Errorf("--log-http path must end with .jsonl")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log-http file: %w", err)
	}
	return &httpLogger{w: f}, nil
}

func (l *httpLogger) Close() error {
	if closer, ok := l.w.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (l *httpLogger) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqBodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))

		reqBody := parseLogBody(reqBodyBytes)

		cap := newCaptureResponseWriter(w)
		next(cap, r)

		l.logExchange(start, time.Since(start), r, reqBody, cap)
	}
}

func parseLogBody(body []byte) any {
	if len(body) == 0 {
		return map[string]any{}
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	return parsed
}

func (l *httpLogger) logExchange(start time.Time, duration time.Duration, r *http.Request, reqBody any, cap *captureResponseWriter) {
	rec := map[string]any{
		"timestamp":   start.UTC().Format(time.RFC3339Nano),
		"duration_ms": duration.Milliseconds(),
		"request": map[string]any{
			"method":  r.Method,
			"path":    r.URL.Path,
			"headers": headerMap(r.Header),
			"body":    reqBody,
		},
		"response": buildResponseLog(cap),
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	rec["index"] = l.index
	l.index++

	data, err := json.Marshal(rec)
	if err != nil {
		return
	}
	fmt.Fprintln(l.w, string(data))
}

func buildResponseLog(cap *captureResponseWriter) map[string]any {
	resp := map[string]any{
		"status":  cap.status,
		"headers": headerMap(cap.header),
		"stream":  cap.isStream,
	}
	if cap.isStream {
		resp["chunks"] = cap.chunks
	} else if cap.body.Len() > 0 {
		resp["body"] = parseLogBody(cap.body.Bytes())
	} else {
		resp["body"] = map[string]any{}
	}
	return resp
}

func headerMap(h http.Header) map[string]any {
	out := make(map[string]any, len(h))
	for k, vv := range h {
		if len(vv) == 1 {
			out[k] = vv[0]
		} else {
			copied := make([]string, len(vv))
			copy(copied, vv)
			out[k] = copied
		}
	}
	return out
}

type captureResponseWriter struct {
	http.ResponseWriter
	status      int
	header      http.Header
	body        bytes.Buffer
	chunks      []string
	isStream    bool
	wroteHeader bool
}

func newCaptureResponseWriter(w http.ResponseWriter) *captureResponseWriter {
	return &captureResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
		header:         make(http.Header),
	}
}

func (c *captureResponseWriter) Header() http.Header {
	return c.header
}

func (c *captureResponseWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true
	c.status = statusCode
	dst := c.ResponseWriter.Header()
	for k, v := range c.header {
		dst[k] = v
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *captureResponseWriter) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	if strings.Contains(c.header.Get("Content-Type"), "text/event-stream") {
		c.isStream = true
		c.chunks = append(c.chunks, string(b))
	} else {
		c.body.Write(b)
	}
	return c.ResponseWriter.Write(b)
}

func (c *captureResponseWriter) Flush() {
	if flusher, ok := c.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func registerHandler(mux *http.ServeMux, path string, fn http.HandlerFunc, logger *httpLogger) {
	if logger != nil {
		fn = logger.wrap(fn)
	}
	mux.HandleFunc(path, fn)
}