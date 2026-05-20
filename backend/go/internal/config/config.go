package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort         string
	DBPath             string
	TMDBAPIKey         string
	TVDBAPIKey         string
	RadarrURL          string
	SonarrURL          string
	RadarrAPIKey       string
	SonarrAPIKey       string
	ScanInterval       int      // hours between scans (0 = disabled)
	MediaLibraryPaths  []string // directories to scan for media files
	MoviesLibraryPaths []string // directories to scan for movies (optional, overrides MediaLibraryPaths if set)
	SkipFolders        []string // folder names to skip during scanning
	MediainfoPath      string   // path to mediainfo binary
	ScanTimeout        int      // timeout in seconds per file
}

func Load() *Config {
	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "./indexarr.db"),
		TMDBAPIKey:         getEnv("TMDB_API_KEY", ""),
		TVDBAPIKey:         getEnv("TVDB_API_KEY", ""),
		RadarrURL:          getEnv("RADARR_URL", ""),
		SonarrURL:          getEnv("SONARR_URL", ""),
		RadarrAPIKey:       getEnv("RADARR_API_KEY", ""),
		SonarrAPIKey:       getEnv("SONARR_API_KEY", ""),
		ScanInterval:       getEnvInt("SCAN_INTERVAL", 24),
		MediaLibraryPaths:  getEnvList("MEDIA_LIBRARY_PATHS", []string{}),
		MoviesLibraryPaths: getEnvList("MOVIES_LIBRARY_PATHS", []string{}),
		SkipFolders:        getEnvList("SKIP_FOLDERS", []string{}),
		MediainfoPath:      getEnv("MEDIAINFO_PATH", "mediainfo"),
		ScanTimeout:        getEnvInt("SCAN_TIMEOUT", 30),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvList(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		paths := strings.Split(value, ",")
		var result []string
		for _, p := range paths {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

// HasRadarrConfig returns true if Radarr API is configured
func (c *Config) HasRadarrConfig() bool {
	return c.RadarrURL != "" && c.RadarrAPIKey != ""
}

// HasSonarrConfig returns true if Sonarr API is configured (for future use)
func (c *Config) HasSonarrConfig() bool {
	return c.SonarrURL != "" && c.SonarrAPIKey != ""
}

// UseFilesystemScan returns true if filesystem scanning should be used
// (when no Radarr/Sonarr config is present but media paths are configured)
func (c *Config) UseFilesystemScan() bool {
	return !c.HasRadarrConfig() && len(c.MediaLibraryPaths) > 0
}

// GetScanMode returns the current scan mode as a string
func (c *Config) GetScanMode() string {
	if c.HasRadarrConfig() {
		return "radarr"
	}
	if len(c.MediaLibraryPaths) > 0 {
		return "filesystem"
	}
	return "disabled"
}
