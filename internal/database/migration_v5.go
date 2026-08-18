// Copyright (c) 2026 Chakib Ben Ziane and GoSuki contributors.
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

// migrateToVersion5 adds durable local metadata overrides. Browser scanners
// continue to own gskbookmarks, while consumers read effective_bookmarks to see
// local title, tag and description replacements when present.
func (db *DB) migrateToVersion5() error {
	log.Debug("DB schema: migrating to v5")
	tx, err := db.Handle.Begin()
	if err != nil {
		return DBError{DBName: db.Name, Err: err}
	}

	if _, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS bookmark_overrides (
			url TEXT PRIMARY KEY,
			metadata TEXT,
			tags TEXT,
			desc TEXT,
			modified INTEGER NOT NULL DEFAULT (strftime('%s'))
		)
	`); err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	if _, err = tx.Exec(QCreateEffectiveView); err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	if err = tx.Commit(); err != nil {
		return DBError{DBName: db.Name, Err: err}
	}
	return nil
}
