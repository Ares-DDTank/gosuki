// Copyright (c) 2026 Chakib Ben Ziane and GoSuki contributors.
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	db "github.com/blob42/gosuki/internal/database"
)

type BookmarkOverrideSet struct {
	Title       *string   `json:"title,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	Description *string   `json:"description,omitempty"`
}

type BookmarkOverrideAppend struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type BookmarkOverrideRequest struct {
	URLs       []string               `json:"urls"`
	Set        BookmarkOverrideSet    `json:"set"`
	Append     BookmarkOverrideAppend `json:"append"`
	AddTags    []string               `json:"add_tags,omitempty"`
	RemoveTags []string               `json:"remove_tags,omitempty"`
	Clear      []string               `json:"clear,omitempty"`
}

type BookmarkOverrideResponse struct {
	Updated int `json:"updated"`
}

func PatchBookmarkOverrides(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		http.Error(w, "bookmark overrides are restricted to localhost", http.StatusForbidden)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	request := BookmarkOverrideRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON request: %s", err), http.StatusBadRequest)
		return
	}
	if len(request.URLs) == 0 || len(request.URLs) > 500 {
		http.Error(w, "urls must contain between 1 and 500 bookmarks", http.StatusBadRequest)
		return
	}

	patch := db.BookmarkOverridePatch{
		URLs:              request.URLs,
		Title:             request.Set.Title,
		Tags:              request.Set.Tags,
		Description:       request.Set.Description,
		AppendTitle:       request.Append.Title,
		AppendDescription: request.Append.Description,
		AddTags:           request.AddTags,
		RemoveTags:        request.RemoveTags,
	}
	for _, field := range request.Clear {
		switch strings.ToLower(strings.TrimSpace(field)) {
		case "title":
			patch.ClearTitle = true
		case "tags":
			patch.ClearTags = true
		case "description":
			patch.ClearDescription = true
		default:
			http.Error(w, fmt.Sprintf("unknown clear field %q", field), http.StatusBadRequest)
			return
		}
	}
	if patch.ClearTitle && (patch.Title != nil || patch.AppendTitle != nil) ||
		patch.ClearTags && (patch.Tags != nil || len(patch.AddTags) > 0 || len(patch.RemoveTags) > 0) ||
		patch.ClearDescription && (patch.Description != nil || patch.AppendDescription != nil) {
		http.Error(w, "a field cannot be set and cleared in the same request", http.StatusBadRequest)
		return
	}
	if patch.Title != nil && patch.AppendTitle != nil ||
		patch.Description != nil && patch.AppendDescription != nil {
		http.Error(w, "a field cannot be set and appended in the same request", http.StatusBadRequest)
		return
	}
	if patch.Title == nil && patch.Tags == nil && patch.Description == nil &&
		patch.AppendTitle == nil && patch.AppendDescription == nil &&
		len(patch.AddTags) == 0 && len(patch.RemoveTags) == 0 && len(request.Clear) == 0 {
		http.Error(w, "request does not contain an override operation", http.StatusBadRequest)
		return
	}

	updated, err := db.ApplyBookmarkOverrides(r.Context(), patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(BookmarkOverrideResponse{Updated: updated}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
