// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import "strings"

const edgeModule = "chrome_edge"

// preferredBookmarkModule preserves the first source by default, but gives an
// Edge source ownership of URLs that are present in more than one browser.
// GoSuki's Chrome-family module names are shaped as
// "chrome_<flavour>_<profile>", hence Edge is "chrome_edge_<profile>".
func preferredBookmarkModule(existing, incoming string) string {
	if existing == "" {
		return incoming
	}
	if isEdgeModule(incoming) && !isEdgeModule(existing) {
		return incoming
	}
	return existing
}

func isEdgeModule(module string) bool {
	return module == edgeModule || strings.HasPrefix(module, edgeModule+"_")
}
