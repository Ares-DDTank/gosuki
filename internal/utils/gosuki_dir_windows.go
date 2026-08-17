package utils

// GetGosukiDataDir returns the platform-appropriate gosuki data directory.
// On Windows this uses the same directory as the config file (%APPDATA%\gosuki).
func GetGosukiDataDir() (string, error) {
	return GetGosukiConfigDir()
}
