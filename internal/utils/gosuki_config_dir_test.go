package utils

import (
	"path/filepath"
	"testing"
)

func TestGetGosukiConfigDirHonorsEnvironmentOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), "gosuki-dev")
	t.Setenv(EnvGosukiConfigHome, want)

	got, err := GetGosukiConfigDir()
	if err != nil {
		t.Fatalf("GetGosukiConfigDir() error = %v", err)
	}
	if got != filepath.Clean(want) {
		t.Fatalf("GetGosukiConfigDir() = %q, want %q", got, filepath.Clean(want))
	}
}
