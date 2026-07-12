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
