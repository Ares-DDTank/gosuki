package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func newOverrideTestDB(t *testing.T, name string) *DB {
	t.Helper()
	database, err := NewDB(name, "", DBTypeInMemoryDSN).Init()
	require.NoError(t, err)
	require.NoError(t, database.InitSchema(context.Background()))
	t.Cleanup(func() { database.Close() })
	return database
}

func seedOverrideBookmark(t *testing.T, database *DB) {
	t.Helper()
	_, err := database.Handle.Exec(`
		INSERT INTO gskbookmarks (URL, metadata, tags, desc, module)
		VALUES ('https://example.test', 'Browser title', ',browser,old,',
			'Browser description', 'firefox_default')
	`)
	require.NoError(t, err)
}

func TestApplyBookmarkOverrides(t *testing.T) {
	cache := newOverrideTestDB(t, "overrides_cache")
	l2 := newOverrideTestDB(t, "overrides_l2")
	seedOverrideBookmark(t, cache)
	seedOverrideBookmark(t, l2)

	title := "Local title"
	description := "Local description"
	updated, err := applyBookmarkOverrides(context.Background(), cache, l2, BookmarkOverridePatch{
		URLs:        []string{"https://example.test"},
		Title:       &title,
		Description: &description,
		AddTags:     []string{"local", "Browser"},
		RemoveTags:  []string{"old"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, updated)

	for _, database := range []*DB{cache, l2} {
		var effective struct {
			Title string `db:"metadata"`
			Tags  string `db:"tags"`
			Desc  string `db:"desc"`
		}
		require.NoError(t, database.Handle.Get(&effective, `
			SELECT metadata, tags, desc FROM effective_bookmarks
			WHERE URL = 'https://example.test'
		`))
		require.Equal(t, "Local title", effective.Title)
		require.Equal(t, ",browser,local,", effective.Tags)
		require.Equal(t, "Local description", effective.Desc)

		var sourceTitle string
		require.NoError(t, database.Handle.Get(&sourceTitle,
			"SELECT metadata FROM gskbookmarks WHERE URL = 'https://example.test'"))
		require.Equal(t, "Browser title", sourceTitle)
	}

	updated, err = applyBookmarkOverrides(context.Background(), cache, l2, BookmarkOverridePatch{
		URLs:             []string{"https://example.test"},
		ClearTitle:       true,
		ClearTags:        true,
		ClearDescription: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, updated)

	var overrideCount int
	require.NoError(t, cache.Handle.Get(&overrideCount, "SELECT COUNT(*) FROM bookmark_overrides"))
	require.Zero(t, overrideCount)
	var restoredTitle string
	require.NoError(t, cache.Handle.Get(&restoredTitle,
		"SELECT metadata FROM effective_bookmarks WHERE URL = 'https://example.test'"))
	require.Equal(t, "Browser title", restoredTitle)
}

func TestApplyBookmarkOverridesIgnoresUnknownURLs(t *testing.T) {
	cache := newOverrideTestDB(t, "overrides_unknown_cache")
	l2 := newOverrideTestDB(t, "overrides_unknown_l2")
	title := "Local title"
	updated, err := applyBookmarkOverrides(context.Background(), cache, l2, BookmarkOverridePatch{
		URLs:  []string{"https://missing.test"},
		Title: &title,
	})
	require.NoError(t, err)
	require.Zero(t, updated)
}

func TestApplyBookmarkOverridesAppendsEffectiveText(t *testing.T) {
	cache := newOverrideTestDB(t, "overrides_append_cache")
	l2 := newOverrideTestDB(t, "overrides_append_l2")
	seedOverrideBookmark(t, cache)
	seedOverrideBookmark(t, l2)

	titleSuffix := "— local"
	descriptionSuffix := "with notes"
	updated, err := applyBookmarkOverrides(context.Background(), cache, l2, BookmarkOverridePatch{
		URLs:              []string{"https://example.test"},
		AppendTitle:       &titleSuffix,
		AppendDescription: &descriptionSuffix,
	})
	require.NoError(t, err)
	require.Equal(t, 1, updated)

	for _, database := range []*DB{cache, l2} {
		var effective struct {
			Title string `db:"metadata"`
			Desc  string `db:"desc"`
		}
		require.NoError(t, database.Handle.Get(&effective, `
			SELECT metadata, desc FROM effective_bookmarks
			WHERE URL = 'https://example.test'
		`))
		require.Equal(t, "Browser title — local", effective.Title)
		require.Equal(t, "Browser description with notes", effective.Desc)
	}
}

func TestAppendText(t *testing.T) {
	require.Equal(t, "base suffix", appendText(" base ", " suffix "))
	require.Equal(t, "suffix", appendText("", "suffix"))
	require.Equal(t, "base", appendText("base", ""))
}
