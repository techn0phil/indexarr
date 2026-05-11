package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort        string
	DBPath            string
	TMDBAPIKey        string
	TVDBAPIKey        string
	RadarrURL         string
	SonarrURL         string
	ScanInterval      int      // hours between scans (0 = disabled)
	MediaLibraryPaths []string // directories to scan for media files
	MediainfoPath     string   // path to mediainfo binary
	ScanTimeout       int      // timeout in seconds per file
}

func Load() *Config {
	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		DBPath:            getEnv("DB_PATH", "./indexarr.db"),
		TMDBAPIKey:        getEnv("TMDB_API_KEY", ""),
		TVDBAPIKey:        getEnv("TVDB_API_KEY", ""),
		RadarrURL:         getEnv("RADARR_URL", ""),
		SonarrURL:         getEnv("SONARR_URL", ""),
		ScanInterval:      getEnvInt("SCAN_INTERVAL", 24),
		MediaLibraryPaths: getEnvList("MEDIA_LIBRARY_PATHS", []string{}),
		MediainfoPath:     getEnv("MEDIAINFO_PATH", "mediainfo"),
		ScanTimeout:       getEnvInt("SCAN_TIMEOUT", 30),
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
