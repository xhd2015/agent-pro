package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	frontend "github.com/xhd2015/agent-pro/frontend-agent-run"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func registerStatic(mux *http.ServeMux, store agentstorage.Store) error {
	dist, err := fs.Sub(frontend.DistFS, "dist")
	if err != nil {
		return err
	}
	indexHTML, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return err
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := dist.Open(path)
		if err != nil {
			path = "index.html"
			f, err = dist.Open(path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()
		if path != "index.html" {
			w.Header().Set("Content-Type", mimeTypeByExtension(filepath.Ext(path)))
			_, _ = io.Copy(w, f)
			return
		}

		body := indexHTML
		if bootstrap := sessionBootstrapJSON(store, r.URL.Path); bootstrap != "" {
			injection := `<script id="agent-run-session-bootstrap" type="application/json">` + bootstrap + `</script>`
			body = bytes.Replace(indexHTML, []byte("</body>"), []byte(injection+"</body>"), 1)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(body)
	})
	return nil
}

func sessionBootstrapJSON(store agentstorage.Store, urlPath string) string {
	runner, sessionID, ok := parseSessionPagePath(urlPath)
	if !ok {
		return ""
	}
	meta, err := store.GetSession(runner, sessionID)
	if err != nil {
		return ""
	}
	events, eventsOffset, err := store.ReadEvents(runner, sessionID, 0)
	if err != nil {
		return ""
	}
	if events == nil {
		events = []types.AgentEvent{}
	}
	payload := map[string]any{
		"session":       meta.Meta,
		"events":        events,
		"events_offset": eventsOffset,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func parseSessionPagePath(urlPath string) (runner, sessionID string, ok bool) {
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	if len(parts) != 3 || parts[0] != "sessions" {
		return "", "", false
	}
	runner = strings.TrimSpace(parts[1])
	sessionID = strings.TrimSpace(parts[2])
	if runner == "" || sessionID == "" {
		return "", "", false
	}
	return runner, sessionID, true
}

func mimeTypeByExtension(ext string) string {
	switch ext {
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".html":
		return "text/html; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}