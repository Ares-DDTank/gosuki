// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import "testing"

func TestPreferredBookmarkModule(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		incoming string
		want     string
	}{
		{
			name:     "edge replaces firefox",
			existing: "firefox_default-release-1",
			incoming: "chrome_edge_用户配置 1",
			want:     "chrome_edge_用户配置 1",
		},
		{
			name:     "edge replaces chrome",
			existing: "chrome_用户1",
			incoming: "chrome_edge_用户配置 1",
			want:     "chrome_edge_用户配置 1",
		},
		{
			name:     "firefox cannot replace edge",
			existing: "chrome_edge_用户配置 1",
			incoming: "firefox_default-release-1",
			want:     "chrome_edge_用户配置 1",
		},
		{
			name:     "first non-edge source remains stable",
			existing: "firefox_default-release-1",
			incoming: "chrome_用户1",
			want:     "firefox_default-release-1",
		},
		{
			name:     "empty source accepts incoming source",
			existing: "",
			incoming: "firefox_default-release-1",
			want:     "firefox_default-release-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preferredBookmarkModule(tt.existing, tt.incoming); got != tt.want {
				t.Fatalf("preferredBookmarkModule(%q, %q) = %q, want %q", tt.existing, tt.incoming, got, tt.want)
			}
		})
	}
}
