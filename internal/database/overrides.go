// Copyright (c) 2026 Chakib Ben Ziane and GoSuki contributors.
// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

// BookmarkOverridePatch describes local-only metadata changes. Nil Set fields
// are left unchanged; Clear fields remove the local value so the browser value
// becomes effective again.
type BookmarkOverridePatch struct {
	URLs              []string
	Title             *string
	Tags              *[]string
	Description       *string
	AppendTitle       *string
	AppendDescription *string
	AddTags           []string
	RemoveTags        []string
	ClearTitle        bool
	ClearTags         bool
	ClearDescription  bool
}

type bookmarkOverrideState struct {
	URL                  string         `db:"url"`
	Metadata             sql.NullString `db:"override_metadata"`
	Tags                 sql.NullString `db:"override_tags"`
	Description          sql.NullString `db:"override_desc"`
	EffectiveTags        string         `db:"effective_tags"`
	EffectiveTitle       string         `db:"effective_title"`
	EffectiveDescription string         `db:"effective_description"`
}

type bookmarkOverrideRecord struct {
	URL         string
	Metadata    sql.NullString
	Tags        sql.NullString
	Description sql.NullString
	Modified    int64
}

// ApplyBookmarkOverrides updates both memory layers and schedules their normal
// debounced backup to disk. Browser scans only update gskbookmarks, so these
// values remain authoritative until explicitly cleared.
func ApplyBookmarkOverrides(ctx context.Context, patch BookmarkOverridePatch) (int, error) {
	if Cache.DB == nil || Cache.Handle == nil || L2Cache.DB == nil || L2Cache.Handle == nil {
		return 0, fmt.Errorf("bookmark caches are not initialized")
	}

	cacheMu.Lock()
	updated, err := applyBookmarkOverrides(ctx, Cache.DB, L2Cache.DB, patch)
	cacheMu.Unlock()
	if err != nil {
		return 0, err
	}
	if updated > 0 {
		ScheduleBackupToDisk()
	}
	return updated, nil
}

func applyBookmarkOverrides(
	ctx context.Context,
	cache *DB,
	l2 *DB,
	patch BookmarkOverridePatch,
) (int, error) {
	urls := uniqueNonEmpty(patch.URLs)
	if len(urls) == 0 {
		return 0, fmt.Errorf("no bookmark URLs supplied")
	}

	records := make([]bookmarkOverrideRecord, 0, len(urls))
	modified := time.Now().Unix()
	for _, url := range urls {
		state := bookmarkOverrideState{}
		err := cache.Handle.GetContext(ctx, &state, `
			SELECT
				b.URL AS url,
				o.metadata AS override_metadata,
				o.tags AS override_tags,
				o.desc AS override_desc,
				e.tags AS effective_tags,
				e.metadata AS effective_title,
				e.desc AS effective_description
			FROM gskbookmarks AS b
			JOIN effective_bookmarks AS e ON e.URL = b.URL
			LEFT JOIN bookmark_overrides AS o ON o.url = b.URL
			WHERE b.URL = ?
		`, url)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return 0, fmt.Errorf("read override state for %q: %w", url, err)
		}

		record := bookmarkOverrideRecord{
			URL:         url,
			Metadata:    state.Metadata,
			Tags:        state.Tags,
			Description: state.Description,
			Modified:    modified,
		}
		if patch.ClearTitle {
			record.Metadata = sql.NullString{}
		} else if patch.Title != nil {
			record.Metadata = sql.NullString{String: *patch.Title, Valid: true}
		} else if patch.AppendTitle != nil {
			record.Metadata = sql.NullString{
				String: appendText(state.EffectiveTitle, *patch.AppendTitle), Valid: true,
			}
		}
		if patch.ClearDescription {
			record.Description = sql.NullString{}
		} else if patch.Description != nil {
			record.Description = sql.NullString{String: *patch.Description, Valid: true}
		} else if patch.AppendDescription != nil {
			record.Description = sql.NullString{
				String: appendText(state.EffectiveDescription, *patch.AppendDescription), Valid: true,
			}
		}

		hasTagOperation := patch.Tags != nil || len(patch.AddTags) > 0 || len(patch.RemoveTags) > 0
		if patch.ClearTags {
			record.Tags = sql.NullString{}
		} else if hasTagOperation {
			base := tagsFromString(state.EffectiveTags, TagSep).Get()
			if patch.Tags != nil {
				base = *patch.Tags
			}
			tags := modifyTags(base, patch.AddTags, patch.RemoveTags)
			record.Tags = sql.NullString{
				String: NewTags(tags, TagSep).PreSanitize().Sort().StringWrap(),
				Valid:  true,
			}
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return 0, nil
	}
	if err := writeBookmarkOverrides(ctx, l2, records); err != nil {
		return 0, fmt.Errorf("update L2 overrides: %w", err)
	}
	if err := writeBookmarkOverrides(ctx, cache, records); err != nil {
		return 0, fmt.Errorf("update L1 overrides: %w", err)
	}
	return len(records), nil
}

func appendText(base, suffix string) string {
	base = strings.TrimSpace(base)
	suffix = strings.TrimSpace(suffix)
	if base == "" {
		return suffix
	}
	if suffix == "" {
		return base
	}
	return base + " " + suffix
}

func writeBookmarkOverrides(ctx context.Context, db *DB, records []bookmarkOverrideRecord) error {
	tx, err := db.Handle.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		if !record.Metadata.Valid && !record.Tags.Valid && !record.Description.Valid {
			if _, err = tx.ExecContext(ctx,
				"DELETE FROM bookmark_overrides WHERE url = ?", record.URL); err != nil {
				return err
			}
			continue
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO bookmark_overrides (url, metadata, tags, desc, modified)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(url) DO UPDATE SET
				metadata = excluded.metadata,
				tags = excluded.tags,
				desc = excluded.desc,
				modified = excluded.modified
		`, record.URL, record.Metadata, record.Tags, record.Description, record.Modified)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func modifyTags(base, add, remove []string) []string {
	removed := make(map[string]bool)
	for _, tag := range remove {
		if tag = strings.TrimSpace(tag); tag != "" {
			removed[strings.ToLower(tag)] = true
		}
	}

	byFoldedName := make(map[string]string)
	for _, tag := range append(base, add...) {
		tag = strings.TrimSpace(tag)
		folded := strings.ToLower(tag)
		if tag == "" || removed[folded] {
			continue
		}
		if _, exists := byFoldedName[folded]; !exists {
			byFoldedName[folded] = tag
		}
	}

	result := make([]string, 0, len(byFoldedName))
	for _, tag := range byFoldedName {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
