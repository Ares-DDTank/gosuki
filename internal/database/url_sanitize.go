// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"html"
	"regexp"
)

var terminatedHTMLReference = regexp.MustCompile(`&(?:#[0-9]+|#[xX][0-9A-Fa-f]+|[A-Za-z][A-Za-z0-9]+);`)

// unescapeTerminatedURLReferences decodes references that are explicitly
// terminated by a semicolon. html.UnescapeString also accepts legacy named
// references without one, which can corrupt ordinary URL query parameters such
// as &timestamp= by interpreting &times as the multiplication sign.
func unescapeTerminatedURLReferences(url string) string {
	return terminatedHTMLReference.ReplaceAllStringFunc(url, html.UnescapeString)
}
