package trace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Store struct {
	dataDir string
}

func NewStore(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

func (s *Store) HandleRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/knowledge/agent-traces")
	path = strings.Trim(path, "/")
	if path == "" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleList(w)
		return
	}
	if strings.HasSuffix(path, "/stop") {
		id := strings.Trim(strings.TrimSuffix(path, "/stop"), "/")
		if id == "" {
			writeJSONError(w, http.StatusNotFound, "agent trace not found")
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleStop(w, id)
		return
	}
	if strings.HasSuffix(path, "/stream") {
		id := strings.Trim(strings.TrimSuffix(path, "/stream"), "/")
		if id == "" {
			writeJSONError(w, http.StatusNotFound, "agent trace not found")
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleStream(w, r, id)
		return
	}
	if strings.Contains(path, "/") {
		writeJSONError(w, http.StatusNotFound, "agent trace not found")
		return
	}
	if r.Method == http.MethodDelete {
		s.handleDelete(w, path)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.handleGet(w, path)
}

func (s *Store) handleList(w http.ResponseWriter) {
	sessions, err := loadAgentTraceSummaries(s.dataDir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

func (s *Store) handleGet(w http.ResponseWriter, id string) {
	detail, err := loadAgentTraceDetail(s.dataDir, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (s *Store) handleStop(w http.ResponseWriter, id string) {
	detail, err := markAgentTraceStopped(s.dataDir, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (s *Store) handleDelete(w http.ResponseWriter, id string) {
	if err := deleteAgentTraceSession(s.dataDir, id); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
