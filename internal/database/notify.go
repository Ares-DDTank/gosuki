// Copyright (c) 2026 Chakib Ben Ziane and GoSuki contributors.
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"fmt"
	"os"
	"time"

	"github.com/blob42/gosuki/internal/utils"
)

// notifyDatabaseChange updates configured marker files after the debounced
// in-memory database has been persisted successfully. Consumers can watch a
// marker without polling or holding the SQLite database open.
func notifyDatabaseChange() {
	// A leading comment also makes the marker safe when a consumer requires a
	// script-like extension (for example, a Keypirinha Live Package).
	content := []byte(fmt.Sprintf("# %d\n", time.Now().UnixNano()))
	for _, configuredPath := range Config.ChangeNotifyFiles {
		markerPath, err := utils.ExpandOnly(configuredPath)
		if err != nil {
			log.Warnf("invalid database change notification path %q: %s", configuredPath, err)
			continue
		}
		if err := os.WriteFile(markerPath, content, 0o644); err != nil {
			log.Warnf("could not update database change notification file %q: %s", markerPath, err)
		}
	}
}
