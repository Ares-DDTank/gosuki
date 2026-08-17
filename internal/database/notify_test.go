package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotifyDatabaseChange(t *testing.T) {
	markerPath := filepath.Join(t.TempDir(), "catalog.changed")
	previous := Config.ChangeNotifyFiles
	Config.ChangeNotifyFiles = []string{markerPath}
	t.Cleanup(func() { Config.ChangeNotifyFiles = previous })

	notifyDatabaseChange()

	content, err := os.ReadFile(markerPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
}

func TestNotifyDatabaseChangeIgnoresInvalidTargets(t *testing.T) {
	previous := Config.ChangeNotifyFiles
	Config.ChangeNotifyFiles = []string{""}
	t.Cleanup(func() { Config.ChangeNotifyFiles = previous })

	require.NotPanics(t, notifyDatabaseChange)
}
