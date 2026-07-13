package main

import (
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func listAllSessions(store agentstorage.Store) ([]agentstorage.SessionMeta, error) {
	list, err := store.ListSessions()
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []agentstorage.SessionMeta{}, nil
	}
	return list, nil
}

// sortSessionsNewestFirst sorts by updated_at desc, then created_at desc, then session_id asc.
func sortSessionsNewestFirst(list []agentstorage.SessionMeta) {
	sort.SliceStable(list, func(i, j int) bool {
		ui := parseSessionTime(list[i].UpdatedAt)
		uj := parseSessionTime(list[j].UpdatedAt)
		if !ui.Equal(uj) {
			return ui.After(uj)
		}
		ci := parseSessionTime(list[i].CreatedAt)
		cj := parseSessionTime(list[j].CreatedAt)
		if !ci.Equal(cj) {
			return ci.After(cj)
		}
		return list[i].SessionID < list[j].SessionID
	})
}

func parseSessionTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

// applySessionLimit applies default limit 10; limit 0 means all; negative treated as 0 (all).
func applySessionLimit(list []agentstorage.SessionMeta, limit int) []agentstorage.SessionMeta {
	if limit == 0 {
		return list
	}
	if limit < 0 {
		return list
	}
	if len(list) <= limit {
		return list
	}
	return list[:limit]
}

// sessionStatusBucket maps a session status to list filter buckets.
// done = finished + idle; running stays running; other statuses count only toward all.
func sessionStatusBucket(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running":
		return "running"
	case "finished", "idle":
		return "done"
	default:
		return ""
	}
}

// filterSessionsByStatusFilter filters by status query: all | running | done.
// Unknown/empty status is treated as all.
func filterSessionsByStatusFilter(list []agentstorage.SessionMeta, status string) []agentstorage.SessionMeta {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" || status == "all" {
		return list
	}
	out := make([]agentstorage.SessionMeta, 0, len(list))
	for _, s := range list {
		bucket := sessionStatusBucket(s.Status)
		if status == "running" && bucket == "running" {
			out = append(out, s)
		} else if status == "done" && bucket == "done" {
			out = append(out, s)
		}
	}
	return out
}

// filterSessionsByQuery keeps sessions whose prompt, id, workspace, or runner
// contain q as a case-insensitive substring. Empty q returns list unchanged.
func filterSessionsByQuery(list []agentstorage.SessionMeta, q string) []agentstorage.SessionMeta {
	q = strings.TrimSpace(q)
	if q == "" {
		return list
	}
	needle := strings.ToLower(q)
	out := make([]agentstorage.SessionMeta, 0, len(list))
	for _, s := range list {
		if sessionMatchesQuery(s, needle) {
			out = append(out, s)
		}
	}
	return out
}

func sessionMatchesQuery(s agentstorage.SessionMeta, needleLower string) bool {
	fields := []string{
		s.InitialPrompt,
		s.SessionID,
		s.Workspace,
		s.Runner,
	}
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), needleLower) {
			return true
		}
	}
	return false
}

// sessionListCounts are chip totals over the full store (ignore q).
type sessionListCounts struct {
	All     int `json:"all"`
	Running int `json:"running"`
	Done    int `json:"done"`
}

func computeSessionListCounts(list []agentstorage.SessionMeta) sessionListCounts {
	c := sessionListCounts{All: len(list)}
	for _, s := range list {
		switch sessionStatusBucket(s.Status) {
		case "running":
			c.Running++
		case "done":
			c.Done++
		}
	}
	return c
}

// pageSessions applies offset/limit. limit <= 0 means return the whole list (has_more=false).
func pageSessions(list []agentstorage.SessionMeta, limit, offset int) (page []agentstorage.SessionMeta, hasMore bool, outLimit, outOffset int) {
	if offset < 0 {
		offset = 0
	}
	if limit < 0 {
		limit = 0
	}
	total := len(list)
	if offset > total {
		offset = total
	}
	if limit == 0 {
		// No pagination: return all remaining from offset (compat when limit omitted).
		page = list[offset:]
		return page, false, 0, offset
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page = list[offset:end]
	hasMore = end < total
	return page, hasMore, limit, offset
}
