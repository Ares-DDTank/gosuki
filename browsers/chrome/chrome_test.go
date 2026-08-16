package chrome

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blob42/gosuki/internal/database"
	"github.com/blob42/gosuki/internal/index"
	"github.com/blob42/gosuki/internal/utils"
	"github.com/blob42/gosuki/pkg/browsers"
	"github.com/blob42/gosuki/pkg/logging"
	"github.com/blob42/gosuki/pkg/modules"
	"github.com/blob42/gosuki/pkg/parsing"
	"github.com/blob42/gosuki/pkg/profiles"
	"github.com/blob42/gosuki/pkg/tree"
)

const statePath = "testdata/Local State"

var ch Chrome

func setupChrome() {
	bufDB, err := database.NewBuffer("chrome_test")
	if err != nil {
		panic(err)
	}
	ch = Chrome{
		ChromeConfig: &ChromeConfig{
			BrowserConfig: &modules.BrowserConfig{
				Name:     "chrome",
				BaseDir:  "",
				BkDir:    "testdata",
				BkFile:   bookmarksFileName,
				BufferDB: bufDB,
				URLIndex: index.NewIndex(),
				NodeTree: &tree.Node{
					Title:  RootNodeName,
					Parent: nil,
					Type:   tree.RootNode,
				},
				UseFileWatcher: true,
				UseHooks:       []string{},
			},
		},
		Counter: &parsing.BrowserCounter{},
	}
}

func setupCache() error {
	if database.Cache != nil && database.Cache.DB != nil {
		if err := database.Cache.DB.Close(); err != nil {
			return err
		}
	}
	cacheDB, err := database.NewDB(database.CacheName, "", database.DBTypeCacheDSN).Init()
	if err != nil {
		return err
	}
	database.Cache = &database.CacheDB{DB: cacheDB}
	return nil
}

func TestMain(m *testing.M) {
	database.RegisterSqliteHooks()

	if err := setupCache(); err != nil {
		log.Fatal(err)
	}

	setupChrome()
	exitVal := m.Run()
	os.Exit(exitVal)
}

var blackholeState *StateData

func TestLoadLocalState(t *testing.T) {
	fullPath, err := utils.ExpandPath(statePath)
	if err != nil {
		t.Fatal(err)
	}
	blackholeState, err = loadLocalState(fullPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetProfiles(t *testing.T) {
	var needle *profiles.Profile

	browsers.AddBrowserDef(browsers.ChromeBrowser(BrowserName, "testdata", "", ""))
	ch := &Chrome{}
	profiles, err := ch.GetProfiles(BrowserName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(profiles), "wrong number of profiles found")

	for _, profile := range profiles {
		if profile.ID == "Default" {
			needle = profile
			break
		}
	}
	assert.NotNil(t, needle, "No profile with ID 'Default' found")
}

func TestRun(t *testing.T) {
	logging.SetLevel(logging.Silent)
	ch.Run()

	// dummy google Bookmarks file url count
	assert.EqualValues(t, 2007, int(ch.URLCount()), "wrong # of parsed urls")

	// 2007 urls and 1909 folders
	assert.EqualValues(t, 2007+1909, int(ch.NodeCount()), "wrong # of parsed nodes")

}

func TestBookmarkPaths(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		wantFiles []string
	}{
		{
			name:      "local bookmarks only",
			files:     []string{bookmarksFileName},
			wantFiles: []string{bookmarksFileName},
		},
		{
			name:      "account bookmarks only",
			files:     []string{accountBookmarksFileName},
			wantFiles: []string{accountBookmarksFileName},
		},
		{
			name:      "local and account bookmarks",
			files:     []string{bookmarksFileName, accountBookmarksFileName},
			wantFiles: []string{bookmarksFileName, accountBookmarksFileName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, fileName := range tt.files {
				assert.NoError(t, os.WriteFile(filepath.Join(dir, fileName), []byte("{}"), 0o600))
			}

			browser := &Chrome{ChromeConfig: NewChromeConfig()}
			browser.BkDir = dir
			paths, err := browser.bookmarkPaths()
			assert.NoError(t, err)

			wantPaths := make([]string, 0, len(tt.wantFiles))
			for _, fileName := range tt.wantFiles {
				wantPaths = append(wantPaths, filepath.Join(dir, fileName))
			}
			assert.Equal(t, wantPaths, paths)
		})
	}
}

func TestSetupWatchersAccountBookmarksOnly(t *testing.T) {
	dir := t.TempDir()
	accountPath := filepath.Join(dir, accountBookmarksFileName)
	assert.NoError(t, os.WriteFile(accountPath, []byte("{}"), 0o600))

	browser := &Chrome{ChromeConfig: NewChromeConfig()}
	browser.BkDir = dir
	if !assert.NoError(t, browser.setupWatchers()) {
		return
	}
	t.Cleanup(func() {
		assert.NoError(t, browser.GetWatcher().W.Close())
	})

	assert.Equal(t, bookmarksFileName, browser.BkFile)
	assert.Equal(t, []string{
		filepath.Join(dir, bookmarksFileName),
		accountPath,
	}, browser.GetWatcher().Watches[0].EventNames)
}

func TestRunAccountBookmarksOnly(t *testing.T) {
	assert.NoError(t, setupCache())
	setupChrome()
	t.Cleanup(func() {
		assert.NoError(t, setupCache())
		setupChrome()
	})
	dir := t.TempDir()
	fixture, err := os.ReadFile(filepath.Join("testdata", bookmarksFileName))
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, accountBookmarksFileName), fixture, 0o600))

	ch.BkDir = dir
	ch.ResetCount()
	ch.Run()

	assert.EqualValues(t, 2007, int(ch.URLCount()), "wrong # of parsed account bookmark urls")
}

func TestPreCount(t *testing.T) {
	assert.NoError(t, ch.PreLoad(&modules.Context{}), "error preloading bookmarks")
	total := ch.Total()
	assert.EqualValues(t, 2007, int(total), "wrong # of url count")
}

func BenchmarkRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ch.Run()
	}
}
