package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnescapeTerminatedURLReferences(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "encoded ampersand",
			url:  "https://example.com/?one=1&amp;two=2",
			want: "https://example.com/?one=1&two=2",
		},
		{
			name: "numeric ampersand",
			url:  "https://example.com/?one=1&#38;two=2",
			want: "https://example.com/?one=1&two=2",
		},
		{
			name: "query parameter resembling legacy entity",
			url:  "https://example.com/?share_tag=s_i&timestamp=1603340978",
			want: "https://example.com/?share_tag=s_i&timestamp=1603340978",
		},
		{
			name: "underscore query parameter",
			url:  "https://example.com/?t=1&_sec_version_=1",
			want: "https://example.com/?t=1&_sec_version_=1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, unescapeTerminatedURLReferences(test.url))
		})
	}
}
