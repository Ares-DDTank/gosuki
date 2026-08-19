package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatchBookmarkOverridesRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty urls", `{"urls":[],"add_tags":["x"]}`},
		{"no operation", `{"urls":["https://example.test"]}`},
		{"unknown clear field", `{"urls":["https://example.test"],"clear":["url"]}`},
		{"set and clear same field", `{"urls":["https://example.test"],"set":{"title":"x"},"clear":["title"]}`},
		{"append and clear same field", `{"urls":["https://example.test"],"append":{"description":"x"},"clear":["description"]}`},
		{"set and append same field", `{"urls":["https://example.test"],"set":{"title":"x"},"append":{"title":"y"}}`},
		{"unknown json field", `{"urls":["https://example.test"],"wat":true}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPatch, "/api/bookmarks/overrides",
				strings.NewReader(test.body))
			request.RemoteAddr = "127.0.0.1:12345"
			response := httptest.NewRecorder()
			PatchBookmarkOverrides(response, request)
			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestPatchBookmarkOverridesRejectsRemoteClients(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/api/bookmarks/overrides",
		strings.NewReader(`{"urls":["https://example.test"],"add_tags":["x"]}`))
	request.RemoteAddr = "192.0.2.1:12345"
	response := httptest.NewRecorder()
	PatchBookmarkOverrides(response, request)
	require.Equal(t, http.StatusForbidden, response.Code)
}
